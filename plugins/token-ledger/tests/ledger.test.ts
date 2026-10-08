import type { On } from 'claude-code'
import { describe, expect, test } from 'claude-code/testing'

import { classify, prepSubcommand } from '../hooks/classify'
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
    expect(prepSubcommand('prep guide 01ABC --json')).toBe('guide')
    expect(prepSubcommand('/usr/local/bin/prep complete 01ABC')).toBe('complete')
    expect(prepSubcommand('go test ./... | grep prep')).toBeUndefined()
    expect(classify('Read', { file_path: '/r/.prep/knowledge/overview.md' }).area).toBe('knowledge')
    expect(classify('Read', { file_path: '/r/.prep/issues/01ABC/context.md' }).area).toBe('issue')
    expect(classify('Read', { file_path: '/r/internal/cli/cli.go' }).area).toBe('code')
    expect(classify('Bash', { command: 'cat .prep/knowledge/components/cli.md' }).area).toBe('knowledge')
    expect(classify('Bash', { command: 'prep show 01ABC' })).toEqual({ target: 'prep show 01ABC', area: 'prep-cli', prep: 'show' })
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
    expect(rows[3]).toMatchObject({ tool: 'Bash', area: 'prep-cli', prep: 'guide' })
    expect(rows[4].contextTokens).toBe(1234)
  })
})

describe('summary', () => {
  test('totals requests and groups tool output by tool and area', () => {
    const text = summarize([
      { type: 'step', t: 0, usage: { input: 10, cacheRead: 100, cacheWrite: 5, output: 7 } },
      { type: 'tool', t: 0, tool: 'Read', area: 'knowledge', resultChars: 800 },
      { type: 'tool', t: 0, tool: 'Bash', area: 'prep-cli', prep: 'guide', resultChars: 400 },
    ])
    expect(text).toContain('1 model requests: input 10, cache read 100, cache write 5, output 7 tokens')
    expect(text).toContain('Read: 1 calls, ~200 tokens')
    expect(text).toContain('knowledge: 1 calls, ~200 tokens')
    expect(text).toContain('guide: 1 calls, ~100 tokens')
  })
})
