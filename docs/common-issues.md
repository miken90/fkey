# Các lỗi thường gặp với bộ gõ Tiếng Việt & Cách khắc phục

Tài liệu này tổng hợp các vấn đề phổ biến nhất mà người dùng gặp phải khi sử dụng các bộ gõ tiếng Việt (UniKey, EVKey, OpenKey, GoNhanh...) và cách xử lý triệt để.

---

## Mục lục

1. [Lỗi chung (Mọi bộ gõ)](#1-lỗi-chung-mọi-bộ-gõ)
2. [Lỗi trên Trình duyệt (Chrome/Edge/Safari)](#2-lỗi-trên-trình-duyệt)
3. [Lỗi trên Microsoft Office (Word/Excel)](#3-lỗi-trên-microsoft-office)
4. [Lỗi trên Terminal & IDE](#4-lỗi-trên-terminal--ide)
5. [Lỗi đặc thù từng bộ gõ](#5-lỗi-đặc-thù-từng-bộ-gõ)
6. [Lời khuyên](#6-lời-khuyên)

---

## 1. Lỗi chung (Mọi bộ gõ)

### 1.1. Gõ ra số thay vì chữ (ă -> a8, ô -> o6)

- **Nguyên nhân**:
  - Đang để chế độ gõ sai (Ví dụ: muốn gõ Telex nhưng đang chọn VNI hoặc ngược lại).
  - Kích hoạt nhầm **NumLock** trên bàn phím laptop nhỏ (biến một phần phím chữ thành số).
- **Khắc phục**:
  - Kiểm tra cài đặt "Kiểu gõ" trong bộ gõ.
  - Tắt `NumLock` (thường là `Fn + NumLock`).

### 1.2. Mất chữ V, E, O khi gõ (gõ `v` ra `d`, `o` ra `ô`)

- **Nguyên nhân**: Cài đặt sai bảng mã hoặc font chữ không hỗ trợ.
- **Khắc phục**:
  - Luôn dùng Bảng mã **Unicode**.
  - Dùng các font phổ biến như **Times New Roman, Arial, Roboto, Open Sans**.

### 1.3. Gõ tiếng Việt bị gạch chân hoặc hiện khung (macOS)

- **Nguyên nhân**: Tính năng đoán từ/sửa lỗi của macOS đang bật.
- **Khắc phục**:
  - Vào `System Settings` -> `Keyboard` -> `Text Input`.
  - Tắt "Correct spelling automatically".

### 1.4. Nhảy ký tự hoặc đảo ngược từ

- **Nguyên nhân**: Xung đột giữa nhiều bộ gõ cùng chạy một lúc (Ví dụ: Cả UniKey và bộ gõ mặc định của Windows/macOS cùng bật).
- **Khắc phục**:
  - Chỉ giữ **MỘT** bộ gõ duy nhất hoạt động.
  - Trên Windows: Gỡ bỏ bàn phím tiếng Việt mặc định (`Vietnamese Keyboard`) trong `Language Settings`, chỉ để `English (US)`.
  - Trên macOS: Chuyển bộ gõ mặc định về `ABC`, chỉ dùng bộ gõ ngoài (như GoNhanh, OpenKey).

---

## 2. Lỗi trên Trình duyệt (Chrome/Edge/Safari/Firefox)

### 2.1. Mất chữ / Dính chữ trên thanh địa chỉ (Omnibox) - Chrome/Edge Chromium

- **Hiện tượng**:
  - Gõ `tiếng việt` -> `tiêng viêt` (mất dấu).
  - Gõ `aa` -> `aâ` (dính chữ, thay vì `â`).
  - Gõ `dd` -> `dđ` (dính chữ).
- **Nguyên nhân**: Cơ chế **Suggestion/Autocomplete** của trình duyệt xung đột với bộ gõ khi đang bỏ dấu.

* **Khắc phục**:
  - **Cách 1 (Khuyên dùng)**: Dùng phím tắt `Ctrl+K` (hoặc `Cmd+L` rồi `Tab`) để tìm kiếm bằng Google Search thay vì gõ trực tiếp URL.
  - **Cách 2**: Tắt gợi ý trong Settings:
    - Chrome: `Settings` -> `You and Google` -> `Sync and Google services` -> tắt **Autocomplete searches and URLs**.
    - Edge: `Settings` -> `Privacy, search, and services` -> `Address bar and search` -> tắt **Search and site suggestions**.
  - **Cách 3**: Dùng **GoNhanh** hoặc **EVKey** (có tính năng "Sửa lỗi Chrome/Excel").
  - **Cách 4**: Xóa lịch sử duyệt web để loại bỏ các gợi ý cũ gây lỗi.

### 2.2. Lỗi gõ trên Facebook Messenger (Safari macOS)

- **Hiện tượng**: Gõ xong Enter bị mất chữ, hoặc nhảy ký tự lung tung khi chat.
- **Nguyên nhân**: Safari xử lý `ContentEditable` khác với Chrome, cộng thêm cơ chế Text Replacement của macOS.
- **Khắc phục**:
  - Khởi động lại Safari.
  - Tắt **Auto-correct**: `System Settings` -> `Keyboard` -> `Text Input` -> `Edit` -> tắt **Correct spelling automatically**.
  - Dùng trình duyệt khác (Chrome/Firefox) cho Facebook nếu lỗi quá nặng.

### 2.3. Lỗi mất chữ trên Firefox

- **Hiện tượng**: Gõ chữ có dấu bị biến mất hoặc hiển thị ô vuông.
- **Nguyên nhân**: Cấu hình font hoặc Hardware Acceleration.
- **Khắc phục**:
  - Tắt **Hardware Acceleration**: `Settings` -> `General` -> `Performance` -> bỏ chọn `Use recommended performance settings` & `Use hardware acceleration`.
  - Chỉnh font: `Settings` -> `General` -> `Fonts` -> `Advanced` -> chọn **Other Writing Systems** -> Fallback Text Encoding là **Vietnamese**.

### 2.4. Google Docs/Sheets bị lặp chữ, mất dấu

- **Hiện tượng**: Gõ `được` -> `đđược`, hoặc gõ `trường` -> `trương`.
- **Nguyên nhân**: Google Docs dùng cơ chế Canvas hoặc xử lý input riêng, không đồng bộ kịp với IME.
- **Khắc phục**:
  - Tắt các Extension (Tiện ích) không cần thiết (chúng có thể làm chậm DOM).
  - Với **EVKey/OpenKey**: Bật tính năng "Sửa lỗi Google Docs".
  - Trong Google Docs: `File` -> `Language` -> Chọn **Tiếng Việt**.

---

## 3. Lỗi trên Microsoft Office (Word/Excel)

### 3.1. Tự động đổi chữ, sửa lỗi không mong muốn

- **Nguyên nhân**: Tính năng **AutoCorrect** của Word/Excel. Ví dụ gõ `i` thành `I`, hoặc `adn` thành `and`.
- **Khắc phục**:
  - Vào `File` -> `Options` -> `Proofing` -> `AutoCorrect Options`.
  - Bỏ chọn các mục tự động sửa (đặc biệt là `Capitalize first letter...` nếu không muốn).

### 3.2. Mất chữ khi chèn thêm vào giữa từ

- **Nguyên nhân**: Phím **Insert** bị kích hoạt (Chế độ `Overtype`).
- **Khắc phục**: Nhấn phím `Insert` trên bàn phím một lần để tắt.

---

## 4. Lỗi trên Terminal & IDE (VS Code, IntelliJ, Terminal)

### 4.1. Không gõ được tiếng Việt trong Terminal

- **Nguyên nhân**: Một số Terminal (đặc biệt trên Windows/Linux) không hỗ trợ IME composition window (cửa sổ gõ).
- **Khắc phục**:
  - Dùng Terminal hiện đại: **Windows Terminal**, **iTerm2** (macOS), **Alacritty**.
  - Trên Windows: Dùng **GoNhanh** hoặc **EVKey** (các bộ gõ này xử lý tốt việc gửi ký tự unicode trực tiếp thay vì qua IME cũ).

### 4.2. Gõ bị nhân đôi chữ trong VS Code/IntelliJ

- **Hiện tượng**: Gõ `được` -> `đđược`.
- **Nguyên nhân**: Xung đột giữa plugin Vim (nếu cài) hoặc cơ chế Auto-complete của IDE.
- **Khắc phục**:
  - Nếu dùng Vim extension: Đảm bảo tắt IME khi thoát Insert Mode (cần config thêm).
  - Restart IDE.

---

## 5. Lỗi đặc thù từng bộ gõ

### 5.1. UniKey

- **Không chạy được với quyền Admin**: Khi bật Task Manager hoặc Regedit, không gõ được tiếng Việt.
  - -> **Fix**: Chuột phải vào UniKey -> **Run as Administrator**.
- **Lỗi bảng mã**: Gõ ra ký tự lạ.
  - -> **Fix**: Nhấn `Ctrl+Shift+F5`, đảm bảo chọn **Unicode dựng sẵn**.

### 5.2. OpenKey (macOS)

- **Mất quyền Accessibility**: Gõ không ra dấu dù đã bật.
  - -> **Fix**: Vào `System Settings` -> `Privacy & Security` -> `Accessibility`. Xóa OpenKey đi và thêm lại (nút `-` rồi nút `+`).
- **Secure Input**: Bị app khác chiếm quyền nhập liệu (ví dụ password field).
  - -> **Fix**: Tắt app đang chiếm quyền hoặc Log out/Log in lại.

### 5.3. IBus-Bamboo (Linux)

- **Không gõ được trong ứng dụng Qt/KDE**.
  - -> **Fix**: Cần set biến môi trường:
    ```bash
    export GTK_IM_MODULE=ibis
    export QT_IM_MODULE=ibus
    export XMODIFIERS=@im=ibus
    ```

---

## 6. Lời khuyên để gõ ổn định

1.  **Dùng Smart Defaults**: Đa số người dùng chỉ cần **Unicode** + **Telex/VNI**. Đừng chỉnh sang các bảng mã lạ (như TCVN3) trừ khi bắt buộc.
2.  **Một vua một cõi**: Trên 1 máy tính chỉ nên bật **1 bộ gõ duy nhất**.
3.  **Quyền ưu tiên**: Trên Windows, nên chạy bộ gõ với quyền Administrator nếu bạn hay làm việc với các phần mềm hệ thống. Trên macOS, luôn cấp quyền Accessibility.
4.  **Cập nhật**: Luôn dùng phiên bản mới nhất của bộ gõ. GoNhanh ra đời để giải quyết tận gốc các vấn đề về hiệu suất và tính ổn định kể trên bằng công nghệ mới.
