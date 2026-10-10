import Foundation

/// `prep flags --json`: every query flag with the values it accepts now
/// and their issue counts.
public struct FlagList: Codable, Sendable {
    public var flags: [FlagInfo]
}

public struct FlagInfo: Codable, Hashable, Sendable {
    public var name: String
    public var kind: String
    public var description: String?
    public var count: Int?
    public var values: [FlagValue]?
}

public struct FlagValue: Codable, Hashable, Sendable {
    public var value: String
    public var count: Int

    public init(value: String, count: Int) {
        self.value = value
        self.count = count
    }
}

/// The sidebar's filter, turned into `prep list` arguments. A saved view
/// comes first; prep applies the other flags on top of it.
public struct IssueFilter: Equatable, Sendable {
    public var view: String?
    public var states: Set<String> = []
    public var kinds: Set<String> = []
    public var priorities: Set<String> = []
    public var tags: Set<String> = []
    public var actionable = false
    public var blocked = false
    public var stale = false
    /// Only this issue's descendants (prep list --under).
    public var under: String?
    public var text = ""

    public init(view: String? = nil) { self.view = view }

    /// How many filters besides the view and the text are set.
    public var activeCount: Int {
        [states, kinds, priorities, tags].filter { !$0.isEmpty }.count
            + [actionable, blocked, stale, under != nil].filter { $0 }.count
    }

    public var isEmpty: Bool { activeCount == 0 && text.trimmingCharacters(in: .whitespaces).isEmpty }

    public var arguments: [String] {
        var args: [String] = []
        if let view, !view.isEmpty { args += ["--view", view] }
        func list(_ flag: String, _ values: Set<String>, order: [String] = []) {
            guard !values.isEmpty else { return }
            let sorted = values.sorted { a, b in
                let ia = order.firstIndex(of: a) ?? Int.max, ib = order.firstIndex(of: b) ?? Int.max
                return ia != ib ? ia < ib : a < b
            }
            args += [flag, sorted.joined(separator: ",")]
        }
        list("--state", states, order: Lifecycle.states)
        list("--kind", kinds)
        list("--priority", priorities, order: Lifecycle.priorities)
        list("--tag", tags)
        if actionable { args.append("--actionable") }
        if blocked { args.append("--blocked") }
        if stale { args.append("--stale") }
        if let under { args += ["--under", under] }
        let t = text.trimmingCharacters(in: .whitespaces)
        if !t.isEmpty { args += ["--text", t] }
        return args
    }
}

public enum Lifecycle {
    public static let states = ["open", "defined", "ready", "in-progress", "done", "dropped"]
    /// The steps an issue walks through, in order (dropped is a way out).
    public static let steps = ["open", "defined", "ready", "in-progress", "done"]
    public static let priorities = ["critical", "high", "medium", "low"]
    public static let kinds = ["code", "research", "decision", "manual"]

    public static func title(_ state: String) -> String {
        switch state {
        case "in-progress": "In Progress"
        default: state.prefix(1).uppercased() + state.dropFirst()
        }
    }
}

/// One line of an issue's work log: `- <time> <actor>: <text>`.
public struct LogEntry: Hashable, Sendable {
    public var date: Date?
    public var actor: String?
    public var text: String
}

public enum WorkLog {
    public static func parse(_ history: String) -> [LogEntry] {
        var out: [LogEntry] = []
        for raw in history.components(separatedBy: "\n") {
            let line = raw.trimmingCharacters(in: .whitespaces)
            guard !line.isEmpty else { continue }
            if line.hasPrefix("- ") {
                let rest = line.dropFirst(2)
                let parts = rest.split(separator: " ", maxSplits: 1)
                if parts.count == 2, let date = PrepDate.parse(String(parts[0])) {
                    let body = parts[1]
                    if let colon = body.range(of: ": ") {
                        out.append(LogEntry(date: date, actor: String(body[..<colon.lowerBound]), text: String(body[colon.upperBound...])))
                    } else {
                        out.append(LogEntry(date: date, actor: nil, text: String(body)))
                    }
                    continue
                }
                out.append(LogEntry(date: nil, actor: nil, text: String(rest)))
            } else if !out.isEmpty {
                out[out.count - 1].text += "\n" + line
            } else {
                out.append(LogEntry(date: nil, actor: nil, text: line))
            }
        }
        return out
    }
}

/// prep IDs are ULIDs: the first ten characters are the creation time in
/// milliseconds, in Crockford base32.
public enum ULID {
    static let alphabet = Array("0123456789ABCDEFGHJKMNPQRSTVWXYZ")

    public static func date(_ id: String) -> Date? {
        guard id.count >= 10 else { return nil }
        var ms: UInt64 = 0
        for ch in id.uppercased().prefix(10) {
            guard let v = alphabet.firstIndex(of: ch) else { return nil }
            ms = ms * 32 + UInt64(v)
        }
        return Date(timeIntervalSince1970: Double(ms) / 1000)
    }
}
