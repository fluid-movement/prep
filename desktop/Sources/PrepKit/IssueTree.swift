import Foundation

/// An issue with its children, for the outline view.
public struct IssueNode: Identifiable, Hashable, Sendable {
    public var issue: IssueSummary
    public var children: [IssueNode]?
    public var id: String { issue.id }
}

public enum IssueTree {
    /// Builds the forest from a flat list: an issue whose parent is not in
    /// the list (filtered out, or none) becomes a root. Order follows the
    /// list, which prep already sorts by priority, then ID.
    public static func build(_ issues: [IssueSummary]) -> [IssueNode] {
        let ids = Set(issues.map(\.id))
        var byParent: [String: [IssueSummary]] = [:]
        var roots: [IssueSummary] = []
        for i in issues {
            if let p = i.parent, ids.contains(p) {
                byParent[p, default: []].append(i)
            } else {
                roots.append(i)
            }
        }
        func node(_ i: IssueSummary) -> IssueNode {
            let kids = byParent[i.id]?.map(node)
            return IssueNode(issue: i, children: kids)
        }
        return roots.map(node)
    }
}

/// Splits a query typed in the filter field into prep list arguments, so
/// `--state open,ready release` becomes ["--state", "open,ready", "--text", "release"].
public enum Query {
    public static func arguments(_ text: String) -> [String] {
        var args: [String] = []
        var words: [String] = []
        let parts = text.split(whereSeparator: \.isWhitespace).map(String.init)
        var k = 0
        while k < parts.count {
            let p = parts[k]
            if p.hasPrefix("--") {
                if p.contains("=") || k + 1 >= parts.count || parts[k + 1].hasPrefix("--") || booleanFlags.contains(p) {
                    args.append(p)
                } else {
                    args += [p, parts[k + 1]]
                    k += 1
                }
            } else {
                words.append(p)
            }
            k += 1
        }
        if !words.isEmpty { args += ["--text", words.joined(separator: " ")] }
        return args
    }

    /// Query flags that take no value.
    static let booleanFlags: Set<String> = ["--stale", "--actionable", "--blocked", "--leaf", "--top", "--tree"]
}
