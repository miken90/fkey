use gonhanh_core::engine::Engine;
use gonhanh_core::engine::Result;

fn main() {
    let mut e = Engine::new();
    println!("=== Testing 'issue' step by step ===\n");
    
    // i = key 34
    let r = e.on_key(34, false, false);
    println!("'i'(key=34) → action={}, bs={}, count={}", r.action, r.backspace, r.count);
    print_chars(&r);
    
    // s = key 1
    let r = e.on_key(1, false, false);
    println!("'s'(key=1) → action={}, bs={}, count={}", r.action, r.backspace, r.count);
    print_chars(&r);
    
    // s = key 1 (second s)
    let r = e.on_key(1, false, false);
    println!("'s'(key=1) → action={}, bs={}, count={}", r.action, r.backspace, r.count);
    print_chars(&r);
    
    // u = key 32
    let r = e.on_key(32, false, false);
    println!("'u'(key=32) → action={}, bs={}, count={}", r.action, r.backspace, r.count);
    print_chars(&r);
    
    // e = key 14
    let r = e.on_key(14, false, false);
    println!("'e'(key=14) → action={}, bs={}, count={}", r.action, r.backspace, r.count);
    print_chars(&r);
}

fn print_chars(r: &Result) {
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
