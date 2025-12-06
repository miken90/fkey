# ⚡ GoNhanh

Bộ gõ tiếng Việt hiệu suất cao, native cho macOS và Windows.

## Features

- ⚡ Siêu nhẹ (~3 MB)
- 🚀 Cực nhanh (~25 MB RAM)
- 🎯 Native macOS SwiftUI
- 🦀 Rust core - an toàn & hiệu quả
- 🔒 Open source - GPL-3.0

## 📁 Structure

```
gonhanh.org/
├── core/           # Rust core engine (cross-platform)
├── platforms/      # Platform-specific apps
│   ├── macos/      # macOS (SwiftUI)
│   └── windows/    # Windows (WPF) - coming soon
└── scripts/        # Build scripts
```

## 🚀 Build

### macOS
```bash
./scripts/build-macos.sh
```

### Core only
```bash
cd core && cargo build --release
```

## 📊 Metrics

- Binary: ~3 MB
- RAM: ~25 MB
- Startup: ~0.2s

## 🛠 Tech Stack

- **Core**: Rust (rdev, enigo)
- **macOS**: SwiftUI
- **Windows**: WPF/WinUI3 (planned)

## 📄 License

GPL-3.0-or-later
