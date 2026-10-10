// swift-tools-version: 6.0
// The prep desktop app for macOS. PrepKit talks to the prep binary through
// its JSON output; PrepDesktop is the SwiftUI app on top. Build the .app
// with `just desktop` from the repository root.
import PackageDescription

let package = Package(
    name: "PrepDesktop",
    platforms: [.macOS(.v14)],
    products: [
        .executable(name: "PrepDesktop", targets: ["PrepDesktop"]),
        .library(name: "PrepKit", targets: ["PrepKit"]),
    ],
    targets: [
        .target(name: "PrepKit"),
        .executableTarget(name: "PrepDesktop", dependencies: ["PrepKit"]),
        .testTarget(name: "PrepKitTests", dependencies: ["PrepKit"]),
    ],
    swiftLanguageModes: [.v5]
)
