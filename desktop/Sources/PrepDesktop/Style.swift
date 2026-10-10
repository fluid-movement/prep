import SwiftUI
import PrepKit

/// Colors and small badges shared by the views, after the TUI's state colors.
enum Style {
    static func color(state: String) -> Color {
        switch state {
        case "open": .gray
        case "defined": .blue
        case "ready": .teal
        case "in-progress": .orange
        case "done": .green
        case "dropped": .secondary
        default: .primary
        }
    }

    static func priorityMark(_ p: String?) -> String? {
        switch p {
        case "critical": "!crit"
        case "high": "!high"
        case "low": "low"
        default: nil
        }
    }
}

struct StateBadge: View {
    var state: String
    var body: some View {
        Text(state)
            .font(.caption.weight(.medium))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(Style.color(state: state).opacity(0.18), in: Capsule())
            .foregroundStyle(Style.color(state: state))
    }
}

struct Pill: View {
    var text: String
    var color: Color = .secondary
    var body: some View {
        Text(text)
            .font(.caption)
            .padding(.horizontal, 5)
            .padding(.vertical, 1)
            .overlay(Capsule().stroke(color.opacity(0.5)))
            .foregroundStyle(color)
    }
}

struct IssueRow: View {
    var issue: IssueSummary
    var body: some View {
        HStack(spacing: 6) {
            Circle()
                .fill(Style.color(state: issue.state))
                .frame(width: 8, height: 8)
                .help(issue.state)
            Text(issue.shortID)
                .font(.system(.caption, design: .monospaced))
                .foregroundStyle(.secondary)
            if let mark = Style.priorityMark(issue.priority) {
                Text(mark).font(.caption.weight(.semibold)).foregroundStyle(.red)
            }
            Text(issue.title)
                .lineLimit(1)
                .strikethrough(issue.state == "dropped")
                .foregroundStyle(issue.isResolved ? .secondary : .primary)
            Spacer(minLength: 4)
            if issue.stale { Pill(text: "stale", color: .yellow) }
            if issue.blocked { Image(systemName: "lock.fill").font(.caption).foregroundStyle(.red).help("blocked") }
            if let p = issue.progress, p.total > 0 {
                Text("\(p.done + p.dropped)/\(p.total)")
                    .font(.caption.monospacedDigit())
                    .foregroundStyle(.secondary)
            }
        }
    }
}
