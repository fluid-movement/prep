// Draws the panel from a snapshot: the current issue or the project overview.

import type { Color, Elements, RenderChildren, RenderElement } from 'claude-code'

import type { PrepGuide, PrepPrime, PrepShow, PrepSnapshot, PrepSummary } from '../types'

type Kit = Pick<Elements['terminal'], 'Box' | 'Text' | 'Markdown'>

// The TUI design system's tokens, mapped onto the theme keys of Claude Code
// so the panel follows the person's light or dark theme.
const ACCENT: Color = 'claude'
const STATE: Record<string, Color> = {
  open: 'inactive',
  defined: 'suggestion',
  ready: 'ide',
  in_progress: 'warning',
  done: 'success',
  dropped: 'subtle',
}

// How much of each section shows; the pane scrolls past the rest.
const REQUIREMENT_LINES = 14
const LIST_ROWS = 8
const HISTORY_LINES = 3

const SOURCE = { pinned: 'pinned', agent: 'following the agent', claimed: 'claimed here' }

export function drawPanel(kit: Kit, snap: PrepSnapshot | null): RenderElement {
  const { Box, Text } = kit
  if (snap === null) return <Text dimColor>Loading prep…</Text>
  if (snap.view === 'error') {
    return (
      <Box flexDirection="column">
        <Text color="error">prep could not be read</Text>
        <Text dimColor>{snap.message}</Text>
      </Box>
    )
  }
  return snap.view === 'issue'
    ? drawIssue(kit, snap.show, snap.guide, snap.issues, SOURCE[snap.source])
    : drawOverview(kit, snap.prime)
}

function drawIssue(kit: Kit, show: PrepShow, guide: PrepGuide, issues: PrepSummary[], source: string): RenderElement {
  const { Box, Text, Markdown } = kit
  const i = show.issue
  const byId = new Map(issues.map(s => [s.id, s]))
  const criteria = i.criteria ?? []
  const checked = criteria.filter(c => c.checked).length
  const next = nextTransition(guide, show.stale)
  const requirement = clip(i.requirement.trim(), REQUIREMENT_LINES)
  const history = (show.history ?? '')
    .split('\n')
    .filter(l => l.trim() !== '')
    .slice(-HISTORY_LINES)

  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold color={ACCENT}>
          {i.title}
        </Text>
        <Text dimColor wrap="truncate-end">
          {[i.id, i.kind, ...(i.tags ?? []).map(t => `#${t}`)].join(' · ')}
        </Text>
        <Text>
          <Text color={STATE[show.state] ?? 'text'}>{show.state}</Text>
          <Text dimColor>{` · step ${guide.step} · ${source}`}</Text>
        </Text>
        {show.stale && <Text color="warning">stale: the requirement changed since its baseline</Text>}
        {show.blocked && <Text color="error">blocked by unfinished dependencies</Text>}
      </Box>

      {next && (
        <Box flexDirection="column">
          <Text>
            <Text dimColor>next </Text>
            <Text color={next.allowed ? 'success' : 'warning'}>{next.command}</Text>
          </Text>
          {(next.unmet ?? []).map(u => (
            <Text color="warning">{`  ${u.message}`}</Text>
          ))}
        </Box>
      )}

      {section(kit, 'Requirement', [
        <Markdown text={requirement.text || '(none)'} />,
        requirement.more > 0 && <Text dimColor>{`… ${requirement.more} more lines`}</Text>,
      ])}

      {(i.open_questions ?? '').trim() !== '' &&
        section(kit, 'Open questions', [<Markdown text={(i.open_questions ?? '').trim()} />], 'warning')}

      {criteria.length > 0 &&
        section(kit, `Acceptance ${checked}/${criteria.length}`, [
          criteria.map((c, n) => (
            <Text>
              <Text color={c.checked ? 'success' : 'inactive'}>{c.checked ? '✓ ' : '○ '}</Text>
              <Text dimColor={c.checked}>{`${n + 1}. ${c.text}`}</Text>
            </Text>
          )),
        ])}

      {show.definition_of_done.length > 0 &&
        section(kit, 'Definition of Done', [show.definition_of_done.map(d => <Text dimColor>{`· ${d}`}</Text>)])}

      {(i.decisions ?? []).length > 0 &&
        section(kit, 'Decisions', [
          (i.decisions ?? []).map(d => (
            <Text>
              <Text dimColor>{`${d.id} `}</Text>
              {d.title}
            </Text>
          )),
        ])}

      {section(kit, 'Surroundings', [
        i.parent ? relation(kit, 'parent', byId.get(i.parent), i.parent) : <Text dimColor>top level</Text>,
        (i.depends_on ?? []).map(d => relation(kit, 'depends on', byId.get(d), d)),
        show.blocks.slice(0, LIST_ROWS).map(b => relation(kit, 'blocks', byId.get(b), b)),
        show.children.slice(0, LIST_ROWS).map(c => relation(kit, 'child', byId.get(c), c)),
        show.children.length > LIST_ROWS && (
          <Text dimColor>{`  … ${show.children.length - LIST_ROWS} more children`}</Text>
        ),
        (guide.knowledge ?? []).slice(0, LIST_ROWS).map(k => (
          <Text wrap="truncate-end">
            <Text dimColor>knowledge </Text>
            {k.title}
          </Text>
        )),
      ])}

      {history.length > 0 && section(kit, 'History', [history.map(l => <Text dimColor>{l.replace(/^- /, '')}</Text>)])}
    </Box>
  )
}

function drawOverview(kit: Kit, p: PrepPrime): RenderElement {
  const { Box, Text } = kit
  const attention = p.stale.length > 0 || p.check.errors > 0 || p.check.warnings > 0
  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold color={ACCENT}>
          prep
        </Text>
        <Text dimColor>{`${p.issues} issues · ${p.actionable} actionable · no issue in focus`}</Text>
      </Box>
      {p.bootstrap && <Text color="warning">{p.bootstrap}</Text>}
      {p.claims.length > 0 &&
        section(kit, 'In progress', [
          p.claims.slice(0, LIST_ROWS).map(c => (
            <Text wrap="truncate-end">
              <Text dimColor>{`${c.id} `}</Text>
              {c.title}
            </Text>
          )),
        ])}
      {p.next.length > 0 && section(kit, 'Actionable', [p.next.slice(0, LIST_ROWS).map(s => row(kit, s))])}
      {attention &&
        section(
          kit,
          'Attention',
          [
            (p.check.errors > 0 || p.check.warnings > 0) && (
              <Text>{`check: ${p.check.errors} errors, ${p.check.warnings} warnings`}</Text>
            ),
            p.stale.slice(0, LIST_ROWS).map(s => row(kit, s, 'stale')),
          ],
          'warning',
        )}
      {p.parents.length > 0 && section(kit, 'Top-level parents', [p.parents.slice(0, LIST_ROWS).map(s => row(kit, s))])}
    </Box>
  )
}

function section(kit: Kit, title: string, body: RenderChildren, color?: Color): RenderElement {
  const { Box, Text } = kit
  return (
    <Box flexDirection="column">
      <Text bold color={color}>
        {title}
      </Text>
      {body}
    </Box>
  )
}

/** One issue as a row: state, title, progress. */
function row(kit: Kit, s: PrepSummary, note?: string): RenderElement {
  const { Text } = kit
  const progress = s.progress ? ` ${s.progress.done}/${s.progress.total}` : ''
  return (
    <Text wrap="truncate-end">
      <Text color={STATE[s.state] ?? 'text'}>{pad(s.state)}</Text>
      {s.title}
      <Text dimColor>{progress + (note ? ` (${note})` : '')}</Text>
    </Text>
  )
}

function relation(kit: Kit, label: string, s: PrepSummary | undefined, id: string): RenderElement {
  const { Text } = kit
  if (!s) return <Text dimColor>{`${label} ${id}`}</Text>
  const progress = s.progress ? ` ${s.progress.done}/${s.progress.total}` : ''
  return (
    <Text wrap="truncate-end">
      <Text dimColor>{`${label} `}</Text>
      <Text color={STATE[s.state] ?? 'text'}>{`${s.state} `}</Text>
      {s.title}
      <Text dimColor>{progress}</Text>
    </Text>
  )
}

function pad(state: string): string {
  return (state + ' ').padEnd(12)
}

/**
 * The transition the step leads to: acknowledging a stale issue, else the
 * first one that moves the issue forward (release and drop are ways out).
 */
export function nextTransition(guide: PrepGuide, stale: boolean): PrepGuide['transitions'][number] | undefined {
  const t = guide.transitions
  if (stale) return t.find(x => x.op === 'ack') ?? t[0]
  return t.find(x => !['ack', 'release', 'drop'].includes(x.op))
}

function clip(text: string, lines: number): { text: string; more: number } {
  const all = text.split('\n')
  if (all.length <= lines) return { text, more: 0 }
  return { text: all.slice(0, lines).join('\n'), more: all.length - lines }
}
