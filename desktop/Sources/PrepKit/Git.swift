import Foundation

/// The few git reads the review screen needs: the commit an issue was
/// completed with and what it changed. Git is versioning here, not prep's
/// records, so reading it directly keeps the CLI contract intact.
public struct Git: Sendable {
    public var root: String
    public var binary = URL(fileURLWithPath: "/usr/bin/git")

    public init(root: String) { self.root = root }

    public func run(_ args: [String]) async throws -> String {
        let bin = binary, root = root
        return try await Task.detached {
            let r = try PrepCLI.runSync(bin, ["-C", root] + args, stdin: nil)
            guard r.status == 0 else {
                throw PrepError(code: "E_GIT", message: String(data: r.err, encoding: .utf8) ?? "git failed")
            }
            return String(data: r.out, encoding: .utf8) ?? ""
        }.value
    }

    /// The commit that completed an issue: its evidence, else the commit
    /// that added its resolution record (completions staged with the work).
    public func completionCommit(issue id: String, evidence: String?) async -> String? {
        if let evidence, !evidence.isEmpty { return evidence }
        let out = try? await run(["log", "-1", "--format=%H", "--", ".prep/issues/\(id)/resolution.md"])
        let hash = out?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return hash.isEmpty ? nil : hash
    }

    public func commit(_ hash: String) async throws -> CommitInfo {
        let sep = "\u{1f}"
        let head = try await run(["show", "-s", "--format=%H\(sep)%an\(sep)%aI\(sep)%s\(sep)%b", hash])
        let f = head.components(separatedBy: sep)
        let patch = try await run(["show", "--format=", "--patch", "--find-renames", "--no-color", "--unified=3", hash])
        return CommitInfo(
            hash: f.first ?? hash,
            author: f.count > 1 ? f[1] : "",
            date: f.count > 2 ? PrepDate.parse(f[2].trimmingCharacters(in: .whitespacesAndNewlines)) : nil,
            subject: f.count > 3 ? f[3] : "",
            body: f.count > 4 ? f[4].trimmingCharacters(in: .whitespacesAndNewlines) : "",
            files: Diff.parse(patch)
        )
    }

    /// The commit's page on GitHub, when origin is a GitHub remote.
    public func webURL(for hash: String) async -> URL? {
        guard let remote = try? await run(["remote", "get-url", "origin"]) else { return nil }
        var r = remote.trimmingCharacters(in: .whitespacesAndNewlines)
        if r.hasPrefix("git@github.com:") { r = "https://github.com/" + r.dropFirst("git@github.com:".count) }
        guard r.hasPrefix("https://github.com/") else { return nil }
        if r.hasSuffix(".git") { r = String(r.dropLast(4)) }
        return URL(string: "\(r)/commit/\(hash)")
    }
}

public struct CommitInfo: Sendable {
    public var hash: String
    public var author: String
    public var date: Date?
    public var subject: String
    public var body: String
    public var files: [FileDiff]
}

public struct FileDiff: Identifiable, Hashable, Sendable {
    public var path: String
    public var oldPath: String?
    public var status: Status
    public var binary = false
    public var hunks: [Hunk] = []
    public var id: String { path }

    public enum Status: String, Sendable { case added, deleted, modified, renamed }

    public var additions: Int { hunks.reduce(0) { $0 + $1.lines.filter { $0.kind == .added }.count } }
    public var deletions: Int { hunks.reduce(0) { $0 + $1.lines.filter { $0.kind == .removed }.count } }

    /// Where the file belongs in a review.
    public var area: Area {
        if path.hasPrefix(".prep/knowledge/") { return .knowledge }
        if path.hasPrefix(".prep/") { return .records }
        return .code
    }

    public enum Area: String, CaseIterable, Sendable { case code, knowledge, records }
}

public struct Hunk: Hashable, Sendable {
    public var header: String
    public var lines: [DiffLine]
}

public struct DiffLine: Hashable, Sendable {
    public enum Kind: Sendable { case context, added, removed }
    public var kind: Kind
    public var text: String
    public var old: Int?
    public var new: Int?
}

public enum Diff {
    /// Parses `git show --patch` output into files and hunks.
    public static func parse(_ patch: String) -> [FileDiff] {
        var files: [FileDiff] = []
        var file: FileDiff?
        var hunk: Hunk?
        var oldNo = 0, newNo = 0

        func flushHunk() {
            if let h = hunk { file?.hunks.append(h) }
            hunk = nil
        }
        func flushFile() {
            flushHunk()
            if let f = file { files.append(f) }
            file = nil
        }

        for line in patch.components(separatedBy: "\n") {
            if line.hasPrefix("diff --git ") {
                flushFile()
                // diff --git a/x b/y
                let parts = line.dropFirst("diff --git ".count).components(separatedBy: " b/")
                let path = parts.count > 1 ? parts.last! : String(line.dropFirst(13))
                file = FileDiff(path: path, status: .modified)
                continue
            }
            guard file != nil else { continue }
            if hunk == nil || line.hasPrefix("@@") {
                if line.hasPrefix("new file") { file?.status = .added; continue }
                if line.hasPrefix("deleted file") { file?.status = .deleted; continue }
                if line.hasPrefix("rename from ") { file?.oldPath = String(line.dropFirst(12)); file?.status = .renamed; continue }
                if line.hasPrefix("Binary files") { file?.binary = true; continue }
                if line.hasPrefix("+++ ") || line.hasPrefix("--- ") || line.hasPrefix("index ") || line.hasPrefix("similarity") || line.hasPrefix("rename to") { continue }
            }
            if line.hasPrefix("@@") {
                flushHunk()
                // @@ -a,b +c,d @@ context
                let nums = line.split(separator: " ")
                if nums.count >= 3 {
                    oldNo = Int(nums[1].dropFirst().split(separator: ",").first ?? "0") ?? 0
                    newNo = Int(nums[2].dropFirst().split(separator: ",").first ?? "0") ?? 0
                }
                hunk = Hunk(header: line, lines: [])
                continue
            }
            guard hunk != nil else { continue }
            if line.hasPrefix("+") {
                hunk?.lines.append(DiffLine(kind: .added, text: String(line.dropFirst()), old: nil, new: newNo))
                newNo += 1
            } else if line.hasPrefix("-") {
                hunk?.lines.append(DiffLine(kind: .removed, text: String(line.dropFirst()), old: oldNo, new: nil))
                oldNo += 1
            } else if line.hasPrefix(" ") {
                hunk?.lines.append(DiffLine(kind: .context, text: String(line.dropFirst()), old: oldNo, new: newNo))
                oldNo += 1
                newNo += 1
            }
        }
        flushFile()
        return files
    }
}
