// The prep side panel: shows the issue the agent works on, or the project at
// a glance, and refreshes whenever anything under .prep changes.

import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { PrepGuide, PrepPrime, PrepShow, PrepSnapshot, PrepSource, PrepSummary } from '../types'
import { issueFromCommand, issueFromPath } from './infer'
import { drawPanel } from './panel'

const PANE = 'prep'
const TITLE = 'prep'

const focus = atom({ plugin: 'prep', key: 'focus' } as const, null)
const pin = atom({ plugin: 'prep', key: 'pin' } as const, null)
const snapshot = atom({ plugin: 'prep', key: 'snapshot' } as const, null)

/** Runs a prep read command and parses its JSON output. */
async function prep<T>($: EngineInterface, args: string[]): Promise<T> {
  const { exitCode, stdout, stderr } = await $.process.run(['prep', ...args, '--json'])
  if (exitCode !== 0) {
    let message = stderr.trim() || stdout.trim()
    try {
      message = JSON.parse(stdout).error?.message ?? message
    } catch {}
    throw new Error(message || `prep ${args[0]} exited ${exitCode}`)
  }
  return JSON.parse(stdout) as T
}

/** Loads what the pane shows: the pinned, followed or claimed issue, else the overview. */
async function load($: EngineInterface): Promise<PrepSnapshot> {
  const pinned = await read($, pin)
  const followed = await read($, focus)
  let ref = pinned ?? followed
  let source: PrepSource = pinned ? 'pinned' : 'agent'
  if (ref === null) {
    const prime = await prep<PrepPrime>($, ['prime'])
    if (prime.claims.length === 0) return { view: 'overview', prime }
    ref = prime.claims[0]!.id
    source = 'claimed'
  }
  const [show, guide, list] = await Promise.all([
    prep<PrepShow>($, ['show', ref]),
    prep<PrepGuide>($, ['guide', ref]),
    prep<{ issues: PrepSummary[] }>($, ['list']),
  ])
  return { view: 'issue', source, show, guide, issues: list.issues }
}

// Refreshes can overlap (a watch line and a tool call); only the newest writes.
let generation = 0
// Set once session.start found a prep project; outside one the module idles.
let active = false

async function refresh($: EngineInterface): Promise<void> {
  const gen = ++generation
  let next: PrepSnapshot
  try {
    next = await load($)
  } catch (err) {
    next = { view: 'error', message: err instanceof Error ? err.message : String(err) }
  }
  if (gen === generation) await update($, snapshot, () => next)
}

/** Makes ref the agent's current issue and redraws when it changed. */
async function follow($: EngineInterface, ref: string | undefined): Promise<void> {
  if (!active || ref === undefined) return
  const before = await read($, focus)
  if (before === ref) return
  await update($, focus, () => ref)
  if ((await read($, pin)) === null) await refresh($)
}

async function isPaneOpen($: EngineInterface): Promise<boolean> {
  return (await $.ui.panes()).some(p => p.id === PANE)
}

/** Opens the pane at session start as the pane option says. */
async function openAtStart($: EngineInterface, mode: string, cwd: string): Promise<void> {
  const remembered = await $.store.get(openKey(cwd))
  if (mode === 'open' || (mode === 'remember' && remembered !== false)) {
    await $.ui.open({ id: PANE, title: TITLE })
  }
}

/** The store key remembering, per project, whether the person left the pane open. */
function openKey(cwd: string): string {
  return `open:${cwd}`
}

// A failing panel hook never changes the call it watched: next(e) replays
// what the call settled to, or runs it when the hook failed before.
const passOn = <E, R>(_$: unknown, e: E, next: (e: E) => R): R => next(e)

export const register: Register = (on, options) => {
  const mode = typeof options.pane === 'string' ? options.pane : 'remember'
  let cwd = ''

  on('session.start', async ($, e, next) => {
    const started = await next(e)
    cwd = e.cwd
    try {
      await prep<PrepPrime>($, ['prime'])
    } catch {
      return started // not a prep project, or no prep binary
    }
    active = true
    await refresh($)
    if (e.isInteractive) await openAtStart($, mode, cwd)
    // prep watch prints a line per change under .prep for the session's life.
    void (async () => {
      try {
        for await (const chunk of $.process.spawn({ argv: ['prep', 'watch'] })) {
          if (chunk.stream === 'stdout') await refresh($)
        }
      } catch (err) {
        $.ui.log(`prep watch stopped: ${err instanceof Error ? err.message : String(err)}`, { to: 'debug' })
      }
    })()
    return started
  })

  // The agent's prep commands and edits under .prep/issues move the focus,
  // once they succeeded.
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const ran = await next(e)
    if (ran.deny === undefined && ran.isError !== true) {
      await follow($, issueFromCommand(e.command, ran.text ?? ''))
    }
    return ran
  }).catch(passOn)
  on('tool.call', { tool: 'Edit' }, async ($, e, next) => {
    const ran = await next(e)
    if (ran.deny === undefined && ran.isError !== true) await follow($, issueFromPath(e.file_path))
    return ran
  }).catch(passOn)
  on('tool.call', { tool: 'Write' }, async ($, e, next) => {
    const ran = await next(e)
    if (ran.deny === undefined && ran.isError !== true) await follow($, issueFromPath(e.file_path))
    return ran
  }).catch(passOn)

  // /prep:focus and /prep:pane are declared in commands/ and answered here.
  on('command.run', { command: 'prep:focus' }, async ($, e) => {
    if (!active) return { text: 'Not a prep project.' }
    const ref = e.args.trim()
    if (ref === '') {
      await update($, pin, () => null)
      await refresh($)
      return { text: 'prep panel follows the agent again.' }
    }
    let id: string
    try {
      id = (await prep<PrepShow>($, ['show', ref])).issue.id
    } catch (err) {
      return { text: `prep panel: ${err instanceof Error ? err.message : String(err)}` }
    }
    await update($, pin, () => id)
    await refresh($)
    if (!(await isPaneOpen($))) await $.ui.open({ id: PANE, title: TITLE })
    return { text: `prep panel pinned to ${id}.` }
  })

  on('command.run', { command: 'prep:pane' }, async $ => {
    if (!active) return { text: 'Not a prep project.' }
    if (await isPaneOpen($)) {
      await $.store.set(openKey(cwd), false)
      await $.ui.close({ id: PANE })
      return { text: 'prep panel closed.' }
    }
    await $.store.set(openKey(cwd), true)
    await $.ui.open({ id: PANE, title: TITLE })
    return { text: 'prep panel opened.' }
  })

  // Closing the pane by hand is remembered for the next session.
  on('ui.close', async ($, e, next) => {
    if (e.id === PANE && e.origin.kind === 'person') await $.store.set(openKey(cwd), false)
    return next(e)
  }).catch(passOn)

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    return drawPanel($.ui.resolve(e), await read($, snapshot))
  })
}
