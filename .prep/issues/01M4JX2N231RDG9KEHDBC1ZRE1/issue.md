---
title: 'Desktop app: Developer ID signing, notarization and a Homebrew cask'
kind: code
parent: 01M4JVP89F3PGDHYETHEPS3YBB
depends_on:
  - 01M4869P6R226F3GCQKAEC5F8J
tags:
  - desktop
---

Release builds of the desktop app are signed with a Developer ID, notarized and stapled, so they open without Gatekeeper warnings, and ship as a Homebrew cask in the tap next to the prep formula. The release workflow builds the app on a macOS runner with the signing certificate and notary credentials as secrets.

## Open questions

- The user's Apple Developer account (needed for the certificate and notarization).
