// The prep panel's state: what prep reported, as the panel draws it.

export type PrepProgress = { done: number; dropped: number; total: number }

/** One issue as prep list --json and prime summarize it. */
export type PrepSummary = {
  id: string
  title: string
  kind: string
  state: string
  stale: boolean
  blocked: boolean
  actionable: boolean
  parent?: string
  depends_on?: string[]
  tags?: string[]
  children: number
  progress?: PrepProgress
  /** Nesting in prep list --tree. */
  depth?: number
}

/** prep show <id> --json, the fields the panel reads. */
export type PrepShow = {
  state: string
  stale: boolean
  blocked: boolean
  children: string[]
  blocks: string[]
  progress?: PrepProgress | null
  definition_of_done: string[]
  history?: string | null
  issue: {
    id: string
    title: string
    kind: string
    tags?: string[]
    parent?: string
    depends_on?: string[]
    requirement: string
    open_questions?: string
    criteria?: { text: string; checked: boolean }[]
    decisions?: { id: string; title: string; date?: string }[]
  }
}

/** prep guide <id> --json, the fields the panel reads. */
export type PrepGuide = {
  step: string
  transitions: {
    op: string
    command: string
    allowed: boolean
    unmet?: { code: string; message: string }[]
  }[]
  knowledge?: { path: string; title: string }[]
}

/** prep prime --json, the fields the panel reads. */
export type PrepPrime = {
  issues: number
  actionable: number
  parents: PrepSummary[]
  next: PrepSummary[]
  stale: PrepSummary[]
  claims: { id: string; title: string; by: string }[]
  check: { errors: number; warnings: number }
  bootstrap?: string
}

/** Why the panel shows the issue it shows. */
export type PrepSource = 'pinned' | 'agent' | 'claimed'

/** What the pane draws: an issue, the project overview, or why neither. */
export type PrepSnapshot =
  | { view: 'issue'; source: PrepSource; show: PrepShow; guide: PrepGuide; issues: PrepSummary[] }
  | { view: 'overview'; prime: PrepPrime }
  | { view: 'error'; message: string }

/** Which view the pane shows: the issue being worked on, or the project. */
export type PrepTab = 'live' | 'project'

/** What the project view draws: the overview and the unresolved issues in tree order. */
export type PrepProjectSnapshot = { prime: PrepPrime; tree: PrepSummary[] } | { error: string }

declare module 'claude-code' {
  interface PluginState {
    prep: {
      /** The issue reference the agent's last prep command or edit named. */
      focus: string | null
      /** The issue /prep:focus pinned; it wins over focus. */
      pin: string | null
      /** The last data loaded for the live view. */
      snapshot: PrepSnapshot | null
      /** The view the pane shows; live unless the person picked project. */
      tab: PrepTab
      /** The last data loaded for the project view, while it is shown. */
      project: PrepProjectSnapshot | null
    }
  }
}
