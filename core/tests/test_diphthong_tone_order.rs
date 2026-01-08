use gonhanh_core::engine::Engine;

fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => 0,
        's' => 1,
        'd' => 2,
        'f' => 3,
        'h' => 4,
        'g' => 5,
        'z' => 6,
        'x' => 7,
        'c' => 8,
        'v' => 9,
        'b' => 11,
        'q' => 12,
        'w' => 13,
        'e' => 14,
        'r' => 15,
        'y' => 16,
        't' => 17,
        'o' => 31,
        'u' => 32,
        'i' => 34,
        'p' => 35,
        'l' => 37,
        'j' => 38,
        'k' => 40,
        'n' => 45,
        'm' => 46,
        _ => 255,
    }
}

fn type_word(engine: &mut Engine, word: &str) -> String {
    engine.clear();
    let mut output = String::new();

    for ch in word.chars() {
        let key = char_to_key(ch);
        if key == 255 {
            output.push(ch);
            continue;
        }
        let result = engine.on_key(key, ch.is_uppercase(), false);
        if result.action == 1 {
            let bs = result.backspace as usize;
            for _ in 0..bs.min(output.len()) {
                output.pop();
            }
            for i in 0..result.count as usize {
                if let Some(c) = char::from_u32(result.chars[i]) {
                    output.push(c);
                }
            }
        } else {
            output.push(ch);
        }
    }
    output
}

fn type_word_with_space(engine: &mut Engine, word: &str) -> String {
    engine.clear();
    let mut output = String::new();

    for ch in word.chars() {
        let key = char_to_key(ch);
        if key == 255 {
            output.push(ch);
            continue;
        }
        let result = engine.on_key(key, ch.is_uppercase(), false);
        if result.action == 1 {
            let bs = result.backspace as usize;
            for _ in 0..bs.min(output.len()) {
                output.pop();
            }
            for i in 0..result.count as usize {
                if let Some(c) = char::from_u32(result.chars[i]) {
                    output.push(c);
                }
            }
        } else {
            output.push(ch);
        }
    }

    // Type space to trigger auto-restore
    let result = engine.on_key(49, false, false);
    if result.action == 1 {
        let bs = result.backspace as usize;
        for _ in 0..bs.min(output.len()) {
            output.pop();
        }
        for i in 0..result.count as usize {
            if let Some(c) = char::from_u32(result.chars[i]) {
                output.push(c);
            }
        }
    } else {
        output.push(' ');
    }
    output
}

#[test]
fn test_diphthong_ia_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0); // Telex
    engine.set_enabled(true);

    println!("=== IA diphthong tone order ===");
    // Standard order: mia + s = mía
    let r1 = type_word(&mut engine, "mias");
    println!("mias → '{}'", r1);

    // Interleaved order: mi + s + a = mía (should be same)
    let r2 = type_word(&mut engine, "misa");
    println!("misa → '{}'", r2);

    // With final consonant - THIS IS THE BUG CASE
    let r3 = type_word(&mut engine, "kiasn");
    println!("kiasn → '{}' (expected: kían)", r3);
    let r4 = type_word_debug(&mut engine, "kisna");
    println!("kisna → '{}' (expected: kían)", r4);

    // Assertions
    assert_eq!(r1, r2, "mias and misa should produce same result");
    assert_eq!(r3, r4, "kiasn and kisna should produce same result");
}

fn type_word_debug(engine: &mut Engine, word: &str) -> String {
    engine.clear();
    let mut output = String::new();

    for ch in word.chars() {
        let key = char_to_key(ch);
        if key == 255 {
            output.push(ch);
            continue;
        }
        let result = engine.on_key(key, ch.is_uppercase(), false);
        if result.action == 1 {
            let bs = result.backspace as usize;
            for _ in 0..bs.min(output.len()) {
                output.pop();
            }
            for i in 0..result.count as usize {
                if let Some(c) = char::from_u32(result.chars[i]) {
                    output.push(c);
                }
            }
        } else {
            output.push(ch);
        }
        println!(
            "  after '{}': output = '{}', backspace = {}",
            ch, output, result.backspace
        );
    }
    output
}

#[test]
fn test_diphthong_ia_auto_restore() {
    let mut engine = Engine::new();
    engine.set_method(0); // Telex
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    println!("=== IA diphthong with auto-restore ===");

    // Test with space to trigger auto-restore
    // "kína" is invalid VN (tone on wrong vowel), should be detected and restored
    let r1 = type_word_with_space(&mut engine, "kiasn");
    println!("kiasn + space → '{}'", r1);
    let r2 = type_word_with_space(&mut engine, "kisna");
    println!("kisna + space → '{}'", r2);

    // "kían" - check if it's valid VN
    let r3 = type_word_with_space(&mut engine, "kian");
    println!("kian + space → '{}' (should be 'kian ')", r3);

    // "miasn" - check what happens
    let r4 = type_word_with_space(&mut engine, "miasn");
    println!("miasn + space → '{}'", r4);
    let r5 = type_word_with_space(&mut engine, "misna");
    println!("misna + space → '{}'", r5);
}

#[test]
fn test_diphthong_oa_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== OA diphthong tone order ===");
    // Standard: hoa + s = hoá
    let r1 = type_word(&mut engine, "hoas");
    println!("hoas → '{}'", r1);

    // Interleaved: ho + s + a = hoá
    let r2 = type_word(&mut engine, "hosa");
    println!("hosa → '{}'", r2);

    // toàn
    let r3 = type_word(&mut engine, "toafn");
    println!("toafn → '{}'", r3);
    let r4 = type_word(&mut engine, "tofan");
    println!("tofan → '{}'", r4);

    // hoàn
    let r5 = type_word(&mut engine, "hoafn");
    println!("hoafn → '{}'", r5);
    let r6 = type_word(&mut engine, "hofan");
    println!("hofan → '{}'", r6);
}

#[test]
fn test_diphthong_ua_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== UA diphthong tone order ===");
    // Standard: mua + s = múa
    let r1 = type_word(&mut engine, "muas");
    println!("muas → '{}'", r1);

    // Interleaved: mu + s + a = múa
    let r2 = type_word(&mut engine, "musa");
    println!("musa → '{}'", r2);

    // tuần
    let r3 = type_word(&mut engine, "tuafn");
    println!("tuafn → '{}'", r3);
    let r4 = type_word(&mut engine, "tufan");
    println!("tufan → '{}'", r4);
}

#[test]
fn test_diphthong_ue_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== UE diphthong tone order ===");
    // huế
    let r1 = type_word(&mut engine, "hues");
    println!("hues → '{}'", r1);
    let r2 = type_word(&mut engine, "huse");
    println!("huse → '{}'", r2);

    // tuệ
    let r3 = type_word(&mut engine, "tueej");
    println!("tueej → '{}'", r3);
    let r4 = type_word(&mut engine, "tueje");
    println!("tueje → '{}'", r4);
}

#[test]
fn test_diphthong_ai_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== AI diphthong tone order ===");
    // mái
    let r1 = type_word(&mut engine, "mais");
    println!("mais → '{}'", r1);
    let r2 = type_word(&mut engine, "masi");
    println!("masi → '{}'", r2);

    // tài
    let r3 = type_word(&mut engine, "taif");
    println!("taif → '{}'", r3);
    let r4 = type_word(&mut engine, "tafi");
    println!("tafi → '{}'", r4);
}

#[test]
fn test_diphthong_ao_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== AO diphthong tone order ===");
    // cáo
    let r1 = type_word(&mut engine, "caos");
    println!("caos → '{}'", r1);
    let r2 = type_word(&mut engine, "caso");
    println!("caso → '{}'", r2);

    // tào
    let r3 = type_word(&mut engine, "taof");
    println!("taof → '{}'", r3);
    let r4 = type_word(&mut engine, "tafo");
    println!("tafo → '{}'", r4);
}

#[test]
fn test_diphthong_oi_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== OI diphthong tone order ===");
    // mói (không có từ này nhưng test pattern)
    let r1 = type_word(&mut engine, "mois");
    println!("mois → '{}'", r1);
    let r2 = type_word(&mut engine, "mosi");
    println!("mosi → '{}'", r2);

    // tối
    let r3 = type_word(&mut engine, "tois");
    println!("tois → '{}'", r3);
    let r4 = type_word(&mut engine, "tosi");
    println!("tosi → '{}'", r4);
}

#[test]
fn test_diphthong_uo_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== UO diphthong tone order ===");
    // muốn
    let r1 = type_word(&mut engine, "muosn");
    println!("muosn → '{}'", r1);
    let r2 = type_word(&mut engine, "musno");
    println!("musno → '{}'", r2);

    // cuốn
    let r3 = type_word(&mut engine, "cuosn");
    println!("cuosn → '{}'", r3);
    let r4 = type_word(&mut engine, "cusno");
    println!("cusno → '{}'", r4);
}

#[test]
fn test_triphthong_uoi_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== UOI triphthong tone order ===");
    // người (uoi in ươi)
    let r1 = type_word(&mut engine, "nguoiws");
    println!("nguoiws → '{}'", r1);
    let r2 = type_word(&mut engine, "nguowis");
    println!("nguowis → '{}'", r2);

    // cười
    let r3 = type_word(&mut engine, "cuowif");
    println!("cuowif → '{}'", r3);
    let r4 = type_word(&mut engine, "cuwowif");
    println!("cuwowif → '{}'", r4);
}

#[test]
fn test_triphthong_oai_tone_order() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);

    println!("=== OAI triphthong tone order ===");
    // ngoài
    let r1 = type_word(&mut engine, "ngoaif");
    println!("ngoaif → '{}'", r1);
    let r2 = type_word(&mut engine, "ngofai");
    println!("ngofai → '{}'", r2);

    // hoài
    let r3 = type_word(&mut engine, "hoaif");
    println!("hoaif → '{}'", r3);
    let r4 = type_word(&mut engine, "hofai");
    println!("hofai → '{}'", r4);
}
