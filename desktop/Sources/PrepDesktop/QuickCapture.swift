import AppKit
import Carbon.HIToolbox
import SwiftUI
import PrepKit

/// A system-wide hotkey. Carbon's hotkey API needs no Accessibility
/// permission, unlike watching every key press.
final class HotKey {
    private var ref: EventHotKeyRef?
    private static var handlers: [UInt32: () -> Void] = [:]
    private static var installed = false

    init(keyCode: Int, modifiers: Int, id: UInt32, handler: @escaping () -> Void) {
        HotKey.handlers[id] = handler
        if !HotKey.installed {
            HotKey.installed = true
            var spec = EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))
            InstallEventHandler(GetApplicationEventTarget(), { _, event, _ in
                var hk = EventHotKeyID()
                GetEventParameter(event, EventParamName(kEventParamDirectObject), EventParamType(typeEventHotKeyID),
                                  nil, MemoryLayout<EventHotKeyID>.size, nil, &hk)
                let id = hk.id
                DispatchQueue.main.async { HotKey.handlers[id]?() }
                return noErr
            }, 1, &spec, nil, nil)
        }
        let hkID = EventHotKeyID(signature: OSType(0x50524550), id: id) // 'PREP'
        RegisterEventHotKey(UInt32(keyCode), UInt32(modifiers), hkID, GetApplicationEventTarget(), 0, &ref)
    }

    deinit {
        if let ref { UnregisterEventHotKey(ref) }
    }
}

/// The floating Quick Capture panel: a title, Enter, done.
@MainActor
final class QuickCapture {
    static let shared = QuickCapture()
    private var panel: NSPanel?
    private var hotKey: HotKey?

    /// ⌃⌥N, from anywhere.
    func registerHotKey() {
        hotKey = HotKey(keyCode: kVK_ANSI_N, modifiers: controlKey | optionKey, id: 1) { [weak self] in
            self?.show()
        }
    }

    func show() {
        if panel == nil {
            let p = NSPanel(contentRect: NSRect(x: 0, y: 0, width: 560, height: 160),
                            styleMask: [.titled, .closable, .fullSizeContentView],
                            backing: .buffered, defer: false)
            p.titleVisibility = .hidden
            p.titlebarAppearsTransparent = true
            p.isMovableByWindowBackground = true
            p.level = .floating
            p.hidesOnDeactivate = false
            p.isReleasedWhenClosed = false
            p.standardWindowButton(.miniaturizeButton)?.isHidden = true
            p.standardWindowButton(.zoomButton)?.isHidden = true
            panel = p
        }
        guard let panel else { return }
        panel.contentView = NSHostingView(rootView: QuickCaptureView { [weak self] in self?.close() })
        panel.setContentSize(panel.contentView!.fittingSize)
        panel.center()
        NSApp.activate(ignoringOtherApps: true)
        panel.makeKeyAndOrderFront(nil)
    }

    func close() { panel?.orderOut(nil) }
}

struct QuickCaptureView: View {
    var close: () -> Void

    @State private var projects: [ProjectRef] = []
    @State private var root: String = UserDefaults.standard.string(forKey: "lastProject") ?? ""
    @State private var parents: [IssueSummary] = []
    @State private var parent: String?
    @State private var title = ""
    @State private var kind = "code"
    @State private var status: String?
    @State private var busy = false
    @FocusState private var focused: Bool

    var cli: PrepCLI? {
        guard let bin = PrepCLI.locate(), !root.isEmpty else { return nil }
        return PrepCLI(binary: bin, root: root)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 10) {
                Image(systemName: "plus.circle.fill").font(.title2).foregroundStyle(Color.accentColor)
                TextField("New issue in \(projectName)…", text: $title)
                    .textFieldStyle(.plain)
                    .font(.title2)
                    .focused($focused)
                    .onSubmit { create(open: false) }
            }
            HStack(spacing: 8) {
                Menu {
                    ForEach(projects) { p in Button(p.name) { root = p.root } }
                } label: {
                    Label(projectName, systemImage: "folder")
                }
                .menuStyle(.borderlessButton)
                .fixedSize()
                Menu {
                    Button("No Parent") { parent = nil }
                    Divider()
                    ForEach(parents) { p in Button(p.title) { parent = p.id } }
                } label: {
                    Label(parent.flatMap { id in parents.first { $0.id == id }?.title } ?? "No parent", systemImage: "list.bullet.indent")
                        .lineLimit(1)
                }
                .menuStyle(.borderlessButton)
                .frame(maxWidth: 220, alignment: .leading)
                .fixedSize(horizontal: true, vertical: false)
                Menu {
                    ForEach(Lifecycle.kinds, id: \.self) { k in
                        Button { kind = k } label: { Label(k.capitalized, systemImage: Style.symbol(kind: k)) }
                    }
                } label: {
                    Label(kind.capitalized, systemImage: Style.symbol(kind: kind))
                }
                .menuStyle(.borderlessButton)
                .fixedSize()
                Spacer()
                if let status {
                    Text(status).font(.caption).foregroundStyle(.secondary)
                } else {
                    Text("↩ Create   ⌘↩ Create & Open   esc Close").font(.caption).foregroundStyle(.tertiary)
                }
            }
            .font(.callout)
        }
        .padding(.horizontal, 18)
        .padding(.top, 26)
        .padding(.bottom, 14)
        .frame(width: 560)
        .background {
            Button("") { create(open: true) }.keyboardShortcut(.return, modifiers: .command).hidden()
            Button("") { close() }.keyboardShortcut(.cancelAction).hidden()
        }
        .disabled(busy)
        .task {
            focused = true
            if let bin = PrepCLI.locate() {
                projects = (try? await PrepCLI(binary: bin, root: nil).projects()) ?? []
                if root.isEmpty { root = projects.first?.root ?? "" }
            }
        }
        .task(id: root) {
            parent = nil
            parents = ((try? await cli?.list(["--state", "open,defined,ready,in-progress"])) ?? [])
                .filter { ($0.children ?? 0) > 0 || $0.kind == "manual" }
        }
    }

    var projectName: String {
        projects.first { $0.root == root }?.name ?? URL(fileURLWithPath: root).lastPathComponent
    }

    func create(open: Bool) {
        let t = title.trimmingCharacters(in: .whitespaces)
        guard !t.isEmpty, let cli else { return }
        busy = true
        Task {
            defer { busy = false }
            do {
                let id = try await cli.create(title: t, kind: kind, parent: parent)
                title = ""
                status = "Created \(id.suffix(6))"
                if open {
                    NotificationCenter.default.post(name: .openIssue, object: nil, userInfo: ["issue": id, "root": root])
                    close()
                } else {
                    try? await Task.sleep(for: .milliseconds(700))
                    status = nil
                    close()
                }
            } catch {
                status = error.localizedDescription
            }
        }
    }
}
