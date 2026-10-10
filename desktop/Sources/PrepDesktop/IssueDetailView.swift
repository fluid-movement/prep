import SwiftUI
import PrepKit

struct IssueDetailView: View {
    @Bindable var model: ProjectModel
    var detail: IssueDetail
    @State private var logText = ""
    @State private var pending: PrepKit.Transition?

    var issue: IssueRecord { detail.issue }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                header
                links
                transitions
                if let r = issue.requirement, !r.isEmpty {
                    DetailSection("Requirement") { MarkdownView(source: r, onLink: model.openLink) }
                }
                criteria
                if let dod = detail.definitionOfDone, !dod.isEmpty {
                    DetailSection("Definition of Done") {
                        VStack(alignment: .leading, spacing: 4) {
                            ForEach(dod, id: \.self) { Label($0, systemImage: "checkmark.seal").labelStyle(.titleAndIcon) }
                        }
                        .foregroundStyle(.secondary)
                    }
                }
                if let c = detail.context, !c.isEmpty {
                    DetailSection("Context") { MarkdownView(source: c, onLink: model.openLink) }
                }
                decisions
                if let f = detail.findings, !f.isEmpty {
                    DetailSection("Findings") { MarkdownView(source: f, onLink: model.openLink) }
                }
                workLog
                if let r = issue.resolution {
                    DetailSection("Resolution") {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("\(r.outcome) by \(r.by ?? "?") at \(r.at ?? "?")")
                            if let e = r.evidence { Text("evidence: \(e)").font(.system(.body, design: .monospaced)) }
                            if let reason = r.reason { Text(reason) }
                        }
                        .foregroundStyle(.secondary)
                    }
                }
            }
            .padding(24)
            .frame(maxWidth: 820, alignment: .leading)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .sheet(item: $pending) { t in
            TransitionSheet(transition: t) { args in
                pending = nil
                if let args { Task { await model.write(args) } }
            }
        }
    }

    var header: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                Text(String(issue.id.suffix(6)))
                    .font(.system(.callout, design: .monospaced))
                    .foregroundStyle(.secondary)
                    .help(issue.id)
                StateBadge(state: detail.state)
                Pill(text: issue.kind)
                if let mark = Style.priorityMark(detail.priority) { Pill(text: mark, color: .red) }
                if detail.stale { Pill(text: "stale", color: .yellow) }
                if detail.blocked { Pill(text: "blocked", color: .red) }
                if detail.actionable { Pill(text: "actionable", color: .green) }
                ForEach(issue.tags ?? [], id: \.self) { Text("#\($0)").font(.caption).foregroundStyle(.secondary) }
                Spacer()
                Button {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(issue.id, forType: .string)
                } label: { Image(systemName: "doc.on.doc") }
                .buttonStyle(.borderless)
                .help("Copy the issue ID")
            }
            Text(issue.title).font(.title.bold()).textSelection(.enabled)
            if let p = detail.progress, p.total > 0 {
                ProgressView(value: Double(p.done + p.dropped), total: Double(p.total)) {
                    Text("\(p.done) done, \(p.dropped) dropped of \(p.total)").font(.caption)
                }
                .frame(maxWidth: 320)
            }
        }
    }

    @ViewBuilder
    var links: some View {
        let groups: [(String, [String])] = [
            ("Parent", issue.parent.map { [$0] } ?? []),
            ("Depends on", issue.dependsOn ?? []),
            ("Children", detail.children ?? []),
            ("Blocks", detail.blocks ?? []),
        ].filter { !$0.1.isEmpty }
        if !groups.isEmpty || !(issue.contextLinks ?? []).isEmpty {
            Grid(alignment: .leading, horizontalSpacing: 12, verticalSpacing: 4) {
                ForEach(groups, id: \.0) { name, ids in
                    GridRow(alignment: .firstTextBaseline) {
                        Text(name).foregroundStyle(.secondary).gridColumnAlignment(.trailing)
                        VStack(alignment: .leading, spacing: 2) {
                            ForEach(ids, id: \.self) { id in
                                Button { model.go(to: id) } label: {
                                    if let s = model.summary(id) { IssueRow(issue: s) } else { Text(id) }
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                }
                if let ks = issue.contextLinks, !ks.isEmpty {
                    GridRow(alignment: .firstTextBaseline) {
                        Text("Knowledge").foregroundStyle(.secondary)
                        VStack(alignment: .leading, spacing: 2) {
                            ForEach(ks, id: \.self) { path in
                                Button(path) { model.openLink(path) }.buttonStyle(.link)
                            }
                        }
                    }
                }
            }
            .font(.callout)
        }
    }

    @ViewBuilder
    var transitions: some View {
        if let g = model.guide, let ts = g.transitions, !ts.isEmpty {
            VStack(alignment: .leading, spacing: 6) {
                if let step = g.step { Text("Step: \(step)").font(.caption).foregroundStyle(.secondary) }
                HStack {
                    ForEach(ts) { t in
                        Button(t.op.capitalized) { run(t) }
                            .disabled(!t.allowed)
                            .help(t.allowed ? t.command : (t.unmet ?? []).map(\.message).joined(separator: "\n"))
                    }
                }
                let blocked = ts.filter { !$0.allowed && $0.op != "define" && $0.op != "ack" }
                ForEach(blocked) { t in
                    ForEach(t.unmet ?? [], id: \.self) { gate in
                        Label("\(t.op): \(gate.message)", systemImage: "exclamationmark.triangle")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .padding(10)
            .background(.quaternary.opacity(0.35), in: RoundedRectangle(cornerRadius: 8))
        }
    }

    /// Every transition is confirmed in a sheet that also takes its flags.
    func run(_ t: PrepKit.Transition) {
        pending = t
    }

    @ViewBuilder
    var criteria: some View {
        if let cs = issue.criteria, !cs.isEmpty {
            DetailSection("Acceptance") {
                VStack(alignment: .leading, spacing: 6) {
                    ForEach(Array(cs.enumerated()), id: \.offset) { k, c in
                        Toggle(isOn: Binding(
                            get: { c.checked ?? false },
                            set: { on in Task { await model.toggle(criterion: k + 1, checked: on) } }
                        )) {
                            Text(Markdown.inline(c.text)).fixedSize(horizontal: false, vertical: true)
                        }
                        .toggleStyle(.checkbox)
                        .disabled(detail.state == "done" || detail.state == "dropped")
                    }
                }
            }
        }
    }

    @ViewBuilder
    var decisions: some View {
        if let ds = issue.decisions, !ds.isEmpty {
            DetailSection("Decisions") {
                VStack(alignment: .leading, spacing: 12) {
                    ForEach(ds, id: \.id) { d in
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                Text("\(d.id): \(d.title)").font(.headline)
                                if d.outcome == true { Pill(text: "outcome", color: .green) }
                                if let s = d.supersedes { Pill(text: "supersedes \(s)") }
                                Spacer()
                                if let date = d.date { Text(date).font(.caption).foregroundStyle(.secondary) }
                            }
                            if let b = d.body { MarkdownView(source: b, onLink: model.openLink) }
                        }
                    }
                }
            }
        }
    }

    var workLog: some View {
        DetailSection("Work log") {
            VStack(alignment: .leading, spacing: 6) {
                if let h = detail.history, !h.isEmpty {
                    MarkdownView(source: h, onLink: model.openLink).font(.callout)
                }
                HStack {
                    TextField("Add to the work log", text: $logText)
                        .textFieldStyle(.roundedBorder)
                        .onSubmit(addLog)
                    Button("Log", action: addLog).disabled(logText.isEmpty)
                }
            }
        }
    }

    func addLog() {
        let text = logText
        logText = ""
        Task { await model.log(text) }
    }
}

/// A titled section of the detail view.
struct DetailSection<Content: View>: View {
    var title: String
    @ViewBuilder var content: Content
    init(_ title: String, @ViewBuilder content: () -> Content) {
        self.title = title
        self.content = content()
    }
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title.uppercased())
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
            content
        }
    }
}

/// Asks for the flags a transition needs (a reason, docs, a commit) before
/// running it.
struct TransitionSheet: View {
    var transition: PrepKit.Transition
    var done: ([String]?) -> Void
    @State private var text: String

    init(transition: PrepKit.Transition, done: @escaping ([String]?) -> Void) {
        self.transition = transition
        self.done = done
        _text = State(initialValue: "")
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(transition.command).font(.headline.monospaced()).fixedSize(horizontal: false, vertical: true)
            ForEach(transition.needs ?? [], id: \.self) { Text($0).font(.callout.monospaced()).foregroundStyle(.secondary) }
            TextField((transition.needs ?? []).isEmpty ? "Optional flags, e.g. --note \"why\"" : (transition.needs ?? []).joined(separator: " "),
                      text: $text, axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .font(.body.monospaced())
                .frame(minWidth: 460)
            HStack {
                Spacer()
                Button("Cancel") { done(nil) }.keyboardShortcut(.cancelAction)
                Button(transition.op.capitalized) {
                    let parts = transition.command.split(separator: " ").map(String.init)
                    done(Array(parts.dropFirst().prefix(2)) + Shell.split(text))
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(20)
    }
}

