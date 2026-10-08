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
 * Shell variables assigned earlier in the command (I=<id> && prep log $I)
 * are expanded; a reference that stays a variable falls back to the last
 * issue ID in the output. Undefined when the command does not move focus.
 */
export function issueFromCommand(command: string, output = ''): string | undefined {
  let found: string | undefined
  for (const { sub, words, vars } of prepCommands(command)) {
    if (sub === 'new') {
      const id = lastId(output)
      if (id) found = id
      continue
    }
    if (!FOLLOWED.has(sub)) continue
    const args = words.map(w => expand(w, vars))
    const ref = firstPositional(args)
    if (ref) found = ref
    else if (args.some(a => a.startsWith('$'))) found = lastId(output) ?? found
  }
  return found
}

/** Whether a Bash command runs prep init. */
export function runsPrepInit(command: string): boolean {
  for (const { sub } of prepCommands(command)) if (sub === 'init') return true
  return false
}

/**
 * The prep commands in a Bash command, in order: the subcommand, the words
 * after it, and the shell variables assigned so far.
 */
function* prepCommands(command: string): Generator<{ sub: string; words: string[]; vars: Map<string, string> }> {
  const vars = new Map<string, string>()
  for (const segment of withoutHeredocs(command).split(/&&|\|\||[;|\n]/)) {
    let words = segment.trim().split(/\s+/).filter(Boolean)
    if (words[0] === 'export') words = words.slice(1)
    // Assignments, alone or leading a command (PREP_ACTOR=... prep log ...).
    let at = 0
    for (; at < words.length; at++) {
      const m = /^([A-Za-z_][A-Za-z0-9_]*)=(.*)$/.exec(words[at] ?? '')
      if (!m) break
      vars.set(m[1]!, expand(unquote(m[2]!), vars))
    }
    if (!/(^|\/)prep$/.test(words[at] ?? '')) continue
    yield { sub: words[at + 1] ?? '', words: words.slice(at + 2), vars }
  }
}

function expand(word: string, vars: Map<string, string>): string {
  return word.replace(/\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?/g, (all, name: string) => vars.get(name) ?? all)
}

function lastId(output: string): string | undefined {
  return [...output.matchAll(new RegExp(ISSUE_ID, 'g'))].pop()?.[1]
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
