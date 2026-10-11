import Foundation
import Testing
@testable import PrepKit

@Suite struct ParsingTests {
    @Test func queryArguments() {
        #expect(Query.arguments("--state open,ready release notes") == ["--state", "open,ready", "--text", "release notes"])
        #expect(Query.arguments("--stale --tag ui") == ["--stale", "--tag", "ui"])
        #expect(Query.arguments("--state=ready") == ["--state=ready"])
        #expect(Query.arguments("") == [])
    }

    @Test func treeNestsChildrenUnderParentsInTheList() {
        let list = [
            IssueSummary(id: "A", title: "a"),
            IssueSummary(id: "B", title: "b", parent: "A"),
            IssueSummary(id: "C", title: "c", parent: "missing"),
            IssueSummary(id: "D", title: "d", parent: "B"),
        ]
        let tree = IssueTree.build(list)
        #expect(tree.map(\.id) == ["A", "C"])
        #expect(tree[0].children?.map(\.id) == ["B"])
        #expect(tree[0].children?[0].children?.map(\.id) == ["D"])
        #expect(tree[1].children == nil)
    }

    @Test func markdownBlocks() {
        let blocks = Markdown.blocks("""
        # Title

        First line
        continues here.

        - [ ] open item
        - [x] done item
          wrapped
        1. one

        ```
        code
        ```
        """)
        #expect(blocks == [
            .heading(level: 1, text: "Title"),
            .paragraph("First line continues here."),
            .bullet(text: "open item", depth: 0, checked: false),
            .bullet(text: "done item wrapped", depth: 0, checked: true),
            .numbered(number: "1", text: "one", depth: 0),
            .code("code"),
        ])
    }

    @Test func filterArguments() {
        var f = IssueFilter(view: "Unresolved")
        #expect(f.arguments == ["--view", "Unresolved"])
        f.states = ["ready", "open"]
        f.priorities = ["low", "high"]
        f.tags = ["ui"]
        f.blocked = true
        f.text = " release "
        #expect(f.arguments == ["--view", "Unresolved", "--state", "open,ready", "--priority", "high,low",
                                "--tag", "ui", "--blocked", "--text", "release"])
        #expect(f.activeCount == 4)
    }

    @Test func workLogEntries() {
        let log = WorkLog.parse("""
        - 2026-10-08T08:33:13Z claude-code/opus-5.5: Did a thing: well.
        - 2026-10-09T10:00:00Z human:ana: Second
          continued
        """)
        #expect(log.count == 2)
        #expect(log[0].actor == "claude-code/opus-5.5")
        #expect(log[0].text == "Did a thing: well.")
        #expect(log[1].text == "Second\ncontinued")
    }

    @Test func ulidTime() {
        let d = ULID.date("01M4JY2S98AKAQ109YACKV5VNV")
        #expect(d != nil)
        #expect(d! > PrepDate.parse("2026-01-01T00:00:00Z")!)
        #expect(d! < PrepDate.parse("2027-01-01T00:00:00Z")!)
    }

    @Test func inboxSortsWhatWaitsOnThePerson() {
        let t0 = Date(timeIntervalSince1970: 1_000_000)
        let issues = [
            IssueSummary(id: "A", title: "open", priority: "medium"),
            IssueSummary(id: "B", title: "questions", openQuestions: true),
            IssueSummary(id: "C", title: "defined, enriched", state: "defined"),
            IssueSummary(id: "D", title: "defined, not enriched", state: "defined"),
            IssueSummary(id: "E", title: "stale", state: "ready", stale: true),
            IssueSummary(id: "F", title: "done by agent", state: "done", resolvedAt: t0.addingTimeInterval(60), resolvedBy: "claude-code/x"),
            IssueSummary(id: "G", title: "done before", state: "done", resolvedAt: t0.addingTimeInterval(-60), resolvedBy: "claude-code/x"),
            IssueSummary(id: "H", title: "done by me", state: "done", resolvedAt: t0.addingTimeInterval(60), resolvedBy: "human:me"),
            IssueSummary(id: "I", title: "release", kind: "manual"),
            IssueSummary(id: "J", title: "high", priority: "high"),
            IssueSummary(id: "K", title: "reviewed", state: "done", resolvedAt: t0.addingTimeInterval(60), resolvedBy: "claude-code/x"),
        ]
        let inbox = Inbox.build(issues, readyAllowed: ["C"], since: t0, reviewed: ["K"])
        #expect(inbox.map { "\($0.kind.rawValue):\($0.issue.id)" } == [
            "review:F", "acknowledge:E", "signOff:C", "questions:B", "define:J", "define:A",
        ])
    }

    @Test func diffParsesFilesHunksAndAreas() {
        let patch = """
        diff --git a/internal/x.go b/internal/x.go
        index 1..2 100644
        --- a/internal/x.go
        +++ b/internal/x.go
        @@ -1,3 +1,4 @@ func main
         keep
        -old
        +new
        +++plus
         tail
        diff --git a/.prep/knowledge/a.md b/.prep/knowledge/a.md
        new file mode 100644
        --- /dev/null
        +++ b/.prep/knowledge/a.md
        @@ -0,0 +1 @@
        +# A
        diff --git a/img.png b/img.png
        Binary files a/img.png and b/img.png differ
        """
        let files = Diff.parse(patch)
        #expect(files.map(\.path) == ["internal/x.go", ".prep/knowledge/a.md", "img.png"])
        #expect(files[0].additions == 2 && files[0].deletions == 1)
        #expect(files[0].hunks[0].lines.map(\.new) == [1, nil, 2, 3, 4])
        #expect(files[0].hunks[0].lines[3].text == "++plus")
        #expect(files[1].status == .added && files[1].area == .knowledge)
        #expect(files[2].binary && files[2].area == .code)
    }

    @Test func agentLaunchScriptQuotes() throws {
        let l = AgentLaunch(root: "/tmp/it's here", issue: "01ABC", title: "Don't break")
        #expect(l.prompt == "Work on prep issue 01ABC (Don't break). Run `prep guide 01ABC` and follow it.")
        #expect(l.script.contains("cd '/tmp/it'\\''s here' || exit 1"))
        #expect(l.script.contains("exec claude 'Work on prep issue 01ABC (Don'\\''t break)."))
        // The quoting survives a real shell.
        let r = try PrepCLI.runSync(URL(fileURLWithPath: "/bin/sh"), ["-c", "printf %s " + AgentLaunch.quote(l.prompt)], stdin: nil)
        #expect(String(data: r.out, encoding: .utf8) == l.prompt)
    }

    @Test func shellSplit() {
        #expect(Shell.split(#"--reason "not needed" --docs /a.md"#) == ["--reason", "not needed", "--docs", "/a.md"])
        #expect(Shell.split("--note ''") == ["--note", ""])
    }

    @Test func projectRootIsTheNearestDirectoryWithPrep() throws {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("root-\(UUID().uuidString)").standardizedFileURL
        let deep = root.appendingPathComponent("a/b")
        try FileManager.default.createDirectory(at: deep, withIntermediateDirectories: true)
        try FileManager.default.createDirectory(at: root.appendingPathComponent(".prep"), withIntermediateDirectories: true)
        #expect(PrepCLI.projectRoot(from: deep.path) == root.path)
        #expect(PrepCLI.projectRoot(from: "/") == nil)
    }

    @Test func datesWithAndWithoutFractions() {
        #expect(PrepDate.parse("2026-10-10T12:15:46.546246Z") != nil)
        #expect(PrepDate.parse("2026-10-10T12:15:46Z") != nil)
    }
}

/// Runs the real prep binary against a fresh project, so the models are
/// checked against the JSON prep prints now. Skipped when prep is missing.
@Suite struct CLITests {
    let cli: PrepCLI?
    let dir: URL

    init() throws {
        dir = FileManager.default.temporaryDirectory.appendingPathComponent("prepkit-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        if let bin = PrepCLI.locate() {
            cli = PrepCLI(binary: bin, root: dir.path, actor: "test/prepkit")
        } else {
            cli = nil
        }
    }

    @Test func roundTripThroughTheCLI() async throws {
        guard let cli else { return }
        setenv("XDG_STATE_HOME", dir.appendingPathComponent("state").path, 1)
        try await cli.write(["init", "--no-bootstrap"])
        try await cli.write(["knowledge", "new", "/overview.md", "--type", "overview", "--title", "Overview",
                             "--description", "The test project.", "--body", "# Overview\n\nTest."])
        let parentID = try await cli.create(title: "Parent", kind: "manual", tags: ["x"], body: "The parent.")
        let parent = try #require(try await cli.list().first)
        #expect(parent.id == parentID)
        _ = try await cli.create(title: "Child", kind: "code", parent: parent.id, priority: "high", body: "Do it.")
        #expect(try await cli.flags().contains { $0.name == "state" })
        let issues = try await cli.list()
        #expect(issues.count == 2)
        let child = try #require(issues.first { $0.title == "Child" })
        #expect(child.parent == parent.id)

        try await cli.write(["criterion", child.id, "--add", "works"])
        try await cli.write(["log", child.id, "--", "started"])
        try await cli.write(["decide", child.id, "--title", "Pick one", "--body", "Because."])
        let d = try await cli.show(child.id)
        #expect(d.issue.criteria?.first?.text == "works")
        #expect(d.issue.criteria?.first?.checked != true)
        #expect(d.issue.decisions?.first?.title == "Pick one")
        #expect(d.history?.contains("started") == true)

        let g = try await cli.guide(child.id)
        #expect(g.transitions?.contains { $0.op == "define" && $0.allowed } == true)

        try await cli.write(["define", child.id])
        #expect(try await cli.show(child.id).state == "defined")

        let entries = try await cli.knowledge()
        #expect(entries.map(\.path) == ["/overview.md"])
        #expect(try await cli.knowledgeEntry("/overview.md")?.text?.contains("Test.") == true)
        _ = try await cli.views()
        let feed = try await cli.activity()
        #expect(feed.events?.isEmpty == false)

        do {
            _ = try await cli.show("nope")
            Issue.record("show of a missing issue succeeded")
        } catch let e as PrepError {
            #expect(e.code.hasPrefix("E_"))
        }
    }

    @Test func watcherReportsWrites() async throws {
        guard let cli else { return }
        setenv("XDG_STATE_HOME", dir.appendingPathComponent("state").path, 1)
        try await cli.write(["init", "--no-bootstrap"])
        let changed = AsyncStream<Void>.makeStream()
        let w = try Watcher(binary: cli.binary, root: dir.path, debounce: 0.05) { changed.continuation.yield() }
        defer { w.stop() }
        try await Task.sleep(for: .milliseconds(300))
        try await cli.write(["knowledge", "new", "/overview.md", "--type", "overview", "--title", "Overview",
                             "--description", "The test project.", "--body", "# Overview\n\nTest."])
        let got = await withTaskGroup(of: Bool.self) { group in
            group.addTask { for await _ in changed.stream { return true }; return false }
            group.addTask { try? await Task.sleep(for: .seconds(5)); return false }
            let first = await group.next() ?? false
            group.cancelAll()
            changed.continuation.finish()
            return first
        }
        #expect(got, "no change reported within 5 s")
    }
}
