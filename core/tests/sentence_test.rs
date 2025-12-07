//! Vietnamese Sentence Tests - Proverbs, Idioms, and Common Phrases
//!
//! STRICT Vietnamese expectations - tests should fail if engine produces wrong output

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
        ' ' => keys::SPACE,
        _ => 255,
    }
}

fn type_sentence(e: &mut Engine, input: &str) -> String {
    let mut screen = String::new();
    for c in input.chars() {
        let key = char_to_key(c);
        if key == keys::SPACE {
            screen.push(' ');
            e.on_key(key, false, false);
            continue;
        }
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
        let result = type_sentence(&mut e, input);
        assert_eq!(
            result, *expected,
            "\n[Telex] '{}'\n→ '{}'\n(expected '{}')",
            input, result, expected
        );
    }
}

fn run_vni(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        e.set_method(1);
        let result = type_sentence(&mut e, input);
        assert_eq!(
            result, *expected,
            "\n[VNI] '{}'\n→ '{}'\n(expected '{}')",
            input, result, expected
        );
    }
}

// ============================================================
// TELEX: GREETINGS
// ============================================================

#[test]
fn telex_greetings() {
    run_telex(&[
        ("xin chaof", "xin chào"),
        ("tamj bieetj", "tạm biệt"),
        ("camr own", "cảm ơn"),
        ("xin looxix", "xin lỗi"),
    ]);
}

// ============================================================
// TELEX: PROVERBS (TỤC NGỮ)
// ============================================================

#[test]
fn telex_proverbs() {
    run_telex(&[
        // Học một biết mười
        ("hocj mootj bieets muwowif", "học một biết mười"),
        // Đi một ngày đàng học một sàng khôn
        (
            "ddi mootj ngayf ddangf hocj mootj sangf khoon",
            "đi một ngày đàng học một sàng khôn",
        ),
        // Tốt gỗ hơn đẹp người
        ("toots goox hown ddepj nguwowif", "tốt gỗ hơn đẹp người"),
        // Uống nước nhớ nguồn
        ("uoongs nuwowcs nhows nguoonf", "uống nước nhớ nguồn"),
        // Nước chảy đá mòn
        ("nuwowcs chayr ddas monf", "nước chảy đá mòn"),
    ]);
}

// ============================================================
// TELEX: IDIOMS (THÀNH NGỮ)
// ============================================================

#[test]
fn telex_idioms() {
    run_telex(&[
        ("an cuw lacj nghieepj", "an cư lạc nghiệp"),
        ("ddoongf taam hieepj luwcj", "đồng tâm hiệp lực"),
        ("thowif gian laf tieenf bacj", "thời gian là tiền bạc"),
    ]);
}

// ============================================================
// TELEX: DAILY CONVERSATIONS
// ============================================================

#[test]
fn telex_daily_conversations() {
    run_telex(&[
        (
            "hoom nay thowif tieets thees naof",
            "hôm nay thời tiết thế nào",
        ),
        ("banj ddi ddaau vaayj", "bạn đi đâu vậy"),
        ("tooi ddang ddi lafm", "tôi đang đi làm"),
        ("mootj ly caf phee nhes", "một ly cà phê nhé"),
        ("bao nhieeu tieenf", "bao nhiêu tiền"),
    ]);
}

// ============================================================
// TELEX: FOOD
// ============================================================

#[test]
fn telex_food() {
    run_telex(&[
        ("cho tooi xem thuwcj ddown", "cho tôi xem thực đơn"),
        (
            "tooi muoons goij mootj phaanf phowr",
            "tôi muốn gọi một phần phở",
        ),
        ("ddoof awn raats ngon", "đồ ăn rất ngon"),
        ("tinhs tieenf nhes", "tính tiền nhé"),
    ]);
}

// ============================================================
// TELEX: EXPRESSIONS
// ============================================================

#[test]
fn telex_expressions() {
    run_telex(&[
        ("khoong sao", "không sao"),
        ("dduwowcj roofif", "được rồi"),
        ("binhf thuwowngf", "bình thường"),
        ("sao cungx dduwowcj", "sao cũng được"),
        ("tuyeetj vowif", "tuyệt vời"),
        ("ddepj quas", "đẹp quá"),
    ]);
}

// ============================================================
// TELEX: POETRY
// ============================================================

#[test]
fn telex_poetry() {
    run_telex(&[
        // Truyện Kiều - Nguyễn Du
        (
            "trawm nawm trong coix nguwowif ta",
            "trăm năm trong cõi người ta",
        ),
        (
            "chuwx taif chuwx meenhj kheos laf ghets nhau",
            "chữ tài chữ mệnh khéo là ghét nhau",
        ),
    ]);
}

// ============================================================
// TELEX: MIXED CASE
// ============================================================

#[test]
fn telex_mixed_case() {
    run_telex(&[
        ("Xin chaof", "Xin chào"),
        ("Vieetj Nam", "Việt Nam"),
        ("VIEETJ NAM", "VIỆT NAM"),
        ("Thanhf phoos Hoof Chis Minh", "Thành phố Hồ Chí Minh"),
    ]);
}

// ============================================================
// TELEX: LONG SENTENCES
// ============================================================

#[test]
fn telex_long_sentences() {
    run_telex(&[
        (
            "vieetj nam laf mootj quoocs gia nawmf owr ddoong nam as",
            "việt nam là một quốc gia nằm ở đông nam á",
        ),
        (
            "thur ddoo cura vieetj nam laf thanhf phoos haf nooij",
            "thủ đô của việt nam là thành phố hà nội",
        ),
    ]);
}

// ============================================================
// VNI TESTS
// VNI: 6=circumflex(^), 7=horn(ơ,ư), 8=breve(ă), 9=đ
// ============================================================

#[test]
fn vni_proverbs() {
    run_vni(&[
        ("ho5c mo65t bie61t mu7o7i2", "học một biết mười"),
        ("uo61ng nu7o71c nho71 nguo62n", "uống nước nhớ nguồn"),
        ("to61t go64 ho7n d9e5p ngu7o7i2", "tốt gỗ hơn đẹp người"),
        ("nu7o71c cha3y d9a1 mo2n", "nước chảy đá mòn"),
    ]);
}

#[test]
fn vni_greetings() {
    run_vni(&[
        ("xin cha2o", "xin chào"),
        ("ta5m bie65t", "tạm biệt"),
        ("ca3m o7n", "cảm ơn"),
    ]);
}

#[test]
fn vni_daily() {
    run_vni(&[
        (
            "ho6m nay tho7i2 tie61t the61 na2o",
            "hôm nay thời tiết thế nào",
        ),
        ("ba5n d9i d9a6u va65y", "bạn đi đâu vậy"),
        ("bao nhie6u tie62n", "bao nhiêu tiền"),
    ]);
}

#[test]
fn vni_mixed_case() {
    run_vni(&[
        ("Xin cha2o", "Xin chào"),
        ("Vie65t Nam", "Việt Nam"),
        ("Tha2nh pho61 Ho62 Chi1 Minh", "Thành phố Hồ Chí Minh"),
    ]);
}

#[test]
fn vni_long_sentences() {
    run_vni(&[
        (
            "vie65t nam la2 mo65t quo61c gia na82m o73 d9o6ng nam a1",
            "việt nam là một quốc gia nằm ở đông nam á",
        ),
        (
            "thu3 d9o6 cu3a vie65t nam la2 tha2nh pho61 ha2 no65i",
            "thủ đô của việt nam là thành phố hà nội",
        ),
    ]);
}
