import SwiftUI
import PrepKit

/// Settings: which prep binary the app runs, and a few display choices.
struct SettingsView: View {
    @AppStorage("showModelRequests") private var showRequests = false
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
