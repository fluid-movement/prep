// The prep side panel: shows the issue the agent works on, or the project at
// a glance, and refreshes whenever anything under .prep changes.

import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type {
  PrepGuide,
  PrepPrime,
  PrepProjectSnapshot,
  PrepShow,
  PrepSnapshot,
  PrepSource,
  PrepSummary,
  PrepTab,
  PrepUsage,
} from '../types'
import { issueFromCommand, issueFromPath } from './infer'
import { drawPane, drawPanel } from './panel'
import { ledgerPath, summarizeLedger } from './usage'

const PANE = 'prep'
const TITLE = 'prep'

const focus = atom({ plugin: 'prep', key: 'focus' } as const, null)
const pin = atom({ plugin: 'prep', key: 'pin' } as const, null)
const snapshot = atom({ plugin: 'prep', key: 'snapshot' } as const, null)
const tab = atom({ plugin: 'prep', key: 'tab' } as const, 'live' as PrepTab)
const project = atom({ plugin: 'prep', key: 'project' } as const, null)
const usage = atom({ plugin: 'prep', key: 'usage' } as const, null)

// token-ledger appends its rows when a turn ends; the usage view reads them
// this long after, so the turn's own rows are in.
const LEDGER_SETTLE_MS = 500

// The project view lists what is not resolved yet.
const UNRESOLVED = ['--state', 'open,defined,ready,in-progress']

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
    const prime = normalPrime(await prep<PrepPrime>($, ['prime']))
    if (prime.claims.length === 0) return { view: 'overview', prime }
    ref = prime.claims[0]!.id
    source = 'claimed'
  }
  const [show, guide, list] = await Promise.all([
    prep<PrepShow>($, ['show', ref]),
    prep<PrepGuide>($, ['guide', ref]),
    prep<{ issues: PrepSummary[] }>($, ['list']),
  ])
  // prep writes null for empty lists (a resolved issue has no transitions).
  guide.transitions ??= []
  guide.knowledge ??= []
  show.blocks ??= []
  show.children ??= []
  show.definition_of_done ??= []
  return { view: 'issue', source, show, guide, issues: list.issues ?? [] }
}

function normalPrime(p: PrepPrime): PrepPrime {
  return { ...p, parents: p.parents ?? [], next: p.next ?? [], stale: p.stale ?? [], claims: p.claims ?? [] }
}

/** Loads the project view: the overview and the unresolved issues in tree order. */
async function loadProject($: EngineInterface): Promise<PrepProjectSnapshot> {
  const [prime, list] = await Promise.all([
    prep<PrepPrime>($, ['prime']),
    prep<{ issues: PrepSummary[] | null }>($, ['list', '--tree', ...UNRESOLVED]),
  ])
  return { prime: normalPrime(prime), tree: list.issues ?? [] }
}

// Refreshes can overlap (a watch line and a tool call); only the newest writes.
let generation = 0
let projectGeneration = 0
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
  if ((await read($, tab)) === 'project') await refreshProject($)
}

async function refreshProject($: EngineInterface): Promise<void> {
  const gen = ++projectGeneration
  let next: PrepProjectSnapshot
  try {
    next = await loadProject($)
  } catch (err) {
    next = { error: err instanceof Error ? err.message : String(err) }
  }
  if (gen === projectGeneration) await update($, project, () => next)
}

/** Loads the usage view: Claude Code's figures for the session, and token-ledger's rows when it wrote any. */
async function loadUsage($: EngineInterface): Promise<PrepUsage> {
  const [sessionId, home, session] = await Promise.all([$.session.id(), $.env.get('HOME'), $.session.usage()])
  const path = ledgerPath(home ?? '', sessionId)
  const hasLedger = home !== undefined && (await $.fs.exists(path))
  const text = hasLedger ? ((await $.fs.read(path)) as string) : ''
  return {
    sessionId,
    contextTokens: session.context.tokens,
    window: session.context.window,
    costUsd: session.cost?.usd,
    hasLedger,
    ...summarizeLedger(text),
  }
}

let usageGeneration = 0

async function refreshUsage($: EngineInterface): Promise<void> {
  const gen = ++usageGeneration
  let next: PrepUsage | { error: string }
  try {
    next = await loadUsage($)
  } catch (err) {
    next = { error: err instanceof Error ? err.message : String(err) }
  }
  if (gen === usageGeneration) await update($, usage, () => next)
}

/** Shows a view; the project and usage views load fresh data as they open. */
async function showTab($: EngineInterface, next: PrepTab): Promise<void> {
  await update($, tab, () => next)
  if (next === 'project') await refreshProject($)
  if (next === 'usage') await refreshUsage($)
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

  // The usage view follows the session turn by turn while it is shown.
  on('turn.complete', async ($, e, next) => {
    const done = await next(e)
    if (active && (await read($, tab)) === 'usage') {
      void $.clock.after(LEDGER_SETTLE_MS, () => refreshUsage($))
    }
    return done
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

  on('command.run', { command: 'prep:pane' }, async ($, e) => {
    if (!active) return { text: 'Not a prep project.' }
    const arg = e.args.trim()
    if (arg === 'live' || arg === 'project' || arg === 'usage') {
      await showTab($, arg)
      await $.store.set(openKey(cwd), true)
      await $.ui.open({ id: PANE, title: TITLE })
      return { text: `prep panel shows the ${arg} view.` }
    }
    if (arg !== '') return { text: `prep panel: unknown view ${arg}; use live, project or usage.` }
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
    const kit = $.ui.resolve(e)
    const [shown, live, proj, used] = [await read($, tab), await read($, snapshot), await read($, project), await read($, usage)]
    try {
      return drawPane(kit, shown, live, proj, used, next => void showTab($, next))
    } catch (err) {
      // Never an empty pane: say what failed instead.
      return drawPanel(kit, { view: 'error', message: err instanceof Error ? err.message : String(err) })
    }
  })
}
