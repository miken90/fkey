//! Character Tests - Single character transformations
//!
//! Tests: marks, tones, đ, revert

mod common;
use common::{run_telex, run_vni};

// ============================================================
// TELEX: ALL SINGLE VOWELS WITH MARKS
// ============================================================

#[test]
fn telex_a_all_marks() {
    run_telex(&[
        ("a", "a"),
        ("as", "á"),
        ("af", "à"),
        ("ar", "ả"),
        ("ax", "ã"),
        ("aj", "ạ"),
    ]);
}

#[test]
fn telex_e_all_marks() {
    run_telex(&[
        ("e", "e"),
        ("es", "é"),
        ("ef", "è"),
        ("er", "ẻ"),
        ("ex", "ẽ"),
        ("ej", "ẹ"),
    ]);
}

#[test]
fn telex_i_all_marks() {
    run_telex(&[
        ("i", "i"),
        ("is", "í"),
        ("if", "ì"),
        ("ir", "ỉ"),
        ("ix", "ĩ"),
        ("ij", "ị"),
    ]);
}

#[test]
fn telex_o_all_marks() {
    run_telex(&[
        ("o", "o"),
        ("os", "ó"),
        ("of", "ò"),
        ("or", "ỏ"),
        ("ox", "õ"),
        ("oj", "ọ"),
    ]);
}

#[test]
fn telex_u_all_marks() {
    run_telex(&[
        ("u", "u"),
        ("us", "ú"),
        ("uf", "ù"),
        ("ur", "ủ"),
        ("ux", "ũ"),
        ("uj", "ụ"),
    ]);
}

#[test]
fn telex_y_all_marks() {
    run_telex(&[
        ("y", "y"),
        ("ys", "ý"),
        ("yf", "ỳ"),
        ("yr", "ỷ"),
        ("yx", "ỹ"),
        ("yj", "ỵ"),
    ]);
}

// ============================================================
// TELEX: MODIFIED VOWELS (â, ê, ô, ă, ơ, ư)
// ============================================================

#[test]
fn telex_a_circumflex_all_marks() {
    run_telex(&[
        ("aa", "â"),
        ("aas", "ấ"),
        ("aaf", "ầ"),
        ("aar", "ẩ"),
        ("aax", "ẫ"),
        ("aaj", "ậ"),
    ]);
}

#[test]
fn telex_e_circumflex_all_marks() {
    run_telex(&[
        ("ee", "ê"),
        ("ees", "ế"),
        ("eef", "ề"),
        ("eer", "ể"),
        ("eex", "ễ"),
        ("eej", "ệ"),
    ]);
}

#[test]
fn telex_o_circumflex_all_marks() {
    run_telex(&[
        ("oo", "ô"),
        ("oos", "ố"),
        ("oof", "ồ"),
        ("oor", "ổ"),
        ("oox", "ỗ"),
        ("ooj", "ộ"),
    ]);
}

#[test]
fn telex_a_breve_all_marks() {
    run_telex(&[
        ("aw", "ă"),
        ("aws", "ắ"),
        ("awf", "ằ"),
        ("awr", "ẳ"),
        ("awx", "ẵ"),
        ("awj", "ặ"),
    ]);
}

#[test]
fn telex_o_horn_all_marks() {
    run_telex(&[
        ("ow", "ơ"),
        ("ows", "ớ"),
        ("owf", "ờ"),
        ("owr", "ở"),
        ("owx", "ỡ"),
        ("owj", "ợ"),
    ]);
}

#[test]
fn telex_u_horn_all_marks() {
    run_telex(&[
        ("uw", "ư"),
        ("uws", "ứ"),
        ("uwf", "ừ"),
        ("uwr", "ử"),
        ("uwx", "ữ"),
        ("uwj", "ự"),
    ]);
}

#[test]
fn telex_d_stroke() {
    run_telex(&[("dd", "đ"), ("DD", "Đ"), ("Dd", "Đ")]);
}

// ============================================================
// TELEX: DELAYED TONE INPUT
// ============================================================

#[test]
fn telex_delayed_tone_input() {
    // Telex delayed mode: tone key can be typed after consonants
    // w applies horn to vowels; for uo compound, applies to BOTH u and o
    run_telex(&[
        // Single vowel cases
        ("tuw", "tư"), // tu + w → tư
        ("tow", "tơ"), // to + w → tơ
        ("taw", "tă"), // ta + w → tă
        // With final consonants - w still finds vowel
        ("tungw", "tưng"), // tung + w → tưng
        ("tongw", "tơng"), // tong + w → tơng
        // uo compound: single w applies horn to BOTH u and o
        ("tuow", "tươ"),       // tuo + w → tươ (w applies to both u and o)
        ("truwowng", "trương"), // standard Telex for trường
    ]);
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
// ============================================================

#[test]
fn vni_a_all_marks() {
    run_vni(&[
        ("a", "a"),
        ("a1", "á"),
        ("a2", "à"),
        ("a3", "ả"),
        ("a4", "ã"),
        ("a5", "ạ"),
    ]);
}

#[test]
fn vni_e_all_marks() {
    run_vni(&[
        ("e", "e"),
        ("e1", "é"),
        ("e2", "è"),
        ("e3", "ẻ"),
        ("e4", "ẽ"),
        ("e5", "ẹ"),
    ]);
}

#[test]
fn vni_i_all_marks() {
    run_vni(&[
        ("i", "i"),
        ("i1", "í"),
        ("i2", "ì"),
        ("i3", "ỉ"),
        ("i4", "ĩ"),
        ("i5", "ị"),
    ]);
}

#[test]
fn vni_o_all_marks() {
    run_vni(&[
        ("o", "o"),
        ("o1", "ó"),
        ("o2", "ò"),
        ("o3", "ỏ"),
        ("o4", "õ"),
        ("o5", "ọ"),
    ]);
}

#[test]
fn vni_u_all_marks() {
    run_vni(&[
        ("u", "u"),
        ("u1", "ú"),
        ("u2", "ù"),
        ("u3", "ủ"),
        ("u4", "ũ"),
        ("u5", "ụ"),
    ]);
}

#[test]
fn vni_y_all_marks() {
    run_vni(&[
        ("y", "y"),
        ("y1", "ý"),
        ("y2", "ỳ"),
        ("y3", "ỷ"),
        ("y4", "ỹ"),
        ("y5", "ỵ"),
    ]);
}

// ============================================================
// VNI: MODIFIED VOWELS
// ============================================================

#[test]
fn vni_a_circumflex_all_marks() {
    run_vni(&[
        ("a6", "â"),
        ("a61", "ấ"),
        ("a62", "ầ"),
        ("a63", "ẩ"),
        ("a64", "ẫ"),
        ("a65", "ậ"),
    ]);
}

#[test]
fn vni_e_circumflex_all_marks() {
    run_vni(&[
        ("e6", "ê"),
        ("e61", "ế"),
        ("e62", "ề"),
        ("e63", "ể"),
        ("e64", "ễ"),
        ("e65", "ệ"),
    ]);
}

#[test]
fn vni_o_circumflex_all_marks() {
    run_vni(&[
        ("o6", "ô"),
        ("o61", "ố"),
        ("o62", "ồ"),
        ("o63", "ổ"),
        ("o64", "ỗ"),
        ("o65", "ộ"),
    ]);
}

#[test]
fn vni_o_horn_all_marks() {
    run_vni(&[
        ("o7", "ơ"),
        ("o71", "ớ"),
        ("o72", "ờ"),
        ("o73", "ở"),
        ("o74", "ỡ"),
        ("o75", "ợ"),
    ]);
}

#[test]
fn vni_u_horn_all_marks() {
    run_vni(&[
        ("u7", "ư"),
        ("u71", "ứ"),
        ("u72", "ừ"),
        ("u73", "ử"),
        ("u74", "ữ"),
        ("u75", "ự"),
    ]);
}

#[test]
fn vni_a_breve_all_marks() {
    run_vni(&[
        ("a8", "ă"),
        ("a81", "ắ"),
        ("a82", "ằ"),
        ("a83", "ẳ"),
        ("a84", "ẵ"),
        ("a85", "ặ"),
    ]);
}

#[test]
fn vni_d_stroke() {
    run_vni(&[("d9", "đ"), ("D9", "Đ")]);
}

#[test]
fn vni_delayed_d_input() {
    run_vni(&[
        ("d9ung1", "đúng"),
        ("du9ng1", "đúng"),
        ("dung91", "đúng"),
        ("dung19", "đúng"),
        ("D9ung1", "Đúng"),
        ("Du9ng1", "Đúng"),
        ("Dung91", "Đúng"),
    ]);
}

// ============================================================
// VNI: DELAYED TONE INPUT
// ============================================================

#[test]
fn vni_delayed_tone_input() {
    run_vni(&[
        ("tu72", "từ"),
        ("to61", "tố"),
        ("ta81", "tắ"),
        ("nu72", "nừ"),
        ("to72", "tờ"),
        ("na82", "nằ"),
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
        ("u77", "u7"),
        ("a88", "a8"),
    ]);
}

// ============================================================
// UPPERCASE
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
        ("U7", "Ư"),
        ("A8", "Ă"),
    ]);
}
