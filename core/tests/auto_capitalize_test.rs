//! Auto-Capitalize Tests
//!
//! Tests for automatic capitalization after sentence-ending punctuation.
//! Feature: Tự viết hoa đầu câu
//!
//! Triggers: . ! ? Enter
//! Default: OFF
//!
//! NOTE: These tests use Vietnamese patterns because the engine transforms
//! characters to Vietnamese. English-like inputs would be converted.

mod common;
use common::telex_auto_capitalize;
use gonhanh_core::data::keys;
use gonhanh_core::engine::Engine;
use gonhanh_core::utils::type_word;

// ============================================================
// BASIC DOT (.) TESTS
// ============================================================

#[test]
fn dot_basic_capitalize() {
    // Basic: dot followed by space and letter → capitalize
    telex_auto_capitalize(&[
        ("chaof. ban", "chào. Ban"),
        ("xin. chaof", "xin. Chào"),
        ("ok. ddi", "ok. Đi"),
    ]);
}

#[test]
fn dot_vietnamese_word() {
    // Vietnamese word after dot
    telex_auto_capitalize(&[
        ("xin chaof. banj", "xin chào. Bạn"),
        ("toots. lams", "tốt. Lám"), // Use 's' for sắc tone
        ("dduwowcj. rooif", "được. Rồi"),
    ]);
}

#[test]
fn dot_multiple_spaces() {
    // Multiple spaces after dot - should still capitalize
    telex_auto_capitalize(&[("ok.  ban", "ok.  Ban"), ("ok.   di", "ok.   Di")]);
}

// ============================================================
// EXCLAMATION MARK (!) TESTS
// ============================================================

#[test]
fn exclamation_basic() {
    // Exclamation mark triggers capitalize
    telex_auto_capitalize(&[("hay! tuyeetj", "hay! Tuyệt"), ("hay! quas", "hay! Quá")]);
}

#[test]
fn exclamation_multiple() {
    // Multiple exclamation marks
    telex_auto_capitalize(&[("hay!! di", "hay!! Di")]);
}

// ============================================================
// QUESTION MARK (?) TESTS
// ============================================================

#[test]
fn question_basic() {
    // Question mark triggers capitalize
    telex_auto_capitalize(&[("sao? taij", "sao? Tại"), ("ddaau? oof", "đâu? Ồ")]);
}

#[test]
fn question_multiple() {
    // Multiple question marks
    telex_auto_capitalize(&[("gi?? di", "gi?? Di")]);
}

// ============================================================
// ENTER KEY TESTS
// ============================================================

#[test]
fn enter_basic() {
    // Enter (represented as newline in type_word would need special handling)
    // Note: Enter is handled in engine but type_word may not simulate it
    // This test verifies the engine logic works
    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Type "hello", press Enter, type "d" (using 'd' to avoid Vietnamese w→ư)
    // We use the engine directly since type_word doesn't handle Enter well
    use gonhanh_core::data::keys;

    // Type "xin" (simple word)
    for &key in &[keys::X, keys::I, keys::N] {
        e.on_key_ext(key, false, false, false);
    }

    // Press Enter (should set pending_capitalize)
    e.on_key_ext(keys::RETURN, false, false, false);

    // Type 'd' - should be capitalized
    let result = e.on_key_ext(keys::D, false, false, false);

    // Check that the result contains uppercase 'D'
    assert_eq!(result.action, 1); // Action::Send
    assert!(result.count > 0);
    let first_char = char::from_u32(result.chars[0]).unwrap();
    assert_eq!(first_char, 'D', "Expected 'D' but got '{}'", first_char);
}

// ============================================================
// NUMBER AFTER DOT (NO CAPITALIZE)
// ============================================================

#[test]
fn number_after_dot_no_capitalize() {
    // Number after dot should NOT trigger capitalize on next letter
    telex_auto_capitalize(&[
        ("1.5k", "1.5k"),               // Decimal number
        ("192.168.1.1", "192.168.1.1"), // IP address
    ]);
}

#[test]
fn number_resets_pending() {
    // After number, next letter should NOT be capitalized
    telex_auto_capitalize(&[
        ("ok. 5k ban", "ok. 5k ban"), // Number resets, "ban" stays lowercase
    ]);
}

// ============================================================
// NON-SENTENCE PUNCTUATION (NO TRIGGER)
// ============================================================

#[test]
fn comma_no_capitalize() {
    // Comma should NOT trigger capitalize
    telex_auto_capitalize(&[("xin, chaof", "xin, chào"), ("ban, toi", "ban, toi")]);
}

#[test]
fn semicolon_no_capitalize() {
    // Semicolon should NOT trigger capitalize
    telex_auto_capitalize(&[("a; b", "a; b")]);
}

#[test]
fn colon_no_capitalize() {
    // Colon should NOT trigger capitalize (Vietnamese doesn't capitalize after colon)
    telex_auto_capitalize(&[("ban: di", "ban: di")]);
}

// ============================================================
// CONSECUTIVE SENTENCES
// ============================================================

#[test]
fn multiple_sentences() {
    // Multiple sentences in a row
    telex_auto_capitalize(&[
        ("a. b. c", "a. B. C"),
        ("xin. chaof. banj", "xin. Chào. Bạn"),
    ]);
}

#[test]
fn mixed_punctuation() {
    // Mixed sentence-ending punctuation
    telex_auto_capitalize(&[("ok. hay! sao? ddi", "ok. Hay! Sao? Đi")]);
}

// ============================================================
// ALREADY UPPERCASE (NO CHANGE)
// ============================================================

#[test]
fn already_uppercase() {
    // If user already typed uppercase, it stays uppercase
    telex_auto_capitalize(&[
        ("ok. Ban", "ok. Ban"), // User typed 'B' with shift
        ("hay! Di", "hay! Di"),
    ]);
}

// ============================================================
// FEATURE OFF (DEFAULT)
// ============================================================

#[test]
fn feature_off_no_capitalize() {
    // When feature is OFF, no auto-capitalize
    let mut e = Engine::new();
    // auto_capitalize defaults to false

    let result = type_word(&mut e, "ok. ban");
    assert_eq!(
        result, "ok. ban",
        "Should NOT capitalize when feature is OFF"
    );
}

// ============================================================
// EDGE CASES
// ============================================================

#[test]
fn empty_after_dot() {
    // Just dot, no following text
    telex_auto_capitalize(&[("ok.", "ok.")]);
}

#[test]
fn dot_at_start() {
    // Dot at very start (edge case)
    telex_auto_capitalize(&[(". di", ". Di")]);
}

#[test]
fn no_space_no_capitalize() {
    // Issue #185: No space after punctuation = no capitalize
    // This fixes "google.Com" problem
    // Note: Using inputs that don't trigger Vietnamese transforms (no vowel combos)
    telex_auto_capitalize(&[
        ("x.y", "x.y"),                 // Simple: no capitalize without space
        ("192.168.1.1", "192.168.1.1"), // IP address stays lowercase
        ("a.b.c", "a.b.c"),             // Multiple dots without spaces
    ]);
}

#[test]
fn abbreviations_known_tradeoff() {
    // Issue #185: Abbreviations like "v.v." should NOT auto-capitalize
    // Previously this was a known trade-off, but now fixed
    telex_auto_capitalize(&[
        ("v.v. tieeps", "v.v. Tiếp"), // Fixed: no capitalize without space
    ]);
}

// ============================================================
// SPECIAL CHARACTERS AFTER PUNCTUATION
// ============================================================

#[test]
fn quote_after_dot() {
    // Quote after dot - quote doesn't reset, next letter capitalizes
    telex_auto_capitalize(&[("ok. \"ban\"", "ok. \"Ban\"")]);
}

#[test]
fn parenthesis_after_dot() {
    // Parenthesis after dot
    telex_auto_capitalize(&[("ok. (di)", "ok. (Di)")]);
}

// ============================================================
// VIETNAMESE DIACRITICS AFTER CAPITALIZE
// ============================================================

#[test]
fn vietnamese_tone_after_capitalize() {
    // Vietnamese word with tone after capitalize trigger
    telex_auto_capitalize(&[
        ("ok. ddis", "ok. Đí"),
        ("ok. nhanh", "ok. Nhanh"),
        ("toots. tuyeetj", "tốt. Tuyệt"), // Use 's' for sắc tone on 'tốt'
    ]);
}

#[test]
fn vietnamese_complex_after_capitalize() {
    // Complex Vietnamese words
    telex_auto_capitalize(&[
        ("ok. nguwowif", "ok. Người"),
        ("ok. dduwowcj", "ok. Được"),
        ("ok. khong", "ok. Khong"), // Use simple 'khong' without double 'g'
    ]);
}

// ============================================================
// DELETE AND RETYPE SCENARIOS
// ============================================================

#[test]
fn delete_retype_restores_capitalize() {
    // Type after ". " → capitalize → delete → retype → should capitalize again
    // Use '<' as backspace in type_word
    telex_auto_capitalize(&[
        ("ok. b<c", "ok. C"),       // Type b (→B), delete, type c (→C)
        ("ok. ban<<<di", "ok. Di"), // Type "ban" (→"Ban"), delete all 3 chars, type "di" (→"Di")
    ]);
}

#[test]
fn delete_partial_no_restore() {
    // Partial delete should NOT restore capitalize (buffer not empty)
    telex_auto_capitalize(&[
        ("ok. ban<n", "ok. Ban"), // Type "ban" (→"Ban"), delete 'n', type 'n' again
    ]);
}

#[test]
fn enter_delete_retype_restores_capitalize() {
    // Enter → type letter → delete → retype → should capitalize again
    use gonhanh_core::data::keys;

    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Type "xin"
    for &key in &[keys::X, keys::I, keys::N] {
        e.on_key_ext(key, false, false, false);
    }

    // Press Enter (should set pending_capitalize)
    e.on_key_ext(keys::RETURN, false, false, false);

    // Type 'd' - should be capitalized to 'D'
    let result = e.on_key_ext(keys::D, false, false, false);
    assert_eq!(result.action, 1); // Action::Send
    let first_char = char::from_u32(result.chars[0]).unwrap();
    assert_eq!(first_char, 'D', "First letter after Enter should be 'D'");

    // Delete 'D' with backspace
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type 'b' - should be capitalized to 'B' again
    let result2 = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(result2.action, 1); // Action::Send
    let second_char = char::from_u32(result2.chars[0]).unwrap();
    assert_eq!(
        second_char, 'B',
        "After delete, next letter should still capitalize to 'B'"
    );
}

// ============================================================
// ARROW KEYS AND NAVIGATION (NO RESET)
// ============================================================

#[test]
fn arrow_keys_preserve_pending() {
    // Arrow keys should NOT reset pending_capitalize
    // Issue #185: Need space after punctuation for capitalize
    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Type "ok." + space (to set pending_capitalize)
    for &key in &[keys::O, keys::K] {
        e.on_key_ext(key, false, false, false);
    }
    e.on_key_ext(keys::DOT, false, false, false);
    e.on_key_ext(keys::SPACE, false, false, false); // Need space for pending_capitalize

    // Press arrow keys - should NOT reset pending
    e.on_key_ext(keys::LEFT, false, false, false);
    e.on_key_ext(keys::RIGHT, false, false, false);
    e.on_key_ext(keys::UP, false, false, false);
    e.on_key_ext(keys::DOWN, false, false, false);

    // Type "c" - should be capitalized
    let r = e.on_key_ext(keys::C, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action after arrows");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'C', "Arrow keys should preserve pending");
}

#[test]
fn tab_preserves_pending() {
    // Tab should NOT reset pending_capitalize
    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Press Enter to set pending
    e.on_key_ext(keys::RETURN, false, false, false);

    // Press Tab - should NOT reset pending
    e.on_key_ext(keys::TAB, false, false, false);

    // Type "a" - should be capitalized
    let r = e.on_key_ext(keys::A, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action after Tab");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'A', "Tab should preserve pending");
}

// ============================================================
// SELECTION DELETE (CLEAR RESTORES PENDING)
// ============================================================

#[test]
fn clear_restores_pending_capitalize() {
    // When user selects text and deletes, clear() should restore pending_capitalize
    // Issue #185: Need space after punctuation for capitalize
    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Type "ok." + space (to set pending_capitalize)
    for &key in &[keys::O, keys::K] {
        e.on_key_ext(key, false, false, false);
    }
    e.on_key_ext(keys::DOT, false, false, false);
    e.on_key_ext(keys::SPACE, false, false, false); // Need space for pending_capitalize

    // Type "ban" - 'b' becomes 'B' due to auto-capitalize
    let r = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(r.action, 1);
    assert_eq!(char::from_u32(r.chars[0]).unwrap(), 'B');
    e.on_key_ext(keys::A, false, false, false);
    e.on_key_ext(keys::N, false, false, false);

    // Simulate selection-delete by calling clear()
    e.clear();

    // After clear, pending_capitalize should be restored
    // Type "c" - should be capitalized to "C"
    let r = e.on_key_ext(keys::C, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action after clear()");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'C', "After clear(), should capitalize");
}

#[test]
fn delete_past_buffer_keeps_pending() {
    // This test verifies behavior when deleting past buffer end.
    // 
    // IMPORTANT: When a break character (like DOT) is typed, both buffer and
    // word_history are cleared. So after "ok." the buffer and history are empty.
    // When SPACE is pressed after DOT, nothing is pushed to history (buffer is empty).
    // When "Ban" is typed, it starts a new word. When deleted, we can't restore "ok"
    // because it was never in history.
    //
    // The capitalize state IS preserved across this deletion because:
    // - When "B" was typed, auto_capitalize_used was set
    // - When buffer becomes empty (after deleting "Ban"), pending_capitalize is restored
    // - When space is deleted, typed_after_space=true prevents resetting pending_capitalize
    //
    // However, since there's no word to restore (history is empty), buffer stays empty.
    // When "c" is typed, the restored_pending_clear check triggers clear(), but 
    // pending_capitalize should still be true.
    
    let mut e = Engine::new();
    e.set_auto_capitalize(true);

    // Type "ok."
    for &key in &[keys::O, keys::K] {
        e.on_key_ext(key, false, false, false);
    }
    e.on_key_ext(keys::DOT, false, false, false);
    // After DOT: buffer=empty, word_history=empty, saw_sentence_ending=true

    // Type space - buffer is empty, nothing pushed to history
    e.on_key_ext(keys::SPACE, false, false, false);
    // After SPACE: pending_capitalize=true (from saw_sentence_ending)

    // Type "ban" - should capitalize to "Ban"
    let r = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(char::from_u32(r.chars[0]).unwrap(), 'B');
    e.on_key_ext(keys::A, false, false, false);
    e.on_key_ext(keys::N, false, false, false);

    // Delete "Ban" (3 chars)
    e.on_key_ext(keys::DELETE, false, false, false);
    e.on_key_ext(keys::DELETE, false, false, false);
    e.on_key_ext(keys::DELETE, false, false, false);
    // After deleting "B": buffer=empty, auto_capitalize_used=true → pending_capitalize=true

    // Delete space (buffer already empty, spaces_after_commit=0)
    // Since spaces_after_commit was 0 (SPACE after empty buffer doesn't set it),
    // this delete doesn't trigger restore logic.
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "c" - should still be capitalized because pending_capitalize was restored
    // when buffer became empty after deleting "B"
    let r = e.on_key_ext(keys::C, false, false, false);
    // Actually, since spaces_after_commit was 0, the "delete space" just sets
    // has_non_letter_prefix=true and resets last_break_key.
    // pending_capitalize should still be true from step when buffer became empty.
    assert_eq!(r.action, 1, "Expected Send action");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'C', "After deleting to period, should capitalize");
}

// ============================================================
// BUG FIX: Delete dot should reset auto-capitalize
// ============================================================

#[test]
fn backspace_delete_dot_resets_capitalize() {
    // Bug scenario: "Hello. " → backspace (delete space) → backspace (delete dot) → "w"
    // Expected: "w" should NOT be capitalized after deleting the dot
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." - sets saw_sentence_ending
    e.on_key_ext(keys::DOT, false, false, false);

    // Type space - sets pending_capitalize
    e.on_key_ext(keys::SPACE, false, false, false);

    // Backspace to delete space
    e.on_key_ext(keys::DELETE, false, false, false);

    // Backspace to delete dot - should reset pending_capitalize
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "w" - should NOT be capitalized
    let r = e.on_key_ext(keys::W, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'ư', "After deleting dot, should NOT capitalize (ư not Ư)");
}

#[test]
fn backspace_delete_question_mark_resets_capitalize() {
    // Same test with question mark
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "?" (Shift+/)
    e.on_key_ext(keys::SLASH, false, true, false);

    // Type space
    e.on_key_ext(keys::SPACE, false, false, false);

    // Backspace to delete space
    e.on_key_ext(keys::DELETE, false, false, false);

    // Backspace to delete question mark
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "w" - should NOT be capitalized (w → ư in Telex)
    let r = e.on_key_ext(keys::W, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'ư', "After deleting ?, should NOT capitalize");
}

#[test]
fn backspace_delete_exclamation_resets_capitalize() {
    // Same test with exclamation mark
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "!" (Shift+1)
    e.on_key_ext(keys::N1, false, true, false);

    // Type space
    e.on_key_ext(keys::SPACE, false, false, false);

    // Backspace to delete space
    e.on_key_ext(keys::DELETE, false, false, false);

    // Backspace to delete exclamation
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "w" - should NOT be capitalized (w → ư in Telex)
    let r = e.on_key_ext(keys::W, false, false, false);
    assert_eq!(r.action, 1, "Expected Send action");
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'ư', "After deleting !, should NOT capitalize");
}

#[test]
fn backspace_delete_dot_no_space_resets_capitalize() {
    // Bug scenario: "ok." (NO space) → backspace (delete dot) → space → "b"
    // Expected: "b" should NOT be capitalized
    // This is different from the with-space test above
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." - sets saw_sentence_ending (but NOT pending_capitalize, that needs space)
    e.on_key_ext(keys::DOT, false, false, false);

    // Backspace to delete dot - should reset saw_sentence_ending
    // Since no space was typed, pending_capitalize is still false
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type space - should NOT set pending_capitalize because saw_sentence_ending was reset
    e.on_key_ext(keys::SPACE, false, false, false);

    // Type "b" - should NOT be capitalized
    let r = e.on_key_ext(keys::B, false, false, false);
    // Either action 0 (no change needed) or action 1 with lowercase 'b'
    if r.action == 1 {
        let ch = char::from_u32(r.chars[0]).unwrap();
        assert!(ch == 'b', "After deleting dot (no space), 'b' should be lowercase, got '{}'", ch);
    }
    // If action is 0, it means the 'b' was passed through unchanged (which is fine)
}

#[test]
fn backspace_delete_dot_direct_no_space() {
    // Simpler test: "ok." → backspace → "b" (no space in between)
    // User might just delete dot and continue typing without space
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." - sets saw_sentence_ending
    e.on_key_ext(keys::DOT, false, false, false);

    // Backspace to delete dot
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "b" directly (no space) - should NOT be capitalized
    // saw_sentence_ending should have been reset by backspace
    let r = e.on_key_ext(keys::B, false, false, false);
    if r.action == 1 {
        let ch = char::from_u32(r.chars[0]).unwrap();
        assert!(ch == 'b', "After deleting dot directly, 'b' should be lowercase, got '{}'", ch);
    }
}

#[test]
fn backspace_delete_after_auto_capitalized_letter() {
    // Bug scenario: "ok." → space → "B" (auto-cap) → delete B → delete space → delete dot → "b"
    // Expected: "b" should NOT be capitalized
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." - sets saw_sentence_ending
    e.on_key_ext(keys::DOT, false, false, false);

    // Type space - sets pending_capitalize
    e.on_key_ext(keys::SPACE, false, false, false);

    // Type "b" - becomes "B" due to auto-capitalize
    let r = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(r.action, 1);
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'B', "First letter after '. ' should be capitalized");

    // Delete "B"
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete space (buffer is now empty, this should reset typed_after_space)
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete dot (buffer is still empty, last_break_key was consumed but typed_after_space is now false)
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type space again
    e.on_key_ext(keys::SPACE, false, false, false);

    // Type "b" - should NOT be capitalized because we deleted the dot
    let r = e.on_key_ext(keys::B, false, false, false);
    if r.action == 1 {
        let ch = char::from_u32(r.chars[0]).unwrap();
        assert!(ch == 'b', "After deleting B, space, and dot, 'b' should be lowercase, got '{}'", ch);
    }
}

#[test]
fn backspace_delete_after_auto_capitalized_letter_no_final_space() {
    // Bug scenario: "ok." → space → "B" (auto-cap) → delete B → delete space → delete dot → "b"
    // WITHOUT typing another space before "b"
    // Expected: "b" should NOT be capitalized
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." - sets saw_sentence_ending
    e.on_key_ext(keys::DOT, false, false, false);

    // Type space - sets pending_capitalize
    e.on_key_ext(keys::SPACE, false, false, false);

    // Type "b" - becomes "B" due to auto-capitalize
    let r = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(r.action, 1);
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'B', "First letter after '. ' should be capitalized");

    // Delete "B"
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete space
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete dot
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "b" DIRECTLY (no space) - should NOT be capitalized
    let r = e.on_key_ext(keys::B, false, false, false);
    if r.action == 1 {
        let ch = char::from_u32(r.chars[0]).unwrap();
        assert!(ch == 'b', "After deleting B, space, and dot directly, 'b' should be lowercase, got '{}'", ch);
    }
}

#[test]
fn backspace_multi_letter_then_delete_dot() {
    // Scenario closer to real usage:
    // "ok." → space → "Ban" (auto-cap "B") → delete n → delete a → delete B → delete space → delete dot → "b"
    // This tests when user types more than one letter before deleting
    let mut e = Engine::new();
    e.set_method(0); // Telex
    e.set_auto_capitalize(true);

    // Type "ok"
    e.on_key_ext(keys::O, false, false, false);
    e.on_key_ext(keys::K, false, false, false);

    // Type "." 
    e.on_key_ext(keys::DOT, false, false, false);

    // Type space
    e.on_key_ext(keys::SPACE, false, false, false);

    // Type "Ban" - "B" is auto-capitalized
    let r = e.on_key_ext(keys::B, false, false, false);
    assert_eq!(r.action, 1);
    let ch = char::from_u32(r.chars[0]).unwrap();
    assert_eq!(ch, 'B', "First letter should be capitalized");
    
    e.on_key_ext(keys::A, false, false, false);
    e.on_key_ext(keys::N, false, false, false);

    // Delete "n"
    e.on_key_ext(keys::DELETE, false, false, false);
    // Delete "a"
    e.on_key_ext(keys::DELETE, false, false, false);
    // Delete "B"
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete space
    e.on_key_ext(keys::DELETE, false, false, false);

    // Delete dot
    e.on_key_ext(keys::DELETE, false, false, false);

    // Type "b" DIRECTLY - should NOT be capitalized
    let r = e.on_key_ext(keys::B, false, false, false);
    if r.action == 1 {
        let ch = char::from_u32(r.chars[0]).unwrap();
        assert!(ch == 'b', "After deleting multi-letter word, space, and dot, 'b' should be lowercase, got '{}'", ch);
    }
}
