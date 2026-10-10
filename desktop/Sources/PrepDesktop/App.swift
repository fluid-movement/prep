import AppKit
import UserNotifications
import SwiftUI
import PrepKit

/// Launch options: a project directory, and the snapshot mode agents use
/// to see the window (see Snapshot.swift).
struct LaunchOptions {
    var project: String?
    var snapshot: String?
    var select: String?
    var tab: Tab = .issues
    var entry: String?
    var size = CGSize(width: 1280, height: 820)
    var appearance: String?
    var sheet: String?

    static func parse(_ args: [String]) -> LaunchOptions {
        var o = LaunchOptions()
        var k = 0
        func value() -> String? {
            k += 1
            return k < args.count ? args[k] : nil
        }
        while k < args.count {
            switch args[k] {
            case "--snapshot": o.snapshot = value()
            case "--select": o.select = value()
            case "--entry": o.entry = value()
            case "--tab": o.tab = value().flatMap { v in Tab.allCases.first { $0.rawValue.lowercased().hasPrefix(v.lowercased()) } } ?? .issues
            case "--appearance": o.appearance = value()
            case "--sheet": o.sheet = value()
            case "--render-icon": _ = value()
            case "--size":
                let parts = (value() ?? "").split(separator: "x").compactMap { Double($0) }
                if parts.count == 2 { o.size = CGSize(width: parts[0], height: parts[1]) }
            case let a where a.hasPrefix("-"):
                // Flags macOS passes to apps (-NSDocumentRevisionsDebugMode, -psn_…) take a value or none.
                if a.hasPrefix("-NS") || a.hasPrefix("-Apple") { k += 1 }
            default:
                var isDir: ObjCBool = false
                if FileManager.default.fileExists(atPath: args[k], isDirectory: &isDir), isDir.boolValue {
                    o.project = URL(fileURLWithPath: args[k]).standardizedFileURL.path
                }
            }
            k += 1
        }
        return o
    }
}

struct PrepApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var delegate

    var body: some Scene {
        WindowGroup {
            WindowRoot()
                .frame(minWidth: 960, minHeight: 600)
        }
        .defaultSize(width: 1280, height: 820)
        .commands {
            CommandGroup(after: .newItem) {
                OpenProjectCommand()
            }
            CommandGroup(before: .sidebar) {
                ViewCommands()
            }
            CommandMenu("Go") {
                GoCommands()
            }
            CommandMenu("Issue") {
                IssueCommands()
            }
        }
        Settings {
            SettingsView()
        }
    }
}

/// Holds a window's model; the first window opens the launch project, or
/// the last one used.
struct WindowRoot: View {
    @State private var model = ProjectModel()
    @State private var started = false

    var body: some View {
        ContentView(model: model)
            .onReceive(NotificationCenter.default.publisher(for: .openIssue)) { n in
                if let id = n.userInfo?["issue"] as? String, model.all[id] != nil { model.go(to: id) }
            }
            .task {
                guard !started else { return }
                started = true
                let launch = LaunchOptions.parse(CommandLine.arguments)
                if let p = launch.project ?? AppDelegate.claimLaunchProject() {
                    await model.open(path: p)
                } else if let last = UserDefaults.standard.string(forKey: "lastProject"),
                          FileManager.default.fileExists(atPath: last + "/.prep") {
                    await model.open(path: last)
                }
            }
    }
}

extension Notification.Name {
    /// A notification was clicked: the userInfo's "issue" should open.
    static let openIssue = Notification.Name("prep.openIssue")
}

final class AppDelegate: NSObject, NSApplicationDelegate, UNUserNotificationCenterDelegate {
    func userNotificationCenter(_ center: UNUserNotificationCenter, didReceive response: UNNotificationResponse) async {
        if let id = response.notification.request.content.userInfo["issue"] as? String {
            await MainActor.run {
                NSApp.activate(ignoringOtherApps: true)
                NotificationCenter.default.post(name: .openIssue, object: nil, userInfo: ["issue": id])
            }
        }
    }

    nonisolated(unsafe) static var launchProject: String?
    nonisolated(unsafe) static var claimed = false

    /// The launch project goes to the first window only.
    static func claimLaunchProject() -> String? {
        if claimed { return nil }
        claimed = true
        return launchProject ?? LaunchOptions.parse(CommandLine.arguments).project
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.regular)
        if Bundle.main.bundleIdentifier != nil { UNUserNotificationCenter.current().delegate = self }
        // The bundle carries AppIcon.icns; a bare build draws it here.
        if Bundle.main.object(forInfoDictionaryKey: "CFBundleIconFile") == nil {
            let r = ImageRenderer(content: AppIconView().frame(width: 512, height: 512))
            if let image = r.nsImage { NSApp.applicationIconImage = image }
        }
        NSApp.activate(ignoringOtherApps: true)
    }

    func applicationWillTerminate(_ notification: Notification) {
        Watcher.stopAll()
    }

    /// `open -a Prep <dir>` hands the directory over here.
    func application(_ application: NSApplication, open urls: [URL]) {
        if let u = urls.first { AppDelegate.launchProject = u.path }
    }
}

struct ProjectModelKey: FocusedValueKey {
    typealias Value = ProjectModel
}

extension FocusedValues {
    var projectModel: ProjectModel? {
        get { self[ProjectModelKey.self] }
        set { self[ProjectModelKey.self] = newValue }
    }
}

struct OpenProjectCommand: View {
    @FocusedValue(\.projectModel) var model
    var body: some View {
        Button("Open Project…") {
            let panel = NSOpenPanel()
            panel.canChooseDirectories = true
            panel.canChooseFiles = false
            if panel.runModal() == .OK, let url = panel.url, let model {
                Task { await model.open(path: url.path) }
            }
        }
        .keyboardShortcut("o")
        .disabled(model == nil)
    }
}

struct ViewCommands: View {
    @FocusedValue(\.projectModel) var model
    var body: some View {
        ForEach(Array(Tab.allCases.enumerated()), id: \.offset) { k, t in
            Button(t.rawValue) { model?.tab = t }
                .keyboardShortcut(KeyEquivalent(Character(String(k + 1))))
        }
        Divider()
        Button(model?.showInspector == true ? "Hide Inspector" : "Show Inspector") { model?.showInspector.toggle() }
            .keyboardShortcut("i", modifiers: [.command, .option])
        Button("Reload") { Task { await model?.reload() } }
            .keyboardShortcut("r")
        Divider()
    }
}

struct GoCommands: View {
    @FocusedValue(\.projectModel) var model
    var body: some View {
        Button("Back") { model?.goBack() }
            .keyboardShortcut("[")
            .disabled(model?.canGoBack != true)
        Button("Forward") { model?.goForward() }
            .keyboardShortcut("]")
            .disabled(model?.canGoForward != true)
        Divider()
        Button("Parent") {
            if let m = model, let id = m.current, let p = m.summary(id)?.parent { m.go(to: p) }
        }
        .keyboardShortcut(.upArrow, modifiers: [.command])
        .disabled(model?.current.flatMap { model?.summary($0)?.parent } == nil)
    }
}

struct IssueCommands: View {
    @FocusedValue(\.projectModel) var model
    var body: some View {
        Button("New Issue…") { model?.newIssue() }
            .keyboardShortcut("n", modifiers: [.command, .shift])
            .disabled(model?.project == nil)
        Button("New Child Issue…") { model?.newIssue(parent: model?.current) }
            .keyboardShortcut("n", modifiers: [.command, .option])
            .disabled(model?.current == nil)
        Divider()
        Button("Rename…") { model?.editing = .title }
            .keyboardShortcut("r", modifiers: [.command, .shift])
            .disabled(model?.detail == nil || model?.detail?.issue.resolution != nil)
        Button("Edit Requirement…") { model?.editing = .requirement }
            .keyboardShortcut("e")
            .disabled(model?.detail == nil || model?.detail?.issue.resolution != nil)
        Button("Edit Context…") { model?.editing = .context }
            .disabled(model?.detail == nil || model?.detail?.issue.resolution != nil)
        Button("Add Criterion…") { model?.editing = .criterion }
            .disabled(model?.detail == nil || model?.detail?.issue.resolution != nil)
        Button("Record Decision…") { model?.editing = .decision }
            .disabled(model?.detail == nil || model?.detail?.issue.resolution != nil)
        Divider()
        Button("Copy ID") { if let id = model?.current { model?.copy(id) } }
            .keyboardShortcut("c", modifiers: [.command, .shift])
            .disabled(model?.current == nil)
        Button("Reveal in Finder") { if let id = model?.current { model?.revealInFinder(id) } }
            .disabled(model?.current == nil)
    }
}
