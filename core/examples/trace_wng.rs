use gonhanh_core::engine::Engine;
use gonhanh_core::utils::type_word;

fn main() {
    println!("Testing 'uwng':");
    let mut e1 = Engine::new();
    let r1 = type_word(&mut e1, "uwng");
    println!("  Result: '{}'", r1);

    println!("\nTesting 'wng':");
    let mut e2 = Engine::new();
    let r2 = type_word(&mut e2, "wng");
    println!("  Result: '{}'", r2);

    println!("\nTesting 'wng ' (with space):");
    let mut e3 = Engine::new();
    let r3 = type_word(&mut e3, "wng ");
    println!("  Result: '{}'", r3);
}
