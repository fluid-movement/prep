import Foundation

// Codable mirrors of the prep CLI's --json output. Only the fields the app
// shows are decoded; everything optional that prep omits when empty is
// optional here too.

/// One row of `prep list --json`.
public struct IssueSummary: Codable, Identifiable, Hashable, Sendable {
    public var id: String
    public var title: String
    public var kind: String
    public var state: String
    public var stale: Bool
    public var blocked: Bool
    public var actionable: Bool
    public var parent: String?
    public var tags: [String]?
    public var priority: String?
    public var children: Int?
    public var progress: Progress?

    public init(id: String, title: String, kind: String = "code", state: String = "open",
                stale: Bool = false, blocked: Bool = false, actionable: Bool = false,
                parent: String? = nil, tags: [String]? = nil, priority: String? = nil,
                children: Int? = nil, progress: Progress? = nil) {
        self.id = id; self.title = title; self.kind = kind; self.state = state
        self.stale = stale; self.blocked = blocked; self.actionable = actionable
        self.parent = parent; self.tags = tags; self.priority = priority
        self.children = children; self.progress = progress
    }

    /// The last six characters, as the TUI shows IDs.
    public var shortID: String { String(id.suffix(6)) }
    public var isResolved: Bool { state == "done" || state == "dropped" }
}

public struct Progress: Codable, Hashable, Sendable {
    public var done: Int
    public var dropped: Int
    public var total: Int
}

public struct IssueList: Codable, Sendable {
    public var issues: [IssueSummary]
}

/// `prep show <id> --json`.
public struct IssueDetail: Codable, Sendable {
    public var actionable: Bool
    public var blocked: Bool
    public var blocks: [String]?
    public var children: [String]?
    public var context: String?
    public var definitionOfDone: [String]?
    public var findings: String?
    public var history: String?
    public var issue: IssueRecord
    public var priority: String?
    public var progress: Progress?
    public var stale: Bool
    public var state: String

    enum CodingKeys: String, CodingKey {
        case actionable, blocked, blocks, children, context, findings, history, issue, priority, progress, stale, state
        case definitionOfDone = "definition_of_done"
    }
}

public struct IssueRecord: Codable, Sendable {
    public var id: String
    public var title: String
    public var kind: String
    public var parent: String?
    public var dependsOn: [String]?
    public var tags: [String]?
    public var requirement: String?
    public var criteria: [Criterion]?
    public var decisions: [Decision]?
    public var contextLinks: [String]?
    public var contextPaths: [String]?
    public var baselines: [Baseline]?
    public var ready: Signature?
    public var claim: Signature?
    public var resolution: Resolution?

    enum CodingKeys: String, CodingKey {
        case id, title, kind, parent, tags, requirement, criteria, decisions, ready, claim, resolution
        case dependsOn = "depends_on"
        case contextLinks = "context_links"
        case contextPaths = "context_paths"
        case baselines
    }
}

public struct Criterion: Codable, Hashable, Sendable {
    public var text: String
    public var checked: Bool?
}

public struct Decision: Codable, Hashable, Sendable {
    public var id: String
    public var title: String
    public var date: String?
    public var body: String?
    public var outcome: Bool?
    public var supersedes: String?
}

public struct Signature: Codable, Hashable, Sendable {
    public var by: String?
    public var at: String?
    public var note: String?
}

public struct Resolution: Codable, Hashable, Sendable {
    public var outcome: String
    public var by: String?
    public var at: String?
    public var evidence: String?
    public var reason: String?
    public var note: String?
    public var documentation: Documentation?
}

public struct Documentation: Codable, Hashable, Sendable {
    public var entries: [String]?
    public var noImpact: String?

    enum CodingKeys: String, CodingKey {
        case entries
        case noImpact = "no_impact"
    }
}

public struct Baseline: Codable, Hashable, Sendable {
    public var name: String
    public var by: String?
    public var at: String?
}

/// `prep guide <id> --json`, reduced to the transitions.
public struct Guide: Codable, Sendable {
    public var step: String?
    public var instructions: [String]?
    public var transitions: [Transition]?
}

public struct Transition: Codable, Hashable, Sendable, Identifiable {
    public var op: String
    public var command: String
    public var allowed: Bool
    public var unmet: [Gate]?
    public var needs: [String]?
    public var id: String { op }

    public init(op: String, command: String, allowed: Bool, unmet: [Gate]? = nil, needs: [String]? = nil) {
        self.op = op
        self.command = command
        self.allowed = allowed
        self.unmet = unmet
        self.needs = needs
    }
}

public struct Gate: Codable, Hashable, Sendable {
    public var code: String
    public var message: String
}

/// `prep knowledge list --json` and `prep knowledge show --json`.
public struct KnowledgeList: Codable, Sendable {
    public var entries: [KnowledgeEntry]
}

public struct KnowledgeEntry: Codable, Identifiable, Hashable, Sendable {
    public var path: String
    public var type: String
    public var title: String
    public var description: String?
    public var status: String?
    public var size: Int?
    public var scope: [String]?
    public var text: String?
    public var id: String { path }
}

/// `prep activity --json`.
public struct ActivityFeed: Codable, Sendable {
    public var events: [ActivityEvent]?
    public var focus: [Focus]?

    public init(events: [ActivityEvent]? = nil, focus: [Focus]? = nil) {
        self.events = events
        self.focus = focus
    }
}

public struct ActivityEvent: Codable, Hashable, Sendable {
    public var at: Date
    public var actor: String?
    public var session: String?
    public var kind: String?
    public var op: String?
    public var issue: String?
    public var verb: String?
    public var target: String?
    public var area: String?
    public var tokens: Tokens?
}

public struct Tokens: Codable, Hashable, Sendable {
    public var input: Int?
    public var output: Int?
}

public struct Focus: Codable, Hashable, Sendable {
    public var agent: String
    public var actor: String?
    public var session: String?
    public var issue: String?
    public var at: Date?
}

/// `prep views --json`.
public struct Views: Codable, Sendable {
    public var views: [String: String]
}

/// `prep projects --json`.
public struct ProjectList: Codable, Sendable {
    public var projects: [ProjectRef]
}

public struct ProjectRef: Codable, Hashable, Identifiable, Sendable {
    public var root: String
    public var name: String
    public var lastUsed: Date?
    public var id: String { root }

    public init(root: String, name: String, lastUsed: Date? = nil) {
        self.root = root; self.name = name; self.lastUsed = lastUsed
    }

    enum CodingKeys: String, CodingKey {
        case root, name
        case lastUsed = "last_used"
    }
}

/// The error object prep prints under --json when a command fails.
public struct CLIErrorBody: Codable, Sendable {
    public struct Inner: Codable, Sendable {
        public var code: String
        public var message: String
    }
    public var ok: Bool?
    public var error: Inner?
}

extension JSONDecoder {
    /// A decoder for prep's JSON: RFC 3339 dates, with or without fractions.
    public static let prep: JSONDecoder = {
        let d = JSONDecoder()
        d.dateDecodingStrategy = .custom { decoder in
            let s = try decoder.singleValueContainer().decode(String.self)
            if let date = PrepDate.parse(s) { return date }
            throw DecodingError.dataCorrupted(.init(codingPath: decoder.codingPath, debugDescription: "bad date \(s)"))
        }
        return d
    }()
}

public enum PrepDate {
    public static func parse(_ s: String) -> Date? {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let d = f.date(from: s) { return d }
        f.formatOptions = [.withInternetDateTime]
        return f.date(from: s)
    }
}
