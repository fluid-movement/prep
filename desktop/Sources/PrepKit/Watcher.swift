import Foundation

/// Follows `prep watch --json` for one project and calls onChange, debounced,
/// whenever prep reports a change. prep's events carry no paths, so the
/// receiver reloads what it shows.
public final class Watcher: @unchecked Sendable {
    private let process = Process()
    private let queue = DispatchQueue(label: "prep.watch")
    private var pending: DispatchWorkItem?
    private let debounce: TimeInterval
    private let onChange: @Sendable () -> Void

    public init(binary: URL, root: String, debounce: TimeInterval = 0.15, onChange: @escaping @Sendable () -> Void) throws {
        self.debounce = debounce
        self.onChange = onChange
        process.executableURL = binary
        process.arguments = ["watch", "--json", "--root", root]
        let out = Pipe()
        process.standardOutput = out
        process.standardError = FileHandle.nullDevice
        process.standardInput = FileHandle.nullDevice
        out.fileHandleForReading.readabilityHandler = { [weak self] h in
            let data = h.availableData
            if data.isEmpty {
                h.readabilityHandler = nil
                return
            }
            self?.changed()
        }
        try process.run()
        Watcher.lock.withLock { Watcher.running.append(Weak(self)) }
    }

    private struct Weak { weak var watcher: Watcher?; init(_ w: Watcher) { watcher = w } }
    private nonisolated(unsafe) static var running: [Weak] = []
    private static let lock = NSLock()

    /// Stops every watcher, for when the app quits. (prep watch also ends
    /// by itself when its parent is gone, which covers crashes.)
    public static func stopAll() {
        let all = lock.withLock { running.compactMap(\.watcher) }
        all.forEach { $0.stop() }
    }

    private func changed() {
        queue.async {
            self.pending?.cancel()
            let onChange = self.onChange
            let item = DispatchWorkItem { onChange() }
            self.pending = item
            self.queue.asyncAfter(deadline: .now() + self.debounce, execute: item)
        }
    }

    public func stop() {
        if process.isRunning { process.terminate() }
    }

    deinit { stop() }
}
