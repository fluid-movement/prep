// Records where a session's tokens go: one JSONL row per model request (its
// usage) and per tool call (what it touched, how much text it returned), each
// with the prep issue being worked on, to
// ~/.claude/token-ledger/<session id>.jsonl. Rows are buffered and appended at
// the end of each turn; /token-ledger summarizes the current session.

import type { EngineInterface, Register } from 'claude-code'

import { classify } from './classify'
import { issueFromCommand, issueFromPath } from './infer'

type Row = Record<string, unknown> & { type: string; t: number }

const FLUSH_AT = 200

let buffer: Row[] = []
let turnId: string | undefined
// The issue being worked on, as the prep panel infers it: the last one a prep
// write or guide named, or whose files were edited.
let issue: string | undefined
let writes: Promise<void> = Promise.resolve()

async function ledgerPath($: EngineInterface): Promise<string> {
  const home = (await $.env.get('HOME')) ?? '.'
  return `${home}/.claude/token-ledger/${await $.session.id()}.jsonl`
}

/** Buffers one row; a failure is logged, never raised into the session. */
async function record($: EngineInterface, row: Omit<Row, 't'>): Promise<void> {
  try {
    buffer.push({ t: await $.clock.now(), ...row } as Row)
    if (buffer.length >= FLUSH_AT) await flush($)
  } catch (err) {
    $.ui.log(`token-ledger: ${String(err)}`)
  }
}

/** Appends the buffered rows; writes are queued so two flushes never interleave. */
function flush($: EngineInterface): Promise<void> {
  const rows = buffer
  buffer = []
  if (rows.length === 0) return writes
  writes = writes.then(async () => {
    const path = await ledgerPath($)
    const before = (await $.fs.exists(path)) ? ((await $.fs.read(path)) as string) : ''
    await $.fs.write(path, before + rows.map(r => JSON.stringify(r)).join('\n') + '\n')
  }).catch(err => $.ui.log(`token-ledger: write failed: ${String(err)}`))
  return writes
}

/** Reads this session's rows: those written and those still buffered. */
async function rows($: EngineInterface): Promise<Row[]> {
  await flush($)
  const path = await ledgerPath($)
  if (!(await $.fs.exists(path))) return []
  const text = (await $.fs.read(path)) as string
  return text.split('\n').filter(Boolean).map(line => JSON.parse(line) as Row)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({ name: 'token-ledger', description: "Summarize this session's token usage and tool calls" })
    issue = undefined
    const isPrep = await $.fs.exists(`${e.cwd}/.prep`)
    await record($, { type: 'session', sessionId: await $.session.id(), cwd: e.cwd, isPrep, interactive: e.isInteractive })
    return next(e)
  })

  on('turn.start', async ($, e, next) => {
    turnId = e.turnId
    await record($, { type: 'turn', turnId: e.turnId, promptChars: e.text.length })
    return next(e)
  })

  on('turn.step', async function* ($, e, next) {
    const r = yield* next(e)
    await record($, {
      type: 'step',
      turnId: e.turnId,
      index: e.index,
      agentId: e.agentId,
      issue,
      messages: e.messageCount,
      model: r.usage?.model ?? e.model,
      usage: r.usage && {
        input: r.usage.input_tokens,
        cacheRead: r.usage.cache_read_input_tokens,
        cacheWrite: r.usage.cache_creation_input_tokens,
        output: r.usage.output_tokens,
      },
      answerChars: r.answer.length,
      tools: r.toolUses.map(u => u.name),
      stop: r.stopReason,
    })
    return r
  })

  // The ledger only observes: if anything here fails, the call goes on as if
  // the plugin were not installed (a call already made replays its result).
  on('tool.call', async ($, e, next) => {
    const started = Date.now()
    const ran = await next(e)
    const ms = Date.now() - started
    if (ran.deny === undefined && ran.isError !== true) issue = issueOf(e as unknown as Record<string, unknown>, ran.text ?? '') ?? issue
    await record($, {
      type: 'tool',
      turnId: e.agentId ? undefined : turnId,
      agentId: e.agentId,
      tool: e.tool,
      issue,
      ...classify(e.tool, e as unknown as Record<string, unknown>),
      resultChars: ran.deny === undefined ? (ran.text?.length ?? 0) : 0,
      isError: ran.deny !== undefined || ran.isError === true,
      ms,
    })
    return ran
  }).catch(($, e, next) => next(e))

  on('turn.complete', async ($, e, next) => {
    const r = await next(e)
    const { context, cost } = await $.session.usage()
    await record($, { type: 'turn.end', turnId: e.turnId, contextTokens: context.tokens, window: context.window, costUsd: cost?.usd })
    await flush($)
    return r
  })

  on('session.end', async ($, e, next) => {
    await record($, { type: 'session.end', reason: e.reason })
    await flush($)
    return next(e)
  })

  on('command.run', { command: 'token-ledger' }, async $ => ({ text: summarize(await rows($)) }))
}

/** The issue a tool call moves the work to, if any. */
function issueOf(call: Record<string, unknown>, output: string): string | undefined {
  const str = (k: string) => (typeof call[k] === 'string' ? (call[k] as string) : undefined)
  switch (call.tool) {
    case 'Bash':
      return issueFromCommand(str('command') ?? '', output)
    case 'Edit':
    case 'Write':
    case 'NotebookEdit': {
      const path = str('file_path') ?? str('notebook_path')
      return path === undefined ? undefined : issueFromPath(path)
    }
  }
  return undefined
}

/** Totals per model request and per tool and prep area, for a quick look. */
export function summarize(all: readonly Row[]): string {
  const steps = all.filter(r => r.type === 'step')
  const tools = all.filter(r => r.type === 'tool')
  const sum = (key: string) => steps.reduce((n, r) => n + (((r.usage as Record<string, number>) ?? {})[key] ?? 0), 0)
  const lines = [
    `${steps.length} model requests: input ${sum('input')}, cache read ${sum('cacheRead')}, cache write ${sum('cacheWrite')}, output ${sum('output')} tokens`,
    `${tools.length} tool calls (result size ≈ chars/4 tokens):`,
  ]
  const group = (key: string) => {
    const by = new Map<string, { n: number; chars: number }>()
    for (const r of tools) {
      const v = r[key] as string | string[] | undefined
      const k = Array.isArray(v) ? v.join(' + ') : v
      if (k === undefined) continue
      const g = by.get(k) ?? { n: 0, chars: 0 }
      g.n++
      g.chars += (r.resultChars as number) ?? 0
      by.set(k, g)
    }
    return [...by].sort((a, b) => b[1].chars - a[1].chars).map(([k, g]) => `  ${k}: ${g.n} calls, ~${Math.round(g.chars / 4)} tokens`)
  }
  lines.push(...group('tool'), 'by prep area:', ...group('area'), 'prep subcommands:', ...group('prep'))
  lines.push('by issue (model tokens include cache reads):', ...byIssue(steps, tools))
  return lines.join('\n')
}

/** Model tokens and tool output per issue worked on; before any issue counts as none. */
function byIssue(steps: readonly Row[], tools: readonly Row[]): string[] {
  const by = new Map<string, { requests: number; tokens: number; calls: number; chars: number }>()
  const get = (r: Row) => {
    const k = (r.issue as string | undefined) ?? '(none)'
    const g = by.get(k) ?? { requests: 0, tokens: 0, calls: 0, chars: 0 }
    by.set(k, g)
    return g
  }
  for (const r of steps) {
    const g = get(r)
    g.requests++
    for (const n of Object.values((r.usage as Record<string, number>) ?? {})) g.tokens += n ?? 0
  }
  for (const r of tools) {
    const g = get(r)
    g.calls++
    g.chars += (r.resultChars as number) ?? 0
  }
  return [...by]
    .sort((a, b) => b[1].tokens - a[1].tokens)
    .map(([k, g]) => `  ${k}: ${g.requests} requests, ${g.tokens} tokens; ${g.calls} tool calls, ~${Math.round(g.chars / 4)} tokens`)
}
