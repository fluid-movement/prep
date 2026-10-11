import SwiftUI
import PrepKit

/// Where the issue sits in the flow of work: what it waits on to the left,
/// the issue in the middle, what waits on it to the right, joined by
/// arrows. A done dependency's arrow is green; an open one, which blocks,
/// is orange.
struct DependencyMap: View {
    var model: ProjectModel
    var detail: IssueDetail

    var upstream: [String] { detail.issue.dependsOn ?? [] }
    var downstream: [String] { detail.blocks ?? [] }

    /// The cross-project dependency a reference names, if it is one.
    func cross(_ ref: String) -> CrossDep? { detail.crossDeps?.first { $0.ref == ref } }

    /// Whether a dependency still blocks: an open local issue, or a found
    /// cross-project one that is not done (unknown ones never block).
    func blocking(_ id: String) -> Bool {
        if let x = cross(id) { return x.blocks }
        return !(model.summary(id)?.isResolved ?? false)
    }

    var body: some View {
        let waiting = upstream.filter(blocking).count
        Card("Dependencies", symbol: "point.3.connected.trianglepath.dotted", accessory: {
            if waiting > 0 {
                Chip(text: "Waiting on \(waiting)", symbol: "lock.fill", color: .orange)
            } else if !upstream.isEmpty {
                Chip(text: "Unblocked", symbol: "lock.open.fill", color: .green)
            }
        }) {
            VStack(spacing: 6) {
                HStack(spacing: 44) {
                    header("Depends on")
                    header("")
                    header("Blocks")
                }
                HStack(alignment: .center, spacing: 44) {
                    column(ids: upstream, side: "L", empty: "Nothing")
                        .frame(maxWidth: .infinity)
                    center.frame(maxWidth: .infinity)
                    column(ids: downstream, side: "R", empty: "Nothing waits on this")
                        .frame(maxWidth: .infinity)
                }
            }
            .padding(.vertical, 4)
            .backgroundPreferenceValue(NodeAnchors.self) { anchors in
                GeometryReader { g in
                    connectors(anchors, in: g)
                }
            }
        }
    }

    var center: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 5) {
                StateIcon(state: detail.state)
                Text("This issue").font(.caption.weight(.semibold)).foregroundStyle(.secondary)
            }
            Text(detail.issue.title).font(.callout.weight(.semibold)).lineLimit(3)
        }
        .padding(10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.accentColor.opacity(0.1), in: RoundedRectangle(cornerRadius: 8))
        .overlay(RoundedRectangle(cornerRadius: 8).strokeBorder(Color.accentColor.opacity(0.7), lineWidth: 1.5))
        .anchorPreference(key: NodeAnchors.self, value: .bounds) { ["C": $0] }
    }

    func header(_ title: String) -> some View {
        Text(title.uppercased()).font(.caption2.weight(.semibold)).foregroundStyle(.tertiary)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    func column(ids: [String], side: String, empty: String) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            if ids.isEmpty {
                Text(empty).font(.callout).foregroundStyle(.tertiary)
                    .padding(10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .overlay(RoundedRectangle(cornerRadius: 8).strokeBorder(style: StrokeStyle(lineWidth: 1, dash: [4, 3])).foregroundStyle(.quaternary))
            }
            ForEach(ids, id: \.self) { id in
                Group {
                    if let x = cross(id) {
                        CrossIssueCard(model: model, dep: x)
                    } else {
                        IssueCard(model: model, id: id)
                    }
                }
                .anchorPreference(key: NodeAnchors.self, value: .bounds) { ["\(side)-\(id)": $0] }
            }
        }
    }

    func connectors(_ anchors: [String: Anchor<CGRect>], in g: GeometryProxy) -> some View {
        guard let c = anchors["C"].map({ g[$0] }) else { return AnyView(EmptyView()) }
        return AnyView(ZStack {
            ForEach(upstream, id: \.self) { id in
                if let r = anchors["L-\(id)"].map({ g[$0] }) {
                    let done = !blocking(id)
                    Arrow(from: CGPoint(x: r.maxX, y: r.midY), to: CGPoint(x: c.minX, y: c.midY))
                        .stroke(done ? Color.green.opacity(0.7) : Color.orange, style: StrokeStyle(lineWidth: 1.6, lineCap: .round, dash: done ? [] : [5, 3]))
                }
            }
            ForEach(downstream, id: \.self) { id in
                if let r = anchors["R-\(id)"].map({ g[$0] }) {
                    Arrow(from: CGPoint(x: c.maxX, y: c.midY), to: CGPoint(x: r.minX, y: r.midY))
                        .stroke(Color.secondary.opacity(0.6), style: StrokeStyle(lineWidth: 1.6, lineCap: .round))
                }
            }
        })
    }
}

struct NodeAnchors: PreferenceKey {
    static let defaultValue: [String: Anchor<CGRect>] = [:]
    static func reduce(value: inout [String: Anchor<CGRect>], nextValue: () -> [String: Anchor<CGRect>]) {
        value.merge(nextValue()) { $1 }
    }
}

/// A curved connector with an arrowhead at its end.
struct Arrow: Shape {
    var from: CGPoint
    var to: CGPoint

    func path(in rect: CGRect) -> Path {
        var p = Path()
        let dx = max(18, (to.x - from.x) / 2)
        p.move(to: from)
        p.addCurve(to: to, control1: CGPoint(x: from.x + dx, y: from.y), control2: CGPoint(x: to.x - dx, y: to.y))
        p.move(to: CGPoint(x: to.x - 6, y: to.y - 4))
        p.addLine(to: to)
        p.addLine(to: CGPoint(x: to.x - 6, y: to.y + 4))
        return p
    }
}

/// A card for another project's issue: the link's name over the ID, its
/// state and title from that project; when the project is not checked out
/// here, where to clone it. Clicking a found one opens it there.
struct CrossIssueCard: View {
    var model: ProjectModel
    var dep: CrossDep

    var body: some View {
        Button {
            if dep.isFound, let root = dep.root { Task { await model.go(to: dep.id, inProject: root) } }
        } label: {
            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 5) {
                    if dep.isFound {
                        StateIcon(state: dep.state ?? "open", size: 11)
                        Text(Lifecycle.title(dep.state ?? "")).font(.caption).foregroundStyle(.secondary)
                    } else {
                        Image(systemName: "questionmark.circle").imageScale(.small).foregroundStyle(.secondary)
                    }
                    Spacer(minLength: 2)
                    Label(dep.link, systemImage: "shippingbox").font(.caption).foregroundStyle(.secondary).labelStyle(.titleAndIcon)
                    ShortID(id: dep.id)
                }
                Text(caption)
                    .font(.callout)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
                    .foregroundStyle(dep.isFound && dep.state != "done" && dep.state != "dropped" ? .primary : .secondary)
            }
            .padding(9)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color(nsColor: .windowBackgroundColor).opacity(0.6), in: RoundedRectangle(cornerRadius: 8))
            .overlay(RoundedRectangle(cornerRadius: 8).strokeBorder(style: StrokeStyle(lineWidth: 1, dash: dep.isFound ? [] : [4, 3])).foregroundStyle(.separator))
        }
        .buttonStyle(HoverButtonStyle(radius: 8))
        .disabled(!dep.isFound)
        .help(dep.isFound ? "Open in \(dep.project ?? dep.link)" : caption)
    }

    var caption: String {
        switch dep.status {
        case "found": return dep.title ?? dep.id
        case "missing": return "No such issue in \(dep.project ?? dep.link)"
        default:
            if let remote = dep.remote { return "Not checked out here · clone \(remote)" }
            return "Not checked out here"
        }
    }
}

/// A compact clickable card for a related issue.
struct IssueCard: View {
    var model: ProjectModel
    var id: String

    var body: some View {
        let s = model.summary(id)
        Button { model.go(to: id) } label: {
            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 5) {
                    StateIcon(state: s?.state ?? "open", size: 11)
                    Text(Lifecycle.title(s?.state ?? "")).font(.caption).foregroundStyle(.secondary)
                    Spacer(minLength: 2)
                    PriorityIcon(priority: s?.priority).imageScale(.small)
                    ShortID(id: id)
                }
                Text(s?.title ?? id)
                    .font(.callout)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
                    .strikethrough(s?.state == "dropped", color: .secondary)
                    .foregroundStyle(s?.isResolved == true ? .secondary : .primary)
            }
            .padding(9)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color(nsColor: .windowBackgroundColor).opacity(0.6), in: RoundedRectangle(cornerRadius: 8))
            .overlay(RoundedRectangle(cornerRadius: 8).strokeBorder(.separator))
        }
        .buttonStyle(HoverButtonStyle(radius: 8))
        .help("Go to \(s?.title ?? id)")
        .contextMenu {
            if let s { IssueMenuItems(model: model, issue: s) }
        }
    }
}

struct IssueMenuItems: View {
    var model: ProjectModel
    var issue: IssueSummary
    var body: some View {
        Button("Open") { model.go(to: issue.id) }
        Button("New Child Issue…") { model.newIssue(parent: issue.id) }
        Divider()
        Button("Copy ID") { model.copy(issue.id) }
        Button("Copy Title") { model.copy(issue.title) }
        Divider()
        Button("Reveal in Finder") { model.revealInFinder(issue.id) }
    }
}

/// A parent's children: progress by state, unresolved work first, resolved
/// work folded away, and a way to add the next child.
struct ChildrenCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    @State private var showResolved = false

    var body: some View {
        let kids = model.children(of: detail.issue.id)
        let open = kids.filter { !$0.isResolved }
        let resolved = kids.filter { $0.isResolved }
        let counts = Lifecycle.states.reversed().map { st in (state: st, count: kids.filter { $0.state == st }.count) }
        Card("Children", symbol: "list.bullet.indent", accessory: {
            HStack(spacing: 10) {
                if !kids.isEmpty {
                    Text("\(resolved.count) of \(kids.count) resolved").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                }
                Button { model.newIssue(parent: detail.issue.id) } label: {
                    Label("Add Child", systemImage: "plus")
                }
                .buttonStyle(.borderless)
                .font(.callout)
                .help("New child issue (⌥⌘N)")
            }
        }) {
            VStack(alignment: .leading, spacing: 10) {
                if kids.isEmpty {
                    Text("No children yet.").foregroundStyle(.secondary)
                } else {
                    StateBar(counts: counts, height: 7)
                    legend(counts)
                    VStack(spacing: 0) {
                        ForEach(open) { ChildRow(model: model, issue: $0) }
                    }
                    if !resolved.isEmpty {
                        DisclosureGroup(isExpanded: $showResolved) {
                            VStack(spacing: 0) {
                                ForEach(resolved) { ChildRow(model: model, issue: $0) }
                            }
                        } label: {
                            Text("Resolved · \(resolved.count)").font(.callout).foregroundStyle(.secondary)
                        }
                    }
                }
            }
        }
    }

    func legend(_ counts: [(state: String, count: Int)]) -> some View {
        HStack(spacing: 12) {
            ForEach(counts.filter { $0.count > 0 }, id: \.state) { c in
                HStack(spacing: 4) {
                    Circle().fill(Style.color(state: c.state).opacity(c.state == "open" ? 0.35 : 0.9)).frame(width: 7, height: 7)
                    Text("\(c.count) \(Lifecycle.title(c.state).lowercased())")
                }
            }
        }
        .font(.caption)
        .foregroundStyle(.secondary)
    }
}

struct ChildRow: View {
    var model: ProjectModel
    var issue: IssueSummary

    var body: some View {
        Button { model.go(to: issue.id) } label: {
            HStack(spacing: 8) {
                StateIcon(state: issue.state)
                Text(issue.title)
                    .lineLimit(1)
                    .strikethrough(issue.state == "dropped", color: .secondary)
                    .foregroundStyle(issue.isResolved ? .secondary : .primary)
                ForEach((issue.tags ?? []).prefix(2), id: \.self) { TagChip(tag: $0) }
                Spacer(minLength: 6)
                if issue.actionable {
                    Image(systemName: "bolt.fill").imageScale(.small).foregroundStyle(.green).help("Actionable")
                }
                if issue.blocked {
                    Image(systemName: "lock.fill").imageScale(.small).foregroundStyle(.secondary).help("Blocked")
                }
                PriorityIcon(priority: issue.priority).imageScale(.small)
                if let p = issue.progress, p.total > 0 {
                    Text("\(p.done + p.dropped)/\(p.total)").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                    ProgressRing(progress: p, size: 12)
                }
                ShortID(id: issue.id)
                Image(systemName: "chevron.right").font(.caption2.weight(.semibold)).foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 6)
        }
        .buttonStyle(HoverButtonStyle(radius: 6))
        .contextMenu { IssueMenuItems(model: model, issue: issue) }
    }
}
