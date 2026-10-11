import AppKit
import PrepKit

extension TerminalApp {
    var isInstalled: Bool { NSWorkspace.shared.urlForApplication(withBundleIdentifier: bundleID) != nil }

    static var installed: [TerminalApp] { allCases.filter(\.isInstalled) }

    /// The chosen terminal, else the first installed one.
    static var preferred: TerminalApp {
        let chosen = UserDefaults.standard.string(forKey: "terminalApp").flatMap(TerminalApp.init(rawValue:))
        if let chosen, chosen.isInstalled { return chosen }
        return installed.first ?? .terminal
    }
}

extension ProjectModel {
    /// Starts the configured agent on an issue in the user's terminal.
    func startAgent(_ id: String) {
        guard let root = project?.root else { return }
        let d = UserDefaults.standard
        let launch = AgentLaunch(
            root: root, issue: id, title: title(id),
            command: d.string(forKey: "agentCommand").flatMap { $0.isEmpty ? nil : $0 } ?? AgentLaunch.defaultCommand,
            template: d.string(forKey: "agentPrompt").flatMap { $0.isEmpty ? nil : $0 } ?? AgentLaunch.defaultTemplate
        )
        do {
            let dir = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask)[0]
                .appendingPathComponent("dev.prep.desktop/agents", isDirectory: true)
            try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
            let script = dir.appendingPathComponent("prep-\(id.suffix(6)).command")
            try launch.script.write(to: script, atomically: true, encoding: .utf8)
            try FileManager.default.setAttributes([.posixPermissions: 0o755], ofItemAtPath: script.path)
            let p = Process()
            p.executableURL = URL(fileURLWithPath: "/usr/bin/open")
            p.arguments = TerminalApp.preferred.openArguments(script: script.path)
            try p.run()
            p.waitUntilExit()
            if p.terminationStatus != 0 { error = "Could not open \(TerminalApp.preferred.rawValue)." }
        } catch {
            self.error = "Could not start the agent: \(error.localizedDescription)"
        }
    }
}
