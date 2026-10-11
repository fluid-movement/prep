import SwiftUI
import PrepKit

/// An issue as a document: where it sits, where it is in its lifecycle and
/// what to do next, how it relates to other work, then its records.
struct IssueDetailView: View {
    @Bindable var model: ProjectModel
    var detail: IssueDetail
    @State private var pending: PrepKit.Transition?

    var issue: IssueRecord { detail.issue }
    var resolved: Bool { issue.resolution != nil }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                if let item = model.currentInboxItem { InboxBanner(model: model, item: item) }
                else if model.tab == .issues, !model.isListed(issue.id) { NotInListBanner(model: model) }
                Breadcrumbs(model: model, issue: issue)
                header
                if let r = issue.resolution { ResolutionBanner(model: model, resolution: r) }
                if issue.resolution?.outcome == "done", issue.kind != "manual" { ChangesCard(model: model, detail: detail) }
                if issue.resolution == nil {
                    LifecycleCard(model: model, detail: detail) { pending = $0 }
                }
                if !(issue.dependsOn ?? []).isEmpty || !(detail.blocks ?? []).isEmpty {
                    DependencyMap(model: model, detail: detail)
                }
                if !(detail.children ?? []).isEmpty || issue.kind == "manual" {
                    ChildrenCard(model: model, detail: detail)
                }
                RequirementCard(model: model, detail: detail)
                AcceptanceCard(model: model, detail: detail)
                if !(detail.context ?? "").isEmpty || (!resolved && detail.state != "open") {
                    ContextCard(model: model, detail: detail)
                }
                if !(issue.decisions ?? []).isEmpty || (!resolved && issue.kind == "decision") {
                    DecisionsCard(model: model, decisions: issue.decisions ?? [], editable: !resolved)
                }
                if !(detail.findings ?? "").isEmpty || (!resolved && issue.kind == "research") {
                    FindingsCard(model: model, detail: detail)
                }
                WorkLogCard(model: model, detail: detail)
            }
            .padding(.horizontal, 28)
            .padding(.vertical, 20)
            .frame(maxWidth: 920, alignment: .leading)
            .frame(maxWidth: .infinity)
        }
        .sheet(item: $pending) { t in
            TransitionSheet(transition: t, issueTitle: issue.title, entries: model.entries) { args in
                pending = nil
                if let args { Task { await model.write(args) } }
            }
        }
    }

    var header: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(issue.title)
                .font(.system(size: 26, weight: .bold))
                .fixedSize(horizontal: false, vertical: true)
                .onTapGesture(count: 2) { if !resolved { model.editing = .title } }
                .help(resolved ? issue.title : "Double-click to rename")
            HStack(spacing: 8) {
                StateBadge(state: detail.state)
                Label(issue.kind.capitalized, systemImage: Style.symbol(kind: issue.kind))
                    .font(.callout).foregroundStyle(.secondary)
                if let p = detail.priority, p != "medium" {
                    HStack(spacing: 3) {
                        PriorityIcon(priority: p)
                        Text(p.capitalized)
                    }
                    .font(.callout).foregroundStyle(Style.priorityColor(p))
                }
                if detail.stale { Chip(text: "Stale", symbol: "exclamationmark.arrow.circlepath", color: .yellow) }
                if detail.blocked { Chip(text: "Blocked", symbol: "lock.fill", color: .orange) }
                if detail.actionable { Chip(text: "Actionable", symbol: "bolt.fill", color: .green) }
                ForEach(issue.tags ?? [], id: \.self) { TagChip(tag: $0) }
                Spacer(minLength: 0)
                if let created = ULID.date(issue.id) {
                    Text("Created \(created.relative)")
                        .font(.caption).foregroundStyle(.tertiary)
                        .help(created.formatted(date: .long, time: .shortened))
                }
            }
        }
    }
}

// MARK: - Where it sits

struct NotInListBanner: View {
    var model: ProjectModel
    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "eye.slash").foregroundStyle(.secondary)
            Text("This issue is hidden by the current filter.").foregroundStyle(.secondary)
            Spacer()
            Button("Show in List") { model.revealInList() }.controlSize(.small)
        }
        .font(.callout)
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(.quaternary.opacity(0.4), in: RoundedRectangle(cornerRadius: 8))
    }
}

/// The parents above the issue, and stepping through its siblings.
struct Breadcrumbs: View {
    var model: ProjectModel
    var issue: IssueRecord

    var body: some View {
        let chain = model.ancestors(of: issue.id)
        HStack(spacing: 4) {
            ForEach(chain) { a in
                Button { model.go(to: a.id) } label: {
                    HStack(spacing: 4) {
                        StateIcon(state: a.state, size: 10)
                        Text(a.title).lineLimit(1)
                    }
                    .padding(.horizontal, 6)
                    .padding(.vertical, 3)
                }
                .buttonStyle(HoverButtonStyle(radius: 6))
                .help("Go to \(a.title)")
                Image(systemName: "chevron.right").font(.system(size: 9, weight: .semibold)).foregroundStyle(.tertiary)
            }
            Text(String(issue.id.suffix(6)))
                .font(.system(.callout, design: .monospaced))
                .foregroundStyle(.secondary)
                .textSelection(.enabled)
                .help(issue.id)
            Spacer(minLength: 8)
            siblings
        }
        .font(.callout)
        .foregroundStyle(.secondary)
    }

    @ViewBuilder
    var siblings: some View {
        if let parent = issue.parent {
            let sibs = model.children(of: parent)
            if sibs.count > 1, let k = sibs.firstIndex(where: { $0.id == issue.id }) {
                HStack(spacing: 2) {
                    Button { model.go(to: sibs[k - 1].id) } label: { Image(systemName: "chevron.up") }
                        .disabled(k == 0)
                        .help(k > 0 ? "Previous sibling: \(sibs[k - 1].title)" : "")
                    Text("\(k + 1) of \(sibs.count)").font(.caption.monospacedDigit()).frame(minWidth: 44)
                    Button { model.go(to: sibs[k + 1].id) } label: { Image(systemName: "chevron.down") }
                        .disabled(k == sibs.count - 1)
                        .help(k < sibs.count - 1 ? "Next sibling: \(sibs[k + 1].title)" : "")
                }
                .buttonStyle(.borderless)
            }
        }
    }
}

struct ResolutionBanner: View {
    var model: ProjectModel
    var resolution: Resolution

    var done: Bool { resolution.outcome == "done" }

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: done ? "checkmark.seal.fill" : "xmark.octagon.fill")
                .font(.title2)
                .foregroundStyle(done ? .green : .secondary)
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 6) {
                    Text(done ? "Done" : "Dropped").font(.headline)
                    if let by = resolution.by { Text("by \(by)").foregroundStyle(.secondary) }
                    if let at = resolution.at.flatMap(PrepDate.parse) {
                        Text(at.relative).foregroundStyle(.secondary).help(at.formatted())
                    }
                }
                if let r = resolution.reason { Text(Markdown.inline(r)).fixedSize(horizontal: false, vertical: true) }
                if let n = resolution.note { Text(Markdown.inline(n)).foregroundStyle(.secondary) }
                FlowLayout(spacing: 6) {
                    if let e = resolution.evidence {
                        Button { model.copy(e) } label: {
                            Chip(text: String(e.prefix(8)), symbol: "point.topleft.down.to.point.bottomright.curvepath", color: .secondary)
                        }
                        .buttonStyle(.plain)
                        .help("Commit \(e) — click to copy")
                    }
                    ForEach(resolution.documentation?.entries ?? [], id: \.self) { path in
                        Button { model.openLink(path) } label: { Chip(text: path, symbol: "book", color: .blue) }
                            .buttonStyle(.plain)
                    }
                    if let none = resolution.documentation?.noImpact {
                        Text("No docs impact: \(none)").font(.caption).foregroundStyle(.secondary)
                    }
                }
            }
            Spacer(minLength: 0)
        }
        .padding(14)
        .background((done ? Color.green : Color.gray).opacity(0.09), in: RoundedRectangle(cornerRadius: Style.cardRadius))
        .overlay(RoundedRectangle(cornerRadius: Style.cardRadius).strokeBorder((done ? Color.green : Color.gray).opacity(0.25)))
    }
}

// MARK: - Lifecycle

/// The lifecycle as steps, with the one action that moves the issue on and
/// what still stands in its way.
struct LifecycleCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    var confirm: (PrepKit.Transition) -> Void

    static let forward = ["open": "define", "defined": "ready", "ready": "claim", "in-progress": "complete"]
    static let verbs = [
        "define": ("Define", "Write the requirement baseline"),
        "ack": ("Acknowledge", "Accept the requirement change; enrichment stays valid"),
        "ready": ("Mark Ready", "Sign off the enrichment against the newest baseline"),
        "claim": ("Start Work", "Claim the issue"),
        "complete": ("Complete", "Finish with evidence and documentation"),
        "release": ("Release", "Stop work without finishing"),
        "drop": ("Drop", "Drop the issue"),
    ]

    var transitions: [PrepKit.Transition] { model.guide?.transitions ?? [] }
    var nextOp: String? { detail.stale ? "ack" : Self.forward[detail.state] }
    var next: PrepKit.Transition? { transitions.first { $0.op == nextOp } }
    var others: [PrepKit.Transition] { transitions.filter { $0.op != nextOp && $0.allowed } }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Stepper(state: detail.state)
            Divider()
            HStack(alignment: .top, spacing: 14) {
                VStack(alignment: .leading, spacing: 6) {
                    if let t = next, let v = Self.verbs[t.op] {
                        Text(t.allowed ? "Next: \(v.1.lowercasedFirst)" : "Before you can \(v.0.lowercased()):")
                            .font(.callout.weight(.medium))
                        if !t.allowed {
                            ForEach(t.unmet ?? [], id: \.self) { g in
                                HStack(spacing: 8) {
                                    Label {
                                        Text(Gates.text(g)).fixedSize(horizontal: false, vertical: true)
                                    } icon: {
                                        Image(systemName: "circle").foregroundStyle(.orange)
                                    }
                                    .help(g.message)
                                    if let (label, target) = Gates.fix(g) {
                                        Button(label) { model.editing = target }
                                            .buttonStyle(.link)
                                    }
                                }
                                .font(.callout)
                            }
                        }
                    } else if let step = model.guide?.step {
                        Text("Step: \(step)").font(.callout.weight(.medium))
                    }
                    if let first = model.guide?.instructions?.first {
                        Text(first).font(.caption).foregroundStyle(.secondary).fixedSize(horizontal: false, vertical: true)
                    }
                }
                Spacer(minLength: 8)
                HStack(spacing: 8) {
                    Button { model.startAgent(detail.issue.id) } label: {
                        Label("Start Agent", systemImage: "sparkles")
                    }
                    .controlSize(.large)
                    .help("Open \(TerminalApp.preferred.rawValue) in the project with the agent on this issue (⌥⌘A)")
                    if !others.isEmpty {
                        Menu {
                            ForEach(others) { t in
                                Button((Self.verbs[t.op]?.0 ?? t.op.capitalized) + "…") { confirm(t) }
                            }
                        } label: {
                            Image(systemName: "ellipsis")
                        }
                        .menuStyle(.borderlessButton)
                        .menuIndicator(.hidden)
                        .fixedSize()
                        .help("Other transitions")
                    }
                    if let t = next, let v = Self.verbs[t.op] {
                        Button {
                            if (t.needs ?? []).isEmpty {
                                Task { await model.write([t.op, detail.issue.id]) }
                            } else {
                                confirm(t)
                            }
                        } label: {
                            Text((t.needs ?? []).isEmpty ? v.0 : v.0 + "…").frame(minWidth: 80)
                        }
                        .buttonStyle(.borderedProminent)
                        .controlSize(.large)
                        .disabled(!t.allowed || model.busy)
                        .keyboardShortcut(.return, modifiers: .command)
                        .help(t.allowed ? "\(t.command) (⌘↩)" : "Not yet: see the list on the left")
                    }
                }
            }
        }
        .padding(14)
        .background(Color(nsColor: .controlBackgroundColor), in: RoundedRectangle(cornerRadius: Style.cardRadius))
        .overlay(RoundedRectangle(cornerRadius: Style.cardRadius).strokeBorder(.separator.opacity(0.6)))
    }

    struct Stepper: View {
        var state: String
        var body: some View {
            let k = Lifecycle.steps.firstIndex(of: state) ?? 0
            HStack(spacing: 0) {
                ForEach(Array(Lifecycle.steps.enumerated()), id: \.offset) { i, step in
                    VStack(spacing: 5) {
                        ZStack {
                            Circle()
                                .fill(i < k ? Style.color(state: step).opacity(0.9) : i == k ? Style.color(state: step).opacity(0.18) : Color.clear)
                            Circle()
                                .strokeBorder(i <= k ? Style.color(state: step) : Color.secondary.opacity(0.35), lineWidth: i == k ? 2 : 1.5)
                            if i < k {
                                Image(systemName: "checkmark").font(.system(size: 9, weight: .heavy)).foregroundStyle(.white)
                            } else if i == k {
                                Circle().fill(Style.color(state: step)).frame(width: 7, height: 7)
                            }
                        }
                        .frame(width: 20, height: 20)
                        Text(Lifecycle.title(step))
                            .font(.caption.weight(i == k ? .semibold : .regular))
                            .foregroundStyle(i == k ? .primary : .secondary)
                            .fixedSize()
                    }
                    if i < Lifecycle.steps.count - 1 {
                        Rectangle()
                            .fill(i < k ? Color.secondary.opacity(0.5) : Color.secondary.opacity(0.18))
                            .frame(height: 2)
                            .padding(.bottom, 18)
                    }
                }
            }
        }
    }
}

extension String {
    var lowercasedFirst: String { prefix(1).lowercased() + dropFirst() }
}

/// Confirms a transition and asks for what it needs as fields, not flags:
/// a reason to drop or release, the documentation a completion touched.
struct TransitionSheet: View {
    var transition: PrepKit.Transition
    var issueTitle: String
    var entries: [KnowledgeEntry] = []
    var done: ([String]?) -> Void

    @State private var reason = ""
    @State private var note = ""
    @State private var commit = ""
    @State private var docs: Set<String> = []
    @State private var noImpact = false
    @State private var docQuery = ""

    var op: String { transition.op }
    var verb: String { LifecycleCard.verbs[op]?.0 ?? op.capitalized }
    var id: String { transition.command.split(separator: " ").dropFirst(2).first.map(String.init) ?? "" }

    var valid: Bool {
        switch op {
        case "drop": return !reason.trimmingCharacters(in: .whitespaces).isEmpty
        case "complete":
            return noImpact ? !reason.trimmingCharacters(in: .whitespaces).isEmpty : !docs.isEmpty
        default: return true
        }
    }

    var arguments: [String] {
        var args = [op, id]
        switch op {
        case "drop", "release":
            if !reason.isEmpty { args += ["--reason", reason] }
        case "complete":
            if !commit.isEmpty { args += ["--commit", commit] }
            if noImpact {
                args += ["--no-impact", reason]
            } else {
                for d in docs.sorted() { args += ["--docs", d] }
            }
            if !note.isEmpty { args += ["--note", note] }
        default:
            if !note.isEmpty { args += ["--note", note] }
        }
        return args
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack(spacing: 12) {
                Image(systemName: op == "drop" ? "xmark.octagon.fill" : op == "complete" ? "checkmark.seal.fill" : "arrow.right.circle.fill")
                    .font(.system(size: 30))
                    .foregroundStyle(op == "drop" ? Color.red : op == "complete" ? Color.green : Color.accentColor)
                VStack(alignment: .leading, spacing: 2) {
                    Text("\(verb) “\(issueTitle)”").font(.headline).lineLimit(2)
                    Text(LifecycleCard.verbs[op]?.1 ?? transition.command).foregroundStyle(.secondary)
                }
            }
            fields
            Text(arguments.map { $0.contains(" ") ? "\"\($0)\"" : $0 }.joined(separator: " ").replacingOccurrences(of: id, with: String(id.suffix(6))))
                .font(.caption.monospaced())
                .foregroundStyle(.tertiary)
                .lineLimit(2)
                .help("The prep command this runs")
            HStack {
                Spacer()
                Button("Cancel") { done(nil) }.keyboardShortcut(.cancelAction)
                Button(verb) { done(arguments) }
                    .keyboardShortcut(.defaultAction)
                    .buttonStyle(.borderedProminent)
                    .tint(op == "drop" ? .red : .accentColor)
                    .disabled(!valid)
            }
        }
        .padding(20)
        .frame(width: 520)
    }

    @ViewBuilder
    var fields: some View {
        switch op {
        case "drop", "release":
            field(op == "drop" ? "Reason" : "Reason (optional)") {
                TextField(op == "drop" ? "Why is it dropped?" : "Why does work stop?", text: $reason, axis: .vertical)
                    .lineLimit(2...4)
            }
        case "complete":
            field("Documentation") {
                VStack(alignment: .leading, spacing: 6) {
                    Picker("", selection: $noImpact) {
                        Text("Knowledge entries updated").tag(false)
                        Text("No documentation impact").tag(true)
                    }
                    .pickerStyle(.segmented)
                    .labelsHidden()
                    if noImpact {
                        TextField("Why nothing in the knowledge base changes", text: $reason)
                    } else {
                        TextField("Filter entries", text: $docQuery)
                        ScrollView {
                            VStack(alignment: .leading, spacing: 4) {
                                ForEach(entries.filter { docQuery.isEmpty || $0.path.localizedCaseInsensitiveContains(docQuery) || $0.title.localizedCaseInsensitiveContains(docQuery) }) { e in
                                    Toggle(isOn: Binding(
                                        get: { docs.contains(e.path) },
                                        set: { on in if on { docs.insert(e.path) } else { docs.remove(e.path) } }
                                    )) {
                                        HStack {
                                            Text(e.title)
                                            Text(e.path).font(.caption.monospaced()).foregroundStyle(.secondary)
                                        }
                                    }
                                    .toggleStyle(.checkbox)
                                }
                            }
                            .frame(maxWidth: .infinity, alignment: .leading)
                        }
                        .frame(height: 150)
                        .padding(6)
                        .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 6))
                        .overlay(RoundedRectangle(cornerRadius: 6).strokeBorder(.separator))
                    }
                }
            }
            field("Commit (optional)") {
                TextField("Leave empty when the next commit carries the work", text: $commit)
                    .font(.body.monospaced())
            }
            field("Note (optional)") { TextField("", text: $note) }
        default:
            field("Note (optional)") { TextField("Anything worth recording with this step", text: $note) }
        }
    }

    func field<Content: View>(_ label: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(label).font(.callout.weight(.medium))
            content().textFieldStyle(.roundedBorder)
        }
    }
}

// MARK: - Records

struct RequirementCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    var body: some View {
        Card("Requirement", symbol: "text.alignleft", accessory: {
            HStack(spacing: 10) {
                if let b = detail.issue.baselines?.last, let at = b.at.flatMap(PrepDate.parse) {
                    Text("defined \(at.relative)").font(.caption).foregroundStyle(.tertiary)
                }
                if detail.issue.resolution == nil { EditButton(label: "Edit") { model.editing = .requirement } }
            }
        }) {
            if let r = detail.issue.requirement, !r.isEmpty {
                MarkdownView(source: r, onLink: model.openLink)
            } else {
                Text("No requirement yet.").foregroundStyle(.secondary)
            }
        }
    }
}

struct AcceptanceCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    @State private var adding = ""

    var criteria: [Criterion] { detail.issue.criteria ?? [] }
    var resolved: Bool { detail.issue.resolution != nil }

    var body: some View {
        let checked = criteria.filter { $0.checked == true }.count
        Card("Acceptance", symbol: "checklist", accessory: {
            if !criteria.isEmpty {
                HStack(spacing: 6) {
                    Gauge(value: Double(checked), in: 0...Double(max(1, criteria.count))) {}
                        .gaugeStyle(.accessoryLinearCapacity)
                        .tint(checked == criteria.count ? .green : .accentColor)
                        .frame(width: 60)
                    Text("\(checked) of \(criteria.count)").font(.caption.monospacedDigit()).foregroundStyle(.secondary)
                }
            }
        }) {
            VStack(alignment: .leading, spacing: 8) {
                if criteria.isEmpty {
                    Text("No acceptance criteria yet.").foregroundStyle(.secondary)
                }
                ForEach(Array(criteria.enumerated()), id: \.offset) { k, c in
                    Toggle(isOn: Binding(
                        get: { c.checked ?? false },
                        set: { on in Task { await model.toggle(criterion: k + 1, checked: on) } }
                    )) {
                        Text(Markdown.inline(c.text))
                            .foregroundStyle(c.checked == true && !resolved ? .secondary : .primary)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                    .toggleStyle(.checkbox)
                    .disabled(resolved || model.busy)
                }
                if !resolved {
                    HStack {
                        Image(systemName: "plus").foregroundStyle(.tertiary)
                        TextField("Add a criterion", text: $adding)
                            .textFieldStyle(.plain)
                            .onSubmit {
                                let t = adding
                                adding = ""
                                Task { await model.addCriterion(t) }
                            }
                    }
                    .padding(.top, 2)
                }
                if let dod = detail.definitionOfDone, !dod.isEmpty {
                    DisclosureGroup {
                        VStack(alignment: .leading, spacing: 3) {
                            ForEach(dod, id: \.self) { Label($0, systemImage: "checkmark.seal").font(.callout) }
                        }
                        .foregroundStyle(.secondary)
                        .padding(.top, 4)
                    } label: {
                        Text("Definition of Done · \(dod.count)").font(.callout).foregroundStyle(.secondary)
                    }
                    .padding(.top, 4)
                }
            }
        }
    }
}

struct ContextCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    var body: some View {
        Card("Context", symbol: "doc.text.magnifyingglass", accessory: {
            if detail.issue.resolution == nil {
                EditButton(label: (detail.context ?? "").isEmpty ? "Write" : "Edit") { model.editing = .context }
            }
        }) {
            VStack(alignment: .leading, spacing: 10) {
                if (detail.context ?? "").isEmpty {
                    Text("How the work will be done: approach, code paths, knowledge to read. Required before the issue can be marked ready.")
                        .foregroundStyle(.secondary)
                } else {
                    MarkdownView(source: detail.context ?? "", onLink: model.openLink)
                }
                let links = detail.issue.contextLinks ?? []
                let paths = detail.issue.contextPaths ?? []
                if !links.isEmpty || !paths.isEmpty {
                    FlowLayout(spacing: 6) {
                        ForEach(links, id: \.self) { path in
                            Button { model.openLink(path) } label: {
                                Chip(text: model.entries.first { $0.path == path }?.title ?? path, symbol: "book", color: .blue)
                            }
                            .buttonStyle(.plain)
                            .help(path)
                        }
                        ForEach(paths, id: \.self) { p in
                            Button { model.copy(p) } label: {
                                Text(p).font(.caption.monospaced())
                                    .padding(.horizontal, 6).padding(.vertical, 2)
                                    .background(.quaternary.opacity(0.6), in: RoundedRectangle(cornerRadius: 4))
                            }
                            .buttonStyle(.plain)
                            .help("Click to copy")
                        }
                    }
                }
            }
        }
    }
}

struct FindingsCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    var body: some View {
        Card("Findings", symbol: "lightbulb", accessory: {
            if detail.issue.resolution == nil {
                EditButton(label: (detail.findings ?? "").isEmpty ? "Write" : "Edit") { model.editing = .findings }
            }
        }) {
            if let f = detail.findings, !f.isEmpty {
                MarkdownView(source: f, onLink: model.openLink)
            } else {
                Text("What the research found. Required to complete a research issue.").foregroundStyle(.secondary)
            }
        }
    }
}

struct DecisionsCard: View {
    var model: ProjectModel
    var decisions: [Decision]
    var editable = true
    var body: some View {
        Card("Decisions", symbol: "arrow.triangle.branch", accessory: {
            HStack(spacing: 10) {
                if !decisions.isEmpty { Text("\(decisions.count)").font(.caption).foregroundStyle(.secondary) }
                if editable {
                    Button { model.editing = .decision } label: { Label("Record", systemImage: "plus") }
                        .buttonStyle(.borderless).font(.callout)
                }
            }
        }) {
            VStack(alignment: .leading, spacing: 14) {
                if decisions.isEmpty {
                    Text("No decisions recorded yet.").foregroundStyle(.secondary)
                }
                ForEach(Array(decisions.enumerated()), id: \.offset) { k, d in
                    if k > 0 { Divider() }
                    let superseded = decisions.contains { $0.supersedes == d.id }
                    VStack(alignment: .leading, spacing: 6) {
                        HStack(spacing: 8) {
                            Text(d.id)
                                .font(.caption.weight(.bold).monospaced())
                                .padding(.horizontal, 6).padding(.vertical, 2)
                                .background(Color.accentColor.opacity(0.15), in: RoundedRectangle(cornerRadius: 4))
                                .foregroundStyle(Color.accentColor)
                            Text(d.title).font(.headline).strikethrough(superseded, color: .secondary)
                            if d.outcome == true { Chip(text: "Outcome", symbol: "flag.checkered", color: .green) }
                            if let s = d.supersedes { Chip(text: "Supersedes \(s)") }
                            if superseded { Chip(text: "Superseded") }
                            Spacer()
                            if let date = d.date { Text(date).font(.caption).foregroundStyle(.tertiary) }
                        }
                        if let b = d.body, !b.isEmpty {
                            MarkdownView(source: b, onLink: model.openLink).opacity(superseded ? 0.6 : 1)
                        }
                    }
                }
            }
        }
    }
}

struct WorkLogCard: View {
    var model: ProjectModel
    var detail: IssueDetail
    @State private var text = ""

    var body: some View {
        let entries = WorkLog.parse(detail.history ?? "")
        Card("Work Log", symbol: "clock.arrow.circlepath", accessory: {
            if !entries.isEmpty { Text("\(entries.count)").font(.caption).foregroundStyle(.secondary) }
        }) {
            VStack(alignment: .leading, spacing: 0) {
                ForEach(Array(entries.enumerated()), id: \.offset) { k, e in
                    HStack(alignment: .top, spacing: 10) {
                        VStack(spacing: 0) {
                            ActorAvatar(actor: e.actor ?? "?", size: 22)
                            if k < entries.count - 1 {
                                Rectangle().fill(.separator).frame(width: 1).frame(maxHeight: .infinity)
                            }
                        }
                        VStack(alignment: .leading, spacing: 2) {
                            HStack(spacing: 6) {
                                Text(e.actor ?? "").font(.caption.weight(.semibold))
                                if let d = e.date {
                                    Text(d.relative).font(.caption).foregroundStyle(.tertiary).help(d.formatted())
                                }
                            }
                            Text(Markdown.inline(e.text)).fixedSize(horizontal: false, vertical: true).textSelection(.enabled)
                        }
                        .padding(.bottom, 12)
                    }
                }
                HStack(spacing: 8) {
                    ActorAvatar(actor: PrepCLI.defaultActor, size: 22)
                    TextField("Add to the work log…", text: $text, axis: .vertical)
                        .textFieldStyle(.roundedBorder)
                        .lineLimit(1...4)
                        .onSubmit(add)
                    Button("Add", action: add).disabled(text.trimmingCharacters(in: .whitespaces).isEmpty)
                }
            }
        }
    }

    func add() {
        let t = text.trimmingCharacters(in: .whitespacesAndNewlines)
        text = ""
        Task { await model.log(t) }
    }
}

/// Lays children out in rows, wrapping like text.
struct FlowLayout: Layout {
    var spacing: CGFloat = 6

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let width = proposal.width ?? .infinity
        var x: CGFloat = 0, y: CGFloat = 0, row: CGFloat = 0, widest: CGFloat = 0
        for s in subviews {
            let size = s.sizeThatFits(.unspecified)
            if x > 0, x + size.width > width {
                y += row + spacing
                x = 0
                row = 0
            }
            x += size.width + spacing
            row = max(row, size.height)
            widest = max(widest, x - spacing)
        }
        return CGSize(width: min(widest, width), height: y + row)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var x = bounds.minX, y = bounds.minY, row: CGFloat = 0
        for s in subviews {
            let size = s.sizeThatFits(.unspecified)
            if x > bounds.minX, x + size.width > bounds.maxX {
                y += row + spacing
                x = bounds.minX
                row = 0
            }
            s.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
            x += size.width + spacing
            row = max(row, size.height)
        }
    }
}
