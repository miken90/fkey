use gonhanh_core::engine::Engine;
use gonhanh_core::utils::type_word;

fn main() {
    let mut e = Engine::new();
    let result = type_word(&mut e, "issue");
    println!("'issue' -> '{}'", result);
    
    let mut e2 = Engine::new();
    let result2 = type_word(&mut e2, "issue ");
    println!("'issue ' -> '{}'", result2);
}
