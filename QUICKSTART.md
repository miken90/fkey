# ⚡ GoNhanh - Quick Start

## 🎯 What You Have

A complete **Rust + SwiftUI** project structure for a Vietnamese input method:

- ✅ **20 files** created
- ✅ **Rust core** with FFI exports
- ✅ **macOS SwiftUI** interface
- ✅ **Build scripts** ready
- ✅ **Documentation** complete

## 📁 Project Location

```
/Users/khaphan/Documents/Work/gonhanh.org/
```

## 🚀 Next Steps (5 minutes)

### 1️⃣ Test Rust Core

```bash
cd core
cargo test
```

Expected: Tests pass ✅

### 2️⃣ Build Rust Library

```bash
cd ..
./scripts/build-core.sh
```

Expected: `platforms/macos/libgonhanh_core.a` created

### 3️⃣ Create Xcode Project

1. Open **Xcode**
2. File → New → Project
3. Choose **macOS → App**
4. Settings:
   - **Product Name**: `GoNhanh`
   - **Organization ID**: `org.gonhanh`
   - **Interface**: `SwiftUI`
   - **Language**: `Swift`
   - **Location**: `/Users/khaphan/Documents/Work/gonhanh.org/platforms/macos/`

5. **IMPORTANT**: When saving, choose "Don't create Git repository"

### 4️⃣ Add Files to Xcode

1. In Xcode, **delete** the auto-generated files:
   - `ContentView.swift`
   - `GoNhanhApp.swift`

2. **Drag** these files from Finder to Xcode:
   - `App.swift`
   - `MenuBar.swift`
   - `SettingsView.swift`
   - `RustBridge.swift`

3. Select **"Copy items if needed"** → Uncheck
4. Click **"Finish"**

### 5️⃣ Link Rust Library

1. Select **GoNhanh** project in navigator
2. Go to **"Build Phases"**
3. Expand **"Link Binary With Libraries"**
4. Click **"+"**
5. Click **"Add Other..."** → **"Add Files..."**
6. Navigate to `platforms/macos/`
7. Select `libgonhanh_core.a`
8. Click **"Open"**

### 6️⃣ Update Info.plist

1. Select `Info.plist` in Xcode
2. Right-click → **"Open As"** → **"Source Code"**
3. Replace contents with `platforms/macos/GoNhanh/Info.plist`

### 7️⃣ Build & Run

1. Press **Cmd + B** to build
2. Press **Cmd + R** to run
3. Look for ⌨️ icon in menu bar!

## ✅ Success Checklist

- [ ] Rust tests pass
- [ ] Rust library builds
- [ ] Xcode project created
- [ ] Swift files added
- [ ] Library linked
- [ ] App builds
- [ ] Menu bar icon appears

## 🐛 Troubleshooting

### "Library not found"
```bash
./scripts/build-core.sh
```

### "Undefined symbols"
Make sure `libgonhanh_core.a` is in **Link Binary With Libraries**

### "Permission denied"
```bash
chmod +x scripts/*.sh
```

## 📚 Documentation

- **Architecture**: `docs/architecture.md`
- **Development**: `docs/development.md`
- **Contributing**: `CONTRIBUTING.md`
- **Structure**: `PROJECT_STRUCTURE.md`

## 🎉 You're Ready!

Your GoNhanh project is set up and ready for development!

Next: Implement full Telex/VNI rules in `core/src/engine.rs`

---

**Questions?** Check `docs/development.md` or open an issue.
