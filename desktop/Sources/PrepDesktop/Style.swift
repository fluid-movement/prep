import SwiftUI
import PrepKit

/// The app's visual vocabulary: one symbol and color per state, priority
/// and kind, used the same way in every list, card and inspector.
enum Style {
    static func color(state: String) -> Color {
        switch state {
        case "open": .secondary
        case "defined": .blue
        case "ready": .teal
        case "in-progress": .orange
        case "done": .green
        case "dropped": .gray
        default: .primary
        }
    }

    /// Open is a dashed outline, each step fills more, done is a check.
    static func symbol(state: String) -> String {
        switch state {
        case "open": "circle.dashed"
        case "defined": "circle"
        case "ready": "circle.circle"
        case "in-progress": "circle.lefthalf.filled"
        case "done": "checkmark.circle.fill"
        case "dropped": "xmark.circle.fill"
        default: "circle"
        }
    }

    static func symbol(kind: String) -> String {
        switch kind {
        case "code": "chevron.left.forwardslash.chevron.right"
        case "research": "magnifyingglass"
        case "decision": "arrow.triangle.branch"
        case "manual": "shippingbox"
        default: "doc"
        }
    }

    static func priorityColor(_ p: String?) -> Color {
        switch p {
        case "critical": .red
        case "high": .orange
        case "low": .secondary
        default: .secondary
        }
    }

    static func symbol(knowledge type: String) -> String {
        switch type {
        case "overview": "map"
        case "component": "cube"
        case "feature": "sparkle"
        case "decision": "signpost.right"
        case "guide": "book"
        default: "doc.text"
        }
    }

    static let cardRadius: CGFloat = 10
}

struct StateIcon: View {
    var state: String
    var size: CGFloat = 13
    var body: some View {
        Image(systemName: Style.symbol(state: state))
            .font(.system(size: size, weight: .semibold))
            .foregroundStyle(Style.color(state: state))
            .frame(width: size + 4)
            .help(Lifecycle.title(state))
            .accessibilityLabel(Lifecycle.title(state))
    }
}

/// Priority as signal bars; critical gets an exclamation mark. Medium,
/// the default, shows nothing so the lists stay quiet.
struct PriorityIcon: View {
    var priority: String?
    var showMedium = false
    var body: some View {
        switch priority {
        case "critical":
            Image(systemName: "exclamationmark.square.fill").foregroundStyle(.red).help("Critical priority")
        case "high":
            Image(systemName: "cellularbars", variableValue: 1).foregroundStyle(.orange).help("High priority")
        case "low":
            Image(systemName: "cellularbars", variableValue: 0.34).foregroundStyle(.secondary).help("Low priority")
        default:
            if showMedium {
                Image(systemName: "cellularbars", variableValue: 0.67).foregroundStyle(.secondary).help("Medium priority")
            }
        }
    }
}

struct StateBadge: View {
    var state: String
    var body: some View {
        Label(Lifecycle.title(state), systemImage: Style.symbol(state: state))
            .font(.callout.weight(.medium))
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(Style.color(state: state).opacity(0.14), in: Capsule())
            .foregroundStyle(Style.color(state: state))
    }
}

struct Chip: View {
    var text: String
    var symbol: String? = nil
    var color: Color = .secondary
    var body: some View {
        HStack(spacing: 3) {
            if let symbol { Image(systemName: symbol).imageScale(.small) }
            Text(text)
        }
        .font(.caption.weight(.medium))
        .padding(.horizontal, 7)
        .padding(.vertical, 2)
        .background(color.opacity(0.12), in: Capsule())
        .foregroundStyle(color)
    }
}

struct TagChip: View {
    var tag: String
    var body: some View {
        Text("#\(tag)")
            .font(.caption.weight(.medium))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(.quaternary.opacity(0.6), in: Capsule())
            .foregroundStyle(.secondary)
    }
}

struct ShortID: View {
    var id: String
    var body: some View {
        Text(String(id.suffix(6)))
            .font(.system(.caption, design: .monospaced))
            .foregroundStyle(.tertiary)
            .help(id)
    }
}

/// A small ring filled by the resolved share of an issue's children.
struct ProgressRing: View {
    var progress: PrepKit.Progress
    var size: CGFloat = 14
    var fraction: Double { progress.total == 0 ? 0 : Double(progress.done + progress.dropped) / Double(progress.total) }
    var body: some View {
        ZStack {
            Circle().stroke(.quaternary, lineWidth: 2.5)
            Circle()
                .trim(from: 0, to: fraction)
                .stroke(fraction >= 1 ? Color.green : Color.accentColor, style: StrokeStyle(lineWidth: 2.5, lineCap: .round))
                .rotationEffect(.degrees(-90))
        }
        .frame(width: size, height: size)
        .help("\(progress.done + progress.dropped) of \(progress.total) children resolved")
    }
}

/// A bar split into one segment per state, in lifecycle order.
struct StateBar: View {
    var counts: [(state: String, count: Int)]
    var height: CGFloat = 6
    var body: some View {
        let total = max(1, counts.reduce(0) { $0 + $1.count })
        GeometryReader { g in
            HStack(spacing: 2) {
                ForEach(counts.filter { $0.count > 0 }, id: \.state) { c in
                    RoundedRectangle(cornerRadius: height / 2)
                        .fill(Style.color(state: c.state).opacity(c.state == "open" ? 0.35 : 0.9))
                        .frame(width: max(height, (g.size.width - CGFloat(counts.count) * 2) * CGFloat(c.count) / CGFloat(total)))
                        .help("\(c.count) \(Lifecycle.title(c.state).lowercased())")
                }
            }
        }
        .frame(height: height)
    }
}

/// A rounded content card with a header, the detail view's building block.
struct Card<Content: View, Accessory: View>: View {
    var title: String
    var symbol: String
    var accessory: Accessory
    var content: Content

    init(_ title: String, symbol: String, @ViewBuilder accessory: () -> Accessory, @ViewBuilder content: () -> Content) {
        self.title = title
        self.symbol = symbol
        self.accessory = accessory()
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 6) {
                Image(systemName: symbol).foregroundStyle(.secondary).frame(width: 16)
                Text(title).font(.headline)
                Spacer()
                accessory
            }
            content
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(nsColor: .controlBackgroundColor), in: RoundedRectangle(cornerRadius: Style.cardRadius))
        .overlay(RoundedRectangle(cornerRadius: Style.cardRadius).strokeBorder(.separator.opacity(0.6)))
    }
}

extension Card where Accessory == EmptyView {
    init(_ title: String, symbol: String, @ViewBuilder content: () -> Content) {
        self.init(title, symbol: symbol, accessory: { EmptyView() }, content: content)
    }
}

/// Initials in a colored circle, standing for an actor (agent or person).
struct ActorAvatar: View {
    var actor: String
    var size: CGFloat = 20
    var isAgent: Bool { !actor.hasPrefix("human:") }
    var label: String {
        if actor.hasPrefix("human:") { return String(actor.dropFirst(6).prefix(1)).uppercased() }
        return String(actor.prefix(1)).uppercased()
    }
    var color: Color {
        let palette: [Color] = [.blue, .purple, .pink, .orange, .teal, .indigo, .mint, .brown]
        let h = actor.unicodeScalars.reduce(0) { ($0 &* 31 &+ Int($1.value)) & 0xffff }
        return palette[h % palette.count]
    }
    var body: some View {
        ZStack {
            Circle().fill(color.gradient)
            if isAgent {
                Image(systemName: "sparkles").font(.system(size: size * 0.5, weight: .bold)).foregroundStyle(.white)
            } else {
                Text(label).font(.system(size: size * 0.5, weight: .bold)).foregroundStyle(.white)
            }
        }
        .frame(width: size, height: size)
        .help(actor)
    }
}

/// A plain button that highlights on hover, for clickable rows and cards.
struct HoverButtonStyle: ButtonStyle {
    var radius: CGFloat = 8
    func makeBody(configuration: Configuration) -> some View {
        HoverBody(configuration: configuration, radius: radius)
    }

    struct HoverBody: View {
        let configuration: Configuration
        var radius: CGFloat
        @State private var hovering = false
        var body: some View {
            configuration.label
                .contentShape(RoundedRectangle(cornerRadius: radius))
                .background(
                    RoundedRectangle(cornerRadius: radius)
                        .fill(Color.primary.opacity(configuration.isPressed ? 0.12 : hovering ? 0.06 : 0))
                )
                .onHover { hovering = $0 }
                .animation(.easeOut(duration: 0.12), value: hovering)
        }
    }
}

extension Date {
    var relative: String {
        let f = RelativeDateTimeFormatter()
        f.unitsStyle = .short
        return f.localizedString(for: self, relativeTo: .now)
    }
}

/// The issue row of every list: state, title, then the quiet details.
struct IssueRow: View {
    var issue: IssueSummary
    var showsID = true

    var body: some View {
        HStack(spacing: 7) {
            StateIcon(state: issue.state)
            Text(issue.title)
                .lineLimit(1)
                .strikethrough(issue.state == "dropped", color: .secondary)
                .foregroundStyle(issue.isResolved ? .secondary : .primary)
            Spacer(minLength: 4)
            if issue.stale {
                Image(systemName: "exclamationmark.arrow.circlepath").foregroundStyle(.yellow)
                    .help("Requirement changed since it was defined")
            }
            if issue.blocked {
                Image(systemName: "lock.fill").imageScale(.small).foregroundStyle(.secondary)
                    .help("Waiting on unresolved dependencies")
            }
            PriorityIcon(priority: issue.priority).imageScale(.small)
            if let p = issue.progress, p.total > 0 {
                ProgressRing(progress: p, size: 12)
            }
            if showsID { ShortID(id: issue.id) }
        }
    }
}
