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

fn type_word(engine: &mut Engine, word: &str) -> String {
    engine.clear();
    let mut output = String::new();
    for ch in word.chars() {
        let key = char_to_key(ch);
        if key == 255 { continue; }
        let result = engine.on_key(key, ch.is_uppercase(), false);
        if result.action == 1 {
            let bs = result.backspace as usize;
            for _ in 0..bs.min(output.len()) { output.pop(); }
            for i in 0..result.count as usize {
                if let Some(c) = char::from_u32(result.chars[i]) { output.push(c); }
            }
        } else {
            output.push(ch);
        }
        println!("  '{}' -> output='{}' (action={}, bs={})", ch, output, result.action, result.backspace);
    }
    output
}

#[test]
fn test_kisna_debug() {
    let mut engine = Engine::new();
    engine.set_method(0);
    engine.set_enabled(true);
    engine.set_english_auto_restore(true);
    
    println!("\n=== kiasn (natural order) ===");
    let r1 = type_word(&mut engine, "kiasn");
    println!("Result: '{}'\n", r1);
    
    println!("=== kisna (interleaved) ===");
    let r2 = type_word(&mut engine, "kisna");
    println!("Result: '{}'\n", r2);
}
