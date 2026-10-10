import SwiftUI
import PrepKit

/// The issue's properties beside the document: what can be changed in
/// place (kind, priority, tags, parent) and when each step happened.
struct IssueInspector: View {
    var model: ProjectModel
    var detail: IssueDetail
    @State private var newTag = ""

    var issue: IssueRecord { detail.issue }
    var resolved: Bool { issue.resolution != nil }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                InspectorSection("Properties") {
                    InspectorRow("State") { StateBadge(state: detail.state) }
                    InspectorRow("Kind") {
                        Menu {
                            ForEach(Lifecycle.kinds, id: \.self) { k in
                                Button { Task { await model.edit(["--kind", k]) } } label: {
                                    Label(k.capitalized, systemImage: Style.symbol(kind: k))
                                }
                            }
                        } label: {
                            Label(issue.kind.capitalized, systemImage: Style.symbol(kind: issue.kind))
                        }
                        .menuStyle(.borderlessButton)
                        .fixedSize()
                        .disabled(resolved)
                    }
                    InspectorRow("Priority") {
                        Menu {
                            ForEach(Lifecycle.priorities, id: \.self) { p in
                                Button(p.capitalized) { Task { await model.edit(["--priority", p]) } }
                            }
                        } label: {
                            HStack(spacing: 4) {
                                PriorityIcon(priority: detail.priority, showMedium: true)
                                Text((detail.priority ?? "medium").capitalized)
                            }
                        }
                        .menuStyle(.borderlessButton)
                        .fixedSize()
                    }
                    InspectorRow("Parent") {
                        Menu {
                            Button("None") { Task { await model.edit(["--parent", ""]) } }
                            Divider()
                            ForEach(parentCandidates) { s in
                                Button(s.title) { Task { await model.edit(["--parent", s.id]) } }
                            }
                        } label: {
                            Text(issue.parent.map { model.title($0) } ?? "None").lineLimit(1)
                        }
                        .menuStyle(.borderlessButton)
                        .disabled(resolved)
                    }
                }

                InspectorSection("Tags") {
                    FlowLayout(spacing: 5) {
                        ForEach(issue.tags ?? [], id: \.self) { t in
                            HStack(spacing: 3) {
                                Text("#\(t)")
                                if !resolved {
                                    Button { setTags((issue.tags ?? []).filter { $0 != t }) } label: {
                                        Image(systemName: "xmark").font(.system(size: 8, weight: .bold))
                                    }
                                    .buttonStyle(.plain)
                                    .help("Remove #\(t)")
                                }
                            }
                            .font(.caption.weight(.medium))
                            .padding(.horizontal, 7).padding(.vertical, 3)
                            .background(.quaternary.opacity(0.6), in: Capsule())
                        }
                    }
                    if !resolved {
                        TextField("Add a tag", text: $newTag)
                            .textFieldStyle(.roundedBorder)
                            .controlSize(.small)
                            .onSubmit {
                                let t = newTag.trimmingCharacters(in: .whitespaces).trimmingCharacters(in: CharacterSet(charactersIn: "#"))
                                newTag = ""
                                if !t.isEmpty, !(issue.tags ?? []).contains(t) { setTags((issue.tags ?? []) + [t]) }
                            }
                    }
                }

                InspectorSection("Timeline") {
                    if let d = ULID.date(issue.id) { row("Created", date: d, by: nil) }
                    if let b = issue.baselines?.last { row("Defined", at: b.at, by: b.by) }
                    if let r = issue.ready { row("Ready", at: r.at, by: r.by) }
                    if let c = issue.claim { row("Claimed", at: c.at, by: c.by) }
                    if let r = issue.resolution { row(r.outcome.capitalized, at: r.at, by: r.by) }
                    if let n = issue.baselines?.count, n > 1 {
                        InspectorRow("Baselines") { Text("\(n)") }
                    }
                }

                InspectorSection("Issue") {
                    Text(issue.id).font(.caption.monospaced()).foregroundStyle(.secondary).textSelection(.enabled)
                    HStack {
                        Button("Copy ID") { model.copy(issue.id) }
                        Button("Reveal in Finder") { model.revealInFinder(issue.id) }
                    }
                    .controlSize(.small)
                }
            }
            .padding(16)
        }
    }

    var parentCandidates: [IssueSummary] {
        model.all.values
            .filter { $0.id != issue.id && !$0.isResolved && (($0.children ?? 0) > 0 || $0.kind == "manual") }
            .sorted { $0.title < $1.title }
    }

    func setTags(_ tags: [String]) { Task { await model.setTags(tags) } }

    func row(_ label: String, at: String?, by: String?) -> some View {
        row(label, date: at.flatMap(PrepDate.parse), by: by)
    }

    func row(_ label: String, date: Date?, by: String?) -> some View {
        InspectorRow(label) {
            VStack(alignment: .trailing, spacing: 1) {
                if let date { Text(date.relative).help(date.formatted()) }
                if let by { Text(by).font(.caption).foregroundStyle(.secondary).lineLimit(1) }
            }
        }
    }
}

/// A titled inspector group, as in Xcode's and Keynote's inspectors.
struct InspectorSection<Content: View>: View {
    var title: String
    var content: Content
    init(_ title: String, @ViewBuilder content: () -> Content) {
        self.title = title
        self.content = content()
    }
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title).font(.caption.weight(.semibold)).foregroundStyle(.secondary)
            content
            Divider().padding(.top, 4)
        }
    }
}

/// A label on the left, its value on the right.
struct InspectorRow<Value: View>: View {
    var label: String
    var value: Value
    init(_ label: String, @ViewBuilder value: () -> Value) {
        self.label = label
        self.value = value()
    }
    var body: some View {
        HStack(alignment: .firstTextBaseline) {
            Text(label).foregroundStyle(.secondary)
            Spacer(minLength: 8)
            value
        }
        .font(.callout)
    }
}

/// Creating an issue: the title first, the rest optional.
struct NewIssueSheet: View {
    var model: ProjectModel
    @State var draft: NewIssueDraft
    @Environment(\.dismiss) private var dismiss
    @FocusState private var titleFocused: Bool

    var parents: [IssueSummary] {
        model.all.values
            .filter { !$0.isResolved && (($0.children ?? 0) > 0 || $0.kind == "manual") }
            .sorted { $0.title < $1.title }
    }

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 10) {
                Image(systemName: "square.and.pencil").font(.title2).foregroundStyle(Color.accentColor)
                VStack(alignment: .leading, spacing: 1) {
                    Text("New Issue").font(.title3.bold())
                    Text(draft.parent.map { "in \(model.title($0))" } ?? "at the top level")
                        .font(.callout).foregroundStyle(.secondary).lineLimit(1)
                }
                Spacer()
            }
            .padding([.horizontal, .top], 20)
            Form {
                TextField("Title", text: $draft.title, prompt: Text("What needs to happen?"))
                    .font(.title3)
                    .focused($titleFocused)
                Picker("Kind", selection: $draft.kind) {
                    ForEach(Lifecycle.kinds, id: \.self) { k in
                        Label(k.capitalized, systemImage: Style.symbol(kind: k)).tag(k)
                    }
                }
                .pickerStyle(.segmented)
                Picker("Priority", selection: $draft.priority) {
                    ForEach(Lifecycle.priorities, id: \.self) { Text($0.capitalized).tag($0) }
                }
                Picker("Parent", selection: $draft.parent) {
                    Text("None").tag(String?.none)
                    if let p = draft.parent, !parents.contains(where: { $0.id == p }) {
                        Text(model.title(p)).tag(String?.some(p))
                    }
                    ForEach(parents) { Text($0.title).tag(String?.some($0.id)) }
                }
                TextField("Tags", text: $draft.tags, prompt: Text("ui, cli"))
                VStack(alignment: .leading, spacing: 4) {
                    Text("Requirement").font(.callout).foregroundStyle(.secondary)
                    TextEditor(text: $draft.body)
                        .font(.body)
                        .frame(minHeight: 140)
                        .scrollContentBackground(.hidden)
                        .padding(6)
                        .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 6))
                        .overlay(RoundedRectangle(cornerRadius: 6).strokeBorder(.separator))
                    Text("Markdown. Questions to settle go under “## Open questions”.")
                        .font(.caption).foregroundStyle(.tertiary)
                }
            }
            .formStyle(.grouped)
            Divider()
            HStack {
                Spacer()
                Button("Cancel") { dismiss() }.keyboardShortcut(.cancelAction)
                Button("Create Issue") {
                    let d = draft
                    Task {
                        if await model.create(d) { dismiss() }
                    }
                }
                .keyboardShortcut(.defaultAction)
                .buttonStyle(.borderedProminent)
                .disabled(draft.title.trimmingCharacters(in: .whitespaces).isEmpty || model.busy)
            }
            .padding(16)
        }
        .frame(width: 560, height: 560)
        .onAppear { titleFocused = true }
    }
}
