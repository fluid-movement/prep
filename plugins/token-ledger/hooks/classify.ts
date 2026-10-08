// What a tool call touched, in the terms the analysis needs: the target (a
// path, a command, a pattern) and, inside a prep project, which part of prep.

export type Area = 'prep-cli' | 'knowledge' | 'issue' | 'code'

export type Target = { target?: string; area?: Area; prep?: string }

const MAX_TARGET = 200

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
      const command = str('command') ?? ''
      const sub = prepSubcommand(command)
      if (sub !== undefined) return { target: clip(command), area: 'prep-cli', prep: sub }
      const path = prepPathIn(command)
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

/** The prep subcommand a shell command runs, if it runs prep first. */
export function prepSubcommand(command: string): string | undefined {
  const m = command.trim().match(/^(?:\S*\/)?prep\s+([a-z][a-z-]*)/)
  return m?.[1]
}

function prepPathIn(command: string): string | undefined {
  return command.match(/\S*\.prep\/\S+/)?.[0]
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
