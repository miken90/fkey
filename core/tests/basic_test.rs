//! Basic Vietnamese IME Tests - Single Character Transforms
//!
//! Tests individual character transformations: marks, tones, đ

use gonhanh_core::data::keys;
use gonhanh_core::engine::{Action, Engine};

fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => keys::A,
        'b' => keys::B,
        'c' => keys::C,
        'd' => keys::D,
        'e' => keys::E,
        'f' => keys::F,
        'g' => keys::G,
        'h' => keys::H,
        'i' => keys::I,
        'j' => keys::J,
        'k' => keys::K,
        'l' => keys::L,
        'm' => keys::M,
        'n' => keys::N,
        'o' => keys::O,
        'p' => keys::P,
        'q' => keys::Q,
        'r' => keys::R,
        's' => keys::S,
        't' => keys::T,
        'u' => keys::U,
        'v' => keys::V,
        'w' => keys::W,
        'x' => keys::X,
        'y' => keys::Y,
        'z' => keys::Z,
        '0' => keys::N0,
        '1' => keys::N1,
        '2' => keys::N2,
        '3' => keys::N3,
        '4' => keys::N4,
        '5' => keys::N5,
        '6' => keys::N6,
        '7' => keys::N7,
        '8' => keys::N8,
        '9' => keys::N9,
        _ => 255,
    }
}

fn type_word(e: &mut Engine, input: &str) -> String {
    let mut screen = String::new();
    for c in input.chars() {
        let key = char_to_key(c);
        let is_caps = c.is_uppercase();
        let r = e.on_key(key, is_caps, false);
        if r.action == Action::Send as u8 {
            for _ in 0..r.backspace {
                screen.pop();
            }
            for i in 0..r.count as usize {
                if let Some(ch) = char::from_u32(r.chars[i]) {
                    screen.push(ch);
                }
            }
        } else if keys::is_letter(key) {
            screen.push(if is_caps {
                c.to_ascii_uppercase()
            } else {
                c.to_ascii_lowercase()
            });
        }
    }
    screen
}

fn run_telex(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        let result = type_word(&mut e, input);
        assert_eq!(
            result, *expected,
            "\n[Telex] '{}' → '{}' (expected '{}')",
            input, result, expected
        );
    }
}

fn run_vni(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        e.set_method(1);
        let result = type_word(&mut e, input);
        assert_eq!(
            result, *expected,
            "\n[VNI] '{}' → '{}' (expected '{}')",
            input, result, expected
        );
    }
}

// ============================================================
// TELEX: ALL SINGLE VOWELS WITH MARKS
// Complete matrix: 6 vowels × 5 marks = 30 combinations
// ============================================================

#[test]
fn telex_a_all_marks() {
    run_telex(&[
        ("a", "a"),  // base
        ("as", "á"), // sắc
        ("af", "à"), // huyền
        ("ar", "ả"), // hỏi
        ("ax", "ã"), // ngã
        ("aj", "ạ"), // nặng
    ]);
}

#[test]
fn telex_e_all_marks() {
    run_telex(&[
        ("e", "e"),  // base
        ("es", "é"), // sắc
        ("ef", "è"), // huyền
        ("er", "ẻ"), // hỏi
        ("ex", "ẽ"), // ngã
        ("ej", "ẹ"), // nặng
    ]);
}

#[test]
fn telex_i_all_marks() {
    run_telex(&[
        ("i", "i"),  // base
        ("is", "í"), // sắc
        ("if", "ì"), // huyền
        ("ir", "ỉ"), // hỏi
        ("ix", "ĩ"), // ngã
        ("ij", "ị"), // nặng
    ]);
}

#[test]
fn telex_o_all_marks() {
    run_telex(&[
        ("o", "o"),  // base
        ("os", "ó"), // sắc
        ("of", "ò"), // huyền
        ("or", "ỏ"), // hỏi
        ("ox", "õ"), // ngã
        ("oj", "ọ"), // nặng
    ]);
}

#[test]
fn telex_u_all_marks() {
    run_telex(&[
        ("u", "u"),  // base
        ("us", "ú"), // sắc
        ("uf", "ù"), // huyền
        ("ur", "ủ"), // hỏi
        ("ux", "ũ"), // ngã
        ("uj", "ụ"), // nặng
    ]);
}

#[test]
fn telex_y_all_marks() {
    run_telex(&[
        ("y", "y"),  // base
        ("ys", "ý"), // sắc
        ("yf", "ỳ"), // huyền
        ("yr", "ỷ"), // hỏi
        ("yx", "ỹ"), // ngã
        ("yj", "ỵ"), // nặng
    ]);
}

// ============================================================
// TELEX: ALL MODIFIED VOWELS (â, ê, ô, ă, ơ, ư) WITH MARKS
// Complete matrix: 6 modified vowels × 6 states = 36 combinations
// ============================================================

#[test]
fn telex_a_circumflex_all_marks() {
    // â = aa
    run_telex(&[
        ("aa", "â"),  // base
        ("aas", "ấ"), // sắc
        ("aaf", "ầ"), // huyền
        ("aar", "ẩ"), // hỏi
        ("aax", "ẫ"), // ngã
        ("aaj", "ậ"), // nặng
    ]);
}

#[test]
fn telex_e_circumflex_all_marks() {
    // ê = ee
    run_telex(&[
        ("ee", "ê"),  // base
        ("ees", "ế"), // sắc
        ("eef", "ề"), // huyền
        ("eer", "ể"), // hỏi
        ("eex", "ễ"), // ngã
        ("eej", "ệ"), // nặng
    ]);
}

#[test]
fn telex_o_circumflex_all_marks() {
    // ô = oo
    run_telex(&[
        ("oo", "ô"),  // base
        ("oos", "ố"), // sắc
        ("oof", "ồ"), // huyền
        ("oor", "ổ"), // hỏi
        ("oox", "ỗ"), // ngã
        ("ooj", "ộ"), // nặng
    ]);
}

#[test]
fn telex_a_breve_all_marks() {
    // ă = aw
    run_telex(&[
        ("aw", "ă"),  // base
        ("aws", "ắ"), // sắc
        ("awf", "ằ"), // huyền
        ("awr", "ẳ"), // hỏi
        ("awx", "ẵ"), // ngã
        ("awj", "ặ"), // nặng
    ]);
}

#[test]
fn telex_o_horn_all_marks() {
    // ơ = ow
    run_telex(&[
        ("ow", "ơ"),  // base
        ("ows", "ớ"), // sắc
        ("owf", "ờ"), // huyền
        ("owr", "ở"), // hỏi
        ("owx", "ỡ"), // ngã
        ("owj", "ợ"), // nặng
    ]);
}

#[test]
fn telex_u_horn_all_marks() {
    // ư = uw
    run_telex(&[
        ("uw", "ư"),  // base
        ("uws", "ứ"), // sắc
        ("uwf", "ừ"), // huyền
        ("uwr", "ử"), // hỏi
        ("uwx", "ữ"), // ngã
        ("uwj", "ự"), // nặng
    ]);
}

#[test]
fn telex_d_stroke() {
    run_telex(&[("dd", "đ"), ("DD", "Đ"), ("Dd", "Đ")]);
}

// ============================================================
// TELEX: DOUBLE-KEY REVERT
// ============================================================

#[test]
fn telex_revert_mark() {
    run_telex(&[
        ("ass", "as"),
        ("aff", "af"),
        ("arr", "ar"),
        ("axx", "ax"),
        ("ajj", "aj"),
    ]);
}

#[test]
fn telex_revert_tone() {
    run_telex(&[
        ("aaa", "aa"),
        ("eee", "ee"),
        ("ooo", "oo"),
        ("aww", "aw"),
        ("oww", "ow"),
        ("uww", "uw"),
    ]);
}

// ============================================================
// VNI: ALL SINGLE VOWELS WITH MARKS
// Complete matrix: 6 vowels × 5 marks = 30 combinations
// ============================================================

#[test]
fn vni_a_all_marks() {
    run_vni(&[
        ("a", "a"),  // base
        ("a1", "á"), // sắc
        ("a2", "à"), // huyền
        ("a3", "ả"), // hỏi
        ("a4", "ã"), // ngã
        ("a5", "ạ"), // nặng
    ]);
}

#[test]
fn vni_e_all_marks() {
    run_vni(&[
        ("e", "e"),  // base
        ("e1", "é"), // sắc
        ("e2", "è"), // huyền
        ("e3", "ẻ"), // hỏi
        ("e4", "ẽ"), // ngã
        ("e5", "ẹ"), // nặng
    ]);
}

#[test]
fn vni_i_all_marks() {
    run_vni(&[
        ("i", "i"),  // base
        ("i1", "í"), // sắc
        ("i2", "ì"), // huyền
        ("i3", "ỉ"), // hỏi
        ("i4", "ĩ"), // ngã
        ("i5", "ị"), // nặng
    ]);
}

#[test]
fn vni_o_all_marks() {
    run_vni(&[
        ("o", "o"),  // base
        ("o1", "ó"), // sắc
        ("o2", "ò"), // huyền
        ("o3", "ỏ"), // hỏi
        ("o4", "õ"), // ngã
        ("o5", "ọ"), // nặng
    ]);
}

#[test]
fn vni_u_all_marks() {
    run_vni(&[
        ("u", "u"),  // base
        ("u1", "ú"), // sắc
        ("u2", "ù"), // huyền
        ("u3", "ủ"), // hỏi
        ("u4", "ũ"), // ngã
        ("u5", "ụ"), // nặng
    ]);
}

#[test]
fn vni_y_all_marks() {
    run_vni(&[
        ("y", "y"),  // base
        ("y1", "ý"), // sắc
        ("y2", "ỳ"), // huyền
        ("y3", "ỷ"), // hỏi
        ("y4", "ỹ"), // ngã
        ("y5", "ỵ"), // nặng
    ]);
}

// ============================================================
// VNI: ALL MODIFIED VOWELS (â, ê, ô, ă, ơ, ư) WITH MARKS
// VNI: 6=circumflex(^), 7=horn(móc), 8=breve(˘), 9=đ
// Complete matrix: 6 modified vowels × 6 states = 36 combinations
// ============================================================

#[test]
fn vni_a_circumflex_all_marks() {
    // â = a + 6
    run_vni(&[
        ("a6", "â"),  // base
        ("a61", "ấ"), // sắc
        ("a62", "ầ"), // huyền
        ("a63", "ẩ"), // hỏi
        ("a64", "ẫ"), // ngã
        ("a65", "ậ"), // nặng
    ]);
}

#[test]
fn vni_e_circumflex_all_marks() {
    // ê = e + 6
    run_vni(&[
        ("e6", "ê"),  // base
        ("e61", "ế"), // sắc
        ("e62", "ề"), // huyền
        ("e63", "ể"), // hỏi
        ("e64", "ễ"), // ngã
        ("e65", "ệ"), // nặng
    ]);
}

#[test]
fn vni_o_circumflex_all_marks() {
    // ô = o + 6
    run_vni(&[
        ("o6", "ô"),  // base
        ("o61", "ố"), // sắc
        ("o62", "ồ"), // huyền
        ("o63", "ổ"), // hỏi
        ("o64", "ỗ"), // ngã
        ("o65", "ộ"), // nặng
    ]);
}

#[test]
fn vni_o_horn_all_marks() {
    // ơ = o + 7 (horn)
    run_vni(&[
        ("o7", "ơ"),  // base
        ("o71", "ớ"), // sắc
        ("o72", "ờ"), // huyền
        ("o73", "ở"), // hỏi
        ("o74", "ỡ"), // ngã
        ("o75", "ợ"), // nặng
    ]);
}

#[test]
fn vni_u_horn_all_marks() {
    // ư = u + 7 (horn)
    run_vni(&[
        ("u7", "ư"),  // base
        ("u71", "ứ"), // sắc
        ("u72", "ừ"), // huyền
        ("u73", "ử"), // hỏi
        ("u74", "ữ"), // ngã
        ("u75", "ự"), // nặng
    ]);
}

#[test]
fn vni_a_breve_all_marks() {
    // ă = a + 8 (breve)
    run_vni(&[
        ("a8", "ă"),  // base
        ("a81", "ắ"), // sắc
        ("a82", "ằ"), // huyền
        ("a83", "ẳ"), // hỏi
        ("a84", "ẵ"), // ngã
        ("a85", "ặ"), // nặng
    ]);
}

#[test]
fn vni_d_stroke() {
    run_vni(&[("d9", "đ"), ("D9", "Đ")]);
}

#[test]
fn vni_delayed_d_input() {
    run_vni(&[
        // VNI delayed đ: 9 can come after other characters
        ("d9ung1", "đúng"), // immediate: d9 first
        ("du9ng1", "đúng"), // delayed: 9 after vowel
        ("dung91", "đúng"), // delayed: 9 after consonant, then mark
        ("dung19", "đúng"), // mark first, then đ
        ("D9ung1", "Đúng"), // uppercase immediate
        ("Du9ng1", "Đúng"), // uppercase delayed
        ("Dung91", "Đúng"), // uppercase delayed after consonant
    ]);
}

// ============================================================
// VNI: DELAYED TONE INPUT (tone after consonant)
// ============================================================

#[test]
fn vni_delayed_tone_input() {
    run_vni(&[
        // VNI: 6=^(â,ê,ô), 7=horn(ơ,ư), 8=breve(ă), 9=đ
        // tone key finds the vowel in buffer
        ("tu72", "từ"), // t + u + 7(ư) + 2(huyền) = từ
        ("to61", "tố"), // t + o + 6(ô) + 1(sắc) = tố
        ("ta81", "tắ"), // t + a + 8(ă) + 1(sắc) = tắ
        ("nu72", "nừ"), // n + u + 7(ư) + 2(huyền) = nừ
        ("to72", "tờ"), // t + o + 7(ơ) + 2(huyền) = tờ
        ("na82", "nằ"), // n + a + 8(ă) + 2(huyền) = nằ
    ]);
}

// ============================================================
// VNI: DOUBLE-KEY REVERT
// ============================================================

#[test]
fn vni_revert_mark() {
    run_vni(&[
        ("a11", "a1"),
        ("a22", "a2"),
        ("a33", "a3"),
        ("a44", "a4"),
        ("a55", "a5"),
    ]);
}

#[test]
fn vni_revert_tone() {
    run_vni(&[
        ("a66", "a6"),
        ("e66", "e6"),
        ("o66", "o6"),
        ("o77", "o7"),
        ("u77", "u7"), // 7=horn for o,u
        ("a88", "a8"), // 8=breve for a
    ]);
}

// ============================================================
// UPPERCASE HANDLING
// ============================================================

#[test]
fn telex_uppercase() {
    run_telex(&[
        ("As", "Á"),
        ("AS", "Á"),
        ("Aa", "Â"),
        ("AA", "Â"),
        ("Aw", "Ă"),
        ("AW", "Ă"),
        ("Ow", "Ơ"),
        ("Uw", "Ư"),
    ]);
}

#[test]
fn vni_uppercase() {
    run_vni(&[
        ("A1", "Á"),
        ("A6", "Â"),
        ("O7", "Ơ"),
        ("U7", "Ư"), // 7=horn
        ("A8", "Ă"), // 8=breve
    ]);
}
