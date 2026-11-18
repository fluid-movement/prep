import type { On } from 'claude-code'
import { describe, expect, test } from 'claude-code/testing'


const START = { cwd: '/repo', surface: 'terminal', isInteractive: true } as const

/** Records prep activity add calls; failing makes prep exit 1 like outside a project. */
function fakePrep(on: On, failing = false, errorAt?: 'session.id' | 'session.usage') {
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
  let session = 'sess-1'
  on('session.id', async () => {
    if (errorAt === 'session.id') throw new Error('session id unavailable')
    return { value: session }
  })
  on('session.root', async () => ({ value: '/repo' }))
  on('session.usage', async () => {
    if (errorAt === 'session.usage') throw new Error('usage unavailable')
    return { value: { context: { tokens: 50_000, window: 200_000 }, cost: { usd: 0.3 } } as never }
  })
  return { sent, env, calls: () => calls, setSession: (id: string) => { session = id } }
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
    expect(prep.sent).toEqual([{ kind: 'session', op: 'started', actor: 'claude-code', session: id }])
  })

  test('reports lifecycle and links a changed session ID', async ($, on) => {
    const prep = fakePrep(on)
    on('turn.start', async () => ({ turnId: 'turn-1' } as never))
    on('session.end', async () => ({ sessionId: 'sess-2' } as never))
    await $.session.start(START)
    await settle()
    prep.setSession('sess-2')
    await $.turn.start({ turnId: 'turn-1' } as never)
    await settle()
    expect(prep.env.PREP_SESSION).toBe('sess-2')
    await $.session.end({} as never)
    await settle()
    expect(prep.sent).toEqual([
      { kind: 'session', op: 'started', actor: 'claude-code', session: 'sess-1' },
      { kind: 'session', op: 'ended', actor: 'claude-code', session: 'sess-1' },
      { kind: 'session', op: 'started', actor: 'claude-code', session: 'sess-2', previous: 'sess-1' },
      { kind: 'session', op: 'ended', actor: 'claude-code', session: 'sess-2' },
    ])
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

  test('request payloads distinguish main and subagents; turn totals are summaries', async ($, on) => {
    const prep = fakePrep(on)
    on('turn.step', async function* (_$, e) {
      return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], usage: { input_tokens: 2, output_tokens: 3, model: 'model' } } as never
    })
    on('turn.complete', async () => ({ text: '' }))
    on('session.end', async () => ({ sessionId: 'sess-1' }))
    await $.session.start(START)
    for await (const _ of $.turn.step({ turnId: 't', index: 0, model: 'model' } as never)) {}
    for await (const _ of $.turn.step({ turnId: 't', index: 1, model: 'model', agentId: 'child' } as never)) {}
    await $.turn.complete({ turnId: 't', answer: '' } as never)
    await $.session.end({} as never)
    const requests = prep.sent.filter(e => e.kind === 'request')
    expect(requests.length).toBe(2)
    expect(requests[0].agent_id).toBe(undefined)
    expect(requests[1].agent_id).toBe('child')
    expect(requests[1].turn_id).toBe('t')
    expect(prep.sent.filter(e => e.kind === 'summary' && e.context).length).toBe(1)
  })

  test('a failed send writes a durable gap independently and stops sending', async ($, on) => {
    const prep = fakePrep(on, true)
    const gaps: Record<string, unknown>[] = []
    on('turn.start', async () => ({ turnId: 't' } as never))
    on('fs.exists', async () => ({ value: true }))
    on('fs.write', async (_$, e) => {
      gaps.push(e as unknown as Record<string, unknown>)
      return { value: undefined }
    })
    await $.session.start(START)
    await settle()
    expect(gaps.length).toBe(1)
    expect(JSON.stringify(gaps[0])).toContain('bridge-disabled')
    expect(JSON.stringify(gaps[0])).toContain('request-gap-sess-1.json')
    const calls = prep.calls()
    await $.turn.start({ turnId: 't' } as never)
    await settle()
    expect(prep.calls()).toBe(calls)
  })

  test('a request without usage is a gap rather than zero tokens', async ($, on) => {
    const prep = fakePrep(on)
    on('turn.step', async function* (_$, e) {
      return { turnId: e.turnId, index: e.index, answer: '', toolUses: [] } as never
    })
    on('session.end', async () => ({ sessionId: 'sess-1' }))
    await $.session.start(START)
    for await (const _ of $.turn.step({ turnId: 't', index: 0, model: 'model', agentId: 'child' } as never)) {}
    await $.session.end({} as never)
    const gaps = prep.sent.filter(e => e.kind === 'gap')
    expect(gaps.length).toBe(1)
    expect(gaps[0].agent_id).toBe('child')
    expect(gaps[0].tokens).toBe(undefined)
    expect(prep.sent.filter(e => e.kind === 'request').length).toBe(0)
  })

  test('empty and zero usage become model gaps; each non-zero counter is a request', async ($, on) => {
    const prep = fakePrep(on)
    const usages = [
      {},
      { model: 'usage-model' },
      { model: 'usage-model', input_tokens: 0, output_tokens: 0, cache_read_input_tokens: 0, cache_creation_input_tokens: 0 },
      { input_tokens: 1 },
      { output_tokens: 1 },
      { cache_read_input_tokens: 1 },
      { cache_creation_input_tokens: 1 },
    ]
    on('turn.step', async function* (_$, e) {
      return { turnId: e.turnId, index: e.index, answer: '', toolUses: [], usage: usages[e.index] } as never
    })
    on('session.end', async () => ({ sessionId: 'sess-1' }))
    await $.session.start(START)
    for (let index = 0; index < usages.length; index++) {
      for await (const _ of $.turn.step({ turnId: 't', index, model: 'event-model', agentId: 'child' } as never)) {}
    }
    await $.session.end({} as never)
    expect(prep.sent.filter(e => e.kind === 'gap')).toEqual([
      { kind: 'gap', actor: 'claude-code', session: 'sess-1', agent_id: 'child', turn_id: 't', op: 'event-model' },
      { kind: 'gap', actor: 'claude-code', session: 'sess-1', agent_id: 'child', turn_id: 't', op: 'usage-model' },
      { kind: 'gap', actor: 'claude-code', session: 'sess-1', agent_id: 'child', turn_id: 't', op: 'usage-model' },
    ])
    expect(prep.sent.filter(e => e.kind === 'request').length).toBe(4)
  })

  test('entry markers precede work, step once per session and complete once per turn', async ($, on) => {
    const prep = fakePrep(on)
    on('turn.start', async () => ({ turnId: 't' } as never))
    on('turn.step', async function* (_$, e) {
      await settle()
      expect(prep.sent.filter(e => e.op === 'turn.step').length).toBe(e.index === 2 ? 2 : 1)
      return { turnId: e.turnId, index: e.index, answer: '', toolUses: [] } as never
    })
    on('turn.complete', async () => {
      await settle()
      expect(prep.sent.filter(e => e.op === 'turn.complete').length).toBe(completed + 1)
      return { text: '' }
    })
    on('session.end', async () => ({ sessionId: 'sess-2' }))
    let completed = 0
    await $.session.start(START)
    for (let index = 0; index < 3; index++) {
      if (index === 2) {
        prep.setSession('sess-2')
        await $.turn.start({ turnId: 't' } as never)
      }
      for await (const _ of $.turn.step({ turnId: 't', index, model: 'model' } as never)) {}
      await $.turn.complete({ turnId: 't', answer: '' } as never)
      completed++
    }
    await $.session.end({} as never)
    expect(prep.sent.filter(e => e.status === 'entered')).toEqual([
      { kind: 'summary', actor: 'claude-code', session: 'sess-1', op: 'turn.step', status: 'entered' },
      { kind: 'summary', actor: 'claude-code', session: 'sess-1', op: 'turn.complete', status: 'entered' },
      { kind: 'summary', actor: 'claude-code', session: 'sess-1', op: 'turn.complete', status: 'entered' },
      { kind: 'summary', actor: 'claude-code', session: 'sess-2', op: 'turn.step', status: 'entered' },
      { kind: 'summary', actor: 'claude-code', session: 'sess-2', op: 'turn.complete', status: 'entered' },
    ])
  })

  test('a turn usage failure leaves its entry marker and a visible error', async ($, on) => {
    const prep = fakePrep(on, false, 'session.usage')
    on('turn.complete', async () => ({ text: '' }))
    await $.session.start(START)
    await $.turn.complete({ turnId: 't', answer: '' } as never)
    await settle()
    expect(prep.sent.filter(e => e.op === 'turn.complete')).toEqual([
      { kind: 'summary', actor: 'claude-code', session: 'sess-1', op: 'turn.complete', status: 'entered' },
      { kind: 'gap', actor: 'claude-code', session: 'sess-1', op: 'turn.complete', status: 'error', verb: 'error', target: 'HooksError: no implementation for session.usage' },
    ])
  })

  test('a session start failure leaves a visible error even before session attribution', async ($, on) => {
    const prep = fakePrep(on, false, 'session.id')
    await $.session.start(START)
    await settle()
    expect(prep.sent).toEqual([
      { kind: 'gap', actor: 'claude-code', op: 'session.start', status: 'error', verb: 'error', target: 'HooksError: no implementation for session.id' },
    ])
  })

  test('a failed request sends an attributed request-interrupted gap', async ($, on) => {
    const prep = fakePrep(on)
    on('turn.step', async function* () {
      throw new Error('request failed')
    })
    on('session.end', async () => ({ sessionId: 'sess-1' }))
    await $.session.start(START)
    let failed = false
    try {
      for await (const _ of $.turn.step({ turnId: 't', index: 0, model: 'model', agentId: 'child' } as never)) {}
    } catch {
      failed = true
    }
    await $.session.end({} as never)
    expect(failed).toBe(true)
    expect(prep.sent.filter(e => e.kind === 'gap')).toEqual([
      { kind: 'gap', actor: 'claude-code', session: 'sess-1', agent_id: 'child', turn_id: 't', op: 'request-interrupted' },
    ])
    expect(prep.sent.filter(e => e.kind === 'request').length).toBe(0)
  })
})
