// Which issue the agent works on, inferred from what it runs and edits.

// Subcommands that make the issue they name current: every write that takes
// an issue, and guide. Reads (show, list, next, ...) only look around.
const FOLLOWED = new Set([
  'guide',
  'edit',
  'context',
  'decide',
  'criterion',
  'dod',
  'findings',
  'log',
  'define',
  'ack',
  'ready',
  'claim',
  'release',
  'complete',
  'drop',
])

// Flags of prep commands that take a value, so their value is not read as
// the issue reference.
const VALUED = new Set([
  '--root',
  '--by',
  '--title',
  '--kind',
  '--parent',
  '--depends-on',
  '--tag',
  '--body',
  '--body-file',
  '--add',
  '--check',
  '--uncheck',
  '--remove',
  '--opt-out',
  '--reason',
  '--note',
  '--supersedes',
  '--commit',
  '--docs',
  '--no-impact',
])

const ISSUE_ID = /\b(\d{8}-\d{6})\b/

/**
 * The issue a Bash command makes current: the reference of the last prep
 * command in it that names one, or for prep new the ID its output reports.
 * Undefined when the command does not move focus.
 */
export function issueFromCommand(command: string, output = ''): string | undefined {
  let found: string | undefined
  for (const segment of withoutHeredocs(command).split(/&&|\|\||[;|\n]/)) {
    const words = segment.trim().split(/\s+/).filter(Boolean)
    // Skip leading environment assignments (PREP_ACTOR=... prep log ...).
    let at = 0
    while (at < words.length && /^[A-Za-z_][A-Za-z0-9_]*=/.test(words[at] ?? '')) at++
    if (!/(^|\/)prep$/.test(words[at] ?? '')) continue
    const sub = words[at + 1] ?? ''
    if (sub === 'new') {
      const id = ISSUE_ID.exec(output)?.[1]
      if (id) found = id
      continue
    }
    if (!FOLLOWED.has(sub)) continue
    const ref = firstPositional(words.slice(at + 2))
    if (ref) found = ref
  }
  return found
}

// Heredoc bodies are text, not commands: a requirement may mention prep guide.
function withoutHeredocs(command: string): string {
  return command.replace(/(<<-?\s*(['"]?)(\w+)\2[^\n]*)\n[\s\S]*?\n\s*\3[ \t]*(?=\n|$)/g, '$1')
}

function firstPositional(args: string[]): string | undefined {
  for (let k = 0; k < args.length; k++) {
    const a = unquote(args[k] ?? '')
    if (a.startsWith('-')) {
      if (VALUED.has(a)) k++
      continue
    }
    return /^[0-9-]+$/.test(a) ? a : undefined
  }
  return undefined
}

function unquote(word: string): string {
  return word.replace(/^['"]|['"]$/g, '')
}

/** The issue an edited file belongs to: a path under .prep/issues/<id>/. */
export function issueFromPath(path: string): string | undefined {
  return /(?:^|\/)\.prep\/issues\/(\d{8}-\d{6})\//.exec(path)?.[1]
}
