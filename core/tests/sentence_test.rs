//! Vietnamese Sentence Tests - Proverbs, Idioms, and Common Phrases
//!
//! Telex rules:
//! - Tones: aa=â, ee=ê, oo=ô, aw=ă, ow=ơ, uw=ư, dd=đ
//! - Marks: s=sắc, f=huyền, r=hỏi, x=ngã, j=nặng
//!
//! IMPORTANT: Only double the vowel if you need the circumflex tone!
//! - tiền = t + ie + e(ê) + n + f → "tieenf"
//! - không = kh + o + o(ô) + ng → "khoong"
//! - cũng = c + u + ng + x → "cungx" (NO double u!)

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
        let result = type_sentence(&mut e, input);
        assert_eq!(result, *expected, "\n[Telex] '{}'\n→ '{}'\n(expected '{}')", input, result, expected);
    }
}

fn run_vni(cases: &[(&str, &str)]) {
    for (input, expected) in cases {
        let mut e = Engine::new();
        e.set_method(1);
        let result = type_sentence(&mut e, input);
        assert_eq!(result, *expected, "\n[VNI] '{}'\n→ '{}'\n(expected '{}')", input, result, expected);
    }
}

// ============================================================
// TELEX: GREETINGS & COMMON PHRASES
// ============================================================

#[test]
fn telex_greetings() {
    run_telex(&[
        ("xin chaof", "xin chào"),           // xin chào
        ("tamj bieetj", "tạm biệt"),         // tạm biệt
        ("camr own", "cảm ơn"),              // cảm ơn
        ("xin looxix", "xin lỗi"),           // xin lỗi (oo=ô, x=ngã)
        ("raats vui dduwowcj gaawpj banj", "rất vui được gặp bạn"),
    ]);
}

#[test]
fn telex_introductions() {
    run_telex(&[
        ("tooi teen laf", "tôi tên là"),
        ("tooi ddeens tufw", "tôi đến từ"),
        ("tooi laf nguwowif vieetj nam", "tôi là người việt nam"),
        ("banj khoer khoong", "bạn khỏe không"),  // khoer = khỏe, khoong = không
        ("tooi khoer", "tôi khỏe"),
    ]);
}

// ============================================================
// TELEX: TỤC NGỮ (Proverbs)
// ============================================================

#[test]
fn telex_proverbs_about_learning() {
    run_telex(&[
        ("hocj hanhf", "học hành"),
        ("hocj mootj bieets muwowif", "học một biết mười"),
        ("ddi mootj ngayf ddangf hocj mootj sangf khoon", "đi một ngày đàng học một sàng khôn"),
        ("khoong thaafy ddoos mayf laf thawngs", "không thầy đố mày là thắng"),
        ("hocj awn hocj nois hocj gois hocj mowr", "học ăn học nói học gói học mở"),
    ]);
}

#[test]
fn telex_proverbs_about_family() {
    run_telex(&[
        ("coong cha nhuw nuis thais sown", "công cha như núi thái sơn"),
        ("nghiax mej nhuw nuwowcs trong nguoonf chayr ra", "nghĩa mẹ như nước trong nguồn chảy ra"),
        ("mootj giotj maus ddaof hown ao hoof nuwowcs lax", "một giọt máu đào hơn ao hồ nước lã"),
        ("anh em nhuw theer chaanf tay", "anh em như thể chân tay"),
    ]);
}

#[test]
fn telex_proverbs_about_work() {
    run_telex(&[
        ("cos coong maif sawts ngayf", "có công mài sắt ngày"),
        ("tay lafm hafm nhai", "tay làm hàm nhai"),
        ("nawngx gif gawtj luas", "nắng gì gặt lúa"),  // aw=ă + x=ngã
        ("muaf xuaan hais vuwown", "mùa xuân hái vườn"),
    ]);
}

#[test]
fn telex_proverbs_about_character() {
    run_telex(&[
        ("toots goox hown ddepj nguwowif", "tốt gỗ hơn đẹp người"),
        ("owr khoong hown lafn laf ddaays", "ở không hơn làn là đấy"),
        ("uoongs nuwowcs nhows nguoonf", "uống nước nhớ nguồn"),  // nhows = nhớ (ow=ơ, s=sắc)
        ("awn quar nhows keer troongf caay", "ăn quả nhớ kẻ trồng cây"),
    ]);
}

#[test]
fn telex_proverbs_about_nature() {
    run_telex(&[
        ("nuwowcs chayr ddas monf", "nước chảy đá mòn"),  // monf = mòn (o, n, f=huyền)
        ("trowif muwaf khi naof mats", "trời mưa khi nào mát"),
        ("las ruwngf veef cooix", "lá rừng về cõi"),
    ]);
}

// ============================================================
// TELEX: THÀNH NGỮ (Idioms)
// ============================================================

#[test]
fn telex_idioms_4_words() {
    run_telex(&[
        ("an cuw lacj nghieepj", "an cư lạc nghiệp"),
        ("ddoongf taam hieepj luwcj", "đồng tâm hiệp lực"),
        ("cos hoofn phucs phaanf", "có hồn phúc phần"),
    ]);
}

#[test]
fn telex_idioms_about_time() {
    run_telex(&[
        ("mootj nawm cos boons muaf", "một năm có bốn mùa"),
        ("thowif gian laf tieenf bacj", "thời gian là tiền bạc"),  // tieenf = tiền
        ("soongs mootj ngayf bieets mootj ngayf", "sống một ngày biết một ngày"),
    ]);
}

#[test]
fn telex_idioms_about_friendship() {
    run_telex(&[
        ("banj bef boons nguwowif", "bạn bè bốn người"),
        ("toots goox hown toots banj", "tốt gỗ hơn tốt bạn"),
        ("gaanf muwcj thif dden gaanf ddenf thif sangs", "gần mực thì đen gần đèn thì sáng"),
    ]);
}

// ============================================================
// TELEX: DAILY CONVERSATIONS
// ============================================================

#[test]
fn telex_daily_conversations() {
    run_telex(&[
        ("hoom nay thowif tieets thees naof", "hôm nay thời tiết thế nào"),
        ("banj ddi ddaau vaayj", "bạn đi đâu vậy"),
        ("tooi ddang ddi lafm", "tôi đang đi làm"),
        ("mootj ly caf phee nhes", "một ly cà phê nhé"),
        ("bao nhieeu tieenf", "bao nhiêu tiền"),  // tieenf = tiền
        ("camr own banj nhieeuf lawms", "cảm ơn bạn nhiều lắm"),
    ]);
}

#[test]
fn telex_food_ordering() {
    run_telex(&[
        ("cho tooi xem thuwcj ddown", "cho tôi xem thực đơn"),
        ("tooi muoons goij mootj phaanf phowr", "tôi muốn gọi một phần phở"),
        ("tooi muoons uoongs traf", "tôi muốn uống trà"),
        ("ddoof awn raats ngon", "đồ ăn rất ngon"),
        ("tinhs tieenf nhes", "tính tiền nhé"),
    ]);
}

// ============================================================
// VNI: PROVERBS
// ============================================================

#[test]
fn vni_proverbs() {
    run_vni(&[
        ("ho5c mo65t bie61t mu8o8i2", "học một biết mười"),
        ("uo61ng nu8o81c nho81 nguo62n", "uống nước nhớ nguồn"),
        ("to61t go64 ho8n d9e5p ngu8o8i2", "tốt gỗ hơn đẹp người"),
        ("nu8o81c cha3y d9a1 mo2n", "nước chảy đá mòn"),
    ]);
}

#[test]
fn vni_greetings() {
    run_vni(&[
        ("xin cha2o", "xin chào"),
        ("ta5m bie65t", "tạm biệt"),
        ("ca3m o8n", "cảm ơn"),
        ("to6i la2 ngu8o8i2 vie65t nam", "tôi là người việt nam"),
    ]);
}

#[test]
fn vni_daily_phrases() {
    run_vni(&[
        ("ho6m nay tho8i2 tie61t the61 na2o", "hôm nay thời tiết thế nào"),
        ("ba5n d9i d9a6u va65y", "bạn đi đâu vậy"),
        ("to6i d9ang d9i la2m", "tôi đang đi làm"),
        ("bao nhie6u tie62n", "bao nhiêu tiền"),
    ]);
}

// ============================================================
// MIXED CASE SENTENCES
// ============================================================

#[test]
fn telex_mixed_case_sentences() {
    run_telex(&[
        ("Xin chaof", "Xin chào"),
        ("Xin Chaof", "Xin Chào"),
        ("Vieetj Nam", "Việt Nam"),
        ("VIEETJ NAM", "VIỆT NAM"),
        ("Thanhf phoos Hoof Chis Minh", "Thành phố Hồ Chí Minh"),
    ]);
}

#[test]
fn vni_mixed_case_sentences() {
    run_vni(&[
        ("Xin cha2o", "Xin chào"),
        ("Vie65t Nam", "Việt Nam"),
        ("Tha2nh pho61 Ho62 Chi1 Minh", "Thành phố Hồ Chí Minh"),
    ]);
}

// ============================================================
// LONG SENTENCES
// ============================================================

#[test]
fn telex_long_sentences() {
    run_telex(&[
        (
            "vieetj nam laf mootj quoocs gia nawmf owr ddoong nam as",
            "việt nam là một quốc gia nằm ở đông nam á"
        ),
        (
            "thur ddoo curar vieetj nam laf thanhf phoos haf nooij",
            "thủ đô của việt nam là thành phố hà nội"
        ),  // curar = của (u, r=hỏi, a)
        (
            "nguwowif vieetj nam raats thaanf thieenj vaf hieeuws khachs",
            "người việt nam rất thân thiện và hiếu khách"
        ),
    ]);
}

#[test]
fn vni_long_sentences() {
    run_vni(&[
        (
            "vie65t nam la2 mo65t quo61c gia na72m o83 d9o6ng nam a1",
            "việt nam là một quốc gia nằm ở đông nam á"
        ),
        (
            "thu3 d9o6 cu3a vie65t nam la2 tha2nh pho61 ha2 no65i",
            "thủ đô của việt nam là thành phố hà nội"
        ),
    ]);
}

// ============================================================
// SPECIAL VIETNAMESE EXPRESSIONS
// ============================================================

#[test]
fn telex_exclamations() {
    run_telex(&[
        ("ooi trowif owi", "ôi trời ơi"),
        ("cheets thaatj", "chết thật"),
        ("tuyeetj vowif", "tuyệt vời"),
        ("kinh khungr", "kinh khủng"),
        ("hay quas", "hay quá"),
        ("ddepj quas", "đẹp quá"),
    ]);
}

#[test]
fn telex_common_expressions() {
    run_telex(&[
        ("khoong sao", "không sao"),
        ("dduwowcj roofif", "được rồi"),
        ("binhf thuwowngf", "bình thường"),
        ("char bieets", "chả biết"),
        ("ai maf bieets", "ai mà biết"),
        ("sao cungx dduwowcj", "sao cũng được"),  // cungx = cũng (NO double u)
    ]);
}

// ============================================================
// POETRY & LITERATURE
// ============================================================

#[test]
fn telex_poetry() {
    run_telex(&[
        (
            "trawm nawm trong cooix nguwowif ta",
            "trăm năm trong cõi người ta"
        ),
        (
            "chuwx taif chuwx meenhj kheos laf ghets nhau",
            "chữ tài chữ mệnh khéo là ghét nhau"
        ),
        (
            "mawtj trowif ddor hawts xuoongs bieenr",
            "mặt trời đỏ hắt xuống biển"
        ),
    ]);
}
