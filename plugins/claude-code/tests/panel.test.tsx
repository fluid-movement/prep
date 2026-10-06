import type { On } from 'claude-code'
import { describe, expect, mock, test } from 'claude-code/testing'
import type { Engine } from 'claude-code/testing'

import { issueFromCommand, issueFromPath } from '../hooks/infer'

// Fixtures: prep's JSON for a project with one parent and one child issue.
const PARENT = '20261005-152616'
const CHILD = '20261005-152621'
const OTHER = '20261005-190543'

const summary = (id: string, title: string, state: string, extra = {}) => ({
  id,
  title,
  kind: 'code',
  state,
  stale: false,
  blocked: false,
  actionable: false,
  children: 0,
  ...extra,
})

const prime = (claims: { id: string; title: string; by: string }[] = []) => ({
  issues: 3,
  actionable: 1,
  parents: [summary(PARENT, 'TUI parent', 'open', { children: 1, progress: { done: 0, dropped: 0, total: 1 } })],
  next: [summary(OTHER, 'Other work', 'ready')],
  stale: [],
  claims,
  check: { errors: 0, warnings: 1 },
})

const show = (id: string) => ({
  state: 'in_progress',
  stale: false,
  blocked: false,
  children: [],
  blocks: [OTHER],
  definition_of_done: ['go test ./... passes'],
  history:
    '- 2026-10-06T10:00:00Z claude-code/2: moved the watcher\n- 2026-10-06T11:00:00Z claude-code/2: wrote the panel',
  issue: {
    id,
    title: id === CHILD ? 'Side panel' : 'Other work',
    kind: 'code',
    tags: ['integration'],
    parent: PARENT,
    requirement: 'A side panel shows what the agent works on.',
    open_questions: id === OTHER ? '- Which colors?' : '',
    criteria: [
      { text: 'prep watch streams', checked: true },
      { text: 'the panel follows the agent', checked: false },
    ],
    decisions: [{ id: 'D1', title: 'Infer the current issue' }],
  },
})

const guide = {
  step: 'implement',
  transitions: [
    { op: 'release', command: `prep release ${CHILD}`, allowed: true },
    {
      op: 'complete',
      command: `prep complete ${CHILD} --commit <ref>`,
      allowed: false,
      unmet: [{ code: 'G_UNCHECKED', message: '1 acceptance criteria unchecked' }],
    },
  ],
  knowledge: [{ path: '/components/claude-code.md', title: 'Claude Code integration' }],
}

const list = {
  issues: [
    summary(PARENT, 'TUI parent', 'open', { children: 1, progress: { done: 0, dropped: 0, total: 1 } }),
    summary(CHILD, 'Side panel', 'in_progress', { parent: PARENT }),
    summary(OTHER, 'Other work', 'ready'),
  ],
}

// prep list --tree with the unresolved states: rows in tree order with depth.
const tree = {
  issues: [
    summary(PARENT, 'TUI parent', 'open', { children: 1, progress: { done: 0, dropped: 0, total: 1 } }),
    summary(CHILD, 'Side panel', 'in_progress', { parent: PARENT, depth: 1, stale: true }),
    summary(OTHER, 'Other work', 'ready', { actionable: true }),
  ],
}

type Project = {
  claims?: { id: string; title: string; by: string }[]
  isPrep?: boolean
  store?: Record<string, unknown>
}

/** Answers the plugin's prep calls from the fixtures and records them. */
function fakePrep(on: On, project: Project = {}) {
  const calls: string[][] = []
  const ok = (v: unknown) => ({
    exitCode: 0,
    stdout: JSON.stringify(v),
    stderr: '',
    isStdoutTruncated: false,
    isStderrTruncated: false,
  })
  on('process.run', async (_$, e) => ({ value: run([...e.argv]) }))
  const run = (argv: string[]) => {
    calls.push(argv)
    if (project.isPrep === false) return { ...ok({}), exitCode: 1, stdout: '', stderr: 'no .prep found' }
    const [, sub, ref] = argv
    const full = [CHILD, OTHER, PARENT].find(id => ref !== undefined && id.endsWith(ref))
    switch (sub) {
      case 'prime':
        return ok(prime(project.claims))
      case 'show':
        return full ? ok(show(full)) : { ...ok({}), exitCode: 1, stdout: '', stderr: `no issue ${ref}` }
      case 'guide':
        // A resolved issue: prep writes null for its empty lists.
        return ok(full === PARENT ? { ...guide, step: 'resolved', transitions: null, knowledge: null } : guide)
      case 'list':
        return ok(argv.includes('--tree') ? tree : list)
    }
    return { ...ok({}), exitCode: 2 }
  }
  on('process.spawn', async function* () {
    return { value: { code: 0, signal: null } }
  })
  on('session.start', async (_$, e) => ({ cwd: e.cwd }))
  mock.store(on, project.store)
  return calls
}

/** Tracks the panes the plugin opens and closes, as the surface would. */
function fakePanes(on: On) {
  const open = new Set<string>()
  on('ui.open', async (_$, e) => {
    open.add(e.id)
    return { value: { isPlaced: true as const } }
  })
  on('ui.close', async (_$, e) => {
    open.delete(e.id)
    return { value: undefined }
  })
  on('ui.panes', async () => ({
    value: [...open].map(id => ({ id, title: id, isShown: true, isFocused: false, isPlaced: true })),
  }))
  return open
}

const START = { cwd: '/repo', surface: 'terminal', isInteractive: true } as const

// One mounted pane per test: it redraws as the plugin's state changes.
const mounted = new WeakMap<object, ReturnType<typeof mount>>()

function pane($: Engine) {
  if (!mounted.has($)) mounted.set($, mount($))
  return mounted.get($)!
}

async function drawn($: Engine) {
  const ui = await pane($)
  return (await ui.findAll({})).map(el => el.text ?? '').join('\n')
}

async function mount($: Engine) {
  return $.ui.mount({
    plugin: 'prep',
    surface: 'terminal',
    component: 'Pane',
    requestId: 'prep',
    props: { title: 'prep', isFocused: false, placement: 'dock', bodyColumns: 60 } as never,
    viewport: { columns: 140, rows: 40 } as never,
  })
}

async function bash($: Engine, command: string) {
  return $.tool.call({ tool: 'Bash', tool_use_id: `t${Math.random()}`, command } as never)
}

describe('inference', () => {
  test('prep writes and guide name the current issue; reads do not', () => {
    expect(issueFromCommand(`prep guide ${CHILD}`)).toBe(CHILD)
    expect(issueFromCommand('prep criterion 152621 --check 2 --by claude-code/2')).toBe('152621')
    expect(issueFromCommand(`cd /repo && PREP_ACTOR=x prep log ${CHILD} "did it"`)).toBe(CHILD)
    expect(issueFromCommand(`prep context --by claude-code/2 ${CHILD} --body-file -`)).toBe(CHILD)
    expect(issueFromCommand(`prep show ${CHILD}`)).toBeUndefined()
    expect(issueFromCommand('prep list --json && prep next')).toBeUndefined()
    expect(issueFromCommand('go test ./... | grep prep')).toBeUndefined()
  })

  test('the last issue-naming prep command wins', () => {
    expect(issueFromCommand(`prep guide ${OTHER}; prep claim ${CHILD}`)).toBe(CHILD)
  })

  test('heredoc bodies are not commands', () => {
    const cmd = `prep context ${CHILD} --body-file - <<'EOF'\nprep guide ${OTHER}\nEOF`
    expect(issueFromCommand(cmd)).toBe(CHILD)
  })

  test('prep new takes the id from its output', () => {
    expect(issueFromCommand('prep new --title T --kind code --json', `{"ok":true,"op":"new","id":"${OTHER}"}`)).toBe(
      OTHER,
    )
    expect(issueFromCommand('prep new --title T --kind code', '')).toBeUndefined()
  })

  test('edits under .prep/issues name their issue', () => {
    expect(issueFromPath(`/repo/.prep/issues/${CHILD}/context.md`)).toBe(CHILD)
    expect(issueFromPath('/repo/.prep/knowledge/overview.md')).toBeUndefined()
    expect(issueFromPath('/repo/internal/cli/watch.go')).toBeUndefined()
  })
})

describe('panel', () => {
  test('without a focus or claim it shows the project overview', async ($, on) => {
    fakePrep(on)
    fakePanes(on)
    await $.session.start(START)
    const text = await drawn($)
    expect(text).toContain('Actionable')
    expect(text).toContain('Other work')
    expect(text).toContain('Top-level parents')
    expect(text).toContain('check: 0 errors, 1 warnings')
  })

  test('falls back to an issue claimed in the repository', async ($, on) => {
    fakePrep(on, { claims: [{ id: CHILD, title: 'Side panel', by: 'claude-code/2' }] })
    fakePanes(on)
    await $.session.start(START)
    const text = await drawn($)
    expect(text).toContain('Side panel')
    expect(text).toContain('claimed here')
  })

  test('follows prep writes and guide, ignores reads and failed commands', async ($, on) => {
    fakePrep(on)
    fakePanes(on)
    let failing = false
    on('tool.call', async () =>
      failing
        ? { result: {} as never, text: 'Exit code 1', isError: true as const }
        : { result: {} as never, text: 'ok' },
    )
    await $.session.start(START)

    await bash($, `prep show ${CHILD} --json`)
    expect(await drawn($)).toContain('no issue in focus')

    await bash($, `prep guide ${CHILD}`)
    let text = await drawn($)
    expect(text).toContain('Side panel')
    expect(text).toContain('following the agent')

    failing = true
    await bash($, `prep claim ${OTHER}`)
    expect(await drawn($)).toContain('Side panel')

    failing = false
    await $.tool.call({
      tool: 'Edit',
      tool_use_id: 'e1',
      file_path: `/repo/.prep/issues/${OTHER}/context.md`,
      old_string: 'a',
      new_string: 'b',
    } as never)
    text = await drawn($)
    expect(text).toContain('Other work')
    expect(text).toContain('Which colors?')
  })

  test('draws a resolved issue, whose guide has no transitions', async ($, on) => {
    fakePrep(on)
    fakePanes(on)
    on('tool.call', async () => ({ result: {} as never, text: 'ok' }))
    await $.session.start(START)
    await bash($, `prep guide ${PARENT}`)
    const text = await drawn($)
    expect(text).toContain('step resolved')
    expect(text).not.toContain('could not be read')
  })

  test('draws status, next step, checklist, DoD, decisions and surroundings', async ($, on) => {
    fakePrep(on)
    fakePanes(on)
    on('tool.call', async () => ({ result: {} as never, text: 'ok' }))
    await $.session.start(START)
    await bash($, `prep guide ${CHILD}`)
    const text = await drawn($)
    for (const want of [
      CHILD,
      '#integration',
      'in_progress',
      'step implement',
      `prep complete ${CHILD} --commit <ref>`,
      '1 acceptance criteria unchecked',
      'A side panel shows what the agent works on.',
      'Acceptance 1/2',
      '2. the panel follows the agent',
      'go test ./... passes',
      'Infer the current issue',
      'TUI parent',
      'Other work',
      'Claude Code integration',
      'wrote the panel',
    ]) {
      expect(text).toContain(want)
    }
  })

  test('/prep:focus pins an issue and without an id follows the agent again', async ($, on) => {
    fakePrep(on)
    const panes = fakePanes(on)
    on('tool.call', async () => ({ result: {} as never, text: 'ok' }))
    await $.session.start(START)
    await bash($, `prep guide ${CHILD}`)

    const pinned = await $.command.run({ command: 'prep:focus', args: '190543' } as never)
    expect(pinned.text).toContain(`pinned to ${OTHER}`)
    expect(panes.has('prep')).toBe(true)
    await bash($, `prep log ${CHILD} more`)
    let text = await drawn($)
    expect(text).toContain('Other work')
    expect(text).toContain('pinned')

    const bad = await $.command.run({ command: 'prep:focus', args: '999999' } as never)
    expect(bad.text).toContain('no issue 999999')

    await $.command.run({ command: 'prep:focus', args: '' } as never)
    text = await drawn($)
    expect(text).toContain('Side panel')
  })

  test('/prep:pane toggles the pane and remembers it', async ($, on) => {
    fakePrep(on)
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(true)
    await $.command.run({ command: 'prep:pane', args: '' } as never)
    expect(panes.has('prep')).toBe(false)
    await $.command.run({ command: 'prep:pane', args: '' } as never)
    expect(panes.has('prep')).toBe(true)
  })

  test('outside a prep project the plugin stays idle', async ($, on) => {
    fakePrep(on, { isPrep: false })
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(false)
    const out = await $.command.run({ command: 'prep:pane', args: '' } as never)
    expect(out.text).toBe('Not a prep project.')
  })
})

describe('views', () => {
  test('the Live and Project tabs switch the view', async ($, on) => {
    const calls = fakePrep(on)
    fakePanes(on)
    on('tool.call', async () => ({ result: {} as never, text: 'ok' }))
    await $.session.start(START)
    await bash($, `prep guide ${CHILD}`)
    expect(await drawn($)).toContain('step implement')
    expect(calls.some(c => c.includes('--tree'))).toBe(false)

    const ui = await pane($)
    await ui.press({ key: 'tab-project' })
    const text = await drawn($)
    expect(text).toContain('Open issues 3')
    expect(text).toContain('check: 0 errors, 1 warnings')
    expect(text).toContain('  in_progress')
    expect(text).toContain('stale')
    expect(text).toContain('actionable')
    expect(text).not.toContain('step implement')

    await ui.press({ key: 'tab-live' })
    expect(await drawn($)).toContain('step implement')
  })

  test('/prep:pane project opens the pane on the project view', async ($, on) => {
    fakePrep(on)
    const panes = fakePanes(on)
    await $.session.start(START)
    await $.command.run({ command: 'prep:pane', args: '' } as never)
    expect(panes.has('prep')).toBe(false)
    const out = await $.command.run({ command: 'prep:pane', args: 'project' } as never)
    expect(out.text).toContain('project view')
    expect(panes.has('prep')).toBe(true)
    expect(await drawn($)).toContain('Open issues 3')
    await $.command.run({ command: 'prep:pane', args: 'live' } as never)
    expect(await drawn($)).toContain('no issue in focus')
    const bad = await $.command.run({ command: 'prep:pane', args: 'board' } as never)
    expect(bad.text).toContain('unknown view board')
  })
})

describe('pane option', () => {
  test('closed keeps the pane closed at session start', { options: { pane: 'closed' } }, async ($, on) => {
    fakePrep(on)
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(false)
  })

  test('remember reopens the pane as the person left it', async ($, on) => {
    fakePrep(on)
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(true)
    await $.command.run({ command: 'prep:pane', args: '' } as never)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(false)
    await $.command.run({ command: 'prep:pane', args: '' } as never)
    panes.clear()
    await $.session.start(START)
    expect(panes.has('prep')).toBe(true)
  })

  test('remember keeps a pane closed by hand closed', { options: { pane: 'remember' } }, async ($, on) => {
    fakePrep(on, { store: { 'open:/repo': false } })
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(false)
  })

  test('open opens the pane even when it was closed by hand', { options: { pane: 'open' } }, async ($, on) => {
    fakePrep(on, { store: { 'open:/repo': false } })
    const panes = fakePanes(on)
    await $.session.start(START)
    expect(panes.has('prep')).toBe(true)
  })
})
