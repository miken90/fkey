# Architecture

The canonical architecture document for this repository is
[`system-architecture.md`](./system-architecture.md).

It covers the two-layer design (Rust core engine + Go/Wails v3 Windows
platform), the C-ABI FFI bridge, keystroke/settings/update data flows, text
injection methods, resilience mechanisms (smart profile caching, panic
recovery, `goSafe`), and remote-desktop passthrough.

For the module map, directory tree, and stats, see
[`codebase-summary.md`](./codebase-summary.md).
