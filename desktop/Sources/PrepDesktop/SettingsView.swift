import SwiftUI
import PrepKit

/// Settings: which prep binary the app runs, and a few display choices.
struct SettingsView: View {
    @AppStorage("showModelRequests") private var showRequests = false
    @AppStorage("terminalApp") private var terminal = TerminalApp.preferred.rawValue
    @AppStorage("agentCommand") private var agentCommand = AgentLaunch.defaultCommand
    @AppStorage("agentPrompt") private var agentPrompt = AgentLaunch.defaultTemplate
    @State private var binary = PrepCLI.locate()
    @State private var version = ""

    var body: some View {
        Form {
            Section("prep") {
                LabeledContent("Binary") {
                    VStack(alignment: .trailing, spacing: 2) {
                        Text(binary?.path ?? "Not found").font(.callout.monospaced()).textSelection(.enabled)
                        if !version.isEmpty { Text(version).font(.caption).foregroundStyle(.secondary) }
                    }
                }
                Text("The app runs the first prep it finds: $PREP_BIN, your PATH, ~/.local/bin, Homebrew, ~/go/bin. Run `just install` in the prep repository to use a development build.")
                    .font(.caption).foregroundStyle(.secondary)
            }
            Section("Start Agent") {
                Picker("Terminal", selection: $terminal) {
                    ForEach(TerminalApp.allCases) { t in
                        Text(t.isInstalled ? t.rawValue : "\(t.rawValue) (not installed)").tag(t.rawValue)
                    }
                }
                TextField("Command", text: $agentCommand, prompt: Text(AgentLaunch.defaultCommand))
                    .font(.body.monospaced())
                TextField("Prompt", text: $agentPrompt, prompt: Text(AgentLaunch.defaultTemplate), axis: .vertical)
                    .lineLimit(2...4)
                Text("{id} and {title} are replaced. The agent starts in the project directory, in a login shell.")
                    .font(.caption).foregroundStyle(.secondary)
                Button("Reset to Defaults") {
                    agentCommand = AgentLaunch.defaultCommand
                    agentPrompt = AgentLaunch.defaultTemplate
                }
            }
            Section("Agents") {
                Toggle("Show model requests in the activity feed", isOn: $showRequests)
            }
        }
        .formStyle(.grouped)
        .frame(width: 520)
        .task {
            guard let binary else { return }
            struct V: Decodable { var version: String }
            if let v = try? await PrepCLI(binary: binary, root: nil).decode(V.self, ["version"]) { version = "prep \(v.version)" }
        }
    }
}
