import type { On } from 'claude-code'
import { describe, expect, test } from 'claude-code/testing'

import { toolEvent } from '../hooks/register'

const START = { cwd: '/repo', surface: 'terminal', isInteractive: true } as const

/** Records prep activity add calls; failing makes prep exit 1 like outside a project. */
function fakePrep(on: On, failing = false) {
  const sent: Record<string, unknown>[] = []
  let calls = 0
  on('process.run', async (_$, e) => {
    calls++
    if (e.argv[1] === 'activity' && e.argv[2] === 'add') {
      for (const line of (e.init?.stdin ?? '').split('\n').filter(Boolean)) sent.push(JSON.parse(line))
    }
    return {
      value: {
        exitCode: failing ? 1 : 0,
        stdout: '',
        stderr: failing ? 'prep: no .prep directory [E_NO_PROJECT]' : '',
        isStdoutTruncated: false,
        isStderrTruncated: false,
      },
    }
  })
  const env: Record<string, string | undefined> = {}
  on('env.set', async (_$, e) => {
    env[e.name] = e.value
    return { value: undefined }
  })
  on('session.start', async (_$, e) => ({ cwd: e.cwd }))
  on('session.id', async () => ({ value: 'sess-1' }))
  on('session.usage', async () => ({ value: { context: { tokens: 50_000, window: 200_000 }, cost: { usd: 0.3 } } as never }))
  return { sent, env, calls: () => calls }
}

async function settle() {
  for (let k = 0; k < 5; k++) await new Promise(r => setTimeout(r, 0))
}

describe('tool events', () => {
  test('name what a call touched, in the feed words and areas', () => {
    expect(toolEvent('Edit', { file_path: '/repo/internal/cli/cli.go' }, 40, false)).toMatchObject({ kind: 'tool', op: 'Edit', verb: 'edited', target: '/repo/internal/cli/cli.go', area: 'code', chars: 40 })
    expect(toolEvent('Read', { file_path: '/repo/.prep/knowledge/overview.md' }, 900, false)).toMatchObject({ verb: 'read', area: 'knowledge' })
    expect(toolEvent('Bash', { command: 'go test ./...\necho done' }, 300, true)).toMatchObject({ verb: 'ran', target: 'go test ./...', area: 'code', failed: true })
    expect(toolEvent('WebFetch', { url: 'https://example.com' }, 10, false)).toMatchObject({ verb: 'fetched', target: 'https://example.com', area: 'other' })
  })

  test("leave the agent's own prep commands to prep", () => {
    expect(toolEvent('Bash', { command: 'cd /repo && prep guide 2NGSFZ --by claude-code/opus' }, 500, false)).toBeUndefined()
  })
})

describe('bridge', () => {
  test('exports the session and sends tool calls to prep activity add', async ($, on) => {
    const prep = fakePrep(on)
    on('tool.call', async () => ({ result: {} as never, text: 'some output' }))
    await $.session.start(START)
    const id = 'sess-1'
    expect(prep.env.PREP_SESSION).toBe(id)
    await $.tool.call({ tool: 'Edit', tool_use_id: 't1', file_path: '/repo/main.go', old_string: 'a', new_string: 'b' } as never)
    await $.tool.call({ tool: 'Bash', tool_use_id: 't2', command: 'prep show 2NGSFZ --by claude-code/x' } as never)
    await settle()
    expect(prep.sent).toHaveLength(1)
    expect(prep.sent[0]).toMatchObject({ kind: 'tool', op: 'Edit', area: 'code', chars: 11, actor: 'claude-code', session: id })
  })

  test('outside a prep project the first failure turns it off', async ($, on) => {
    const prep = fakePrep(on, true)
    on('tool.call', async () => ({ result: {} as never, text: 'x' }))
    await $.session.start(START)
    await $.tool.call({ tool: 'Edit', tool_use_id: 't1', file_path: '/a.go', old_string: 'a', new_string: 'b' } as never)
    await settle()
    const after = prep.calls()
    await $.tool.call({ tool: 'Edit', tool_use_id: 't2', file_path: '/b.go', old_string: 'a', new_string: 'b' } as never)
    await settle()
    expect(prep.calls()).toBe(after)
  })
})
