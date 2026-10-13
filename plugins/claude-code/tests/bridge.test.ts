import type { On } from 'claude-code'
import { describe, expect, test } from 'claude-code/testing'


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

describe('bridge', () => {
  test('exports the session and leaves tool events to the shared command', async ($, on) => {
    const prep = fakePrep(on)
    on('tool.call', async () => ({ result: {} as never, text: 'some output' }))
    await $.session.start(START)
    const id = 'sess-1'
    expect(prep.env.PREP_SESSION).toBe(id)
    await $.tool.call({ tool: 'Edit', tool_use_id: 't1', file_path: '/repo/main.go', old_string: 'a', new_string: 'b' } as never)
    await $.tool.call({ tool: 'Bash', tool_use_id: 't2', command: 'prep show 2NGSFZ --by claude-code/x' } as never)
    await settle()
    expect(prep.sent).toHaveLength(0)
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
