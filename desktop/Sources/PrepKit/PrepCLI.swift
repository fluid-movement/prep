import Foundation

/// A failed prep command: prep's own error under --json, or the process
/// failing before it could print one.
public struct PrepError: Error, LocalizedError, Sendable {
    public var code: String
    public var message: String
    public var errorDescription: String? { "\(message) [\(code)]" }
}

/// Runs the prep binary. Every call passes --json and --root, so the app
/// only ever sees prep's JSON contract and never reads .prep itself.
public struct PrepCLI: Sendable {
    public var binary: URL
    public var root: String?
    /// The actor writes are recorded under. Reads pass no actor, so the
    /// app's reads stay out of the activity stream.
    public var actor: String

    public init(binary: URL, root: String?, actor: String = PrepCLI.defaultActor) {
        self.binary = binary
        self.root = root
        self.actor = actor
    }

    public static var defaultActor: String { "human:\(NSUserName())" }

    /// Finds the prep binary. Apps opened from Finder get a minimal PATH,
    /// so the usual install locations are tried before asking a login shell.
    public static func locate(environment: [String: String] = ProcessInfo.processInfo.environment) -> URL? {
        let fm = FileManager.default
        if let p = environment["PREP_BIN"], fm.isExecutableFile(atPath: p) {
            return URL(fileURLWithPath: p)
        }
        let home = fm.homeDirectoryForCurrentUser.path
        var candidates = (environment["PATH"] ?? "").split(separator: ":").map { "\($0)/prep" }
        candidates += ["\(home)/.local/bin/prep", "/opt/homebrew/bin/prep", "/usr/local/bin/prep", "\(home)/go/bin/prep"]
        for c in candidates where fm.isExecutableFile(atPath: c) {
            return URL(fileURLWithPath: c)
        }
        if let out = try? runSync(URL(fileURLWithPath: "/bin/zsh"), ["-lc", "command -v prep"], stdin: nil).out,
           let p = String(data: out, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines),
           !p.isEmpty, fm.isExecutableFile(atPath: p) {
            return URL(fileURLWithPath: p)
        }
        return nil
    }

    /// The project root containing path: the nearest directory with a
    /// .prep directory, found the way prep finds it. Reads no records.
    public static func projectRoot(from path: String) -> String? {
        var url = URL(fileURLWithPath: path).standardizedFileURL
        let fm = FileManager.default
        while true {
            var isDir: ObjCBool = false
            if fm.fileExists(atPath: url.appendingPathComponent(".prep").path, isDirectory: &isDir), isDir.boolValue {
                return url.path
            }
            let parent = url.deletingLastPathComponent()
            if parent.path == url.path { return nil }
            url = parent
        }
    }

    // MARK: Reads

    public func list(_ query: [String] = []) async throws -> [IssueSummary] {
        try await decode(IssueList.self, ["list"] + query).issues
    }

    public func show(_ id: String) async throws -> IssueDetail {
        try await decode(IssueDetail.self, ["show", id])
    }

    public func guide(_ id: String) async throws -> Guide {
        try await decode(Guide.self, ["guide", id])
    }

    public func knowledge() async throws -> [KnowledgeEntry] {
        try await decode(KnowledgeList.self, ["knowledge", "list"]).entries
    }

    public func knowledgeEntry(_ path: String) async throws -> KnowledgeEntry? {
        try await decode(KnowledgeList.self, ["knowledge", "show", path]).entries.first
    }

    public func activity(max: Int = 200) async throws -> ActivityFeed {
        try await decode(ActivityFeed.self, ["activity", "--max", String(max)])
    }

    public func views() async throws -> [String: String] {
        try await decode(Views.self, ["views"]).views
    }

    public func flags() async throws -> [FlagInfo] {
        try await decode(FlagList.self, ["flags"]).flags
    }

    public func projects() async throws -> [ProjectRef] {
        try await decode(ProjectList.self, ["projects"]).projects
    }

    // MARK: Writes

    /// Runs a write command as the app's actor; returns what prep printed.
    @discardableResult
    public func write(_ args: [String], stdin: String? = nil) async throws -> Data {
        try await run(["--by", actor] + args, stdin: stdin)
    }

    /// Creates an issue and returns its ID.
    public func create(title: String, kind: String, parent: String? = nil, priority: String? = nil,
                       tags: [String] = [], body: String = "") async throws -> String {
        var args = ["new", "--title", title, "--kind", kind]
        if let parent { args += ["--parent", parent] }
        if let priority, priority != "medium" { args += ["--priority", priority] }
        for t in tags { args += ["--tag", t] }
        args += ["--body-file", "-"]
        let data = try await write(args, stdin: body)
        struct Created: Decodable { var id: String }
        return try JSONDecoder.prep.decode(Created.self, from: data).id
    }

    // MARK: Plumbing

    public func decode<T: Decodable>(_ type: T.Type, _ args: [String]) async throws -> T {
        let data = try await run(args)
        do {
            return try JSONDecoder.prep.decode(T.self, from: data)
        } catch {
            throw PrepError(code: "E_DECODE", message: "prep \(args.first ?? ""): \(error)")
        }
    }

    public func run(_ args: [String], stdin: String? = nil) async throws -> Data {
        // Global flags go first: after a "--" prep takes everything literally.
        var full = ["--json"]
        if let root { full += ["--root", root] }
        full += args
        let binary = self.binary
        let input = stdin.map { Data($0.utf8) }
        return try await Task.detached {
            let r = try PrepCLI.runSync(binary, full, stdin: input)
            if r.status != 0 {
                if let body = try? JSONDecoder.prep.decode(CLIErrorBody.self, from: r.out), let e = body.error {
                    throw PrepError(code: e.code, message: e.message)
                }
                let msg = String(data: r.err.isEmpty ? r.out : r.err, encoding: .utf8) ?? ""
                throw PrepError(code: "E_EXIT", message: msg.trimmingCharacters(in: .whitespacesAndNewlines))
            }
            return r.out
        }.value
    }

    /// Runs a process to completion, reading stdout and stderr concurrently
    /// so neither pipe can fill up and stall it.
    static func runSync(_ binary: URL, _ args: [String], stdin: Data?) throws -> (status: Int32, out: Data, err: Data) {
        let p = Process()
        p.executableURL = binary
        p.arguments = args
        var env = ProcessInfo.processInfo.environment
        env["PREP_ACTOR"] = nil
        env["PREP_SESSION"] = nil
        p.environment = env
        let out = Pipe(), err = Pipe()
        p.standardOutput = out
        p.standardError = err
        let inPipe = stdin.map { _ in Pipe() }
        p.standardInput = inPipe ?? FileHandle.nullDevice
        try p.run()
        if let inPipe, let stdin {
            inPipe.fileHandleForWriting.write(stdin)
            try? inPipe.fileHandleForWriting.close()
        }
        var errData = Data()
        let group = DispatchGroup()
        group.enter()
        DispatchQueue.global().async {
            errData = err.fileHandleForReading.readDataToEndOfFile()
            group.leave()
        }
        let outData = out.fileHandleForReading.readDataToEndOfFile()
        group.wait()
        p.waitUntilExit()
        return (p.terminationStatus, outData, errData)
    }
}
