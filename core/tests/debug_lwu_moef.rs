use gonhanh_core::engine::Engine;
use gonhanh_core::engine::validation;

fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => 0, 's' => 1, 'd' => 2, 'f' => 3, 'h' => 4, 'g' => 5,
        'z' => 6, 'x' => 7, 'c' => 8, 'v' => 9, 'b' => 11, 'q' => 12,
        'w' => 13, 'e' => 14, 'r' => 15, 'y' => 16, 't' => 17, 'o' => 31,
        'u' => 32, 'i' => 34, 'p' => 35, 'l' => 37, 'j' => 38, 'k' => 40,
        'n' => 45, 'm' => 46, _ => 255,
    }
}

fn type_word(engine: &mut Engine, word: &str) -> String {
    engine.clear();
    let mut output = String::new();
    println!("\n=== Typing '{}' ===", word);
    for ch in word.chars() {
        let key = char_to_key(ch);
        if key == 255 { output.push(ch); continue; }
        let result = engine.on_key(key, ch.is_uppercase(), false);
        if result.action == 1 {
            for _ in 0..(result.backspace as usize).min(output.len()) { output.pop(); }
            for i in 0..result.count as usize {
                if let Some(c) = char::from_u32(result.chars[i]) { output.push(c); }
            }
        } else {
            output.push(ch);
        }
        println!("  '{}' -> output='{}' (action={}, bs={})", ch, output, result.action, result.backspace);
    }
    // Space to trigger auto-restore
    let result = engine.on_key(49, false, false);
    if result.action == 1 {
        for _ in 0..(result.backspace as usize).min(output.len()) { output.pop(); }
        for i in 0..result.count as usize {
            if let Some(c) = char::from_u32(result.chars[i]) { output.push(c); }
        }
    } else {
        output.push(' ');
    }
    println!("  SPACE -> FINAL='{}'", output);
    output
}

#[test]
fn test_lwu() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    let result = type_word(&mut engine, "lwu");
    assert_eq!(result, "lưu ", "lwu should stay as lưu (valid Vietnamese)");
}

#[test]
fn test_moef() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    let result = type_word(&mut engine, "moef");
    assert_eq!(result, "moè ", "moef should stay as moè");
}

#[test]
fn test_related_patterns() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    // Other patterns that might have similar issues
    let cases = [
        ("lwu", "lưu "),    // w → ư, valid VN
        ("luu", "lưu "),    // uu → ưu? Actually in Telex uu doesn't produce ưu
        ("moef", "moè "),   // f → huyền, should keep VN
        ("boef", "boè "),   // similar pattern
        ("soef", "soè "),   // similar pattern
    ];

    for (input, expected) in cases {
        engine.clear();
        let result = type_word(&mut engine, input);
        println!("{} -> '{}' (expected '{}')", input, result, expected);
    }
}

#[test]
fn test_validation() {
    // Key constants from engine/keys.rs
    const M: u16 = 46;
    const B: u16 = 11;
    const S: u16 = 1;
    const O: u16 = 31;
    const E: u16 = 14;
    const L: u16 = 37;
    const U: u16 = 32;
    const W: u16 = 13;

    const HORN: u8 = 2;  // From tone constants

    // Test moe vs boe vs soe
    let moe_keys = vec![M, O, E];
    let moe_tones = vec![0, 0, 0];
    println!("moe validation: {}", validation::is_valid_with_tones(&moe_keys, &moe_tones));

    let boe_keys = vec![B, O, E];
    let boe_tones = vec![0, 0, 0];
    println!("boe validation: {}", validation::is_valid_with_tones(&boe_keys, &boe_tones));

    let soe_keys = vec![S, O, E];
    let soe_tones = vec![0, 0, 0];
    println!("soe validation: {}", validation::is_valid_with_tones(&soe_keys, &soe_tones));

    // Test luu with horn
    let luu_keys = vec![L, U, U];
    let luu_tones = vec![0, HORN, 0];  // First U has horn = Ư
    println!("lưu validation: {}", validation::is_valid_with_tones(&luu_keys, &luu_tones));
}
