# ⚡ GoNhanh

[![CI](https://github.com/user/gonhanh/actions/workflows/ci.yml/badge.svg)](https://github.com/user/gonhanh/actions/workflows/ci.yml)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)

Bộ gõ tiếng Việt hiệu suất cao, native cho macOS và Windows.

## ✨ Features

| | |
|---|---|
| ⚡ **Siêu nhẹ** | ~3 MB binary |
| 🚀 **Cực nhanh** | ~25 MB RAM, khởi động 0.2s |
| 🎯 **Native UI** | SwiftUI (macOS), WPF (Windows) |
| 🦀 **Rust core** | An toàn, hiệu quả, cross-platform |
| 🔒 **Open source** | GPL-3.0 |

## 📥 Installation

### macOS

```bash
# Homebrew (coming soon)
brew install gonhanh

# Manual
# Download from Releases page
```

## 🛠 Build from source

**Prerequisites:** Rust 1.70+, Xcode 15+ (macOS)

```bash
# Clone
git clone https://github.com/user/gonhanh.org
cd gonhanh.org

# Build
./scripts/build-macos.sh

# Or build core only
cd core && cargo build --release
```

## 📁 Structure

```
gonhanh.org/
├── core/           # Rust core (cross-platform)
├── platforms/
│   ├── macos/      # SwiftUI app
│   └── windows/    # WPF (planned)
└── scripts/        # Build scripts
```

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 License

[GPL-3.0-or-later](LICENSE)
