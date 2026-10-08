// What a tool call touched, in the terms the analysis needs: the target (a
// path, a command, a pattern) and, inside a prep project, which part of prep.

export type Area = 'prep-cli' | 'knowledge' | 'issue' | 'code'

export type Target = { target?: string; area?: Area; prep?: string[] }

// Long enough that the target can be classified again afterwards; heredoc
// bodies (text the agent writes) are dropped before clipping.
const MAX_TARGET = 1000

/** Names what a tool call worked on and which part of prep it belongs to. */
export function classify(tool: string, input: Record<string, unknown>): Target {
  const str = (k: string) => (typeof input[k] === 'string' ? (input[k] as string) : undefined)
  switch (tool) {
    case 'Read':
    case 'Edit':
    case 'Write':
    case 'NotebookEdit': {
      const path = str('file_path') ?? str('notebook_path')
      return path === undefined ? {} : { target: path, area: areaOfPath(path) }
    }
    case 'Bash': {
      const command = withoutHeredocs(str('command') ?? '')
      const subs = prepSubcommands(command)
      if (subs.length > 0) return { target: clip(command), area: readsKnowledge(command) ? 'knowledge' : 'prep-cli', prep: subs }
      const path = prepPathIn(command) ?? readPathIn(command)
      return { target: clip(command), area: path ? areaOfPath(path) : undefined }
    }
    case 'Grep':
    case 'Glob': {
      const path = str('path')
      return { target: clip(`${str('pattern') ?? ''}${path ? ` in ${path}` : ''}`), area: path ? areaOfPath(path) : undefined }
    }
    case 'Skill':
      return { target: str('skill') }
    case 'Agent':
      return { target: str('subagent_type') ?? str('description') }
    default:
      return {}
  }
}

/**
 * The prep subcommands a shell command runs, in order: prep wherever it
 * starts a command in the line (after cd ... &&, assignments, export, or
 * a shell keyword such as do).
 */
export function prepSubcommands(command: string): string[] {
  const subs: string[] = []
  for (const segment of withoutHeredocs(command).split(/&&|\|\||[;|\n]/)) {
    let words = segment.trim().split(/\s+/).filter(Boolean)
    // Shell keywords before a command, as in for ...; do prep knowledge show $e; done.
    while (['do', 'then', 'else', 'time', '!', '{', '('].includes(words[0] ?? '')) words = words.slice(1)
    if (words[0] === 'export') words = words.slice(1)
    let at = 0
    while (at < words.length && /^[A-Za-z_][A-Za-z0-9_]*=/.test(words[at]!)) at++
    if (!/^(?:\S*\/)?prep$/.test(words[at] ?? '')) continue
    const sub = words[at + 1]
    if (sub !== undefined && /^[a-z][a-z-]*$/.test(sub)) subs.push(sub)
  }
  return subs
}

/** Whether a shell command reads the knowledge base through prep knowledge list, find or show. */
export function readsKnowledge(command: string): boolean {
  return /(?:^|[\s/;&|])prep\s+knowledge\s+(?:list|find|show)\b/.test(withoutHeredocs(command))
}

// Heredoc bodies are text, not commands, and can be long: keep the marker.
function withoutHeredocs(command: string): string {
  return command.replace(/(<<-?\s*(['"]?)(\w+)\2[^\n]*)\n[\s\S]*?\n\s*\3[ \t]*(?=\n|$)/g, '$1 …')
}

function prepPathIn(command: string): string | undefined {
  return command.match(/\S*\.prep\/\S+/)?.[0]
}

/**
 * The file a shell command reads, when it is a plain read: git show <rev>:<path>
 * anywhere in it, or cat, head, tail or sed starting a command.
 */
export function readPathIn(command: string): string | undefined {
  const shown = /\bgit\s+show\s+[^\s:]*:([^\s;|&]+)/.exec(command)?.[1]
  if (shown) return shown
  for (const segment of command.split(/&&|\|\||[;|\n]/)) {
    const words = segment.trim().split(/\s+/)
    if (!['cat', 'head', 'tail', 'sed'].includes(words[0] ?? '')) continue
    // The first word after options and their counts; sed's comes after its script.
    const args = words.slice(1).filter(w => !w.startsWith('-') && !/^\d+$/.test(w))
    const path = words[0] === 'sed' ? args[1] : args[0]
    if (path) return path.replace(/^['"]|['"]$/g, '')
  }
  return undefined
}

/** knowledge or issue for paths under .prep, code otherwise. */
export function areaOfPath(path: string): Area {
  if (/(^|\/)\.prep\/knowledge(\/|$)/.test(path)) return 'knowledge'
  if (/(^|\/)\.prep\//.test(path)) return 'issue'
  return 'code'
}

function clip(s: string): string {
  return s.length > MAX_TARGET ? `${s.slice(0, MAX_TARGET)}…` : s
}
