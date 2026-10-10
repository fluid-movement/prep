import SwiftUI
import PrepKit

/// Renders prep's markdown records: block structure from PrepKit's parser,
/// inline markup from AttributedString.
struct MarkdownView: View {
    var source: String
    var onLink: ((String) -> Void)? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            ForEach(Array(Markdown.blocks(source).enumerated()), id: \.offset) { _, block in
                blockView(block)
            }
        }
        .textSelection(.enabled)
        .environment(\.openURL, OpenURLAction { url in
            if let onLink, url.scheme == nil {
                onLink(url.absoluteString)
                return .handled
            }
            return .systemAction
        })
    }

    @ViewBuilder
    func blockView(_ block: MarkdownBlock) -> some View {
        switch block {
        case let .heading(level, text):
            Text(Markdown.inline(text))
                .font(level <= 1 ? .title2.bold() : level == 2 ? .title3.bold() : .headline)
                .padding(.top, 4)
        case let .paragraph(text):
            Text(Markdown.inline(text)).fixedSize(horizontal: false, vertical: true)
        case let .bullet(text, depth, checked):
            HStack(alignment: .firstTextBaseline, spacing: 6) {
                if let checked {
                    Image(systemName: checked ? "checkmark.square" : "square").foregroundStyle(.secondary)
                } else {
                    Text("•").foregroundStyle(.secondary)
                }
                Text(Markdown.inline(text)).fixedSize(horizontal: false, vertical: true)
            }
            .padding(.leading, CGFloat(depth) * 16)
        case let .numbered(number, text, depth):
            HStack(alignment: .firstTextBaseline, spacing: 6) {
                Text("\(number).").foregroundStyle(.secondary).monospacedDigit()
                Text(Markdown.inline(text)).fixedSize(horizontal: false, vertical: true)
            }
            .padding(.leading, CGFloat(depth) * 16)
        case let .quote(text):
            Text(Markdown.inline(text))
                .foregroundStyle(.secondary)
                .padding(.leading, 10)
                .overlay(alignment: .leading) { Rectangle().fill(.quaternary).frame(width: 3) }
        case let .code(text):
            ScrollView(.horizontal) {
                Text(text).font(.system(.callout, design: .monospaced)).padding(8)
            }
            .background(.quaternary.opacity(0.5), in: RoundedRectangle(cornerRadius: 6))
        case .rule:
            Divider()
        }
    }
}
