# GoNhanh

[![CI](https://github.com/khaphanspace/gonhanh.org/actions/workflows/ci.yml/badge.svg)](https://github.com/khaphanspace/gonhanh.org/actions/workflows/ci.yml)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)

Bộ gõ tiếng Việt hiệu suất cao, hiện đại và đa nền tảng.

Sức mạnh của **Rust** (Core Engine) kết hợp với **Native UI** (SwiftUI) mang lại trải nghiệm gõ phím mượt mà, ổn định và bảo mật tuyệt đối.

## Về dự án

GoNhanh được xây dựng với mục tiêu trở thành bộ gõ tiếng Việt **hoàn thiện nhất** về trải nghiệm người dùng và hiệu năng kỹ thuật:

- **Hiệu suất vượt trội**: Core engine viết bằng Rust, xử lý tức thì, không độ trễ.
- **Đa nền tảng**: Kiến trúc Hybrid thông minh, sẵn sàng cho macOS, Windows và Linux.
- **Hiện đại**: Hỗ trợ Unicode chuẩn, logic đặt dấu ngữ âm học, không phụ thuộc bảng mã cũ.
- **Tinh tế**: Giao diện Native, không dùng Electron/Qt, tiết kiệm tài nguyên hệ thống.
- **Cài là dùng**: Cấu hình tối ưu sẵn (Smart Defaults), không cần thiết lập phức tạp.

## Động lực

Dự án lấy cảm hứng từ **EVKey** - hình mẫu về sự tối giản và hiệu quả.

Xuất phát từ nhu cầu của một lập trình viên macOS cần sự ổn định tuyệt đối trên các công cụ hiện đại (như Claude Code, Terminal), GoNhanh ra đời để tiếp nối di sản đó khi các công cụ cũ dừng cập nhật. Chúng tôi cam kết xây dựng một sản phẩm **Mã nguồn mở - Miễn phí - Của cộng đồng**.

## Ba Không

- 🚫 **Không thu phí**: Miễn phí trọn đời, không có bản "Premium".
- 🚫 **Không rác**: Không quảng cáo, không popup, không tính năng thừa thãi.
- 🚫 **Không theo dõi**: Offline 100%, không thu thập dữ liệu, mã nguồn minh bạch.

## Cam kết phát triển

- ✅ **Duy trì dài hạn**: Cam kết hỗ trợ ít nhất đến 2030.
- ✅ **Chất lượng cao**: Quy trình phát triển nghiêm ngặt (CI/CD, High Test Coverage).
- ✅ **Tôn trọng người dùng**: Lắng nghe phản hồi và fix bug nhanh chóng (48h).

## So sánh

|              | GoNhanh | OpenKey | UniKey | EVKey |
| ------------ | :-----: | :-----: | :----: | :---: |
| Năm          |  2025   |  2019   |  2000  | 2018  |
| Miễn phí     |   ✅    |   ✅    |   ✅   |  ✅   |
| Open source  |   ✅    |   ✅    |   ⚠️   |  ✅   |
| Chỉ Unicode  |   ✅    |   ❌    |   ❌   |  ❌   |
| macOS native | SwiftUI |  Obj-C  |   Qt   |  Qt   |
| Engine       |  Rust   |   C++   |  C++   |  C++  |

Nếu cần chuyển mã hay dùng bảng mã cũ, dùng UniKey/EVKey/OpenKey.

## Cách hoạt động

Engine dựa trên ngữ âm học tiếng Việt thay vì bảng tra cứu:

```
Âm tiết = [Phụ âm đầu] + [Âm đệm] + Nguyên âm chính + [Âm cuối] + Thanh điệu
```

Thuật toán đặt dấu theo quy tắc ngữ âm. Hỗ trợ cả kiểu cũ (`hoà`) và kiểu mới (`hòa`).

Chi tiết: [docs/vietnamese-language-system.md](docs/vietnamese-language-system.md)

## Kiến trúc

```
┌─────────────────────────────────────┐
│         Platform UI Layer           │
│  ┌──────────┐      ┌──────────┐    │
│  │  macOS   │      │ Windows  │    │
│  │ SwiftUI  │      │   WPF    │    │
│  └─────┬────┘      └────┬─────┘    │
└────────┼────────────────┼──────────┘
         │    FFI (C ABI) │
┌────────▼────────────────▼──────────┐
│         Rust Core Library          │
│  ┌─────────────────────────────┐   │
│  │  Engine (Telex/VNI)         │   │
│  │  - Buffer management        │   │
│  │  - Phonology-based rules    │   │
│  │  - Unicode output           │   │
│  └─────────────────────────────┘   │
└────────────────────────────────────┘
```

- macOS: SwiftUI (done)
- Windows: WPF (planned)

## Cài đặt

```bash
git clone https://github.com/khaphanspace/gonhanh.org
cd gonhanh.org
make build
cp -r platforms/macos/build/Release/GoNhanh.app /Applications/
```

Lần đầu chạy cần cấp quyền Accessibility trong System Settings.

## Phát triển

```bash
make test    # 99 tests
make build   # Build tất cả
make clean   # Xóa build artifacts
```

Xem thêm: [docs/development.md](docs/development.md) · [docs/architecture.md](docs/architecture.md)

## Roadmap

| Version | Khi nào | Gì                  |
| ------- | ------- | ------------------- |
| 0.1     | Q1 2025 | macOS beta          |
| 0.2     | Q2 2025 | Stable, auto-update |
| 0.3     | Q3 2025 | Windows             |
| 1.0     | Q4 2025 | Production          |

## Credits

- [UniKey](https://www.unikey.org/)
- [OpenKey](https://github.com/tuyenvm/OpenKey)
- [EVKey](https://evkeyvn.com/)

## License

[GPL-3.0-or-later](LICENSE)
