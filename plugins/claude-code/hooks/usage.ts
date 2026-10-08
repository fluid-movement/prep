// The Usage view's figures: the session's context and cost from Claude Code,
// and where its tokens went from the token-ledger plugin's file for this
// session, ~/.claude/token-ledger/<session id>.jsonl, when that is installed.

import type { PrepUsage, PrepUsageArea } from '../types'

const AREAS = ['knowledge', 'issue', 'prep-cli', 'code'] as const
const TOP_READS = 5

type LedgerRow = {
  type?: string
  area?: string
  target?: string
  resultChars?: number
  agentId?: string
  usage?: { input?: number; cacheRead?: number; cacheWrite?: number; output?: number }
}

/** The ledger file token-ledger writes for a session. */
export function ledgerPath(home: string, sessionId: string): string {
  return `${home}/.claude/token-ledger/${sessionId}.jsonl`
}

/** Totals the ledger's rows: tokens per request, tool output per prep area, the largest knowledge reads. */
export function summarizeLedger(text: string): Pick<PrepUsage, 'requests' | 'tokens' | 'areas' | 'topReads'> {
  const rows: LedgerRow[] = []
  for (const line of text.split('\n')) {
    if (line.trim() === '') continue
    try {
      rows.push(JSON.parse(line) as LedgerRow)
    } catch {} // a line cut by a write in progress
  }
  const tokens = { input: 0, cacheRead: 0, cacheWrite: 0, output: 0 }
  let requests = 0
  const areas = new Map<string, PrepUsageArea>(AREAS.map(a => [a, { area: a, calls: 0, tokens: 0 }]))
  const reads: { target: string; tokens: number }[] = []
  for (const r of rows) {
    if (r.type === 'step') {
      requests++
      tokens.input += r.usage?.input ?? 0
      tokens.cacheRead += r.usage?.cacheRead ?? 0
      tokens.cacheWrite += r.usage?.cacheWrite ?? 0
      tokens.output += r.usage?.output ?? 0
    } else if (r.type === 'tool' && r.area !== undefined) {
      const a = areas.get(r.area)
      if (a === undefined) continue
      const t = Math.round((r.resultChars ?? 0) / 4)
      a.calls++
      a.tokens += t
      if (r.area === 'knowledge' && r.target !== undefined) reads.push({ target: r.target, tokens: t })
    }
  }
  reads.sort((x, y) => y.tokens - x.tokens)
  return { requests, tokens, areas: [...areas.values()], topReads: reads.slice(0, TOP_READS) }
}

/** 1234 → 1.2k, 1234567 → 1.2M. */
export function compact(n: number): string {
  if (n < 1000) return String(n)
  if (n < 1_000_000) return `${(n / 1000).toFixed(n < 10_000 ? 1 : 0)}k`
  return `${(n / 1_000_000).toFixed(1)}M`
}
