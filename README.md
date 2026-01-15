<h1 align="center">
  <img src="assets/logo.png" alt="FKey Logo" width="128" height="128"><br>
  FKey
</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Windows-0078D6?logo=windows&logoColor=white" />
  <img src="https://img.shields.io/badge/License-BSD--3--Clause-blue.svg" alt="License: BSD-3-Clause">
  <img src="https://img.shields.io/github/v/release/miken90/fkey" alt="Release">
</p>

<p align="center">
  <strong>Bộ gõ tiếng Việt miễn phí, nhanh, nhẹ cho Windows.</strong><br>
  ~5MB · Không cần cài đặt · Không quảng cáo · Không thu thập dữ liệu
</p>

---

## 📥 Tải về

| Nền tảng | Tải xuống | Kích thước |
|:--------:|:---------:|:----------:|
| **Windows** | [📥 FKey-portable.zip](https://github.com/miken90/fkey/releases/latest) | ~5 MB |

### Cài đặt

1. Tải và giải nén `FKey-vX.X.X-portable.zip`
2. Chạy `FKey.exe`
3. App chạy trong system tray (khay hệ thống)

---

## ✨ Tính năng

### 🔥 Highlight

| Tính năng | Mô tả |
|-----------|-------|
| ⚡ **Siêu nhẹ** | ~5MB portable, ~10MB RAM |
| 🔍 **Mọi app** | Chrome, VS Code, Terminal, Discord, Slack... |
| 🔤 **Auto-restore tiếng Anh** | `text` `expect` `user` → tự khôi phục khi nhấn Space |
| ⎋ **ESC khôi phục** | Gõ sai → nhấn ESC → về lại chữ gốc |
| 🔠 **Tự viết hoa** | Đầu câu tự động viết hoa |

### 📋 Đầy đủ

- ⌨️ **Telex & VNI** — Chọn kiểu gõ quen thuộc
- 🎯 **Đặt dấu chuẩn** — `hoà`, `khoẻ`, `thuỷ`
- ✂️ **Gõ tắt** — `vn` → `Việt Nam`
- 🚀 **Auto-start** — Khởi động cùng Windows
- 🔧 **Phím tắt tùy chỉnh** — Ctrl+Space hoặc tuỳ ý

### 🛡️ Cam kết

- 🚫 **Không thu phí** — Miễn phí mãi mãi
- 🚫 **Không quảng cáo** — Không popup
- 🚫 **Không theo dõi** — Offline 100%, mã nguồn mở

---

## 🔧 Dành cho Developer

### Tech Stack

| Layer | Công nghệ |
|-------|-----------|
| **Core Engine** | Rust (zero dependencies) |
| **Windows App** | Go + Wails v3 + WebView2 |
| **Testing** | 700+ tests |

### Build

```powershell
# Build Rust core
cd core
cargo build --release

# Build Windows app
cd platforms/windows-wails
.\build.ps1 -Release -Version "2.0.0"
```

### Test

```powershell
# Rust tests
cd core
cargo test

# Go tests
cd platforms/windows-wails
go test ./...
```

---

## 🙏 Lời cảm ơn

FKey được phát triển dựa trên nền tảng của dự án **[Gõ Nhanh](https://github.com/khaphanspace/gonhanh.org)** bởi **Kha Phan**.

Cảm ơn Kha Phan và cộng đồng Gõ Nhanh đã tạo ra engine xử lý tiếng Việt tuyệt vời. FKey tiếp nối sứ mệnh mang đến bộ gõ chất lượng cao, miễn phí cho người Việt.

---

## 📄 License

[BSD-3-Clause](LICENSE)
