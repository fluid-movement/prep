import Foundation

/// What waits on the person: prep gives the requirement to the user, so
/// settling questions, defining, signing off, acknowledging changes and
/// checking finished work are theirs. Agents do the rest.
public enum InboxKind: String, CaseIterable, Sendable {
    case review, acknowledge, signOff, questions, define

    public var title: String {
        switch self {
        case .review: "Review Completed Work"
        case .acknowledge: "Acknowledge Changes"
        case .signOff: "Sign Off"
        case .questions: "Answer Questions"
        case .define: "Define"
        }
    }

    public var reason: String {
        switch self {
        case .review: "Finished by an agent"
        case .acknowledge: "Requirement changed since it was defined"
        case .signOff: "Enrichment is complete; mark it ready"
        case .questions: "Has open questions in the requirement"
        case .define: "Confirm the requirement and define it"
        }
    }

    public var symbol: String {
        switch self {
        case .review: "checkmark.seal"
        case .acknowledge: "exclamationmark.arrow.circlepath"
        case .signOff: "signature"
        case .questions: "questionmark.bubble"
        case .define: "text.badge.checkmark"
        }
    }
}

public struct InboxItem: Identifiable, Hashable, Sendable {
    public var kind: InboxKind
    public var issue: IssueSummary
    public var id: String { "\(kind.rawValue):\(issue.id)" }
}

public enum Inbox {
    /// Builds the inbox from the full issue list.
    /// - readyAllowed: defined issues whose ready transition is allowed now.
    /// - since: completions before this time are not reviewed (the first time
    ///   the app saw the project), so an old project does not flood the inbox.
    /// - reviewed: completions the person already looked at.
    public static func build(_ issues: [IssueSummary], readyAllowed: Set<String>, since: Date, reviewed: Set<String>) -> [InboxItem] {
        var out: [InboxItem] = []
        for i in issues {
            if i.isResolved {
                if i.state == "done", let at = i.resolvedAt, at > since, !reviewed.contains(i.id),
                   !(i.resolvedBy ?? "").hasPrefix("human:") {
                    out.append(InboxItem(kind: .review, issue: i))
                }
                continue
            }
            if i.stale {
                out.append(InboxItem(kind: .acknowledge, issue: i))
            } else if i.state == "defined", readyAllowed.contains(i.id) {
                out.append(InboxItem(kind: .signOff, issue: i))
            } else if i.state == "open", i.kind != "manual" {
                out.append(InboxItem(kind: i.openQuestions == true ? .questions : .define, issue: i))
            }
        }
        let order = InboxKind.allCases
        return out.sorted { a, b in
            if a.kind != b.kind { return order.firstIndex(of: a.kind)! < order.firstIndex(of: b.kind)! }
            if a.kind == .review { return (a.issue.resolvedAt ?? .distantPast) > (b.issue.resolvedAt ?? .distantPast) }
            let pa = Lifecycle.priorities.firstIndex(of: a.issue.priority ?? "medium") ?? 2
            let pb = Lifecycle.priorities.firstIndex(of: b.issue.priority ?? "medium") ?? 2
            return pa != pb ? pa < pb : a.issue.id < b.issue.id
        }
    }
}
