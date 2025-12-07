//! Vietnamese Word Tests
//!
//! Tests common Vietnamese words with various vowel patterns and mark placements

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
// TELEX: SINGLE VOWEL WORDS
// ============================================================

#[test]
fn telex_single_vowel() {
    run_telex(&[
        // Basic words
        ("ba", "ba"),       // three
        ("me", "me"),       // mother (informal)
        ("di", "di"),       // go (command)
        ("bo", "bo"),       // abandon
        ("tu", "tu"),       // monk

        // With marks
        ("mej", "mẹ"),      // mother
        ("bos", "bó"),      // bundle
        ("cus", "cú"),      // owl
        ("cos", "có"),      // have
        ("laf", "là"),      // is
        ("gif", "gì"),      // what
        ("ddif", "đì"),     // (sound)
        ("ddi", "đi"),      // go
    ]);
}

// ============================================================
// TELEX: TWO VOWEL WORDS - CLOSED SYLLABLE (có phụ âm cuối)
// ============================================================

#[test]
fn telex_two_vowels_closed() {
    run_telex(&[
        // Mark on 2nd vowel (standard rule)
        ("toans", "toán"),      // math
        ("hoanf", "hoàn"),      // complete
        ("tieens", "tiến"),     // progress
        ("muoons", "muốn"),     // want
        ("bieenr", "biển"),     // sea
        ("nguoonf", "nguồn"),   // source
        ("cuoocj", "cuộc"),     // event
        ("buoonf", "buồn"),     // sad
        ("thuoocj", "thuộc"),   // belong
        ("chuoongs", "chuống"), // (sound of bells)
    ]);
}

// ============================================================
// TELEX: TWO VOWEL WORDS - OPEN SYLLABLE (không có phụ âm cuối)
// ============================================================

#[test]
fn telex_two_vowels_open_glide_vowel() {
    // Glide + Vowel: oa, oe, uy → mark on 2nd (modern)
    run_telex(&[
        ("hoaf", "hoà"),        // peace
        ("hoas", "hoá"),        // transform
        ("hoax", "hoã"),        // (rare)
        ("hoaj", "hoạ"),        // painting
        ("hoer", "hoẻ"),        // healthy
        ("khoej", "khoẹ"),      // (onomatopoeia)
        ("huyx", "huỹ"),        // (name)
        ("quyf", "quỳ"),        // kneel
        ("quys", "quý"),        // precious
    ]);
}

#[test]
fn telex_two_vowels_open_vowel_glide() {
    // Vowel + Glide: ai, ao, au, oi, ui → mark on 1st
    run_telex(&[
        ("mais", "mái"),        // roof
        ("maif", "mài"),        // sharpen
        ("hair", "hải"),        // sea (adj)
        ("taij", "tại"),        // at/because
        ("caor", "cảo"),        // manuscript
        ("sauj", "sạu"),        // (rare)
        ("daus", "dáu"),        // (rare)
        ("ddaus", "đáu"),       // (rare)
        ("dois", "dói"),        // hungry (dialect)
        ("ddois", "đói"),       // hungry
        ("tuis", "túi"),        // bag
        ("muij", "mụi"),        // (dialect)
    ]);
}

// ============================================================
// TELEX: COMPOUND VOWELS (ươ, iê, uô)
// ============================================================

#[test]
fn telex_compound_uo() {
    // ươ pattern: người, mười, trường
    run_telex(&[
        ("nguwowif", "người"),      // person
        ("muwowif", "mười"),        // ten
        ("truwowngf", "trường"),    // school
        ("luwowix", "lưỡi"),        // tongue
        ("dduwowngf", "đường"),     // road
        ("ruwowuj", "rượu"),        // alcohol
        ("buwowms", "bướm"),        // butterfly
        ("suwowngs", "sướng"),      // happy
    ]);
}

#[test]
fn telex_compound_ie() {
    // iê pattern: việt, tiếng, biển
    run_telex(&[
        ("vieetj", "việt"),         // Vietnam
        ("tieengs", "tiếng"),       // language/sound
        ("bieenr", "biển"),         // sea
        ("ddieeuf", "điều"),        // thing/matter
        ("nhieeux", "nhiễu"),       // disturb
        ("chieeuf", "chiều"),       // afternoon
        ("mieengs", "miếng"),       // piece
        ("kieenj", "kiện"),         // lawsuit
    ]);
}

#[test]
fn telex_compound_uo_circumflex() {
    // uô pattern: muốn, cuộc, buồn
    run_telex(&[
        ("muoons", "muốn"),         // want
        ("cuoocj", "cuộc"),         // event
        ("buoonf", "buồn"),         // sad
        ("thuoocj", "thuộc"),       // belong
        ("luoonf", "luồn"),         // slip through
        ("chuoongs", "chuống"),     // (bells)
        ("nuoocj", "nuộc"),         // (dialect)
    ]);
}

// ============================================================
// TELEX: THREE VOWEL WORDS
// ============================================================

#[test]
fn telex_three_vowels() {
    run_telex(&[
        // ươi: mark on ơ
        ("nguwowif", "người"),
        ("muwowif", "mười"),
        ("cuwowif", "cười"),        // laugh
        ("tuwowif", "tười"),        // (rare)

        // uyê: mark on ê
        ("khuyeens", "khuyến"),     // encourage
        ("nguyeenx", "nguyễn"),     // Nguyen
        ("tuyeenr", "tuyển"),       // select
        ("chuyeenj", "chuyện"),     // story

        // oai: mark on a
        ("ngoais", "ngoái"),        // turn back
        ("khoair", "khoải"),        // (rare)
        ("loaij", "loại"),          // type/kind
        ("xoaif", "xoài"),          // mango (tree)

        // ươu: mark on ơ
        ("ruwowuj", "rượu"),        // alcohol
        ("huwowuj", "hượu"),        // (rare)
    ]);
}

// ============================================================
// TELEX: đ COMBINATIONS
// ============================================================

#[test]
fn telex_d_words() {
    run_telex(&[
        ("ddi", "đi"),              // go
        ("ddeens", "đến"),          // arrive
        ("dduwowngf", "đường"),     // road
        ("ddangf", "đàng"),         // (there)
        ("ddieeuf", "điều"),        // thing
        ("ddoongf", "đồng"),        // field/copper
        ("ddaatj", "đật"),          // (rare)
        ("ddeemf", "đềm"),          // (rare)
        ("ddays", "đáy"),           // bottom
        ("ddoois", "đối"),          // opposite
        ("ddepj", "đẹp"),           // beautiful
        ("ddor", "đỏ"),             // red
        ("dduwngf", "đừng"),        // don't
    ]);
}

// ============================================================
// TELEX: COMMON WORDS BY CATEGORY
// ============================================================

#[test]
fn telex_pronouns() {
    run_telex(&[
        ("tooi", "tôi"),            // I
        ("banj", "bạn"),            // you (friend)
        ("anh", "anh"),             // brother/you (male)
        ("chij", "chị"),            // sister/you (female)
        ("em", "em"),               // younger sibling/you
        ("nos", "nó"),              // it/he/she (informal)
        ("chungs", "chúng"),        // they/we
        ("hoj", "họ"),              // they
        ("minhf", "mình"),          // self/I
    ]);
}

#[test]
fn telex_verbs() {
    run_telex(&[
        ("laf", "là"),              // is
        ("cos", "có"),              // have
        ("ddi", "đi"),              // go
        ("ddeens", "đến"),          // arrive
        ("veef", "về"),             // return
        ("awn", "ăn"),              // eat
        ("uoongs", "uống"),         // drink
        ("ngur", "ngủ"),            // sleep
        ("lafm", "làm"),            // do
        ("nois", "nói"),            // speak
        ("nghix", "nghĩ"),          // think
        ("bieets", "biết"),         // know
        ("hieeur", "hiểu"),         // understand
        ("yeeu", "yêu"),            // love
        ("thichs", "thích"),        // like
        ("ghets", "ghét"),          // hate
        ("soongj", "sộng"),         // (rare)
    ]);
}

#[test]
fn telex_nouns() {
    run_telex(&[
        ("nhaf", "nhà"),            // house
        ("truwowngf", "trường"),    // school
        ("beenhj", "bệnh"),         // sick
        ("vieenj", "viện"),         // institute
        ("coong", "công"),          // public
        ("nuwowcs", "nước"),        // water/country
        ("tieenf", "tiền"),         // money
        ("sach", "sach"),           // book (no mark)
        ("sachs", "sách"),          // book
        ("viees", "viế"),           // (incomplete, testing)
        ("baof", "bào"),            // (plane tool)
    ]);
}

#[test]
fn telex_numbers() {
    run_telex(&[
        ("mootj", "một"),           // one
        ("hai", "hai"),             // two
        ("ba", "ba"),               // three
        ("boons", "bốn"),           // four
        ("nawm", "năm"),            // five
        ("saus", "sáu"),            // six
        ("bayr", "bảy"),            // seven
        ("tams", "tám"),            // eight
        ("chins", "chín"),          // nine
        ("muwowif", "mười"),        // ten
        ("trawm", "trăm"),          // hundred
        ("nghin", "nghin"),         // thousand (no mark)
        ("nghinf", "nghìn"),        // thousand
    ]);
}

#[test]
fn telex_adjectives() {
    run_telex(&[
        ("toots", "tốt"),           // good
        ("xaaus", "xấu"),           // bad/ugly
        ("ddepj", "đẹp"),           // beautiful
        ("lowns", "lớn"),           // big
        ("nhor", "nhỏ"),            // small
        ("caof", "cào"),            // scratch (verb) / high
        ("thaps", "tháp"),          // tower
        ("daif", "dài"),            // long
        ("ngawns", "ngắn"),         // short
        ("noongf", "nồng"),         // intense (smell)
        ("lanhj", "lạnh"),          // cold
        ("noongs", "nống"),         // (raise price)
    ]);
}

// ============================================================
// VNI: COMMON WORDS
// ============================================================

#[test]
fn vni_pronouns() {
    run_vni(&[
        ("to6i", "tôi"),            // I
        ("ba5n", "bạn"),            // you
        ("chi5", "chị"),            // sister
        ("no1", "nó"),              // it
        ("ho5", "họ"),              // they
    ]);
}

#[test]
fn vni_verbs() {
    // VNI: 6=^, 7=horn(ơ,ư), 8=breve(ă), 9=đ
    run_vni(&[
        ("la2", "là"),              // is
        ("co1", "có"),              // have
        ("d9i", "đi"),              // go
        ("d9e61n", "đến"),          // arrive
        ("ve62", "về"),             // return
        ("a8n", "ăn"),              // eat (8=breve)
        ("uo61ng", "uống"),         // drink
        ("ngu3", "ngủ"),            // sleep
        ("la2m", "làm"),            // do
        ("no1i", "nói"),            // speak
        ("nghi4", "nghĩ"),          // think
        ("bie61t", "biết"),         // know
        ("hie63u", "hiểu"),         // understand
        ("ye6u", "yêu"),            // love
    ]);
}

#[test]
fn vni_compound_vowels() {
    // VNI: 6=^, 7=horn(ơ,ư), 8=breve(ă), 9=đ
    run_vni(&[
        // ươ pattern (7=horn)
        ("ngu7o7i2", "người"),      // person
        ("mu7o7i2", "mười"),        // ten
        ("tru7o7ng2", "trường"),    // school
        ("lu7o7i4", "lưỡi"),        // tongue
        ("d9u7o7ng2", "đường"),     // road

        // iê pattern
        ("vie65t", "việt"),         // Vietnam
        ("tie61ng", "tiếng"),       // language
        ("bie63n", "biển"),         // sea
        ("d9ie62u", "điều"),        // thing
        ("nhie64u", "nhiễu"),       // disturb

        // uô pattern
        ("muo61n", "muốn"),         // want
        ("cuo65c", "cuộc"),         // event
        ("buo62n", "buồn"),         // sad
        ("thuo65c", "thuộc"),       // belong
    ]);
}

#[test]
fn vni_three_vowels() {
    // VNI: 6=^, 7=horn(ơ,ư), 8=breve(ă), 9=đ
    run_vni(&[
        ("ngu7o7i1", "ngưới"),      // (variant)
        ("khuye63n", "khuyển"),     // dog (literary)
        ("nguye64n", "nguyễn"),     // Nguyen
        ("ngoa1i", "ngoái"),        // turn back
        ("ru7o7u5", "rượu"),        // alcohol
    ]);
}

// ============================================================
// UPPERCASE WORDS
// ============================================================

#[test]
fn telex_uppercase_words() {
    run_telex(&[
        ("Chaof", "Chào"),          // Hello
        ("CHAOF", "CHÀO"),          // HELLO
        ("Nguwowif", "Người"),      // Person
        ("NGUWOWIF", "NGƯỜI"),      // PERSON
        ("Vieetj", "Việt"),         // Vietnam
        ("DDaats", "Đất"),          // Earth
        ("DDAATS", "ĐẤT"),          // EARTH
    ]);
}

#[test]
fn vni_uppercase_words() {
    // VNI: 6=^, 7=horn(ơ,ư), 8=breve(ă), 9=đ
    run_vni(&[
        ("Cha2o", "Chào"),          // Hello
        ("CHA2O", "CHÀO"),          // HELLO
        ("Ngu7o7i2", "Người"),      // Person (7=horn)
        ("Vie65t", "Việt"),         // Vietnam
        ("D9a61t", "Đất"),          // Earth
    ]);
}

// ============================================================
// EDGE CASES
// ============================================================

#[test]
fn telex_consonant_only() {
    run_telex(&[
        ("bcs", "bcs"),             // no vowel to mark
        ("ddf", "đf"),              // dd→đ, f passes through
        ("xyz", "xyz"),             // all consonants
    ]);
}

#[test]
fn telex_consonant_clusters() {
    run_telex(&[
        ("nguyeenx", "nguyễn"),     // ng + uyễn
        ("nhuwngx", "những"),       // nh + ưng + ngã
        ("phaatj", "phật"),         // ph + ât + nặng
        ("chauj", "chạu"),          // ch + au + nặng
        ("khoongf", "khồng"),       // kh + ô + ng + huyền
        ("ghees", "ghế"),           // gh + ê + sắc
        ("truwowcs", "trước"),      // tr + ươ + c + sắc
    ]);
}

#[test]
fn vni_delayed_tone() {
    // VNI allows typing tone after multiple chars
    // VNI: 6=^, 7=horn(ơ,ư), 8=breve(ă), 9=đ
    run_vni(&[
        ("toi6", "tôi"),            // 6 finds 'o' not 'i'
        ("toi61", "tối"),           // tôi + sắc
        ("nguoi7", "nguơi"),        // 7=horn finds 'o'
        ("nguoi72", "nguời"),       // nguơi + huyền
        ("duong7", "duơng"),        // 7=horn on 'o'
        ("duong72", "duờng"),       // + huyền
        ("muon6", "muôn"),          // 6 on 'o'
        ("muon61", "muốn"),         // + sắc
    ]);
}
