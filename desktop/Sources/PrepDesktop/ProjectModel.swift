import Foundation
import AppKit
import Observation
import PrepKit

enum Tab: String, CaseIterable, Identifiable {
    case issues = "Issues", knowledge = "Knowledge", agent = "Agent"
    var id: String { rawValue }
    var symbol: String {
        switch self {
        case .issues: "checklist"
        case .knowledge: "books.vertical"
        case .agent: "sparkles"
        }
    }
}

/// One window's project: what it shows, loaded through the prep CLI and
/// reloaded whenever `prep watch` reports a change.
@MainActor
@Observable
final class ProjectModel {
    let binary: URL?
    private(set) var project: ProjectRef?
    private(set) var cli: PrepCLI?
    private var watcher: Watcher?

    var tab: Tab = .issues
    var projects: [ProjectRef] = []

    // Issues
    var query = ""
    var view = "Unresolved"
    var views: [String: String] = [:]
    private(set) var all: [String: IssueSummary] = [:]
    private(set) var tree: [IssueNode] = []
    private(set) var count = 0
    var collapsed: Set<String> = []
    var selection: String? {
        didSet { if selection != oldValue { Task { await loadDetail() } } }
    }
    private(set) var detail: IssueDetail?
    private(set) var guide: Guide?

    // Knowledge
    private(set) var entries: [KnowledgeEntry] = []
    var entrySelection: String? {
        didSet { if entrySelection != oldValue { Task { await loadEntry() } } }
    }
    private(set) var entry: KnowledgeEntry?

    // Agent
    private(set) var activity = ActivityFeed(events: [], focus: [])

    var error: String?
    private(set) var loading = false
    private(set) var loaded = false

    init(binary: URL? = PrepCLI.locate()) {
        self.binary = binary
        if binary == nil {
            error = "The prep binary was not found. Install it, or set PREP_BIN to its path."
        }
    }

    var viewNames: [String] {
        views.keys.sorted { a, b in
            if a == "All" { return true }
            if b == "All" { return false }
            return a < b
        }
    }

    // MARK: Projects

    /// prep's registry, merged with the projects this app opened itself
    /// (older prep binaries have no registry).
    func refreshProjects() async {
        guard let binary else { return }
        let registry = (try? await PrepCLI(binary: binary, root: nil).projects()) ?? []
        let fm = FileManager.default
        let own = (UserDefaults.standard.stringArray(forKey: "recentProjects") ?? [])
            .filter { root in fm.fileExists(atPath: root + "/.prep") && !registry.contains { $0.root == root } }
            .map { ProjectRef(root: $0, name: URL(fileURLWithPath: $0).lastPathComponent) }
        projects = registry + own
    }

    private func remember(_ root: String) {
        var recent = UserDefaults.standard.stringArray(forKey: "recentProjects") ?? []
        recent.removeAll { $0 == root }
        recent.insert(root, at: 0)
        UserDefaults.standard.set(Array(recent.prefix(20)), forKey: "recentProjects")
        UserDefaults.standard.set(root, forKey: "lastProject")
    }

    /// Opens the project at path (any directory inside it) in this window.
    func open(path: String) async {
        guard let binary else { return }
        let cli = PrepCLI(binary: binary, root: path)
        do {
            // prep records the project in its registry on any successful command.
            _ = try await cli.views()
        } catch {
            self.error = "\(path) is not a prep project: \(error.localizedDescription)"
            return
        }
        await refreshProjects()
        // The project root: prep finds it from any directory inside.
        let root = PrepCLI.projectRoot(from: path) ?? path
        let ref = projects.first { $0.root == root }
            ?? ProjectRef(root: root, name: URL(fileURLWithPath: root).lastPathComponent)
        watcher?.stop()
        project = ref
        self.cli = PrepCLI(binary: binary, root: ref.root)
        remember(ref.root)
        await refreshProjects()
        selection = nil
        detail = nil
        guide = nil
        entrySelection = nil
        entry = nil
        error = nil
        await reload()
        watcher = try? Watcher(binary: binary, root: ref.root) { [weak self] in
            Task { @MainActor in await self?.reload() }
        }
    }

    // MARK: Loading

    func reload() async {
        guard let cli else { return }
        loading = true
        defer { loading = false; loaded = true }
        do {
            async let everything = cli.list()
            async let v = cli.views()
            async let k = cli.knowledge()
            async let a = cli.activity()
            let list = try await everything
            views = try await v
            all = Dictionary(list.map { ($0.id, $0) }, uniquingKeysWith: { a, _ in a })
            let filtered = try await cli.list(currentQuery)
            tree = IssueTree.build(filtered)
            count = filtered.count
            entries = try await k
            activity = (try? await a) ?? activity
            error = nil
        } catch {
            self.error = error.localizedDescription
        }
        await loadDetail()
        await loadEntry()
    }

    var currentQuery: [String] {
        var args = Query.arguments(query)
        if let q = views[view], !q.isEmpty, query.isEmpty {
            args = Query.arguments(q)
        }
        return args
    }

    func loadDetail() async {
        guard let cli, let id = selection else {
            detail = nil
            guide = nil
            return
        }
        do {
            async let d = cli.show(id)
            async let g = cli.guide(id)
            detail = try await d
            guide = try? await g
        } catch {
            self.error = error.localizedDescription
        }
    }

    func loadEntry() async {
        guard let cli, let path = entrySelection else {
            entry = nil
            return
        }
        entry = try? await cli.knowledgeEntry(path)
    }

    func title(_ id: String) -> String { all[id]?.title ?? id }
    func summary(_ id: String) -> IssueSummary? { all[id] }

    func go(to id: String) {
        tab = .issues
        if !tree.contains(where: { $0.contains(id) }) {
            query = ""
            view = "All"
            Task { await reload() }
        }
        selection = id
    }

    // MARK: Writes

    /// Runs a write; the watcher reloads after it lands, this reloads now.
    func write(_ args: [String], stdin: String? = nil) async {
        guard let cli else { return }
        do {
            try await cli.write(args, stdin: stdin)
            error = nil
        } catch {
            self.error = error.localizedDescription
        }
        await reload()
    }

    func toggle(criterion n: Int, checked: Bool) async {
        guard let id = selection else { return }
        await write(["criterion", id, checked ? "--check" : "--uncheck", String(n)])
    }

    func log(_ text: String) async {
        guard let id = selection, !text.isEmpty else { return }
        await write(["log", id, "--", text])
    }
}

extension IssueNode {
    func contains(_ id: String) -> Bool {
        issue.id == id || (children ?? []).contains { $0.contains(id) }
    }
}

extension ProjectModel {
    /// Follows a link in a record: a knowledge entry path, an issue ID, or a URL.
    func openLink(_ link: String) {
        let path = link.split(separator: "#").first.map(String.init) ?? link
        if path.hasPrefix("/"), path.hasSuffix(".md") {
            tab = .knowledge
            entrySelection = path
        } else if path.count >= 6, path.allSatisfy({ $0.isLetter || $0.isNumber }),
                  let id = all.keys.first(where: { $0.hasSuffix(path.uppercased()) }) {
            go(to: id)
        } else if let url = URL(string: link), url.scheme != nil {
            NSWorkspace.shared.open(url)
        }
    }
}
