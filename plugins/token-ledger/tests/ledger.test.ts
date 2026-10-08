import type { On } from 'claude-code'
import { describe, expect, test } from 'claude-code/testing'

import { classify, prepSubcommands } from '../hooks/classify'
import { summarize } from '../hooks/register'

/** An in-memory file system and the session facts the ledger asks for. */
function fakeHost(on: On) {
  const files = new Map<string, string>()
  on('fs.exists', async (_$, e) => ({ value: files.has(e.path) || e.path.endsWith('/.prep') }))
  on('fs.read', async (_$, e) => ({ value: files.get(e.path) ?? '' }) as never)
  on('fs.write', async (_$, e) => {
    files.set(e.path, e.text)
    return { value: undefined }
  })
  on('env.get', async () => ({ value: '/home/u' }))
  on('session.id', async () => ({ value: 's1' }))
  on('session.usage', async () => ({ value: { startedAt: 0, context: { tokens: 1234, window: 200000 }, rateLimits: [] } }) as never)
  on('clock.now', async () => ({ value: 1 }))
  on('command.register', async (_$, e) => ({ value: { command: e.name } }) as never)
  on('session.start', async (_$, e) => ({ cwd: e.cwd }))
  on('turn.start', async (_$, e) => ({ turnId: e.turnId }))
  on('turn.complete', async () => ({ text: '' }) as never)
  on('tool.call', async () => ({ result: 'x'.repeat(400), text: 'x'.repeat(400) }) as never)
  return files
}

describe('classify', () => {
  test('prep commands, knowledge and issue paths are told apart', () => {
    expect(prepSubcommands('prep guide 01ABC --json')).toEqual(['guide'])
    expect(prepSubcommands('/usr/local/bin/prep complete 01ABC')).toEqual(['complete'])
    expect(prepSubcommands('go test ./... | grep prep')).toEqual([])
    expect(prepSubcommands('cd /r && prep guide 01ABC')).toEqual(['guide'])
    expect(prepSubcommands('I=01ABC B=x && prep criterion $I --add y && prep ready $I')).toEqual(['criterion', 'ready'])
    expect(prepSubcommands('cd /r && PREP_ACTOR=x prep show 01ABC | head')).toEqual(['show'])
    expect(prepSubcommands("echo prep guide; cat <<'EOF'\nprep init\nEOF")).toEqual([])
    expect(classify('Read', { file_path: '/r/.prep/knowledge/overview.md' }).area).toBe('knowledge')
    expect(classify('Read', { file_path: '/r/.prep/issues/01ABC/context.md' }).area).toBe('issue')
    expect(classify('Read', { file_path: '/r/internal/cli/cli.go' }).area).toBe('code')
    expect(classify('Bash', { command: 'cat .prep/knowledge/components/cli.md' }).area).toBe('knowledge')
    const ctx = classify('Bash', { command: "prep context 01ABC --body-file - <<'EOF'\nlong text\nEOF" })
    expect(ctx).toEqual({ target: "prep context 01ABC --body-file - <<'EOF' …", area: 'prep-cli', prep: ['context'] })
    expect(classify('Bash', { command: 'prep show 01ABC' })).toEqual({ target: 'prep show 01ABC', area: 'prep-cli', prep: ['show'] })
  })
})

describe('ledger', () => {
  test('tool calls are recorded and written when the turn completes', async ($, on) => {
    const files = fakeHost(on)
    await $.session.start({ cwd: '/r', surface: 'terminal', isInteractive: true })
    await $.turn.start({ text: 'hi', turnId: 'turn1' })
    await $.tool.call({ tool: 'Read', tool_use_id: 't1', file_path: '/r/.prep/knowledge/overview.md' } as never)
    await $.tool.call({ tool: 'Bash', tool_use_id: 't2', command: 'prep guide 01ABC' } as never)
    expect(files.size).toBe(0)
    await $.turn.complete({ turnId: 'turn1', answer: '', reason: 'answer' } as never)

    const rows = files.get('/home/u/.claude/token-ledger/s1.jsonl')!.trim().split('\n').map(l => JSON.parse(l))
    expect(rows.map(r => r.type)).toEqual(['session', 'turn', 'tool', 'tool', 'turn.end'])
    expect(rows[0].isPrep).toBe(true)
    expect(rows[2]).toMatchObject({ tool: 'Read', area: 'knowledge', resultChars: 400, turnId: 'turn1' })
    expect(rows[3]).toMatchObject({ tool: 'Bash', area: 'prep-cli', prep: ['guide'] })
    expect(rows[4].contextTokens).toBe(1234)
  })
})

describe('issues', () => {
  test('rows carry the issue a prep write, guide or issue edit moved to', async ($, on) => {
    const files = fakeHost(on)
    const A = '01M4AHPY6HZ0ZFMMH349V1V2TC'
    const B = '01M4AJ9DN4CK5S4WX2N68VBR98'
    await $.session.start({ cwd: '/r', surface: 'terminal', isInteractive: true })
    await $.turn.start({ text: 'hi', turnId: 'turn1' })
    await $.tool.call({ tool: 'Read', tool_use_id: 't1', file_path: '/r/README.md' } as never)
    await $.tool.call({ tool: 'Bash', tool_use_id: 't2', command: `prep guide ${A}` } as never)
    await $.tool.call({ tool: 'Bash', tool_use_id: 't3', command: `prep show ${B}` } as never)
    await $.tool.call({ tool: 'Edit', tool_use_id: 't4', file_path: `/r/.prep/issues/${B}/context.md` } as never)
    await $.turn.complete({ turnId: 'turn1', answer: '', reason: 'answer' } as never)

    const rows = files.get('/home/u/.claude/token-ledger/s1.jsonl')!.trim().split('\n').map(l => JSON.parse(l))
    expect(rows.filter(r => r.type === 'tool').map(r => r.issue)).toEqual([undefined, A, A, B])
  })
})

describe('summary', () => {
  test('totals requests and groups tool output by tool, area and issue', () => {
    const text = summarize([
      { type: 'step', t: 0, usage: { input: 10, cacheRead: 100, cacheWrite: 5, output: 7 } },
      { type: 'step', t: 0, issue: 'A', usage: { input: 1, cacheRead: 1000, cacheWrite: 0, output: 9 } },
      { type: 'tool', t: 0, tool: 'Read', area: 'knowledge', issue: 'A', resultChars: 800 },
      { type: 'tool', t: 0, tool: 'Bash', area: 'prep-cli', prep: ['guide'], resultChars: 400 },
      { type: 'tool', t: 0, tool: 'Bash', area: 'prep-cli', prep: ['ready', 'claim'], resultChars: 40 },
    ])
    expect(text).toContain('2 model requests: input 11, cache read 1100, cache write 5, output 16 tokens')
    expect(text).toContain('Read: 1 calls, ~200 tokens')
    expect(text).toContain('knowledge: 1 calls, ~200 tokens')
    expect(text).toContain('guide: 1 calls, ~100 tokens')
    expect(text).toContain('ready + claim: 1 calls, ~10 tokens')
    expect(text).toContain('A: 1 requests, 1010 tokens; 1 tool calls, ~200 tokens')
    expect(text).toContain('(none): 1 requests, 122 tokens; 2 tool calls, ~110 tokens')
  })
})
