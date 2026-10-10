import AppKit
import PrepKit
import UserNotifications

/// macOS notifications for what newly needs the person, and the Dock badge.
/// Notifications need a bundle, so a bare `swift run` build only badges.
@MainActor
final class Notifier {
    static let shared = Notifier()
    private var authorized = false
    private var asked = false

    private var center: UNUserNotificationCenter? {
        Bundle.main.bundleIdentifier == nil ? nil : UNUserNotificationCenter.current()
    }

    func badge(_ count: Int) {
        NSApp?.dockTile.badgeLabel = count > 0 ? String(count) : nil
    }

    func announce(_ items: [InboxItem], project: String) {
        // While the window is in front the inbox itself shows it.
        guard !items.isEmpty, NSApp?.isActive != true, let center else { return }
        Task {
            if !asked {
                asked = true
                authorized = (try? await center.requestAuthorization(options: [.alert, .sound, .badge])) ?? false
            }
            guard authorized else { return }
            for item in items.prefix(5) {
                let content = UNMutableNotificationContent()
                content.title = item.kind == .review ? "Done: \(item.issue.title)" : item.issue.title
                content.subtitle = project
                content.body = item.kind == .review
                    ? "\(item.issue.resolvedBy ?? "An agent") finished it. Review the work."
                    : item.kind.reason
                content.sound = .default
                content.userInfo = ["issue": item.issue.id]
                try? await center.add(UNNotificationRequest(identifier: item.id, content: content, trigger: nil))
            }
        }
    }
}
