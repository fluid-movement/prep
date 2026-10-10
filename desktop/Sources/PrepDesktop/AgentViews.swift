import SwiftUI
import PrepKit

/// One agent session as the activity stream knows it.
struct AgentSession: Identifiable {
    var id: String
    var actor: String
    var issue: String?
    var lastSeen: Date?
    var events: Int

    var isLive: Bool { lastSeen.map { Date.now.timeIntervalSince($0) < 180 } ?? false }
    var shortSession: String {
        let raw = id.hasPrefix("session:") ? String(id.dropFirst(8)) : id
        return String(raw.prefix(6))
    }
}

extension ProjectModel {
    /// A path relative to the project, else with ~ for the home directory.
    func displayPath(_ p: String) -> String {
        if let root = project?.root, p.hasPrefix(root + "/") { return String(p.dropFirst(root.count + 1)) }
        return p.replacingOccurrences(of: NSHomeDirectory(), with: "~")
    }

    var sessions: [AgentSession] {
        let events = activity.events ?? []
        var out: [AgentSession] = []
        for f in activity.focus ?? [] {
            let mine = events.filter { ($0.session ?? "") == (f.session ?? "") && f.session != nil }
            out.append(AgentSession(id: f.agent, actor: f.actor ?? f.agent, issue: f.issue,
                                    lastSeen: mine.last?.at ?? f.at, events: mine.count))
        }
        return out.sorted { ($0.lastSeen ?? .distantPast) > ($1.lastSeen ?? .distantPast) }
    }
}

struct AgentSidebar: View {
    @Bindable var model: ProjectModel

    var body: some View {
        let sessions = model.sessions
        if sessions.isEmpty {
            ContentUnavailableView("No Agents Yet", systemImage: "sparkles",
                                   description: Text("Agents appear here once they run prep in this project."))
        } else {
            List(selection: $model.agentSelection) {
                Section("Sessions") {
                    ForEach(sessions) { s in
                        VStack(alignment: .leading, spacing: 6) {
                            HStack(spacing: 8) {
                                ActorAvatar(actor: s.actor, size: 22)
                                VStack(alignment: .leading, spacing: 0) {
                                    HStack(spacing: 5) {
                                        Text(s.actor).font(.callout.weight(.medium)).lineLimit(1)
                                        Text(s.shortSession).font(.caption.monospaced()).foregroundStyle(.tertiary)
                                    }
                                    if let seen = s.lastSeen {
                                        Text(s.isLive ? "active now" : seen.relative)
                                            .font(.caption).foregroundStyle(s.isLive ? .green : .secondary)
                                    }
                                }
                                Spacer()
                                if s.isLive {
                                    Circle().fill(.green).frame(width: 8, height: 8).help("Active in the last 3 minutes")
                                }
                            }
                            if let id = s.issue, let issue = model.summary(id) {
                                Button { model.go(to: id) } label: {
                                    HStack(spacing: 5) {
                                        StateIcon(state: issue.state, size: 10)
                                        Text(issue.title).lineLimit(1).font(.caption)
                                    }
                                    .padding(.horizontal, 6).padding(.vertical, 3)
                                }
                                .buttonStyle(HoverButtonStyle(radius: 5))
                                .padding(.leading, 26)
                                .help("Working on \(issue.title)")
                            }
                        }
                        .padding(.vertical, 3)
                        .tag(s.id)
                    }
                }
            }
            .listStyle(.sidebar)
        }
    }
}

/// The activity stream, newest first, grouped by day.
struct AgentFeedView: View {
    var model: ProjectModel
    @AppStorage("showModelRequests") private var showRequests = false

    var events: [ActivityEvent] {
        let all = (model.activity.events ?? []).reversed()
        let shown = all.filter { showRequests || $0.kind != "request" }
        guard let sel = model.agentSelection,
              let session = model.activity.focus?.first(where: { $0.agent == sel })?.session else { return Array(shown) }
        return shown.filter { $0.session == session }
    }

    var body: some View {
        let evs = events
        if evs.isEmpty {
            ContentUnavailableView("No Activity", systemImage: "waveform.path.ecg",
                                   description: Text("What agents do with prep in this project shows up here, live."))
        } else {
            let days = Dictionary(grouping: evs) { Calendar.current.startOfDay(for: $0.at) }
            List {
                ForEach(days.keys.sorted(by: >), id: \.self) { day in
                    Section(day.formatted(date: .complete, time: .omitted)) {
                        ForEach(Array((days[day] ?? []).enumerated()), id: \.offset) { _, e in
                            EventRow(model: model, event: e)
                        }
                    }
                }
            }
            .listStyle(.inset)
            .safeAreaInset(edge: .top, spacing: 0) {
                HStack(spacing: 12) {
                    if model.agentSelection != nil {
                        Label("One session", systemImage: "line.3.horizontal.decrease.circle.fill")
                            .foregroundStyle(Color.accentColor)
                        Button("Show All") { model.agentSelection = nil }.buttonStyle(.link)
                    } else {
                        Text("All sessions").foregroundStyle(.secondary)
                    }
                    Spacer()
                    Toggle("Model requests", isOn: $showRequests).toggleStyle(.checkbox)
                        .help("Show each model request with its token count")
                }
                .font(.callout)
                .padding(.horizontal, 16).padding(.vertical, 8)
                .background(.bar)
                .overlay(alignment: .bottom) { Divider() }
            }
        }
    }
}

struct EventRow: View {
    var model: ProjectModel
    var event: ActivityEvent

    var symbol: String {
        switch event.kind {
        case "request": return "bubble.left"
        case "focus": return "scope"
        case "tool":
            switch event.verb {
            case "read": return "doc.text"
            case "edited", "wrote": return "pencil"
            case "ran": return "terminal"
            case "searched": return "magnifyingglass"
            default: return "wrench.and.screwdriver"
            }
        default: return "checklist"
        }
    }

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 10) {
            Text(event.at, format: .dateTime.hour().minute().second())
                .font(.caption.monospacedDigit())
                .foregroundStyle(.tertiary)
                .frame(width: 64, alignment: .leading)
            Image(systemName: symbol)
                .foregroundStyle(event.kind == "prep" ? Color.accentColor : .secondary)
                .frame(width: 16)
            if event.kind == "request" {
                Text("Model request").foregroundStyle(.secondary)
                if let op = event.op { Text(op).font(.caption.monospaced()).foregroundStyle(.tertiary) }
                if let t = event.tokens, let out = t.output {
                    Text("\(out.formatted()) tokens out").font(.caption).foregroundStyle(.tertiary)
                }
            } else {
                Text(event.verb ?? event.op ?? event.kind ?? "").fontWeight(event.kind == "prep" ? .medium : .regular)
                if let id = event.issue {
                    Button { model.go(to: id) } label: {
                        HStack(spacing: 4) {
                            if let s = model.summary(id) { StateIcon(state: s.state, size: 10) }
                            Text(event.target ?? model.title(id)).lineLimit(1)
                        }
                    }
                    .buttonStyle(.link)
                } else if let t = event.target {
                    Text(model.displayPath(t)).font(event.kind == "tool" ? .callout.monospaced() : .callout)
                        .foregroundStyle(.secondary).lineLimit(1).truncationMode(.middle)
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.vertical, 1)
    }
}
