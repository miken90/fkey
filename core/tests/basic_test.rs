//! Basic Vietnamese IME Tests - Single Character Transforms
//!
//! Tests individual character transformations: marks, tones, đ

use gonhanh_core::data::keys;
use gonhanh_core::engine::{Action, Engine};

fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => keys::A, 'b' => keys::B, 'c' => keys::C, 'd' => keys::D,
        'e' => keys::E, 'f' => keys::F, 'g' => keys::G, 'h' => keys::H,
        'i' => keys::I, 'j' => keys::J, 'k' => keys::K, 'l' => keys::L,
        'm' => keys::M, 'n' => keys::N, 'o' => keys::O, 'p' => keys::P,
        'q' => keys::Q, 'r' => keys::R, 's' => keys::S, 't' => keys::T,
        'u' => keys::U, 'v' => keys::V, 'w' => keys::W, 'x' => keys::X,
        'y' => keys::Y, 'z' => keys::Z,
        '0' => keys::N0, '1' => keys::N1, '2' => keys::N2, '3' => keys::N3,
        '4' => keys::N4, '5' => keys::N5, '6' => keys::N6, '7' => keys::N7,
        '8' => keys::N8, '9' => keys::N9,
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
            for _ in 0..r.backspace { screen.pop(); }
            for i in 0..r.count as usize {
                if let Some(ch) = char::from_u32(r.chars[i]) { screen.push(ch); }
            }
        } else if keys::is_letter(key) {
            screen.push(if is_caps { c.to_ascii_uppercase() } else { c.to_ascii_lowercase() });
        }
    }
    screen
}

fn run_telex(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        let result = type_word(&mut e, input);
        assert_eq!(result, *expected, "\n[Telex] '{}' → '{}' (expected '{}')", input, result, expected);
    }
}

fn run_vni(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        e.set_method(1);
        let result = type_word(&mut e, input);
        assert_eq!(result, *expected, "\n[VNI] '{}' → '{}' (expected '{}')", input, result, expected);
    }
}

// ============================================================
// TELEX: MARKS (sắc, huyền, hỏi, ngã, nặng)
// ============================================================

#[test]
fn telex_mark_sac() {
    run_telex(&[
        ("as", "á"), ("es", "é"), ("is", "í"),
        ("os", "ó"), ("us", "ú"), ("ys", "ý"),
    ]);
}

#[test]
fn telex_mark_huyen() {
    run_telex(&[
        ("af", "à"), ("ef", "è"), ("if", "ì"),
        ("of", "ò"), ("uf", "ù"), ("yf", "ỳ"),
    ]);
}

#[test]
fn telex_mark_hoi() {
    run_telex(&[
        ("ar", "ả"), ("er", "ẻ"), ("ir", "ỉ"),
        ("or", "ỏ"), ("ur", "ủ"), ("yr", "ỷ"),
    ]);
}

#[test]
fn telex_mark_nga() {
    run_telex(&[
        ("ax", "ã"), ("ex", "ẽ"), ("ix", "ĩ"),
        ("ox", "õ"), ("ux", "ũ"), ("yx", "ỹ"),
    ]);
}

#[test]
fn telex_mark_nang() {
    run_telex(&[
        ("aj", "ạ"), ("ej", "ẹ"), ("ij", "ị"),
        ("oj", "ọ"), ("uj", "ụ"), ("yj", "ỵ"),
    ]);
}

// ============================================================
// TELEX: TONES (circumflex ^, breve ˘, horn)
// ============================================================

#[test]
fn telex_tone_circumflex() {
    run_telex(&[
        ("aa", "â"), ("ee", "ê"), ("oo", "ô"),
    ]);
}

#[test]
fn telex_tone_breve_horn() {
    run_telex(&[
        ("aw", "ă"),  // a + breve
        ("ow", "ơ"),  // o + horn
        ("uw", "ư"),  // u + horn
    ]);
}

#[test]
fn telex_d_stroke() {
    run_telex(&[
        ("dd", "đ"),
        ("DD", "Đ"),
        ("Dd", "Đ"),
    ]);
}

// ============================================================
// TELEX: COMBINED TONE + MARK
// ============================================================

#[test]
fn telex_circumflex_with_marks() {
    run_telex(&[
        // â + marks
        ("aas", "ấ"), ("aaf", "ầ"), ("aar", "ẩ"), ("aax", "ẫ"), ("aaj", "ậ"),
        // ê + marks
        ("ees", "ế"), ("eef", "ề"), ("eer", "ể"), ("eex", "ễ"), ("eej", "ệ"),
        // ô + marks
        ("oos", "ố"), ("oof", "ồ"), ("oor", "ổ"), ("oox", "ỗ"), ("ooj", "ộ"),
    ]);
}

#[test]
fn telex_breve_horn_with_marks() {
    run_telex(&[
        // ă + marks
        ("aws", "ắ"), ("awf", "ằ"), ("awr", "ẳ"), ("awx", "ẵ"), ("awj", "ặ"),
        // ơ + marks
        ("ows", "ớ"), ("owf", "ờ"), ("owr", "ở"), ("owx", "ỡ"), ("owj", "ợ"),
        // ư + marks
        ("uws", "ứ"), ("uwf", "ừ"), ("uwr", "ử"), ("uwx", "ữ"), ("uwj", "ự"),
    ]);
}

// ============================================================
// TELEX: DOUBLE-KEY REVERT
// ============================================================

#[test]
fn telex_revert_mark() {
    run_telex(&[
        ("ass", "as"), ("aff", "af"), ("arr", "ar"),
        ("axx", "ax"), ("ajj", "aj"),
    ]);
}

#[test]
fn telex_revert_tone() {
    run_telex(&[
        ("aaa", "aa"), ("eee", "ee"), ("ooo", "oo"),
        ("aww", "aw"), ("oww", "ow"), ("uww", "uw"),
    ]);
}

// ============================================================
// VNI: MARKS (1=sắc, 2=huyền, 3=hỏi, 4=ngã, 5=nặng)
// ============================================================

#[test]
fn vni_mark_sac() {
    run_vni(&[
        ("a1", "á"), ("e1", "é"), ("i1", "í"),
        ("o1", "ó"), ("u1", "ú"), ("y1", "ý"),
    ]);
}

#[test]
fn vni_mark_huyen() {
    run_vni(&[
        ("a2", "à"), ("e2", "è"), ("i2", "ì"),
        ("o2", "ò"), ("u2", "ù"), ("y2", "ỳ"),
    ]);
}

#[test]
fn vni_mark_hoi() {
    run_vni(&[
        ("a3", "ả"), ("e3", "ẻ"), ("i3", "ỉ"),
        ("o3", "ỏ"), ("u3", "ủ"), ("y3", "ỷ"),
    ]);
}

#[test]
fn vni_mark_nga() {
    run_vni(&[
        ("a4", "ã"), ("e4", "ẽ"), ("i4", "ĩ"),
        ("o4", "õ"), ("u4", "ũ"), ("y4", "ỹ"),
    ]);
}

#[test]
fn vni_mark_nang() {
    run_vni(&[
        ("a5", "ạ"), ("e5", "ẹ"), ("i5", "ị"),
        ("o5", "ọ"), ("u5", "ụ"), ("y5", "ỵ"),
    ]);
}

// ============================================================
// VNI: TONES (6=^, 7=ă, 8=ơ/ư, 9=đ)
// ============================================================

#[test]
fn vni_tone_circumflex() {
    run_vni(&[
        ("a6", "â"), ("e6", "ê"), ("o6", "ô"),
    ]);
}

#[test]
fn vni_tone_breve_horn() {
    run_vni(&[
        ("a7", "ă"),  // a + breve
        ("o8", "ơ"),  // o + horn
        ("u8", "ư"),  // u + horn
    ]);
}

#[test]
fn vni_d_stroke() {
    run_vni(&[
        ("d9", "đ"),
        ("D9", "Đ"),
    ]);
}

// ============================================================
// VNI: COMBINED TONE + MARK
// ============================================================

#[test]
fn vni_circumflex_with_marks() {
    run_vni(&[
        // â + marks
        ("a61", "ấ"), ("a62", "ầ"), ("a63", "ẩ"), ("a64", "ẫ"), ("a65", "ậ"),
        // ê + marks
        ("e61", "ế"), ("e62", "ề"), ("e63", "ể"), ("e64", "ễ"), ("e65", "ệ"),
        // ô + marks
        ("o61", "ố"), ("o62", "ồ"), ("o63", "ổ"), ("o64", "ỗ"), ("o65", "ộ"),
    ]);
}

#[test]
fn vni_breve_horn_with_marks() {
    run_vni(&[
        // ă + marks
        ("a71", "ắ"), ("a72", "ằ"), ("a73", "ẳ"), ("a74", "ẵ"), ("a75", "ặ"),
        // ơ + marks
        ("o81", "ớ"), ("o82", "ờ"), ("o83", "ở"), ("o84", "ỡ"), ("o85", "ợ"),
        // ư + marks
        ("u81", "ứ"), ("u82", "ừ"), ("u83", "ử"), ("u84", "ữ"), ("u85", "ự"),
    ]);
}

// ============================================================
// VNI: DOUBLE-KEY REVERT
// ============================================================

#[test]
fn vni_revert_mark() {
    run_vni(&[
        ("a11", "a1"), ("a22", "a2"), ("a33", "a3"),
        ("a44", "a4"), ("a55", "a5"),
    ]);
}

#[test]
fn vni_revert_tone() {
    run_vni(&[
        ("a66", "a6"), ("e66", "e6"), ("o66", "o6"),
        ("a77", "a7"), ("o88", "o8"), ("u88", "u8"),
    ]);
}

// ============================================================
// UPPERCASE HANDLING
// ============================================================

#[test]
fn telex_uppercase() {
    run_telex(&[
        ("As", "Á"), ("AS", "Á"),
        ("Aa", "Â"), ("AA", "Â"),
        ("Aw", "Ă"), ("AW", "Ă"),
    ]);
}

#[test]
fn vni_uppercase() {
    run_vni(&[
        ("A1", "Á"), ("A6", "Â"), ("A7", "Ă"),
        ("O8", "Ơ"), ("U8", "Ư"),
    ]);
}
