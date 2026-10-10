import Foundation

/// Splits a flag line the way a shell would for plain and quoted words.
public enum Shell {
    public static func split(_ s: String) -> [String] {
        var out: [String] = []
        var cur = ""
        var quote: Character? = nil
        var any = false
        for ch in s {
            if let q = quote {
                if ch == q { quote = nil } else { cur.append(ch) }
            } else if ch == "\"" || ch == "'" {
                quote = ch
                any = true
            } else if ch.isWhitespace {
                if any || !cur.isEmpty { out.append(cur) }
                cur = ""
                any = false
            } else {
                cur.append(ch)
            }
        }
        if any || !cur.isEmpty { out.append(cur) }
        return out
    }
}
