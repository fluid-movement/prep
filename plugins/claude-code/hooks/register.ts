// The activity bridge: tells prep what this Claude Code session does that
// prep cannot see itself: each model request's tokens, context fill and cost.
// Shared command hooks report tool calls. PREP_SESSION ties both sides together, and
// PREP_ACTOR names the agent where it gives no --by.

import type { EngineInterface, Register } from 'claude-code'
type Event = Record<string, unknown> & { kind: 'tool' | 'request' | 'session' | 'gap' | 'summary' }

// Events wait here until the next send; a send is one prep call.
let pending: Event[] = []
let sending: Promise<void> = Promise.resolve()
// Off outside a prep project or without the binary: the first failed send
// says so, and the session stops trying.
let enabled = true
let session: string | undefined
let lastCost = 0
let stepped = false

const ACTOR = 'claude-code'

// Independent of the prep process: a broken/missing binary must not hide loss.
// One durable marker means coverage remains unknown after the bridge turns off.
async function disable($: EngineInterface, reason: string, id = session): Promise<void> {
  enabled = false
  pending = []
  $.ui.log(`prep activity bridge off: ${reason}`, { to: 'debug' })
  try {
    const root = await $.session.root()
    if (!await $.fs.exists(`${root}/.prep`)) return
    if (!await $.fs.exists(`${root}/.prep/local`)) {
      const made = await $.process.run(['mkdir', '-p', `${root}/.prep/local`], { timeoutMs: 1000 })
      if (made.exitCode !== 0) throw new Error(made.stderr)
    }
    await $.fs.write(`${root}/.prep/local/request-gap-${encodeURIComponent(id ?? 'unknown')}.json`,
      JSON.stringify({ at: new Date().toISOString(), actor: ACTOR, session: id, kind: 'gap', op: 'bridge-disabled' }) + '\n')
  } catch (err) {
    $.ui.log(`prep activity gap could not be recorded: ${String(err)}`, { to: 'debug' })
  }
}

/**
 * Exports the session and the actor for the agent's shell, so its prep calls
 * carry them: prep records reads only for a named actor, and the bridge
 * leaves prep commands to prep. An explicit --by still wins.
 */
async function joinSession($: EngineInterface): Promise<void> {
  const id = await $.session.id()
  if (id === session) return
  const previous = session
  if (previous) {
    pending.push({ kind: 'session', op: 'ended', session: previous })
    void send($)
  }
  session = id
  stepped = false
  pending.push({ kind: 'session', op: 'started', session: id, previous })
  void send($)
  lastCost = 0
  await $.env.set('PREP_SESSION', id)
  await $.env.set('PREP_ACTOR', ACTOR)
}

/** Sends what is pending to prep activity add, never awaited by a call. */
function send($: EngineInterface): Promise<void> {
  if (!enabled || pending.length === 0) return sending
  const events = pending.map(e => ({ actor: ACTOR, session, ...e }))
  pending = []
  sending = sending
    .then(async () => {
      if (!enabled) return
      const stdin = events.map(e => JSON.stringify(e)).join('\n') + '\n'
      const { exitCode, stderr } = await $.process.run(['prep', 'activity', 'add'], { stdin, timeoutMs: 5000 })
      if (exitCode !== 0) {
        await disable($, stderr.trim() || `exit ${exitCode}`, events[0].session as string | undefined)
      }
    })
    .catch(async err => {
      await disable($, String(err), events[0].session as string | undefined)
    })
  return sending
}

function marker($: EngineInterface, op: string): void {
  if (!enabled) return
  pending.push({ kind: 'summary', op, status: 'entered' })
  void send($)
}

async function reportError($: EngineInterface, op: string, error: unknown): Promise<void> {
  if (!enabled) return
  pending.push({ kind: 'gap', op, status: 'error', verb: 'error', target: String(error) })
  await send($)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    try {
      const started = await next(e)
      enabled = true
      await joinSession($)
      return started
    } catch (error) {
      await reportError($, 'session.start', error)
      throw error
    }
  }).catch(($, e, next) => next(e))

  // Claude Code can move to a new session ID without session.start.
  on('turn.start', async ($, e, next) => {
    try {
      await joinSession($)
      return await next(e)
    } catch (error) {
      await reportError($, 'turn.start', error)
      throw error
    }
  }).catch(($, e, next) => next(e))

  // Tool activity is reported by the shared after-tool command in hooks.json.
  // This module retains session attribution and request/token/cost reporting.

  on('turn.step', async function* ($, e, next) {
    if (!stepped && enabled) {
      stepped = true
      marker($, 'turn.step')
    }
    // Optional fields vary by engine version. Never invent request identity.
    const extra = e as typeof e & { agentType?: string, requestId?: string, messageId?: string }
    const attribution = { agent_id: e.agentId, agent_type: extra.agentType, turn_id: e.turnId,
      request_id: extra.requestId ?? extra.messageId }
    let recorded = false
    try {
      const r = yield* next(e)
      if (enabled) {
        const tokens = {
          input: r.usage?.input_tokens ?? 0,
          output: r.usage?.output_tokens ?? 0,
          cache_read: r.usage?.cache_read_input_tokens ?? 0,
          cache_write: r.usage?.cache_creation_input_tokens ?? 0,
        }
        const hasUsage = Object.values(tokens).some(count => count !== 0)
        pending.push({
          kind: hasUsage ? 'request' : 'gap',
          ...attribution,
          op: r.usage?.model ?? e.model,
          tokens: hasUsage ? tokens : undefined,
        })
        recorded = true
        void send($)
      }
      return r
    } finally {
      if (!recorded && enabled) {
        pending.push({ kind: 'gap', ...attribution, op: 'request-interrupted' })
        void send($)
      }
    }
  })

  // The turn's end carries the context fill and what the turn cost.
  on('turn.complete', async ($, e, next) => {
    marker($, 'turn.complete')
    try {
      const done = await next(e)
      if (enabled) {
        const { context, cost } = await $.session.usage()
        const total = cost?.usd ?? 0
        pending.push({ kind: 'summary', context: { tokens: context.tokens, window: context.window }, cost_usd: Math.max(0, total - lastCost) || undefined })
        lastCost = total
        void send($)
      }
      return done
    } catch (error) {
      await reportError($, 'turn.complete', error)
      throw error
    }
  }).catch(($, e, next) => next(e))

  on('session.end', async ($, e, next) => {
    try {
      if (session) pending.push({ kind: 'session', op: 'ended', session })
      // Awaited so the queue, including ended, is flushed before the session exits.
      await send($)
      session = undefined
      return await next(e)
    } catch (error) {
      await reportError($, 'session.end', error)
      throw error
    }
  }).catch(($, e, next) => next(e))
}
