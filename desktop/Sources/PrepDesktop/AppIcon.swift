import AppKit
import SwiftUI

/// The app icon, drawn in SwiftUI: a macOS squircle with a checklist whose
/// last item is still open — work tracked, with the next step waiting.
/// `PrepDesktop --render-icon <dir>` writes the iconset that `just desktop`
/// turns into AppIcon.icns.
struct AppIconView: View {
    var body: some View {
        GeometryReader { g in
            let s = g.size.width
            ZStack {
                // macOS icons: an 824 pt body on a 1024 pt canvas, with a soft shadow.
                RoundedRectangle(cornerRadius: s * 0.185, style: .continuous)
                    .fill(LinearGradient(colors: [Color(red: 0.36, green: 0.42, blue: 0.98), Color(red: 0.18, green: 0.16, blue: 0.62)],
                                         startPoint: .top, endPoint: .bottom))
                    .overlay(
                        RoundedRectangle(cornerRadius: s * 0.185, style: .continuous)
                            .strokeBorder(.white.opacity(0.18), lineWidth: s * 0.004)
                    )
                    .frame(width: s * 0.805, height: s * 0.805)
                    .shadow(color: .black.opacity(0.35), radius: s * 0.02, y: s * 0.012)
                VStack(alignment: .leading, spacing: s * 0.05) {
                    row(s, done: true, width: 0.30)
                    row(s, done: true, width: 0.24)
                    row(s, done: false, width: 0.28)
                }
                .offset(x: -s * 0.01)
            }
            .frame(width: s, height: s)
        }
        .aspectRatio(1, contentMode: .fit)
    }

    func row(_ s: CGFloat, done: Bool, width: CGFloat) -> some View {
        HStack(spacing: s * 0.05) {
            ZStack {
                Circle()
                    .fill(done ? Color.white : Color.clear)
                    .overlay(Circle().strokeBorder(.white, style: StrokeStyle(lineWidth: s * 0.018, dash: done ? [] : [s * 0.03, s * 0.022])))
                if done {
                    Image(systemName: "checkmark")
                        .font(.system(size: s * 0.06, weight: .black))
                        .foregroundStyle(Color(red: 0.25, green: 0.27, blue: 0.85))
                }
            }
            .frame(width: s * 0.11, height: s * 0.11)
            Capsule()
                .fill(.white.opacity(done ? 0.55 : 0.95))
                .frame(width: s * width, height: s * 0.045)
        }
    }
}

enum IconRenderer {
    /// Writes icon_<n>x<n>[@2x].png files for iconutil.
    @MainActor
    static func render(to dir: String) -> Never {
        let fm = FileManager.default
        try? fm.createDirectory(atPath: dir, withIntermediateDirectories: true)
        for base in [16, 32, 128, 256, 512] {
            for scale in [1, 2] {
                let px = base * scale
                let renderer = ImageRenderer(content: AppIconView().frame(width: CGFloat(px), height: CGFloat(px)))
                renderer.scale = 1
                guard let cg = renderer.cgImage else { continue }
                let rep = NSBitmapImageRep(cgImage: cg)
                let name = scale == 1 ? "icon_\(base)x\(base).png" : "icon_\(base)x\(base)@2x.png"
                try? rep.representation(using: .png, properties: [:])?.write(to: URL(fileURLWithPath: dir).appendingPathComponent(name))
            }
        }
        print(dir)
        exit(0)
    }
}
