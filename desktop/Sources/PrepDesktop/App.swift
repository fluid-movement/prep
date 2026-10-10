import AppKit
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
            case "--tab": o.tab = value().flatMap { v in Tab.allCases.first { $0.rawValue.lowercased() == v.lowercased() } } ?? .issues
            case "--appearance": o.appearance = value()
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
        }
        .defaultSize(width: 1280, height: 820)
        .commands {
            CommandGroup(after: .newItem) {
                OpenProjectCommand()
            }
            CommandMenu("Go") {
                TabCommands()
            }
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

final class AppDelegate: NSObject, NSApplicationDelegate {
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

struct TabCommands: View {
    @FocusedValue(\.projectModel) var model
    var body: some View {
        ForEach(Array(Tab.allCases.enumerated()), id: \.offset) { k, t in
            Button(t.rawValue) { model?.tab = t }
                .keyboardShortcut(KeyEquivalent(Character(String(k + 1))))
        }
        Divider()
        Button("Reload") { Task { await model?.reload() } }
            .keyboardShortcut("r")
    }
}
