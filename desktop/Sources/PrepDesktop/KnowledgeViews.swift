import SwiftUI
import PrepKit

/// The knowledge base: entries by type, searched by title, path and
/// description as you type.
struct KnowledgeSidebar: View {
    @Bindable var model: ProjectModel

    var filtered: [KnowledgeEntry] {
        let q = model.knowledgeQuery.trimmingCharacters(in: .whitespaces).lowercased()
        guard !q.isEmpty else { return model.entries }
        return model.entries.filter {
            $0.title.lowercased().contains(q) || $0.path.lowercased().contains(q)
                || ($0.description ?? "").lowercased().contains(q)
        }
    }

    static let typeOrder = ["overview", "feature", "component", "decision", "guide"]

    var body: some View {
        let groups = Dictionary(grouping: filtered, by: \.type)
        let types = groups.keys.sorted {
            (Self.typeOrder.firstIndex(of: $0) ?? .max, $0) < (Self.typeOrder.firstIndex(of: $1) ?? .max, $1)
        }
        VStack(spacing: 0) {
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass").foregroundStyle(.secondary)
                TextField("Search knowledge", text: $model.knowledgeQuery).textFieldStyle(.plain)
                if !model.knowledgeQuery.isEmpty {
                    Button { model.knowledgeQuery = "" } label: {
                        Image(systemName: "xmark.circle.fill").foregroundStyle(.secondary)
                    }
                    .buttonStyle(.borderless)
                }
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 5)
            .background(.quaternary.opacity(0.5), in: RoundedRectangle(cornerRadius: 7))
            .padding(10)
            Divider()
            if filtered.isEmpty {
                ContentUnavailableView.search(text: model.knowledgeQuery)
            } else {
                List(selection: $model.entrySelection) {
                    ForEach(types, id: \.self) { type in
                        Section(type.capitalized + "s") {
                            ForEach(groups[type] ?? []) { e in
                                HStack(spacing: 8) {
                                    Image(systemName: Style.symbol(knowledge: e.type))
                                        .foregroundStyle(.secondary)
                                        .frame(width: 16)
                                    VStack(alignment: .leading, spacing: 1) {
                                        Text(e.title).lineLimit(1)
                                        Text(e.path).font(.caption.monospaced()).foregroundStyle(.secondary).lineLimit(1)
                                    }
                                    Spacer(minLength: 0)
                                    if let s = e.status, s != "stable" {
                                        Chip(text: s, color: s == "deprecated" ? .red : .orange)
                                    }
                                }
                                .tag(e.path)
                                .help(e.description ?? e.path)
                                .contextMenu {
                                    Button("Copy Path") { model.copy(e.path) }
                                }
                            }
                        }
                    }
                }
                .listStyle(.sidebar)
            }
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            HStack {
                Text("\(model.entries.count) entries").font(.caption).foregroundStyle(.secondary)
                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 7)
            .background(.bar)
            .overlay(alignment: .top) { Divider() }
        }
    }
}

struct KnowledgeEntryView: View {
    var model: ProjectModel
    var entry: KnowledgeEntry

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                VStack(alignment: .leading, spacing: 10) {
                    HStack(spacing: 10) {
                        Image(systemName: Style.symbol(knowledge: entry.type))
                            .font(.system(size: 22))
                            .foregroundStyle(Color.accentColor)
                            .frame(width: 40, height: 40)
                            .background(Color.accentColor.opacity(0.12), in: RoundedRectangle(cornerRadius: 9))
                        VStack(alignment: .leading, spacing: 2) {
                            Text(entry.type.capitalized).font(.caption.weight(.semibold)).foregroundStyle(.secondary)
                            Text(entry.path).font(.callout.monospaced()).foregroundStyle(.secondary).textSelection(.enabled)
                        }
                        Spacer()
                        if let s = entry.status { Chip(text: s, color: s == "stable" ? .green : .orange) }
                        if let size = entry.size {
                            Text("≈\(size / 4) tokens").font(.caption).foregroundStyle(.tertiary)
                        }
                    }
                    if let d = entry.description {
                        Text(Markdown.inline(d)).font(.title3).foregroundStyle(.secondary).fixedSize(horizontal: false, vertical: true)
                    }
                    if let scope = entry.scope, !scope.isEmpty {
                        FlowLayout(spacing: 5) {
                            Image(systemName: "scope").foregroundStyle(.tertiary).font(.caption)
                            ForEach(scope, id: \.self) {
                                Text($0).font(.caption.monospaced())
                                    .padding(.horizontal, 6).padding(.vertical, 2)
                                    .background(.quaternary.opacity(0.6), in: RoundedRectangle(cornerRadius: 4))
                            }
                        }
                    }
                }
                Divider()
                MarkdownView(source: bodyText, onLink: model.openLink)
                    .font(.body)
                    .lineSpacing(2)
            }
            .padding(.horizontal, 32)
            .padding(.vertical, 24)
            .frame(maxWidth: 780, alignment: .leading)
            .frame(maxWidth: .infinity)
        }
    }

    /// The body without its leading title heading, which the header shows.
    var bodyText: String {
        let text = entry.text ?? ""
        let lines = text.components(separatedBy: "\n")
        if let first = lines.first, first.hasPrefix("# ") {
            return lines.dropFirst().joined(separator: "\n")
        }
        return text
    }
}
