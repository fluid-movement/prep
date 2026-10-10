import AppKit

let launch = LaunchOptions.parse(CommandLine.arguments)
if launch.snapshot != nil {
    MainActor.assumeIsolated { Snapshot.run(launch) }
} else {
    PrepApp.main()
}
