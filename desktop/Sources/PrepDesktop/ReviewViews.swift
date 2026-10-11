import SwiftUI
import PrepKit

/// What a completed issue changed: its commit's code, the knowledge
/// entries it updated, and prep's own records folded away — so "done"
/// can be checked, not just believed.
struct ChangesCard: View {
    var model: ProjectModel
    var detail: IssueDetail

    @State private var commit: CommitInfo?
    @State private var web: URL?
    @State private var failure: String?
    @State private var loading = true

    var body: some View {
        Card("Changes", symbol: "plusminus.circle", accessory: { accessory }) {
            if loading {
                HStack { ProgressView().controlSize(.small); Text("Reading the commit…").foregroundStyle(.secondary) }
            } else if let commit {
                VStack(alignment: .leading, spacing: 12) {
                    header(commit)
                    let groups = Dictionary(grouping: commit.files, by: \.area)
                    if let code = groups[.code], !code.isEmpty {
                        FileGroup(title: "Code", symbol: "chevron.left.forwardslash.chevron.right", files: code,
                                  expandFirst: code.count <= 4)
                    }
                    if let docs = groups[.knowledge], !docs.isEmpty {
                        FileGroup(title: "Knowledge", symbol: "book", files: docs, expandFirst: true, onOpen: { path in
                            model.openLink("/" + path.dropFirst(".prep/knowledge/".count))
                        })
                    }
                    if let records = groups[.records], !records.isEmpty {
                        DisclosureGroup {
                            FileGroup(title: nil, symbol: "", files: records, expandFirst: false)
                        } label: {
                            Text("prep records · \(records.count) files").font(.callout).foregroundStyle(.secondary)
                        }
                    }
                    if commit.files.isEmpty { Text("The commit changed no files.").foregroundStyle(.secondary) }
                }
            } else {
                Text(failure ?? "No commit found for this issue.").foregroundStyle(.secondary)
            }
        }
        .task(id: detail.issue.id) { await load() }
    }

    @ViewBuilder
    var accessory: some View {
        if let commit {
            HStack(spacing: 10) {
                let add = commit.files.filter { $0.area != .records }.reduce(0) { $0 + $1.additions }
                let del = commit.files.filter { $0.area != .records }.reduce(0) { $0 + $1.deletions }
                Text("+\(add)").foregroundStyle(.green).font(.caption.monospacedDigit().weight(.semibold))
                Text("−\(del)").foregroundStyle(.red).font(.caption.monospacedDigit().weight(.semibold))
                if let web {
                    Button { NSWorkspace.shared.open(web) } label: { Label("GitHub", systemImage: "arrow.up.right.square") }
                        .buttonStyle(.borderless).font(.callout)
                        .help(web.absoluteString)
                }
            }
        }
    }

    func header(_ c: CommitInfo) -> some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "point.topleft.down.to.point.bottomright.curvepath")
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 2) {
                Text(c.subject).font(.callout.weight(.semibold)).textSelection(.enabled)
                HStack(spacing: 6) {
                    Button { model.copy(c.hash) } label: {
                        Text(String(c.hash.prefix(10))).font(.caption.monospaced())
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(.secondary)
                    .help("Copy \(c.hash)")
                    Text("·").foregroundStyle(.tertiary)
                    Text(c.author).font(.caption).foregroundStyle(.secondary)
                    if let d = c.date { Text(d.relative).font(.caption).foregroundStyle(.tertiary).help(d.formatted()) }
                }
            }
        }
    }

    func load() async {
        loading = true
        defer { loading = false }
        commit = nil
        guard let root = model.project?.root else { return }
        let git = Git(root: root)
        guard let hash = await git.completionCommit(issue: detail.issue.id, evidence: detail.issue.resolution?.evidence) else {
            failure = "No commit carries this completion yet."
            return
        }
        do {
            commit = try await git.commit(hash)
            web = await git.webURL(for: commit?.hash ?? hash)
        } catch {
            failure = "Could not read commit \(hash.prefix(10)): \(error.localizedDescription)"
        }
    }
}

struct FileGroup: View {
    var title: String?
    var symbol: String
    var files: [FileDiff]
    var expandFirst: Bool
    var onOpen: ((String) -> Void)? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            if let title {
                Label("\(title) · \(files.count)", systemImage: symbol)
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
            }
            ForEach(Array(files.enumerated()), id: \.element.id) { k, f in
                FileDiffRow(file: f, expanded: expandFirst && k < 4, onOpen: onOpen)
            }
        }
    }
}

struct FileDiffRow: View {
    var file: FileDiff
    @State var expanded: Bool
    var onOpen: ((String) -> Void)?
    @State private var showAll = false

    static let limit = 300

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Button { withAnimation(.easeOut(duration: 0.15)) { expanded.toggle() } } label: {
                HStack(spacing: 8) {
                    Image(systemName: "chevron.right")
                        .font(.caption2.weight(.bold))
                        .rotationEffect(.degrees(expanded ? 90 : 0))
                        .foregroundStyle(.tertiary)
                    statusIcon
                    pathText
                    Spacer(minLength: 8)
                    if file.binary {
                        Text("binary").font(.caption).foregroundStyle(.secondary)
                    } else {
                        ChangeBar(additions: file.additions, deletions: file.deletions)
                    }
                    if let onOpen {
                        Button { onOpen(file.path) } label: { Image(systemName: "book") }
                            .buttonStyle(.borderless)
                            .help("Open the entry")
                    }
                }
                .padding(.horizontal, 6)
                .padding(.vertical, 5)
            }
            .buttonStyle(HoverButtonStyle(radius: 6))
            if expanded, !file.binary {
                DiffLinesView(hunks: file.hunks, limit: showAll ? Int.max : Self.limit)
                let total = file.hunks.reduce(0) { $0 + $1.lines.count }
                if total > Self.limit, !showAll {
                    Button("Show all \(total) lines") { showAll = true }
                        .buttonStyle(.link)
                        .font(.caption)
                        .padding(.leading, 6)
                }
            }
        }
    }

    var statusIcon: some View {
        let (symbol, color): (String, Color) = switch file.status {
        case .added: ("plus.circle.fill", .green)
        case .deleted: ("minus.circle.fill", .red)
        case .renamed: ("arrow.right.circle.fill", .blue)
        case .modified: ("pencil.circle.fill", .orange)
        }
        return Image(systemName: symbol).foregroundStyle(color).help(file.status.rawValue.capitalized)
    }

    var pathText: some View {
        let url = URL(fileURLWithPath: file.path)
        let dir = url.deletingLastPathComponent().relativePath
        return HStack(spacing: 0) {
            if let old = file.oldPath { Text(old + " → ").foregroundStyle(.secondary) }
            if dir != "." { Text(dir + "/").foregroundStyle(.secondary) }
            Text(url.lastPathComponent).fontWeight(.medium)
        }
        .font(.callout.monospaced())
        .lineLimit(1)
        .truncationMode(.head)
    }
}

/// Additions and deletions as numbers and five little blocks, like GitHub.
struct ChangeBar: View {
    var additions: Int
    var deletions: Int
    var body: some View {
        let total = max(1, additions + deletions)
        let green = Int((Double(additions) / Double(total) * 5).rounded())
        HStack(spacing: 6) {
            Text("+\(additions)").foregroundStyle(.green)
            Text("−\(deletions)").foregroundStyle(.red)
            HStack(spacing: 1.5) {
                ForEach(0..<5, id: \.self) { k in
                    RoundedRectangle(cornerRadius: 1.5)
                        .fill(k < green ? Color.green : additions + deletions == 0 ? Color.secondary.opacity(0.3) : Color.red)
                        .frame(width: 7, height: 7)
                }
            }
        }
        .font(.caption.monospacedDigit())
    }
}

struct DiffLinesView: View {
    var hunks: [Hunk]
    var limit: Int

    var body: some View {
        let rows = flattened
        ScrollView(.horizontal, showsIndicators: false) {
            VStack(alignment: .leading, spacing: 0) {
                ForEach(Array(rows.prefix(limit).enumerated()), id: \.offset) { _, row in
                    switch row {
                    case .header(let h):
                        Text(h)
                            .font(.caption.monospaced())
                            .foregroundStyle(.blue)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 3)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .background(Color.blue.opacity(0.07))
                    case .line(let l):
                        HStack(spacing: 0) {
                            Text(l.old.map(String.init) ?? "").frame(width: 38, alignment: .trailing)
                            Text(l.new.map(String.init) ?? "").frame(width: 38, alignment: .trailing)
                            Text(l.kind == .added ? " + " : l.kind == .removed ? " − " : "   ")
                                .foregroundStyle(l.kind == .added ? .green : l.kind == .removed ? .red : .secondary)
                            Text(l.text.isEmpty ? " " : l.text)
                                .foregroundStyle(.primary)
                                .fixedSize()
                        }
                        .font(.system(size: 11.5, design: .monospaced))
                        .foregroundStyle(.tertiary)
                        .padding(.vertical, 0.5)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(l.kind == .added ? Color.green.opacity(0.11) : l.kind == .removed ? Color.red.opacity(0.11) : .clear)
                    }
                }
            }
            .textSelection(.enabled)
        }
        .background(Color(nsColor: .textBackgroundColor).opacity(0.5))
        .clipShape(RoundedRectangle(cornerRadius: 6))
        .overlay(RoundedRectangle(cornerRadius: 6).strokeBorder(.separator.opacity(0.6)))
        .padding(.vertical, 4)
    }

    enum Row { case header(String), line(DiffLine) }

    var flattened: [Row] {
        hunks.flatMap { h in [Row.header(h.header)] + h.lines.map(Row.line) }
    }
}
