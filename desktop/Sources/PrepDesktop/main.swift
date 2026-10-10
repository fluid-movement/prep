import AppKit

let launch = LaunchOptions.parse(CommandLine.arguments)
if let k = CommandLine.arguments.firstIndex(of: "--render-icon"), k + 1 < CommandLine.arguments.count {
    MainActor.assumeIsolated { IconRenderer.render(to: CommandLine.arguments[k + 1]) }
} else if launch.snapshot != nil {
    MainActor.assumeIsolated { Snapshot.run(launch) }
} else {
    PrepApp.main()
}
