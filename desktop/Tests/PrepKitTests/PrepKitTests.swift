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
        try await cli.write(["new", "--title", "Parent", "--kind", "manual"])
        let parent = try #require(try await cli.list().first)
        try await cli.write(["new", "--title", "Child", "--kind", "code", "--parent", parent.id, "--body", "Do it."])
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
