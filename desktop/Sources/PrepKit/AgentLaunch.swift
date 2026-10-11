import Foundation

/// Starting a coding agent on an issue in the user's own terminal: a small
/// login-shell script that enters the project and runs the agent with a
/// prompt that points it at `prep guide`, which tells it everything else.
public struct AgentLaunch: Sendable {
    public var root: String
    public var issue: String
    public var title: String
    /// The agent command, e.g. "claude".
    public var command: String
    /// The prompt; {id} and {title} are replaced.
    public var template: String

    public static let defaultCommand = "claude"
    public static let defaultTemplate = "Work on prep issue {id} ({title}). Run `prep guide {id}` and follow it."

    public init(root: String, issue: String, title: String, command: String = defaultCommand, template: String = defaultTemplate) {
        self.root = root
        self.issue = issue
        self.title = title
        self.command = command
        self.template = template
    }

    public var prompt: String {
        template.replacingOccurrences(of: "{id}", with: issue).replacingOccurrences(of: "{title}", with: title)
    }

    /// A zsh login script, so the user's PATH (and their agent) is there.
    public var script: String {
        """
        #!/bin/zsh -l
        cd \(Self.quote(root)) || exit 1
        clear
        exec \(command) \(Self.quote(prompt))

        """
    }

    /// Quotes a string for the shell: single quotes, with ' as '\\''.
    public static func quote(_ s: String) -> String {
        "'" + s.replacingOccurrences(of: "'", with: "'\\''") + "'"
    }
}

/// The terminals the app can start an agent in.
public enum TerminalApp: String, CaseIterable, Identifiable, Sendable {
    case terminal = "Terminal", iterm = "iTerm", ghostty = "Ghostty"
    public var id: String { rawValue }

    public var bundleID: String {
        switch self {
        case .terminal: "com.apple.Terminal"
        case .iterm: "com.googlecode.iterm2"
        case .ghostty: "com.mitchellh.ghostty"
        }
    }

    /// The `open` arguments that run script in this terminal.
    public func openArguments(script: String) -> [String] {
        switch self {
        case .terminal: ["-b", bundleID, script]
        case .iterm: ["-b", bundleID, script]
        case .ghostty: ["-na", "Ghostty", "--args", "-e", script]
        }
    }
}
