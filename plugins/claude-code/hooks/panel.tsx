// Draws the panel: tabs over the live view (the current issue, or the
// overview while there is none), the project view (overview and issue tree)
// or the usage view (this session's tokens, by prep area).

import type { Color, Elements, RenderChildren, RenderElement } from 'claude-code'

import type {
  PrepGuide,
  PrepPrime,
  PrepProjectSnapshot,
  PrepShow,
  PrepSnapshot,
  PrepSummary,
  PrepTab,
  PrepUsage,
} from '../types'
import { compact } from './usage'

type Kit = Pick<Elements['terminal'], 'Box' | 'Text' | 'Markdown' | 'Button'>

// The TUI design system's tokens, mapped onto the theme keys of Claude Code
// so the panel follows the person's light or dark theme.
const ACCENT: Color = 'claude'
const STATE: Record<string, Color> = {
  open: 'inactive',
  defined: 'suggestion',
  ready: 'ide',
  'in-progress': 'warning',
  done: 'success',
  dropped: 'subtle',
}

// How much of each section shows; the pane scrolls past the rest.
const REQUIREMENT_LINES = 14
const LIST_ROWS = 8
const TREE_ROWS = 200
const HISTORY_LINES = 3

const SOURCE = { pinned: 'pinned', agent: 'following the agent', claimed: 'claimed here' }

/** The tab row and the chosen view; onTab switches views. */
export function drawPane(
  kit: Kit,
  tab: PrepTab,
  live: PrepSnapshot | null,
  project: PrepProjectSnapshot | null,
  usage: PrepUsage | { error: string } | null,
  onTab: (tab: PrepTab) => void,
): RenderElement {
  const { Box, Button } = kit
  // Buttons that look like buttons ([ Live ] [ Project ]); the active one
  // is the primary. Views switch by click or /prep:pane, no hotkeys.
  const tabButton = (t: PrepTab, label: string) => (
    <Button
      key={`tab-${t}`}
      label={label}
      variant={t === tab ? 'primary' : undefined}
      dimColor={t !== tab}
      onPress={() => onTab(t)}
    />
  )
  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="row" gap={2}>
        {tabButton('live', 'Live')}
        {tabButton('project', 'Project')}
        {tabButton('usage', 'Usage')}
      </Box>
      {tab === 'project'
        ? drawProjectView(kit, project)
        : tab === 'usage'
          ? drawUsageView(kit, usage)
          : drawPanel(kit, live)}
    </Box>
  )
}

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
          {priorityMark(kit, show.priority, ' · ')}
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

function drawProjectView(kit: Kit, project: PrepProjectSnapshot | null): RenderElement {
  const { Box, Text } = kit
  if (project === null) return <Text dimColor>Loading prep…</Text>
  if ('error' in project) return drawPanel(kit, { view: 'error', message: project.error })
  const { prime: p, tree } = project
  const check =
    p.check.errors + p.check.warnings === 0
      ? 'check ok'
      : `check: ${p.check.errors} errors, ${p.check.warnings} warnings`
  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold color={ACCENT}>
          prep
        </Text>
        <Text dimColor>{`${p.issues} issues · ${p.actionable} actionable · ${check}`}</Text>
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
      {p.stale.length > 0 &&
        section(kit, 'Attention', [p.stale.slice(0, LIST_ROWS).map(s => row(kit, s, 'stale'))], 'warning')}
      {section(kit, `Open issues ${tree.length}`, [
        tree.length === 0 && <Text dimColor>Nothing open.</Text>,
        tree.slice(0, TREE_ROWS).map(s => treeRow(kit, s)),
        tree.length > TREE_ROWS && <Text dimColor>{`… ${tree.length - TREE_ROWS} more`}</Text>,
      ])}
    </Box>
  )
}

const AREA_LABEL: Record<string, string> = {
  knowledge: 'knowledge base',
  issue: 'issue files',
  'prep-cli': 'prep commands',
  code: 'code',
}

function drawUsageView(kit: Kit, u: PrepUsage | { error: string } | null): RenderElement {
  const { Box, Text } = kit
  if (u === null) return <Text dimColor>Loading usage…</Text>
  if ('error' in u) return drawPanel(kit, { view: 'error', message: u.error })
  const context =
    u.contextTokens !== undefined && u.window !== undefined
      ? `context ${compact(u.contextTokens)} / ${compact(u.window)} (${Math.round((u.contextTokens / u.window) * 100)}%)`
      : 'context unknown'
  const cost = u.costUsd !== undefined ? ` · $${u.costUsd.toFixed(2)}` : ''
  return (
    <Box flexDirection="column" gap={1}>
      <Box flexDirection="column">
        <Text bold color={ACCENT}>
          This session
        </Text>
        <Text dimColor wrap="truncate-end">{`${context}${cost} · ${u.sessionId}`}</Text>
      </Box>
      {!u.hasLedger ? (
        <Text dimColor>
          Per-request tokens and reads by area come from the token-ledger plugin, after its first turn in this session:
          /plugin install token-ledger@prep
        </Text>
      ) : (
        [
          section(kit, `Requests ${u.requests}`, [
            <Text>{`input ${compact(u.tokens.input)} · cache read ${compact(u.tokens.cacheRead)}`}</Text>,
            <Text>{`cache write ${compact(u.tokens.cacheWrite)} · output ${compact(u.tokens.output)}`}</Text>,
          ]),
          section(kit, 'Read back by area (≈ tokens)', [
            u.areas.map(a => (
              <Text>
                {`${(AREA_LABEL[a.area] ?? a.area).padEnd(16)}${compact(a.tokens).padStart(7)}`}
                <Text dimColor>{`  ${a.calls} calls`}</Text>
              </Text>
            )),
          ]),
          u.topReads.length > 0 &&
            section(kit, 'Largest knowledge reads', [
              u.topReads.map(r => (
                <Text wrap="truncate-end">
                  {`${compact(r.tokens).padStart(6)} `}
                  <Text dimColor>{r.target.replace(/^.*\.prep\/knowledge/, '')}</Text>
                </Text>
              )),
            ]),
        ]
      )}
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

// Only deviations from medium draw the eye, in the TUI's tones.
const PRIORITY: Record<string, { label: string; color: Color }> = {
  critical: { label: '!crit', color: 'error' },
  high: { label: '!high', color: 'warning' },
  low: { label: 'low', color: 'subtle' },
}

/** The priority before a title (or after a separator); nothing for medium. */
function priorityMark(kit: Kit, priority: string | undefined, before = ''): RenderElement | null {
  const { Text } = kit
  const p = PRIORITY[priority ?? 'medium']
  if (!p) return null
  return before ? <Text color={p.color}>{`${before}${p.label}`}</Text> : <Text color={p.color}>{`${p.label} `}</Text>
}

/** One issue as a row: state, title, progress. */
function row(kit: Kit, s: PrepSummary, note?: string): RenderElement {
  const { Text } = kit
  const progress = s.progress ? ` ${s.progress.done}/${s.progress.total}` : ''
  return (
    <Text wrap="truncate-end">
      <Text color={STATE[s.state] ?? 'text'}>{pad(s.state)}</Text>
      {priorityMark(kit, s.priority)}
      {s.title}
      <Text dimColor>{progress + (note ? ` (${note})` : '')}</Text>
    </Text>
  )
}

/** One row of the issue tree: indented by depth, with marks for what needs attention. */
function treeRow(kit: Kit, s: PrepSummary): RenderElement {
  const { Text } = kit
  const progress = s.progress ? ` ${s.progress.done + s.progress.dropped}/${s.progress.total}` : ''
  const marks = [s.stale && 'stale', s.blocked && 'blocked', s.actionable && 'actionable'].filter(Boolean)
  return (
    <Text wrap="truncate-end">
      {'  '.repeat(s.depth ?? 0)}
      <Text color={STATE[s.state] ?? 'text'}>{pad(s.state)}</Text>
      {priorityMark(kit, s.priority)}
      {s.title}
      <Text dimColor>{progress}</Text>
      {marks.length > 0 && <Text color={s.stale || s.blocked ? 'warning' : 'success'}>{` ${marks.join(' ')}`}</Text>}
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
      {priorityMark(kit, s.priority)}
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
