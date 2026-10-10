import SwiftUI
import PrepKit

/// The issue list: search, filter menus with live counts, and the tree or
/// a flat list. Every filter is a prep query flag, so what the list shows
/// is what `prep list` would print.
struct IssueSidebar: View {
    @Bindable var model: ProjectModel

    var body: some View {
        VStack(spacing: 0) {
            FilterBar(model: model)
            Divider()
            if model.listed.isEmpty, !model.loading {
                ContentUnavailableView {
                    Label("No Matching Issues", systemImage: "line.3.horizontal.decrease.circle")
                } description: {
                    Text(model.filter.isEmpty ? "This view has no issues." : "Try fewer filters.")
                } actions: {
                    if !model.filter.isEmpty {
                        Button("Clear Filters") { model.filter = IssueFilter(view: model.filter.view) }
                    }
                }
            } else {
                list
            }
        }
        .safeAreaInset(edge: .bottom, spacing: 0) { footer }
    }

    var list: some View {
        List(selection: $model.selection) {
            switch model.listMode {
            case .tree:
                ForEach(model.tree) { node in IssueOutlineRow(model: model, node: node) }
            case .flat:
                ForEach(model.listed) { issue in
                    IssueRow(issue: issue).tag(issue.id).issueContextMenu(model: model, issue: issue)
                }
            }
        }
        .listStyle(.sidebar)
    }

    var footer: some View {
        HStack(spacing: 8) {
            Text(countText).font(.caption).foregroundStyle(.secondary)
            if model.loading { ProgressView().controlSize(.mini) }
            Spacer()
            Picker("Layout", selection: $model.listMode) {
                Image(systemName: "list.bullet.indent").tag(ListMode.tree).help("Show the hierarchy")
                Image(systemName: "list.bullet").tag(ListMode.flat).help("Show a flat list, by priority")
            }
            .pickerStyle(.segmented)
            .labelsHidden()
            .fixedSize()
            .controlSize(.small)
            Button {
                model.newIssue()
            } label: {
                Image(systemName: "plus")
            }
            .buttonStyle(.borderless)
            .help("New Issue (⇧⌘N)")
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 7)
        .background(.bar)
        .overlay(alignment: .top) { Divider() }
    }

    var countText: String {
        let n = model.listed.count
        let open = model.listed.filter { !$0.isResolved }.count
        if open == n { return n == 1 ? "1 issue" : "\(n) issues" }
        return "\(n) issues, \(open) unresolved"
    }
}

/// A tree row: parents start expanded and remember being collapsed.
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
                IssueRow(issue: node.issue).issueContextMenu(model: model, issue: node.issue)
            }
            .tag(node.id)
        } else {
            IssueRow(issue: node.issue).tag(node.id).issueContextMenu(model: model, issue: node.issue)
        }
    }
}

extension View {
    func issueContextMenu(model: ProjectModel, issue: IssueSummary) -> some View {
        contextMenu {
            Button("Open") { model.go(to: issue.id) }
            Button("New Child Issue…") { model.newIssue(parent: issue.id) }
            Divider()
            Button("Copy ID") { model.copy(issue.id) }
            Button("Copy Title") { model.copy(issue.title) }
            Button("Copy as Link") { model.copy("\(issue.title) (\(issue.id))") }
            Divider()
            Button("Show Only Its Subtree") { model.filter.under = issue.id }
            Button("Reveal in Finder") { model.revealInFinder(issue.id) }
        }
    }
}

/// Search and the filter menus above the list.
struct FilterBar: View {
    @Bindable var model: ProjectModel
    @FocusState private var searching: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass").foregroundStyle(.secondary)
                TextField("Search titles and requirements", text: $model.filter.text)
                    .textFieldStyle(.plain)
                    .focused($searching)
                if !model.filter.text.isEmpty {
                    Button { model.filter.text = "" } label: {
                        Image(systemName: "xmark.circle.fill").foregroundStyle(.secondary)
                    }
                    .buttonStyle(.borderless)
                    .help("Clear the search")
                }
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 5)
            .background(.quaternary.opacity(0.5), in: RoundedRectangle(cornerRadius: 7))
            .background {
                // ⌘F and / focus the search, as / does in the TUI.
                Button("") { searching = true }.keyboardShortcut("f").hidden()
                Button("") { searching = true }.keyboardShortcut("/", modifiers: []).hidden()
            }

            FlowLayout(spacing: 6) {
                    viewMenu
                    FilterMenu(title: "State", selection: $model.filter.states,
                               values: values("state", order: Lifecycle.states),
                               label: { Lifecycle.title($0) }, symbol: { Style.symbol(state: $0) },
                               color: { Style.color(state: $0) })
                    FilterMenu(title: "Priority", selection: $model.filter.priorities,
                               values: values("priority", order: Lifecycle.priorities),
                               label: { $0.capitalized }, symbol: { _ in "cellularbars" }, color: { Style.priorityColor($0) })
                    FilterMenu(title: "Kind", selection: $model.filter.kinds,
                               values: values("kind", order: Lifecycle.kinds),
                               label: { $0.capitalized }, symbol: { Style.symbol(kind: $0) }, color: { _ in .secondary })
                    FilterMenu(title: "Tag", selection: $model.filter.tags,
                               values: values("tag", order: []),
                               label: { "#\($0)" }, symbol: { _ in "number" }, color: { _ in .secondary })
                    moreMenu
                    if let under = model.filter.under {
                        Button { model.filter.under = nil } label: {
                            HStack(spacing: 4) {
                                Image(systemName: "arrow.turn.down.right").imageScale(.small)
                                Text(model.title(under)).lineLimit(1).frame(maxWidth: 140)
                                Image(systemName: "xmark").font(.system(size: 8, weight: .bold))
                            }
                            .font(.caption.weight(.medium))
                            .padding(.horizontal, 9)
                            .padding(.vertical, 4)
                            .background(Color.accentColor.opacity(0.18), in: Capsule())
                            .foregroundStyle(Color.accentColor)
                        }
                        .buttonStyle(.plain)
                        .help("Showing only this issue's subtree; click to show everything")
                    }
                    if model.filter.activeCount > 0 {
                        Button("Clear") {
                            var f = IssueFilter(view: model.filter.view)
                            f.text = model.filter.text
                            model.filter = f
                        }
                        .buttonStyle(.borderless)
                        .font(.caption)
                        .help("Remove every filter except the view and the search")
                    }
            }
        }
        .padding(.horizontal, 10)
        .padding(.top, 8)
        .padding(.bottom, 8)
    }

    func values(_ flag: String, order: [String]) -> [FlagValue] {
        let vs: [FlagValue] = model.flags[flag]?.values ?? order.map { FlagValue(value: $0, count: 0) }
        if order.isEmpty {
            return vs.sorted { (a: FlagValue, b: FlagValue) -> Bool in
                a.count != b.count ? a.count > b.count : a.value < b.value
            }
        }
        func rank(_ v: FlagValue) -> Int { order.firstIndex(of: v.value) ?? Int.max }
        return vs.sorted { rank($0) < rank($1) }
    }

    var viewMenu: some View {
        Menu {
            ForEach(model.viewNames, id: \.self) { name in
                Toggle(isOn: Binding(get: { model.filter.view == name }, set: { on in if on { model.filter.view = name } })) {
                    Text(name)
                    if let q = model.views[name], !q.isEmpty { Text(q) }
                }
            }
        } label: {
            FilterChipLabel(title: model.filter.view ?? "All", symbol: "rectangle.stack", active: false)
        }
        .menuStyle(.borderlessButton)
        .menuIndicator(.hidden)
        .fixedSize()
        .help("Saved views from .prep/config.yaml")
    }

    var moreMenu: some View {
        let f = model.flags
        let n = [model.filter.actionable, model.filter.blocked, model.filter.stale].filter { $0 }.count
        return Menu {
            Toggle(isOn: $model.filter.actionable) {
                Text("Actionable")
                Text(countLabel(f["actionable"]?.count) + "ready, not stale, dependencies done, unclaimed")
            }
            Toggle(isOn: $model.filter.blocked) {
                Text("Blocked")
                Text(countLabel(f["blocked"]?.count) + "has unresolved dependencies")
            }
            Toggle(isOn: $model.filter.stale) {
                Text("Stale")
                Text(countLabel(f["stale"]?.count) + "requirement changed since defined")
            }
        } label: {
            FilterChipLabel(title: n > 0 ? "More · \(n)" : "More", symbol: "slider.horizontal.3", active: n > 0)
        }
        .menuStyle(.borderlessButton)
        .menuIndicator(.hidden)
        .fixedSize()
    }

    func countLabel(_ n: Int?) -> String { n.map { "\($0) · " } ?? "" }
}

/// A multi-select filter menu: each value with its symbol and issue count.
struct FilterMenu: View {
    var title: String
    @Binding var selection: Set<String>
    var values: [FlagValue]
    var label: (String) -> String
    var symbol: (String) -> String
    var color: (String) -> Color

    var body: some View {
        Menu {
            ForEach(values, id: \.value) { v in
                Toggle(isOn: Binding(
                    get: { selection.contains(v.value) },
                    set: { on in if on { selection.insert(v.value) } else { selection.remove(v.value) } }
                )) {
                    Label("\(label(v.value))  \(v.count)", systemImage: symbol(v.value))
                }
            }
            if !selection.isEmpty {
                Divider()
                Button("Clear \(title)") { selection = [] }
            }
        } label: {
            FilterChipLabel(title: chipTitle, symbol: nil, active: !selection.isEmpty)
        }
        .menuStyle(.borderlessButton)
        .menuIndicator(.hidden)
        .fixedSize()
        .disabled(values.isEmpty)
    }

    var chipTitle: String {
        switch selection.count {
        case 0: title
        case 1: label(selection.first!)
        default: "\(title) · \(selection.count)"
        }
    }
}

struct FilterChipLabel: View {
    var title: String
    var symbol: String?
    var active: Bool

    var body: some View {
        HStack(spacing: 4) {
            if let symbol { Image(systemName: symbol).imageScale(.small) }
            Text(title).lineLimit(1)
            Image(systemName: "chevron.down").font(.system(size: 8, weight: .bold)).opacity(0.6)
        }
        .font(.caption.weight(.medium))
        .padding(.horizontal, 9)
        .padding(.vertical, 4)
        .background(active ? Color.accentColor.opacity(0.18) : Color.primary.opacity(0.06), in: Capsule())
        .foregroundStyle(active ? Color.accentColor : Color.primary)
        .contentShape(Capsule())
    }
}
