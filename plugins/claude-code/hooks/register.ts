// The activity bridge: tells prep what this Claude Code session does that
// prep cannot see itself, so prep tui's Agent screen shows it: each tool
// call (what it touched, how much it returned) and each model request (its
// tokens, the context fill and the cost). The agent's own prep commands are
// left out; prep records those. PREP_SESSION ties both sides together, and
// PREP_ACTOR names the agent where it gives no --by.

import type { EngineInterface, Register } from 'claude-code'

import { classify, prepSubcommands } from './classify'

type Event = Record<string, unknown> & { kind: 'tool' | 'request' }

// Events wait here until the next send; a send is one prep call.
let pending: Event[] = []
let sending: Promise<void> = Promise.resolve()
// Off outside a prep project or without the binary: the first failed send
// says so, and the session stops trying.
let enabled = true
let session: string | undefined
let lastCost = 0

const ACTOR = 'claude-code'
const MAX_TARGET = 200

/**
 * Exports the session and the actor for the agent's shell, so its prep calls
 * carry them: prep records reads only for a named actor, and the bridge
 * leaves prep commands to prep. An explicit --by still wins.
 */
async function joinSession($: EngineInterface): Promise<void> {
  const id = await $.session.id()
  if (id === session) return
  session = id
  lastCost = 0
  await $.env.set('PREP_SESSION', id)
  await $.env.set('PREP_ACTOR', ACTOR)
}

/** Sends what is pending to prep activity add, never awaited by a call. */
function send($: EngineInterface): Promise<void> {
  if (!enabled || pending.length === 0) return sending
  const events = pending.map(e => ({ ...e, actor: ACTOR, session }))
  pending = []
  sending = sending
    .then(async () => {
      const stdin = events.map(e => JSON.stringify(e)).join('\n') + '\n'
      const { exitCode, stderr } = await $.process.run(['prep', 'activity', 'add'], { stdin, timeoutMs: 5000 })
      if (exitCode !== 0) {
        enabled = false
        $.ui.log(`prep activity bridge off: ${stderr.trim() || `exit ${exitCode}`}`, { to: 'debug' })
      }
    })
    .catch(err => {
      enabled = false
      $.ui.log(`prep activity bridge off: ${String(err)}`, { to: 'debug' })
    })
  return sending
}

/** What a tool call did, in the feed's words. */
export function toolEvent(tool: string, input: Record<string, unknown>, chars: number, failed: boolean): Event | undefined {
  const t = classify(tool, input)
  if (tool === 'Bash' && prepSubcommands(String(input.command ?? '')).length > 0) return undefined // prep records it
  // A shell command that names no file is still work on the code (builds, tests).
  const area = t.area === 'prep-cli' ? 'prep' : (t.area ?? (tool === 'Bash' ? 'code' : 'other'))
  const verbs: Record<string, string> = {
    Read: 'read',
    Edit: 'edited',
    Write: 'wrote',
    NotebookEdit: 'edited',
    Bash: 'ran',
    Grep: 'searched',
    Glob: 'looked for files',
    WebFetch: 'fetched',
    WebSearch: 'searched the web',
    Agent: 'asked a subagent',
    Skill: 'loaded a skill',
  }
  let target = t.target ?? (typeof input.url === 'string' ? input.url : typeof input.query === 'string' ? input.query : undefined)
  if (target !== undefined) {
    target = target.split('\n')[0]!
    if (target.length > MAX_TARGET) target = `${target.slice(0, MAX_TARGET - 1)}…`
  }
  return { kind: 'tool', op: tool, verb: verbs[tool] ?? tool.toLowerCase(), target, area, chars, failed: failed || undefined }
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    const started = await next(e)
    enabled = true
    await joinSession($)
    return started
  }).catch(($, e, next) => next(e))

  // Claude Code can move to a new session ID without session.start.
  on('turn.start', async ($, e, next) => {
    await joinSession($)
    return next(e)
  }).catch(($, e, next) => next(e))

  // The bridge only observes: a failure here never changes the call.
  on('tool.call', async ($, e, next) => {
    const ran = await next(e)
    const ev = toolEvent(e.tool, e as unknown as Record<string, unknown>, ran.deny === undefined ? (ran.text?.length ?? 0) : 0, ran.deny !== undefined || ran.isError === true)
    if (ev && enabled) {
      pending.push(ev)
      void send($)
    }
    return ran
  }).catch(($, e, next) => next(e))

  on('turn.step', async function* ($, e, next) {
    const r = yield* next(e)
    if (r.usage && enabled) {
      pending.push({
        kind: 'request',
        op: r.usage.model ?? e.model,
        tokens: {
          input: r.usage.input_tokens ?? 0,
          output: r.usage.output_tokens ?? 0,
          cache_read: r.usage.cache_read_input_tokens ?? 0,
          cache_write: r.usage.cache_creation_input_tokens ?? 0,
        },
      })
    }
    return r
  })

  // The turn's end carries the context fill and what the turn cost.
  on('turn.complete', async ($, e, next) => {
    const done = await next(e)
    if (enabled) {
      const { context, cost } = await $.session.usage()
      const total = cost?.usd ?? 0
      pending.push({ kind: 'request', context: { tokens: context.tokens, window: context.window }, cost_usd: Math.max(0, total - lastCost) || undefined })
      lastCost = total
      void send($)
    }
    return done
  }).catch(($, e, next) => next(e))

  on('session.end', async ($, e, next) => {
    await send($)
    return next(e)
  }).catch(($, e, next) => next(e))
}
