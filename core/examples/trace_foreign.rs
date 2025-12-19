use gonhanh_core::data::keys;
use gonhanh_core::engine::validation::is_foreign_word_pattern;

fn main() {
    // Test case: buffer = [U, N], key = N (simulating "wn" after w→ư)
    let buffer_keys = vec![keys::U, keys::N];
    let buffer_tones = vec![2, 0]; // 2 = horn for ư, 0 = none

    println!("Testing is_foreign_word_pattern([U, N], [HORN, NONE], N):");
    let result = is_foreign_word_pattern(&buffer_keys, &buffer_tones, keys::N);
    println!("  Result: {}", result);

    // What about just [U] + N?
    let buffer_keys2 = vec![keys::U];
    let buffer_tones2 = vec![2];
    println!("\nTesting is_foreign_word_pattern([U], [HORN], N):");
    let result2 = is_foreign_word_pattern(&buffer_keys2, &buffer_tones2, keys::N);
    println!("  Result: {}", result2);
}
