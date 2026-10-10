import SwiftUI
import PrepKit

/// One window: a project's issues, knowledge and agents, with history,
/// the project switcher and the inspector in the toolbar.
struct ContentView: View {
    @Bindable var model: ProjectModel
    @State private var opening = false

    var body: some View {
        Group {
            if model.project == nil {
                WelcomeView(model: model, opening: $opening)
            } else {
                NavigationSplitView {
                    sidebar
                        .navigationSplitViewColumnWidth(min: 280, ideal: 380, max: 560)
                } detail: {
                    detailPane
                        .inspector(isPresented: inspectorShown) {
                            if let d = model.detail {
                                IssueInspector(model: model, detail: d)
                                    .inspectorColumnWidth(min: 240, ideal: 280, max: 360)
                            }
                        }
                }
            }
        }
        .toolbar { toolbar }
        .navigationTitle(model.project?.name ?? "prep")
        .navigationSubtitle(subtitle)
        .fileImporter(isPresented: $opening, allowedContentTypes: [.folder]) { result in
            if case let .success(url) = result {
                Task { await model.open(path: url.path) }
            }
        }
        .sheet(item: $model.creating) { draft in
            NewIssueSheet(model: model, draft: draft)
        }
        .sheet(item: $model.editing) { target in
            EditorSheet(model: model, target: target)
        }
        .overlay(alignment: .bottom) {
            if let e = model.error {
                ErrorBanner(message: e) { model.error = nil }
                    .transition(.move(edge: .bottom).combined(with: .opacity))
            }
        }
        .animation(.easeOut(duration: 0.2), value: model.error)
        .focusedSceneValue(\.projectModel, model)
        .task { await model.refreshProjects() }
    }

    /// The inspector belongs to an issue; it hides on the other screens.
    var inspectorShown: Binding<Bool> {
        Binding(
            get: { model.showInspector && model.tab == .issues && model.detail != nil },
            set: { model.showInspector = $0 }
        )
    }

    var subtitle: String {
        switch model.tab {
        case .issues:
            let open = model.all.values.filter { !$0.isResolved }.count
            return "\(open) unresolved of \(model.all.count)"
        case .knowledge: return "\(model.entries.count) knowledge entries"
        case .agent:
            let live = model.sessions.filter(\.isLive).count
            return live > 0 ? "\(live) agent\(live == 1 ? "" : "s") active" : "No agent active"
        }
    }

    @ViewBuilder
    var sidebar: some View {
        switch model.tab {
        case .issues: IssueSidebar(model: model)
        case .knowledge: KnowledgeSidebar(model: model)
        case .agent: AgentSidebar(model: model)
        }
    }

    @ViewBuilder
    var detailPane: some View {
        switch model.tab {
        case .issues:
            if let d = model.detail {
                IssueDetailView(model: model, detail: d)
            } else {
                ContentUnavailableView {
                    Label("No Issue Selected", systemImage: "checklist")
                } description: {
                    Text("Pick an issue from the list, or create one.")
                } actions: {
                    Button("New Issue") { model.newIssue() }
                }
            }
        case .knowledge:
            if let e = model.entry {
                KnowledgeEntryView(model: model, entry: e)
            } else {
                ContentUnavailableView("No Entry Selected", systemImage: "books.vertical",
                                       description: Text("The knowledge base is what agents read before they work."))
            }
        case .agent:
            AgentFeedView(model: model)
        }
    }

    @ToolbarContentBuilder
    var toolbar: some ToolbarContent {
        ToolbarItemGroup(placement: .navigation) {
            ControlGroup {
                Button { model.goBack() } label: { Label("Back", systemImage: "chevron.left") }
                    .disabled(!model.canGoBack)
                    .help("Back (⌘[)")
                Button { model.goForward() } label: { Label("Forward", systemImage: "chevron.right") }
                    .disabled(!model.canGoForward)
                    .help("Forward (⌘])")
            }
            .controlGroupStyle(.navigation)
            ProjectSwitcher(model: model, opening: $opening)
        }
        ToolbarItem(placement: .principal) {
            Picker("Screen", selection: $model.tab) {
                ForEach(Tab.allCases) { t in Label(t.rawValue, systemImage: t.symbol).tag(t) }
            }
            .pickerStyle(.segmented)
            .labelStyle(.titleAndIcon)
            .disabled(model.project == nil)
            .help("Issues ⌘1 · Knowledge ⌘2 · Agents ⌘3")
        }
        ToolbarItemGroup(placement: .primaryAction) {
            Button { model.newIssue(parent: nil) } label: { Label("New Issue", systemImage: "square.and.pencil") }
                .disabled(model.project == nil)
                .help("New Issue (⇧⌘N)")
            Button { model.showInspector.toggle() } label: { Label("Inspector", systemImage: "sidebar.right") }
                .disabled(model.tab != .issues || model.detail == nil)
                .help("Show or hide the inspector (⌥⌘I)")
        }
    }
}

/// The toolbar's project menu: recent projects, then opening another.
struct ProjectSwitcher: View {
    var model: ProjectModel
    @Binding var opening: Bool

    var body: some View {
        Menu {
            Section("Recent Projects") {
                ForEach(model.projects) { p in
                    Button {
                        Task { await model.open(path: p.root) }
                    } label: {
                        if p.root == model.project?.root {
                            Label(p.name, systemImage: "checkmark")
                        } else {
                            Text(p.name)
                        }
                        Text(p.root.replacingOccurrences(of: NSHomeDirectory(), with: "~"))
                    }
                }
            }
            Divider()
            Button("Open Project…") { opening = true }
            if let root = model.project?.root {
                Button("Show in Finder") { NSWorkspace.shared.activateFileViewerSelecting([URL(fileURLWithPath: root)]) }
            }
        } label: {
            Label(model.project?.name ?? "Open Project", systemImage: "folder")
                .labelStyle(.titleAndIcon)
        }
        .help("Switch project")
        .task { await model.refreshProjects() }
    }
}

struct WelcomeView: View {
    var model: ProjectModel
    @Binding var opening: Bool

    var body: some View {
        VStack(spacing: 24) {
            VStack(spacing: 10) {
                Image(systemName: "checklist")
                    .font(.system(size: 44, weight: .medium))
                    .foregroundStyle(Color.accentColor.gradient)
                    .frame(width: 84, height: 84)
                    .background(Color.accentColor.opacity(0.12), in: RoundedRectangle(cornerRadius: 20))
                Text("Welcome to prep").font(.largeTitle.bold())
                Text("Open a project to see its issues, knowledge and agents.")
                    .foregroundStyle(.secondary)
            }
            Button {
                opening = true
            } label: {
                Label("Open Project…", systemImage: "folder").frame(minWidth: 180)
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.large)
            .keyboardShortcut("o")
            if !model.projects.isEmpty {
                VStack(alignment: .leading, spacing: 2) {
                    Text("RECENT").font(.caption2.weight(.semibold)).foregroundStyle(.tertiary).padding(.leading, 10)
                    ForEach(model.projects.prefix(8)) { p in
                        Button {
                            Task { await model.open(path: p.root) }
                        } label: {
                            HStack(spacing: 10) {
                                Image(systemName: "folder.fill").foregroundStyle(Color.accentColor)
                                VStack(alignment: .leading, spacing: 1) {
                                    Text(p.name).font(.callout.weight(.medium))
                                    Text(p.root.replacingOccurrences(of: NSHomeDirectory(), with: "~"))
                                        .font(.caption).foregroundStyle(.secondary).lineLimit(1).truncationMode(.middle)
                                }
                                Spacer()
                                if let d = p.lastUsed { Text(d.relative).font(.caption).foregroundStyle(.tertiary) }
                            }
                            .padding(.horizontal, 10).padding(.vertical, 7)
                        }
                        .buttonStyle(HoverButtonStyle())
                    }
                }
                .frame(width: 420)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(.background)
    }
}

struct ErrorBanner: View {
    var message: String
    var dismiss: () -> Void
    var body: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "exclamationmark.triangle.fill").foregroundStyle(.orange)
            Text(message).textSelection(.enabled).fixedSize(horizontal: false, vertical: true)
            Spacer(minLength: 8)
            Button(action: dismiss) { Image(systemName: "xmark") }
                .buttonStyle(.borderless)
                .keyboardShortcut(.cancelAction)
        }
        .font(.callout)
        .padding(12)
        .frame(maxWidth: 640)
        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 10))
        .overlay(RoundedRectangle(cornerRadius: 10).strokeBorder(.separator))
        .shadow(color: .black.opacity(0.15), radius: 12, y: 4)
        .padding()
    }
}
