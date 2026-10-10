import SwiftUI
import PrepKit

/// One window: a project's issues, knowledge and agent activity, with the
/// project switcher in the toolbar.
struct ContentView: View {
    @Bindable var model: ProjectModel
    @State private var opening = false

    var body: some View {
        Group {
            if model.project == nil {
                WelcomeView(model: model, opening: $opening)
            } else {
                NavigationSplitView {
                    Sidebar(model: model)
                        .navigationSplitViewColumnWidth(min: 300, ideal: 420, max: 640)
                } detail: {
                    detailPane
                }
            }
        }
        .toolbar {
            ToolbarItem(placement: .navigation) {
                ProjectSwitcher(model: model, opening: $opening)
            }
            ToolbarItem(placement: .principal) {
                Picker("Screen", selection: $model.tab) {
                    ForEach(Tab.allCases) { t in Label(t.rawValue, systemImage: t.symbol).tag(t) }
                }
                .pickerStyle(.segmented)
                .disabled(model.project == nil)
            }
            ToolbarItem {
                Button { Task { await model.reload() } } label: { Label("Reload", systemImage: "arrow.clockwise") }
                    .disabled(model.project == nil)
            }
        }
        .navigationTitle(model.project?.name ?? "prep")
        .navigationSubtitle(model.project?.root ?? "")
        .fileImporter(isPresented: $opening, allowedContentTypes: [.folder]) { result in
            if case let .success(url) = result {
                Task { await model.open(path: url.path) }
            }
        }
        .overlay(alignment: .bottom) {
            if let e = model.error {
                ErrorBanner(message: e) { model.error = nil }
            }
        }
        .focusedSceneValue(\.projectModel, model)
        .task { await model.refreshProjects() }
    }

    @ViewBuilder
    var detailPane: some View {
        switch model.tab {
        case .issues:
            if let d = model.detail {
                IssueDetailView(model: model, detail: d).id(d.issue.id)
            } else {
                Placeholder(text: "Select an issue", symbol: "checklist")
            }
        case .knowledge:
            if let e = model.entry {
                KnowledgeEntryView(model: model, entry: e)
            } else {
                Placeholder(text: "Select a knowledge entry", symbol: "books.vertical")
            }
        case .agent:
            AgentFeedView(model: model)
        }
    }
}

struct Sidebar: View {
    @Bindable var model: ProjectModel

    var body: some View {
        switch model.tab {
        case .issues: IssueSidebar(model: model)
        case .knowledge: KnowledgeSidebar(model: model)
        case .agent: AgentSidebar(model: model)
        }
    }
}

struct IssueSidebar: View {
    @Bindable var model: ProjectModel
    @FocusState private var filterFocused: Bool

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 6) {
                Menu {
                    ForEach(model.viewNames, id: \.self) { name in
                        Button {
                            model.view = name
                            model.query = ""
                            Task { await model.reload() }
                        } label: {
                            if name == model.view { Label(name, systemImage: "checkmark") } else { Text(name) }
                        }
                    }
                } label: {
                    Label(model.view, systemImage: "line.3.horizontal.decrease.circle")
                }
                .menuStyle(.borderlessButton)
                .fixedSize()
                .help("Saved views from .prep/config.yaml")
                TextField("Filter: words or --state ready --tag ui", text: $model.query)
                    .textFieldStyle(.roundedBorder)
                    .focused($filterFocused)
                    .onSubmit { Task { await model.reload() } }
            }
            .padding(8)
            Divider()
            List(selection: $model.selection) {
                ForEach(model.tree) { node in IssueOutlineRow(model: model, node: node) }
            }
            .listStyle(.sidebar)
            HStack {
                Text("\(model.count) issues").font(.caption).foregroundStyle(.secondary)
                Spacer()
                if model.loading { ProgressView().controlSize(.small) }
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 4)
        }
        .background {
            // / focuses the filter, as in the TUI.
            Button("") { filterFocused = true }
                .keyboardShortcut("/", modifiers: [])
                .hidden()
        }
    }
}

/// A tree row: parents start expanded, and remember being collapsed.
struct IssueOutlineRow: View {
    var model: ProjectModel
    var node: IssueNode

    var body: some View {
        if let kids = node.children, !kids.isEmpty {
            DisclosureGroup(isExpanded: Binding(
                get: { !model.collapsed.contains(node.id) },
                set: { open in
                    if open { model.collapsed.remove(node.id) } else { model.collapsed.insert(node.id) }
                }
            )) {
                ForEach(kids) { IssueOutlineRow(model: model, node: $0) }
            } label: {
                IssueRow(issue: node.issue)
            }
            .tag(node.id)
        } else {
            IssueRow(issue: node.issue).tag(node.id)
        }
    }
}

struct KnowledgeSidebar: View {
    @Bindable var model: ProjectModel

    var body: some View {
        let groups = Dictionary(grouping: model.entries, by: \.type)
        List(selection: $model.entrySelection) {
            ForEach(groups.keys.sorted(), id: \.self) { type in
                SwiftUI.Section(type) {
                    ForEach(groups[type] ?? []) { e in
                        VStack(alignment: .leading, spacing: 2) {
                            Text(e.title).lineLimit(1)
                            Text(e.path).font(.caption.monospaced()).foregroundStyle(.secondary).lineLimit(1)
                        }
                        .tag(e.path)
                    }
                }
            }
        }
        .listStyle(.sidebar)
    }
}

struct KnowledgeEntryView: View {
    var model: ProjectModel
    var entry: KnowledgeEntry

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                HStack(spacing: 8) {
                    Text(entry.path).font(.callout.monospaced()).foregroundStyle(.secondary)
                    Pill(text: entry.type)
                    if let s = entry.status { Pill(text: s) }
                    if let size = entry.size { Text("~\(size / 4) tok").font(.caption).foregroundStyle(.secondary) }
                }
                if let d = entry.description { Text(d).foregroundStyle(.secondary) }
                Divider()
                MarkdownView(source: entry.text ?? "", onLink: model.openLink)
            }
            .padding(24)
            .frame(maxWidth: 820, alignment: .leading)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
}

struct AgentSidebar: View {
    var model: ProjectModel

    var body: some View {
        List {
            SwiftUI.Section("Agents and their current issue") {
                ForEach(model.activity.focus ?? [], id: \.agent) { f in
                    Button {
                        if let id = f.issue { model.go(to: id) }
                    } label: {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(f.actor ?? f.agent).font(.callout.weight(.medium))
                            if let id = f.issue {
                                if let s = model.summary(id) { IssueRow(issue: s) } else { Text(id).font(.caption) }
                            }
                            if let at = f.at { Text(at, style: .relative).font(.caption).foregroundStyle(.secondary) }
                        }
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        .listStyle(.sidebar)
    }
}

struct AgentFeedView: View {
    var model: ProjectModel

    var body: some View {
        let events = Array((model.activity.events ?? []).reversed())
        if events.isEmpty {
            Placeholder(text: "No agent activity yet", symbol: "sparkles")
        } else {
            List(Array(events.enumerated()), id: \.offset) { _, e in
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text(e.at, format: .dateTime.hour().minute().second())
                        .font(.caption.monospacedDigit())
                        .foregroundStyle(.secondary)
                        .frame(width: 70, alignment: .leading)
                    Text(e.actor ?? "").font(.caption).foregroundStyle(.secondary).frame(width: 150, alignment: .leading).lineLimit(1)
                    Text(e.verb ?? e.kind ?? e.op ?? "").foregroundStyle(e.verb == nil ? .secondary : .primary)
                    if let target = e.target {
                        if let id = e.issue {
                            Button(target) { model.go(to: id) }.buttonStyle(.link).lineLimit(1)
                        } else {
                            Text(target).foregroundStyle(.secondary).lineLimit(1)
                        }
                    }
                }
            }
        }
    }
}

/// The toolbar's project dropdown: recent projects from `prep projects`.
struct ProjectSwitcher: View {
    var model: ProjectModel
    @Binding var opening: Bool

    var body: some View {
        Menu {
            ForEach(model.projects) { p in
                Button {
                    Task { await model.open(path: p.root) }
                } label: {
                    if p.root == model.project?.root {
                        Label(p.name, systemImage: "checkmark")
                    } else {
                        Text(p.name)
                    }
                }
                .help(p.root)
            }
            if !model.projects.isEmpty { Divider() }
            Button("Open Project…") { opening = true }
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
        VStack(spacing: 16) {
            Image(systemName: "checklist").font(.system(size: 48)).foregroundStyle(.secondary)
            Text("Open a prep project").font(.title2)
            Button("Open Project…") { opening = true }.keyboardShortcut("o")
            if !model.projects.isEmpty {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Recent").font(.caption).foregroundStyle(.secondary)
                    ForEach(model.projects.prefix(8)) { p in
                        Button {
                            Task { await model.open(path: p.root) }
                        } label: {
                            HStack {
                                Text(p.name)
                                Text(p.root).font(.caption).foregroundStyle(.secondary)
                            }
                        }
                        .buttonStyle(.link)
                    }
                }
                .padding(.top, 8)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct Placeholder: View {
    var text: String
    var symbol: String
    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: symbol).font(.largeTitle).foregroundStyle(.tertiary)
            Text(text).foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct ErrorBanner: View {
    var message: String
    var dismiss: () -> Void
    var body: some View {
        HStack(alignment: .top) {
            Image(systemName: "exclamationmark.octagon.fill").foregroundStyle(.red)
            Text(message).textSelection(.enabled).fixedSize(horizontal: false, vertical: true)
            Spacer()
            Button(action: dismiss) { Image(systemName: "xmark") }.buttonStyle(.borderless)
        }
        .padding(10)
        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 8))
        .padding()
    }
}
