//! Comprehensive tests for Vietnamese IME Engine

use gonhanh_core::data::keys;
use gonhanh_core::engine::{Action, Engine};

// ============================================================
// Helper
// ============================================================

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
        ' ' => keys::SPACE,
        _ => 0,
    }
}

fn type_string(e: &mut Engine, s: &str) -> Vec<gonhanh_core::engine::Result> {
    s.chars()
        .map(|c| {
            let key = char_to_key(c);
            let caps = c.is_uppercase();
            e.on_key(key, caps, false)
        })
        .collect()
}

fn get_output_char(r: &gonhanh_core::engine::Result) -> Option<char> {
    if r.action == Action::Send as u8 && r.count > 0 {
        char::from_u32(r.chars[0])
    } else {
        None
    }
}

// ============================================================
// TELEX: Marks (s/f/r/x/j)
// ============================================================

#[test]
fn telex_mark_sac() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "as");
    assert_eq!(get_output_char(&results[1]), Some('á'));
}

#[test]
fn telex_mark_huyen() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "af");
    assert_eq!(get_output_char(&results[1]), Some('à'));
}

#[test]
fn telex_mark_hoi() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ar");
    assert_eq!(get_output_char(&results[1]), Some('ả'));
}

#[test]
fn telex_mark_nga() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ax");
    assert_eq!(get_output_char(&results[1]), Some('ã'));
}

#[test]
fn telex_mark_nang() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "aj");
    assert_eq!(get_output_char(&results[1]), Some('ạ'));
}

// ============================================================
// TELEX: Tones (aa/ee/oo/aw/ow/uw)
// ============================================================

#[test]
fn telex_tone_aa() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "aa");
    assert_eq!(get_output_char(&results[1]), Some('â'));
}

#[test]
fn telex_tone_ee() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ee");
    assert_eq!(get_output_char(&results[1]), Some('ê'));
}

#[test]
fn telex_tone_oo() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "oo");
    assert_eq!(get_output_char(&results[1]), Some('ô'));
}

#[test]
fn telex_tone_aw() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "aw");
    assert_eq!(get_output_char(&results[1]), Some('ă'));
}

#[test]
fn telex_tone_ow() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ow");
    assert_eq!(get_output_char(&results[1]), Some('ơ'));
}

#[test]
fn telex_tone_uw() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "uw");
    assert_eq!(get_output_char(&results[1]), Some('ư'));
}

// ============================================================
// TELEX: dd -> đ
// ============================================================

#[test]
fn telex_dd() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "dd");
    assert_eq!(get_output_char(&results[1]), Some('đ'));
}

// ============================================================
// TELEX: Combined (tone + mark)
// ============================================================

#[test]
fn telex_aas() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "aas");
    // aa -> â, then s -> ấ
    assert_eq!(get_output_char(&results[2]), Some('ấ'));
}

#[test]
fn telex_ees() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ees");
    assert_eq!(get_output_char(&results[2]), Some('ế'));
}

#[test]
fn telex_ooj() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "ooj");
    assert_eq!(get_output_char(&results[2]), Some('ộ'));
}

#[test]
fn telex_uws() {
    let mut e = Engine::new();
    e.set_method(0);

    let results = type_string(&mut e, "uws");
    assert_eq!(get_output_char(&results[2]), Some('ứ'));
}

// ============================================================
// VNI: Marks (1-5)
// ============================================================

#[test]
fn vni_mark_sac() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a1");
    assert_eq!(get_output_char(&results[1]), Some('á'));
}

#[test]
fn vni_mark_huyen() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a2");
    assert_eq!(get_output_char(&results[1]), Some('à'));
}

#[test]
fn vni_mark_hoi() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a3");
    assert_eq!(get_output_char(&results[1]), Some('ả'));
}

#[test]
fn vni_mark_nga() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a4");
    assert_eq!(get_output_char(&results[1]), Some('ã'));
}

#[test]
fn vni_mark_nang() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a5");
    assert_eq!(get_output_char(&results[1]), Some('ạ'));
}

// ============================================================
// VNI: Tones (6/7/8)
// ============================================================

#[test]
fn vni_tone_a6() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a6");
    assert_eq!(get_output_char(&results[1]), Some('â'));
}

#[test]
fn vni_tone_e6() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "e6");
    assert_eq!(get_output_char(&results[1]), Some('ê'));
}

#[test]
fn vni_tone_o6() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "o6");
    assert_eq!(get_output_char(&results[1]), Some('ô'));
}

#[test]
fn vni_tone_a7() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a7");
    assert_eq!(get_output_char(&results[1]), Some('ă'));
}

#[test]
fn vni_tone_o8() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "o8");
    assert_eq!(get_output_char(&results[1]), Some('ơ'));
}

#[test]
fn vni_tone_u8() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "u8");
    assert_eq!(get_output_char(&results[1]), Some('ư'));
}

// ============================================================
// VNI: d9 -> đ
// ============================================================

#[test]
fn vni_d9() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "d9");
    assert_eq!(get_output_char(&results[1]), Some('đ'));
}

// ============================================================
// VNI: Combined
// ============================================================

#[test]
fn vni_a61() {
    let mut e = Engine::new();
    e.set_method(1);

    let results = type_string(&mut e, "a61");
    assert_eq!(get_output_char(&results[2]), Some('ấ'));
}

// ============================================================
// Uppercase
// ============================================================

#[test]
fn telex_uppercase_as() {
    let mut e = Engine::new();
    e.set_method(0);

    // Type 'A' (caps) then 's'
    e.on_key(keys::A, true, false);
    let r = e.on_key(keys::S, false, false);
    assert_eq!(get_output_char(&r), Some('Á'));
}

#[test]
fn vni_uppercase_a1() {
    let mut e = Engine::new();
    e.set_method(1);

    e.on_key(keys::A, true, false);
    let r = e.on_key(keys::N1, false, false);
    assert_eq!(get_output_char(&r), Some('Á'));
}

// ============================================================
// Edge Cases
// ============================================================

#[test]
fn space_clears_buffer() {
    let mut e = Engine::new();
    e.set_method(0);

    type_string(&mut e, "a ");
    let r = e.on_key(keys::S, false, false);
    // s after space should NOT trigger mark
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn backspace_removes_from_buffer() {
    let mut e = Engine::new();
    e.set_method(0);

    e.on_key(keys::A, false, false);
    e.on_key(keys::DELETE, false, false);
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn ctrl_clears_buffer() {
    let mut e = Engine::new();
    e.set_method(0);

    e.on_key(keys::A, false, false);
    e.on_key(keys::C, false, true); // Ctrl+C
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn disabled_engine() {
    let mut e = Engine::new();
    e.set_method(0);
    e.set_enabled(false);

    let r = e.on_key(keys::A, false, false);
    assert_eq!(r.action, Action::None as u8);
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// Word tests
// ============================================================

#[test]
fn word_chao() {
    let mut e = Engine::new();
    e.set_method(0);

    // "chaof" -> chào
    type_string(&mut e, "cha");
    let r = e.on_key(keys::F, false, false);
    assert_eq!(get_output_char(&r), Some('à'));
}

#[test]
fn word_viet() {
    let mut e = Engine::new();
    e.set_method(0);

    // "vieetj" -> việt
    type_string(&mut e, "vi");
    let r1 = type_string(&mut e, "ee");
    assert_eq!(get_output_char(&r1[1]), Some('ê'));

    let r2 = e.on_key(keys::J, false, false);
    assert_eq!(get_output_char(&r2), Some('ệ'));
}

// ============================================================
// All vowels with marks
// ============================================================

#[test]
fn all_vowels_sac() {
    let mut e = Engine::new();
    e.set_method(0);

    let vowels = [
        ("as", 'á'), ("es", 'é'), ("is", 'í'),
        ("os", 'ó'), ("us", 'ú'), ("ys", 'ý'),
    ];

    for (input, expected) in vowels {
        e.clear();
        let results = type_string(&mut e, input);
        assert_eq!(
            get_output_char(&results[1]),
            Some(expected),
            "{} should produce {}", input, expected
        );
    }
}

#[test]
fn all_tones_a() {
    let mut e = Engine::new();
    e.set_method(0);

    // Test all tones on 'a'
    e.clear();
    let r = type_string(&mut e, "aa");
    assert_eq!(get_output_char(&r[1]), Some('â'), "aa -> â");

    e.clear();
    let r = type_string(&mut e, "aw");
    assert_eq!(get_output_char(&r[1]), Some('ă'), "aw -> ă");
}
