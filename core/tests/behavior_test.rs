//! Behavior Tests - Real-world typing scenarios
//!
//! Tests for common user behaviors:
//! - Typing mistakes and corrections (backspace)
//! - Retyping after errors
//! - Mixed typing patterns
//! - Edge cases encountered in daily use

mod common;
use common::{test_telex, test_vni};

// ============================================================
// BACKSPACE: Xóa ký tự và gõ lại
// ============================================================

#[test]
fn telex_backspace_and_retype() {
    // vieet = viêt, backspace xóa t → viê, +s(sắc) → viế
    test_telex("vieet<s", "viế");

    // chaof = chào, backspace xóa char cuối trong buffer
    // Engine giữ dấu huyền trên a, +o → chào (engine behavior)
    test_telex("chaof<o", "chào");
}

#[test]
fn telex_backspace_mid_word() {
    // Gõ sai giữa từ, xóa và sửa
    test_telex("toi<as", "toá"); // to + i + backspace + á = toá
}

#[test]
fn vni_backspace_and_retype() {
    // a1 = á, backspace xóa á → empty, 2 không có vowel → không output
    // Engine behavior: backspace xóa cả char khỏi buffer
    test_vni("a1<a2", "à"); // á + backspace + a + 2 = à (gõ lại a trước)
    test_vni("o6<o7", "ơ"); // ô + backspace + o + 7 = ơ (gõ lại o trước)
}

// ============================================================
// TYPO: Gõ nhầm thứ tự phím
// ============================================================

#[test]
fn telex_wrong_order_mark_before_vowel() {
    // Người dùng có thể gõ dấu trước nguyên âm
    // Engine chỉ xử lý khi có nguyên âm trong buffer
    test_telex("sa", "sa"); // s trước, không có vowel -> pass through
    test_telex("as", "á"); // a trước, s sau -> á
}

#[test]
fn telex_double_mark() {
    // Gõ dấu 2 lần -> revert
    test_telex("ass", "as");
    test_telex("aff", "af");
    test_telex("arr", "ar");
}

#[test]
fn telex_double_tone() {
    // Gõ tone 2 lần -> revert
    test_telex("aaa", "aa");
    test_telex("ooo", "oo");
    test_telex("aww", "aw");
}

// ============================================================
// MIXED: Kết hợp nhiều thao tác
// ============================================================

#[test]
fn telex_change_mark_mid_word() {
    // Đổi dấu giữa chừng: gõ sắc rồi đổi sang huyền
    // asf = a + s(sắc→á) + f(huyền thay thế sắc→à)
    // Engine behavior: f thay thế dấu sắc bằng huyền
    test_telex("asf", "à");
}

#[test]
fn telex_tone_then_mark() {
    // Thêm tone (^) rồi thêm mark (sắc)
    test_telex("aas", "ấ");
    test_telex("ees", "ế");
    test_telex("oos", "ố");
}

#[test]
fn telex_mark_then_tone() {
    // Thêm mark trước, tone sau
    test_telex("asa", "ấ"); // á + a = ấ
    test_telex("oso", "ố"); // ó + o = ố (nếu engine hỗ trợ)
}

// ============================================================
// COMMON TYPOS: Lỗi thường gặp khi gõ nhanh
// ============================================================

#[test]
fn telex_common_words_with_typos() {
    // "việt" - vieetj (j = nặng), không phải vieets (s = sắc)
    test_telex("vieetj", "việt");
    // vieets = viết (với sắc)
    test_telex("vieets", "viết");

    // "được" - ươ cần uwow hoặc uow, đúng cách gõ: dduwowcj
    test_telex("dduwowcj", "được");
}

#[test]
fn vni_common_words() {
    // "việt" với VNI
    test_vni("vie65t", "việt");

    // "được" với VNI
    test_vni("d9u7o7c5", "được");
}

// ============================================================
// EDGE CASES: Các trường hợp biên
// ============================================================

#[test]
fn telex_only_consonants() {
    // Chỉ gõ phụ âm, không có nguyên âm
    test_telex("bcd", "bcd");
    test_telex("xyz", "xyz");
}

#[test]
fn telex_mark_without_vowel() {
    // Gõ dấu khi không có nguyên âm trong buffer
    test_telex("bs", "bs"); // không có vowel, s là letter thường
    test_telex("ts", "ts");
}

#[test]
fn telex_multiple_backspace() {
    // Xóa nhiều ký tự liên tiếp
    test_telex("abcd<<<", "a");
    // vieets = viết (4 chars: v,i,ế,t), <<< xóa 3 chars → v, +ng → vng
    test_telex("vieets<<<ng", "vng");
}

#[test]
fn telex_empty_after_backspace() {
    // Xóa hết rồi gõ lại
    test_telex("a<b", "b");
    test_telex("ab<<cd", "cd");
}

// ============================================================
// CONTINUOUS TYPING: Gõ liên tục nhiều từ
// ============================================================

#[test]
fn telex_word_boundary() {
    // Sau khi gõ xong một từ, buffer nên được clear
    // khi gặp ký tự không phải letter (space, punctuation)
    // Hiện tại test với single word
    test_telex("xin", "xin");
    test_telex("chaof", "chào");
}

// ============================================================
// CAPITALIZATION: Chữ hoa
// ============================================================

#[test]
fn telex_caps_mid_word() {
    // Caps ở giữa từ: viEets = v+i+Ê+t+s(sắc) = viẾt
    test_telex("viEets", "viẾt");
}

#[test]
fn telex_all_caps() {
    // VIEETS = V+I+Ê+T+S(sắc) = VIẾT, để có VIỆT cần VIEETJ
    test_telex("VIEETJ", "VIỆT");
    test_telex("VIEETS", "VIẾT");
    test_telex("DDUWOWNGF", "ĐƯỜNG");
}

#[test]
fn vni_all_caps() {
    test_vni("VIE65T", "VIỆT");
    test_vni("D9U7O7NG2", "ĐƯỜNG");
}

// ============================================================
// RAPID TYPING: Gõ nhanh, có thể nhầm
// ============================================================

#[test]
fn telex_rapid_typing_patterns() {
    // Patterns thường gặp khi gõ nhanh
    // ngoafif: ngoa+f(huyền)=ngoà, +i=ngoài, +f áp dụng lên i = ngoài (f không revert vì i chưa có dấu)
    test_telex("ngoafif", "ngoài");
    test_telex("nguwowif", "người");
}

#[test]
fn vni_rapid_typing() {
    test_vni("ngu7o72i2", "người");
    test_vni("to6i1", "tối");
}
