import Foundation

/// A block of a markdown document. prep's records are plain CommonMark with
/// headings, paragraphs, lists, quotes and fenced code; that is all the app
/// renders. Inline markup is left to AttributedString(markdown:).
public enum MarkdownBlock: Hashable, Sendable {
    case heading(level: Int, text: String)
    case paragraph(String)
    case bullet(text: String, depth: Int, checked: Bool?)
    case numbered(number: String, text: String, depth: Int)
    case quote(String)
    case code(String)
    case rule
}

public enum Markdown {
    public static func blocks(_ source: String) -> [MarkdownBlock] {
        var out: [MarkdownBlock] = []
        var para: [String] = []
        var code: [String]? = nil
        func flush() {
            if !para.isEmpty {
                out.append(.paragraph(para.joined(separator: " ")))
                para = []
            }
        }
        for raw in source.components(separatedBy: "\n") {
            let line = raw.replacingOccurrences(of: "\t", with: "    ")
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            if code != nil {
                if trimmed.hasPrefix("```") {
                    out.append(.code(code!.joined(separator: "\n")))
                    code = nil
                } else {
                    code!.append(line)
                }
                continue
            }
            if trimmed.hasPrefix("```") {
                flush()
                code = []
                continue
            }
            if trimmed.isEmpty {
                flush()
                continue
            }
            let indent = line.prefix(while: { $0 == " " }).count
            if let h = heading(trimmed) {
                flush()
                out.append(h)
            } else if trimmed == "---" || trimmed == "***" {
                flush()
                out.append(.rule)
            } else if let b = bullet(trimmed, depth: indent / 2) {
                flush()
                out.append(b)
            } else if let n = numbered(trimmed, depth: indent / 2) {
                flush()
                out.append(n)
            } else if trimmed.hasPrefix(">") {
                flush()
                out.append(.quote(String(trimmed.dropFirst()).trimmingCharacters(in: .whitespaces)))
            } else if indent >= 2, let last = out.last, para.isEmpty {
                // A continuation line of a list item.
                switch last {
                case let .bullet(text, depth, checked):
                    out[out.count - 1] = .bullet(text: text + " " + trimmed, depth: depth, checked: checked)
                case let .numbered(number, text, depth):
                    out[out.count - 1] = .numbered(number: number, text: text + " " + trimmed, depth: depth)
                default:
                    para.append(trimmed)
                }
            } else {
                para.append(trimmed)
            }
        }
        if let code { out.append(.code(code.joined(separator: "\n"))) }
        flush()
        return out
    }

    static func heading(_ s: String) -> MarkdownBlock? {
        let hashes = s.prefix(while: { $0 == "#" }).count
        guard (1...6).contains(hashes), s.dropFirst(hashes).first == " " else { return nil }
        return .heading(level: hashes, text: String(s.dropFirst(hashes + 1)))
    }

    static func bullet(_ s: String, depth: Int) -> MarkdownBlock? {
        guard s.hasPrefix("- ") || s.hasPrefix("* ") || s.hasPrefix("+ ") else { return nil }
        var text = String(s.dropFirst(2))
        var checked: Bool? = nil
        if text.hasPrefix("[ ] ") { checked = false; text = String(text.dropFirst(4)) }
        else if text.lowercased().hasPrefix("[x] ") { checked = true; text = String(text.dropFirst(4)) }
        return .bullet(text: text, depth: depth, checked: checked)
    }

    static func numbered(_ s: String, depth: Int) -> MarkdownBlock? {
        let digits = s.prefix(while: \.isNumber)
        guard !digits.isEmpty, s.dropFirst(digits.count).hasPrefix(". ") else { return nil }
        return .numbered(number: String(digits), text: String(s.dropFirst(digits.count + 2)), depth: depth)
    }

    /// Inline markdown (emphasis, code, links) as an attributed string.
    public static func inline(_ s: String) -> AttributedString {
        (try? AttributedString(markdown: s, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))) ?? AttributedString(s)
    }
}
