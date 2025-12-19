use gonhanh_core::engine::Engine;

fn main() {
    let mut e = Engine::new();
    println!("=== Testing 'wng' step by step ===\n");

    // w = key 13
    let r = e.on_key(13, false, false);
    println!(
        "'w'(key=13) → action={}, bs={}, count={}",
        r.action, r.backspace, r.count
    );
    if r.count > 0 {
        print!("  chars: ");
        for i in 0..r.count as usize {
            if let Some(c) = char::from_u32(r.chars[i]) {
                print!("{}", c);
            }
        }
        println!();
    }

    // n = key 45
    let r = e.on_key(45, false, false);
    println!(
        "'n'(key=45) → action={}, bs={}, count={}",
        r.action, r.backspace, r.count
    );
    if r.count > 0 {
        print!("  chars: ");
        for i in 0..r.count as usize {
            if let Some(c) = char::from_u32(r.chars[i]) {
                print!("{}", c);
            }
        }
        println!();
    }

    // g = key 5
    let r = e.on_key(5, false, false);
    println!(
        "'g'(key=5) → action={}, bs={}, count={}",
        r.action, r.backspace, r.count
    );
    if r.count > 0 {
        print!("  chars: ");
        for i in 0..r.count as usize {
            if let Some(c) = char::from_u32(r.chars[i]) {
                print!("{}", c);
            }
        }
        println!();
    }
}
