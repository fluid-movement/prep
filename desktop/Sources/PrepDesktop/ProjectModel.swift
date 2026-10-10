import AppKit
import Foundation
import Observation
import PrepKit

enum Tab: String, CaseIterable, Identifiable {
    case inbox = "Inbox", issues = "Issues", knowledge = "Knowledge", agent = "Agents"
    var id: String { rawValue }
    var symbol: String {
        switch self {
        case .inbox: "tray"
        case .issues: "checklist"
        case .knowledge: "books.vertical"
        case .agent: "sparkles"
        }
    }
}

enum ListMode: String, CaseIterable {
    case tree, flat
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

    var tab: Tab = .inbox
    var projects: [ProjectRef] = []

    // Issues
    var filter = IssueFilter(view: "Unresolved") {
        didSet { if filter != oldValue { Task { await reloadList() } } }
    }
    var listMode: ListMode = .tree
    private(set) var views: [String: String] = [:]
    private(set) var flags: [String: FlagInfo] = [:]
    private(set) var all: [String: IssueSummary] = [:]
    private(set) var listed: [IssueSummary] = []
    private(set) var tree: [IssueNode] = []
    var collapsed: Set<String> = []

    /// The issue the detail shows. It may be outside the filtered list.
    private(set) var current: String?
    private(set) var detail: IssueDetail?
    private(set) var guide: Guide?
    private var back: [String] = []
    private var forward: [String] = []

    // Inbox
    private(set) var inbox: [InboxItem] = []
    private var inboxLoaded = false

    // Knowledge
    private(set) var entries: [KnowledgeEntry] = []
    var knowledgeQuery = ""
    var entrySelection: String? {
        didSet { if entrySelection != oldValue { Task { await loadEntry() } } }
    }
    private(set) var entry: KnowledgeEntry?

    // Agents
    private(set) var activity = ActivityFeed(events: [], focus: [])
    var agentSelection: String?

    var error: String?
    private(set) var loading = false
    private(set) var busy = false
    var showInspector = true
    var creating: NewIssueDraft?
    var editing: EditTarget?

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
        let root = PrepCLI.projectRoot(from: path) ?? path
        let cli = PrepCLI(binary: binary, root: root)
        do {
            // prep records the project in its registry on any successful command.
            views = try await cli.views()
        } catch {
            self.error = "\(path) is not a prep project: \(error.localizedDescription)"
            return
        }
        watcher?.stop()
        self.cli = cli
        remember(root)
        await refreshProjects()
        project = projects.first { $0.root == root }
            ?? ProjectRef(root: root, name: URL(fileURLWithPath: root).lastPathComponent)
        current = nil
        back = []
        forward = []
        detail = nil
        guide = nil
        entrySelection = nil
        entry = nil
        error = nil
        let view = views["Unresolved"] != nil ? "Unresolved" : nil
        if filter != IssueFilter(view: view) { filter = IssueFilter(view: view) }
        await reload()
        watcher = try? Watcher(binary: binary, root: root) { [weak self] in
            Task { @MainActor in await self?.reload() }
        }
    }

    // MARK: Loading

    func reload() async {
        guard let cli else { return }
        loading = true
        defer { loading = false }
        do {
            async let everything = cli.list()
            async let v = cli.views()
            async let f = cli.flags()
            async let k = cli.knowledge()
            async let a = cli.activity(max: 400)
            let list = try await everything
            all = Dictionary(list.map { ($0.id, $0) }, uniquingKeysWith: { a, _ in a })
            views = try await v
            flags = Dictionary(((try? await f) ?? []).map { ($0.name, $0) }, uniquingKeysWith: { a, _ in a })
            entries = try await k
            activity = (try? await a) ?? activity
            error = nil
        } catch {
            self.error = error.localizedDescription
        }
        await reloadList()
        await reloadInbox()
        await loadDetail()
        await loadEntry()
    }

    /// Rebuilds the inbox, and tells the person about what newly needs them.
    func reloadInbox() async {
        guard let cli, let root = project?.root else { return }
        let defined = all.values.filter { $0.state == "defined" && !$0.stale }.map(\.id)
        var allowed: Set<String> = []
        await withTaskGroup(of: (String, Bool).self) { group in
            for id in defined {
                group.addTask {
                    let g = try? await cli.guide(id)
                    return (id, g?.transitions?.contains { $0.op == "ready" && $0.allowed } ?? false)
                }
            }
            for await (id, ok) in group where ok { allowed.insert(id) }
        }
        let defaults = UserDefaults.standard
        let sinceKey = "inboxSince:\(root)"
        if defaults.object(forKey: sinceKey) == nil { defaults.set(Date.now, forKey: sinceKey) }
        let since = defaults.object(forKey: sinceKey) as? Date ?? .now
        let reviewed = Set(defaults.stringArray(forKey: "reviewed:\(root)") ?? [])
        let fresh = Inbox.build(Array(all.values), readyAllowed: allowed, since: since, reviewed: reviewed)
        if inboxLoaded {
            let before = Set(inbox.map(\.id))
            Notifier.shared.announce(fresh.filter { !before.contains($0.id) && $0.kind != .define }, project: project?.name ?? "")
        }
        inbox = fresh
        inboxLoaded = true
        Notifier.shared.badge(fresh.filter { $0.kind != .define }.count)
    }

    /// The inbox item for the current issue, if it waits on the person.
    var currentInboxItem: InboxItem? {
        guard let id = current else { return nil }
        return inbox.first { $0.issue.id == id }
    }

    func markReviewed(_ ids: [String]) {
        guard let root = project?.root else { return }
        let key = "reviewed:\(root)"
        let old = UserDefaults.standard.stringArray(forKey: key) ?? []
        UserDefaults.standard.set(Array(Set(old + ids)), forKey: key)
        Task { await reloadInbox() }
    }

    /// The one action that clears an inbox item.
    func resolve(_ item: InboxItem) {
        switch item.kind {
        case .review: markReviewed([item.issue.id])
        case .acknowledge: Task { await write(["ack", item.issue.id]) }
        case .signOff: Task { await write(["ready", item.issue.id]) }
        case .define: Task { await write(["define", item.issue.id]) }
        case .questions:
            show(item.issue.id)
            editing = .requirement
        }
    }

    func reloadList() async {
        guard let cli else { return }
        do {
            listed = try await cli.list(filter.arguments)
            tree = IssueTree.build(listed)
        } catch {
            self.error = error.localizedDescription
        }
    }

    func loadDetail() async {
        guard let cli, let id = current else {
            detail = nil
            guide = nil
            return
        }
        do {
            async let d = cli.show(id)
            async let g = cli.guide(id)
            let shown = try await d
            let guided = try? await g
            guard current == id else { return }
            detail = shown
            guide = guided
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
    func isListed(_ id: String) -> Bool { listed.contains { $0.id == id } }

    /// The chain of parents above an issue, outermost first.
    func ancestors(of id: String) -> [IssueSummary] {
        var chain: [IssueSummary] = []
        var seen: Set<String> = [id]
        var next = all[id]?.parent
        while let p = next, !seen.contains(p), let s = all[p] {
            chain.insert(s, at: 0)
            seen.insert(p)
            next = s.parent
        }
        return chain
    }

    func children(of id: String) -> [IssueSummary] {
        all.values.filter { $0.parent == id }.sorted { a, b in
            let ra = Lifecycle.states.firstIndex(of: a.state) ?? 0, rb = Lifecycle.states.firstIndex(of: b.state) ?? 0
            if a.isResolved != b.isResolved { return !a.isResolved }
            if ra != rb { return ra > rb }
            return a.id < b.id
        }
    }

    // MARK: Navigation

    /// The list selection: picking a row shows it and records history.
    var selection: String? {
        get { current }
        set { if let id = newValue { show(id) } }
    }

    func show(_ id: String, recording: Bool = true) {
        guard id != current else { return }
        if recording, let c = current {
            back.append(c)
            forward.removeAll()
        }
        current = id
        if let p = all[id]?.parent { collapsed.remove(p) }
        Task { await loadDetail() }
    }

    /// Follows a link to an issue from anywhere in the window.
    func go(to id: String) {
        if tab != .inbox || !inbox.contains(where: { $0.issue.id == id }) { tab = .issues }
        show(id)
    }

    var canGoBack: Bool { !back.isEmpty }
    var canGoForward: Bool { !forward.isEmpty }

    func goBack() {
        guard let id = back.popLast() else { return }
        if let c = current { forward.append(c) }
        tab = .issues
        show(id, recording: false)
    }

    func goForward() {
        guard let id = forward.popLast() else { return }
        if let c = current { back.append(c) }
        tab = .issues
        show(id, recording: false)
    }

    /// Shows the current issue in the list by clearing the filter.
    func revealInList() {
        filter = IssueFilter(view: viewNames.contains("All") ? "All" : nil)
    }

    // MARK: Writes

    /// Runs a write; the watcher reloads after it lands, this reloads now.
    @discardableResult
    func write(_ args: [String], stdin: String? = nil) async -> Bool {
        guard let cli else { return false }
        busy = true
        defer { busy = false }
        do {
            try await cli.write(args, stdin: stdin)
            error = nil
            await reload()
            return true
        } catch {
            self.error = error.localizedDescription
            return false
        }
    }

    func toggle(criterion n: Int, checked: Bool) async {
        guard let id = current else { return }
        await write(["criterion", id, checked ? "--check" : "--uncheck", String(n)])
    }

    func addCriterion(_ text: String) async {
        guard let id = current, !text.isEmpty else { return }
        await write(["criterion", id, "--add", text])
    }

    func log(_ text: String) async {
        guard let id = current, !text.isEmpty else { return }
        await write(["log", id, "--", text])
    }

    func edit(_ flags: [String]) async {
        guard let id = current else { return }
        await write(["edit", id] + flags)
    }

    func setTags(_ tags: [String]) async {
        guard let id = current else { return }
        let args = tags.isEmpty ? ["--tag", ""] : tags.flatMap { ["--tag", $0] }
        await write(["edit", id] + args)
    }

    func newIssue(parent: String? = nil) {
        creating = NewIssueDraft(parent: parent)
    }

    func create(_ d: NewIssueDraft) async -> Bool {
        guard let cli else { return false }
        busy = true
        defer { busy = false }
        do {
            let tags = d.tags.split(whereSeparator: { $0 == "," || $0 == " " }).map(String.init)
            let id = try await cli.create(title: d.title, kind: d.kind, parent: d.parent, priority: d.priority,
                                          tags: tags, body: d.body)
            await reload()
            go(to: id)
            return true
        } catch {
            self.error = error.localizedDescription
            return false
        }
    }
}

/// The New Issue sheet's fields.
struct NewIssueDraft: Identifiable {
    let id = UUID()
    var title = ""
    var kind = "code"
    var priority = "medium"
    var parent: String?
    var tags = ""
    var body = ""
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

    func revealInFinder(_ id: String) {
        guard let root = project?.root else { return }
        let url = URL(fileURLWithPath: root).appendingPathComponent(".prep/issues/\(id)")
        NSWorkspace.shared.activateFileViewerSelecting([url])
    }

    func copy(_ s: String) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(s, forType: .string)
    }
}
