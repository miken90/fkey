use gonhanh_core::engine::Engine;

fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => 0, 's' => 1, 'd' => 2, 'f' => 3, 'h' => 4, 'g' => 5,
        'z' => 6, 'x' => 7, 'c' => 8, 'v' => 9, 'b' => 11, 'q' => 12,
        'w' => 13, 'e' => 14, 'r' => 15, 'y' => 16, 't' => 17, 'o' => 31,
        'u' => 32, 'i' => 34, 'p' => 35, 'l' => 37, 'j' => 38, 'k' => 40,
        'n' => 45, 'm' => 46, _ => 255,
    }
}

fn type_word_debug(engine: &mut Engine, word: &str) -> String {
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

    // Space
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
fn test_gusta() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    let result = type_word_debug(&mut engine, "gusta");
    assert_eq!(result, "gusta ", "gusta should restore");
}

#[test]
fn test_many_words() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);

    // Words that might have character jumping issue
    let words = [
        "gusta", "vista", "krista", "mista", "lista",
        "costa", "pasta", "fiesta", "siesta",
        "busta", "justa", "musta", "rusta",
    ];

    for word in words {
        engine.clear();
        let result = type_word_debug(&mut engine, word);
        let expected = format!("{} ", word);
        println!("{} -> '{}' (expected '{}')", word, result, expected);
    }
}
