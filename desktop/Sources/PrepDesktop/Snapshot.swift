import AppKit
import SwiftUI
import PrepKit

/// Renders a window offscreen into a PNG and exits, so an agent can see the
/// app without a person, a display grab or extra macOS permissions:
///
///     PrepDesktop <project> --snapshot out.png [--select <id>] [--tab knowledge --entry /overview.md] [--size 1280x820] [--appearance dark]
///
/// The window is real (NSHostingView in an NSWindow, never ordered front),
/// loaded through the prep CLI like a normal window, and drawn with
/// cacheDisplay, which captures AppKit-backed controls that ImageRenderer
/// leaves blank.
@MainActor
enum Snapshot {
    static func run(_ o: LaunchOptions) -> Never {
        let app = NSApplication.shared
        app.setActivationPolicy(.prohibited)
        // One appearance for the app and the window, so the stand-in
        // panels and the views resolve their colors alike.
        let dark = o.appearance.map { $0 == "dark" }
            ?? (UserDefaults.standard.string(forKey: "AppleInterfaceStyle") == "Dark")
        let appearance = NSAppearance(named: dark ? .darkAqua : .aqua)
        app.appearance = appearance
        let model = ProjectModel()
        let window = NSWindow(contentRect: NSRect(origin: .zero, size: o.size),
                              styleMask: [.titled, .closable, .resizable, .miniaturizable, .fullSizeContentView],
                              backing: .buffered, defer: false)
        window.appearance = appearance
        let host = NSHostingView(rootView: AnyView(ContentView(model: model)))
        window.contentView = host
        window.setContentSize(o.size)
        // On the window server but off every screen: lists tile their rows
        // and the window is never seen.
        window.setFrameOrigin(NSPoint(x: -30000, y: -30000))
        window.orderFrontRegardless()

        Task { @MainActor in
            guard let project = o.project else { fail("snapshot needs a project directory") }
            await model.open(path: project)
            model.tab = o.tab
            if let id = o.select {
                let full = model.all.keys.first { $0.hasSuffix(id.uppercased()) } ?? id
                model.selection = full
                await model.loadDetail()
            }
            if let e = o.entry {
                model.entrySelection = e
                await model.loadEntry()
            }
            // --sheet renders a sheet's view on its own, at its own size.
            if let name = o.sheet {
                let sheet: AnyView
                switch name {
                case "new": sheet = AnyView(NewIssueSheet(model: model, draft: NewIssueDraft(parent: model.current)))
                case "capture": sheet = AnyView(QuickCaptureView {})
                case "drop", "complete", "release":
                    let t = model.guide?.transitions?.first { $0.op == name }
                        ?? PrepKit.Transition(op: name, command: "prep \(name) <id>", allowed: true)
                    sheet = AnyView(TransitionSheet(transition: t, issueTitle: model.detail?.issue.title ?? "", entries: model.entries) { _ in })
                default:
                    guard let target = EditTarget(rawValue: name) else { fail("unknown sheet \(name)") }
                    sheet = AnyView(EditorSheet(model: model, target: target))
                }
                host.rootView = AnyView(sheet.background(.background))
                let size = host.fittingSize
                window.setContentSize(NSSize(width: max(size.width, 300), height: max(size.height, 200)))
            }
            // --wait gives views that load on their own (diffs) time.
            if o.wait > 0 { try? await Task.sleep(for: .milliseconds(o.wait)) }
            // Let SwiftUI lay out the loaded state.
            for _ in 0..<5 {
                try? await Task.sleep(for: .milliseconds(150))
                host.layoutSubtreeIfNeeded()
            }
            if ProcessInfo.processInfo.environment["PREP_SNAPSHOT_TREE"] != nil, let root = window.contentView?.superview {
                dump(root, depth: 0)
            }
            write(window: window, to: o.snapshot!)
            if let e = model.error { FileHandle.standardError.write(Data("model error: \(e)\n".utf8)) }
            Watcher.stopAll()
            exit(0)
        }
        app.run()
        exit(0)
    }

    static func write(window: NSWindow, to path: String) {
        guard let frame = window.contentView?.superview else { fail("no view") }
        frame.layoutSubtreeIfNeeded()
        guard let rep = frame.bitmapImageRepForCachingDisplay(in: frame.bounds) else { fail("no bitmap") }
        frame.cacheDisplay(in: frame.bounds, to: rep)
        // Liquid Glass (macOS 26) is composited by the window server, so
        // cacheDisplay draws each glass panel as a blank shape and skips what
        // it holds. Paint a flat stand-in for every panel, then draw the
        // panel's own content views on top, parents before nested panels.
        guard let ctx = NSGraphicsContext(bitmapImageRep: rep) else { fail("no context") }
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = ctx
        frame.effectiveAppearance.performAsCurrentDrawingAppearance {
            let all = descendants(of: frame)
            let glasses = all.filter(isGlass)
            let inGlass = Set(glasses.flatMap(descendants(of:)).map(ObjectIdentifier.init))
            // Scroll views with an edge effect skip their rows when cached as
            // a whole: clear each list to its background, then draw its
            // visible rows one by one. Lists outside glass go first, since
            // they extend under the floating panels.
            @MainActor func lists(inside: Bool) {
                for case let clip as NSClipView in all where inGlass.contains(ObjectIdentifier(clip)) == inside {
                    guard let doc = clip.documentView else { continue }
                    let rows = descendants(of: doc).filter(isRow)
                    let rect = clip.convert(clip.bounds, to: frame)
                    // Lists sit on the control background, documents on the window's.
                    (rows.isEmpty ? NSColor.windowBackgroundColor : NSColor.controlBackgroundColor).setFill()
                    rect.fill()
                    NSGraphicsContext.saveGraphicsState()
                    NSBezierPath(rect: rect).addClip()
                    if rows.isEmpty {
                        draw(doc, into: frame, rect: doc.visibleRect)
                    } else {
                        for row in rows { draw(row, into: frame, rect: row.visibleRect) }
                    }
                    NSGraphicsContext.restoreGraphicsState()
                }
            }
            lists(inside: false)
            for glass in glasses {
                let rect = glass.convert(glass.bounds, to: frame)
                let radius = min(rect.height / 2, 16)
                NSColor.controlBackgroundColor.setFill()
                NSBezierPath(roundedRect: rect, xRadius: radius, yRadius: radius).fill()
                NSColor.separatorColor.setStroke()
                NSBezierPath(roundedRect: rect.insetBy(dx: 0.5, dy: 0.5), xRadius: radius, yRadius: radius).stroke()
                for holder in glass.subviews where "\(type(of: holder))".contains("ContentHolder") {
                    for content in holder.subviews where !content.isHidden {
                        draw(content, into: frame)
                    }
                }
            }
            lists(inside: true)
            // Toolbar items and the title sit beside their glass platters.
            for item in all where ["NSToolbarItemViewer", "NSToolbarTitleView"].contains(where: "\(type(of: item))".hasPrefix) {
                draw(item, into: frame)
            }
        }
        NSGraphicsContext.restoreGraphicsState()
        guard let png = rep.representation(using: .png, properties: [:]) else { fail("no png") }
        do {
            try png.write(to: URL(fileURLWithPath: path))
            print(path)
        } catch {
            fail("\(error)")
        }
    }

    static func isRow(_ v: NSView) -> Bool {
        "\(type(of: v))".contains("TableRowView")
    }

    static func isGlass(_ v: NSView) -> Bool {
        "\(type(of: v))".contains("GlassEffectView")
    }

    static func descendants(of v: NSView) -> [NSView] {
        v.subviews.flatMap { [$0] + descendants(of: $0) }
    }

    /// Draws view at its place in frame's coordinates.
    static func draw(_ view: NSView, into frame: NSView, rect: NSRect? = nil) {
        let rect = rect ?? view.bounds
        guard rect.width >= 1, rect.height >= 1,
              let r = view.bitmapImageRepForCachingDisplay(in: rect) else { return }
        view.cacheDisplay(in: rect, to: r)
        let image = NSImage(size: rect.size)
        image.addRepresentation(r)
        image.draw(in: view.convert(rect, to: frame), from: .zero, operation: .sourceOver, fraction: 1,
                   respectFlipped: true, hints: nil)
    }

    static func dump(_ v: NSView, depth: Int) {
        if depth > 14 { return }
        print(String(repeating: "  ", count: depth) + "\(type(of: v)) \(v.frame.integral) layer=\(v.layer.map { "\(type(of: $0))" } ?? "-")")
        for s in v.subviews { dump(s, depth: depth + 1) }
    }

    static func fail(_ msg: String) -> Never {
        FileHandle.standardError.write(Data("snapshot: \(msg)\n".utf8))
        exit(1)
    }
}
