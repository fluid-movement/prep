import SwiftUI
import PrepKit

/// What needs the person, grouped by what to do, each with its action.
struct InboxSidebar: View {
    @Bindable var model: ProjectModel
    @State private var showDefine = false

    var body: some View {
        let groups = Dictionary(grouping: model.inbox, by: \.kind)
        if model.inbox.isEmpty {
            ContentUnavailableView("Nothing Needs You", systemImage: "tray",
                                   description: Text("Questions to answer, sign-offs and finished work to review show up here."))
        } else {
            List(selection: $model.selection) {
                ForEach(InboxKind.allCases, id: \.self) { kind in
                    if let items = groups[kind], !items.isEmpty {
                        Section {
                            let shown = kind == .define && !showDefine ? Array(items.prefix(5)) : items
                            ForEach(shown) { item in
                                InboxRow(model: model, item: item).tag(item.issue.id)
                            }
                            if kind == .define, items.count > 5 {
                                Button(showDefine ? "Show Fewer" : "Show All \(items.count)") { showDefine.toggle() }
                                    .buttonStyle(.link)
                                    .font(.caption)
                            }
                        } header: {
                            HStack {
                                Label(kind.title, systemImage: kind.symbol)
                                Text("\(items.count)").foregroundStyle(.tertiary)
                                Spacer()
                                if kind == .review, items.count > 1 {
                                    Button("Mark All Reviewed") { model.markReviewed(items.map(\.issue.id)) }
                                        .buttonStyle(.link)
                                        .font(.caption)
                                }
                            }
                        }
                    }
                }
            }
            .listStyle(.sidebar)
        }
    }
}

struct InboxRow: View {
    var model: ProjectModel
    var item: InboxItem
    @State private var hovering = false

    var actionTitle: String {
        switch item.kind {
        case .review: "Reviewed"
        case .acknowledge: "Acknowledge"
        case .signOff: "Mark Ready"
        case .questions: "Answer…"
        case .define: "Define"
        }
    }

    var detail: String {
        switch item.kind {
        case .review:
            var s = "Done"
            if let by = item.issue.resolvedBy { s += " by \(by)" }
            if let at = item.issue.resolvedAt { s += ", \(at.relative)" }
            return s
        default:
            if let p = item.issue.parent { return "\(item.kind.reason) · in \(model.title(p))" }
            return item.kind.reason
        }
    }

    var body: some View {
        HStack(spacing: 8) {
            StateIcon(state: item.issue.state)
            VStack(alignment: .leading, spacing: 2) {
                Text(item.issue.title).lineLimit(1)
                Text(detail).font(.caption).foregroundStyle(.secondary).lineLimit(1)
            }
            Spacer(minLength: 6)
            PriorityIcon(priority: item.issue.priority).imageScale(.small)
            if hovering {
                Button(actionTitle) { model.resolve(item) }
                    .controlSize(.small)
                    .disabled(model.busy)
                    .help(item.kind.reason)
            }
        }
        .padding(.vertical, 2)
        .contentShape(Rectangle())
        .onHover { hovering = $0 }
        .issueContextMenu(model: model, issue: item.issue)
    }
}

/// Shown on an issue that waits on the person, above its record.
struct InboxBanner: View {
    var model: ProjectModel
    var item: InboxItem

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: item.kind.symbol).font(.title3).foregroundStyle(Color.accentColor)
            VStack(alignment: .leading, spacing: 1) {
                Text(item.kind == .review ? "Review the finished work" : "This needs you").font(.callout.weight(.semibold))
                Text(item.kind.reason).font(.caption).foregroundStyle(.secondary)
            }
            Spacer()
            Button(InboxRow(model: model, item: item).actionTitle) { model.resolve(item) }
                .buttonStyle(.borderedProminent)
                .disabled(model.busy)
        }
        .padding(12)
        .background(Color.accentColor.opacity(0.08), in: RoundedRectangle(cornerRadius: Style.cardRadius))
        .overlay(RoundedRectangle(cornerRadius: Style.cardRadius).strokeBorder(Color.accentColor.opacity(0.3)))
    }
}
