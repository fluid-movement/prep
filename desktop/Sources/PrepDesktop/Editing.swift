import SwiftUI
import PrepKit

/// What an editor sheet writes: each case is one prep command.
enum EditTarget: String, Identifiable {
    case title, requirement, context, findings, criterion, decision
    var id: String { rawValue }

    var heading: String {
        switch self {
        case .title: "Rename Issue"
        case .requirement: "Edit Requirement"
        case .context: "Edit Context"
        case .findings: "Edit Findings"
        case .criterion: "Add Acceptance Criterion"
        case .decision: "Record a Decision"
        }
    }

    var hint: String {
        switch self {
        case .title: "A short, specific title."
        case .requirement: "What and why, in markdown. Questions to settle go under “## Open questions”; the issue can be defined once that section is empty."
        case .context: "How: the implementation context. Link knowledge entries as [Title](/path.md) and name code paths in backticks."
        case .findings: "What the research found, in markdown."
        case .criterion: "One checkable statement that is true when the issue is done."
        case .decision: "The decision and its rationale, with the alternatives considered."
        }
    }

    var multiline: Bool { self == .requirement || self == .context || self == .findings || self == .decision }
}

/// Friendlier words for prep's gates, and what fixes each one.
enum Gates {
    static func text(_ g: Gate) -> String {
        switch g.code {
        case "G_OPEN_QUESTIONS": "Settle the open questions in the requirement"
        case "G_REQUIREMENT": "Write the requirement"
        case "G_CRITERIA": "Add at least one acceptance criterion"
        case "G_CONTEXT": "Write the implementation context"
        case "G_DEPS": "Finish the issues it depends on"
        case "G_UNCHECKED": "Check off every acceptance criterion"
        case "G_CHILDREN": "Resolve its children"
        case "G_FINDINGS": "Write the findings"
        case "G_OUTCOME": "Record the decision's outcome"
        default: g.message
        }
    }

    static func fix(_ g: Gate) -> (String, EditTarget)? {
        switch g.code {
        case "G_OPEN_QUESTIONS", "G_REQUIREMENT": ("Edit Requirement", .requirement)
        case "G_CRITERIA": ("Add Criterion", .criterion)
        case "G_CONTEXT": ("Write Context", .context)
        case "G_FINDINGS": ("Write Findings", .findings)
        case "G_OUTCOME": ("Record Decision", .decision)
        default: nil
        }
    }
}

extension ProjectModel {
    func initialText(_ t: EditTarget) -> String {
        switch t {
        case .title: detail?.issue.title ?? ""
        case .requirement: detail?.issue.requirement ?? ""
        case .context: detail?.context ?? ""
        case .findings: detail?.findings ?? ""
        case .criterion, .decision: ""
        }
    }

    func save(_ t: EditTarget, text: String, title: String, outcome: Bool) async -> Bool {
        guard let id = current else { return false }
        switch t {
        case .title: return await write(["edit", id, "--title", text])
        case .requirement: return await write(["edit", id, "--body-file", "-"], stdin: text)
        case .context: return await write(["context", id, "--body-file", "-"], stdin: text)
        case .findings: return await write(["findings", id, "--body-file", "-"], stdin: text)
        case .criterion: return await write(["criterion", id, "--add", text])
        case .decision:
            var args = ["decide", id, "--title", title, "--body-file", "-"]
            if outcome { args.append("--outcome") }
            return await write(args, stdin: text)
        }
    }
}

struct EditorSheet: View {
    var model: ProjectModel
    var target: EditTarget
    @State private var text: String
    @State private var decisionTitle = ""
    @State private var outcome = false
    @Environment(\.dismiss) private var dismiss

    init(model: ProjectModel, target: EditTarget) {
        self.model = model
        self.target = target
        _text = State(initialValue: model.initialText(target))
    }

    var canSave: Bool {
        let t = text.trimmingCharacters(in: .whitespacesAndNewlines)
        if target == .decision { return !decisionTitle.trimmingCharacters(in: .whitespaces).isEmpty }
        return !t.isEmpty || target == .context || target == .findings
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            VStack(alignment: .leading, spacing: 3) {
                Text(target.heading).font(.title3.bold())
                if let t = model.detail?.issue.title { Text(t).foregroundStyle(.secondary).lineLimit(1) }
            }
            if target == .decision {
                TextField("Decision", text: $decisionTitle, prompt: Text("What was decided"))
                    .textFieldStyle(.roundedBorder)
                    .font(.title3)
            }
            if target.multiline {
                TextEditor(text: $text)
                    .font(.system(.body, design: .monospaced))
                    .scrollContentBackground(.hidden)
                    .padding(8)
                    .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 8))
                    .overlay(RoundedRectangle(cornerRadius: 8).strokeBorder(.separator))
                    .frame(minHeight: 260)
            } else {
                TextField(target.heading, text: $text)
                    .textFieldStyle(.roundedBorder)
                    .font(.title3)
                    .onSubmit(save)
            }
            Text(target.hint).font(.caption).foregroundStyle(.secondary).fixedSize(horizontal: false, vertical: true)
            HStack {
                if target == .decision, model.detail?.issue.kind == "decision" {
                    Toggle("This is the outcome", isOn: $outcome).toggleStyle(.checkbox)
                }
                Spacer()
                Button("Cancel") { dismiss() }.keyboardShortcut(.cancelAction)
                Button(target == .criterion || target == .decision ? "Add" : "Save", action: save)
                    .keyboardShortcut(target.multiline ? KeyboardShortcut(.return, modifiers: .command) : .defaultAction)
                    .buttonStyle(.borderedProminent)
                    .disabled(!canSave || model.busy)
                    .help(target.multiline ? "⌘↩" : "↩")
            }
        }
        .padding(20)
        .frame(width: target.multiline ? 680 : 520)
    }

    func save() {
        guard canSave else { return }
        let t = text, title = decisionTitle, o = outcome
        Task {
            if await model.save(target, text: t, title: title, outcome: o) { dismiss() }
        }
    }
}

/// A small pencil button for a card header.
struct EditButton: View {
    var label: String
    var action: () -> Void
    var body: some View {
        Button(action: action) {
            Label(label, systemImage: "pencil").labelStyle(.titleAndIcon)
        }
        .buttonStyle(.borderless)
        .font(.callout)
    }
}
