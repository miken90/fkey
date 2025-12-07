//! Edge Case Tests - Negative paths and boundary conditions
//!
//! Tests for:
//! - Invalid inputs
//! - Engine state transitions
//! - Boundary conditions
//! - Error recovery

mod common;
use common::type_word;
use gonhanh_core::data::keys;
use gonhanh_core::engine::{Action, Engine};

// ============================================================
// ENGINE STATE: Enabled/Disabled
// ============================================================

#[test]
fn disabled_engine_passes_through() {
    let mut e = Engine::new();
    e.set_enabled(false);

    // When disabled, all keys should pass through (action=None)
    let r = e.on_key(keys::A, false, false);
    assert_eq!(r.action, Action::None as u8);

    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn re_enable_engine_works() {
    let mut e = Engine::new();

    // Disable then re-enable
    e.set_enabled(false);
    e.set_enabled(true);

    // After re-enable, typing should work
    // Note: single 'a' returns Action::None (no transformation yet)
    // Only after tone/mark key does it return Action::Send
    let result = type_word(&mut e, "as");
    assert_eq!(result, "á");
}

// ============================================================
// CTRL/CMD: Pass through modifier keys
// ============================================================

#[test]
fn ctrl_key_passes_through() {
    let mut e = Engine::new();

    // Ctrl+A should pass through (action=None)
    let r = e.on_key(keys::A, false, true);
    assert_eq!(r.action, Action::None as u8);

    // Buffer should be cleared after ctrl key
    let r = e.on_key(keys::S, false, false);
    // 's' alone should not produce tone
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn ctrl_clears_buffer() {
    let mut e = Engine::new();

    // Type 'a', then Ctrl+something, then 's'
    e.on_key(keys::A, false, false);
    e.on_key(keys::C, false, true); // Ctrl+C - clears buffer
    let r = e.on_key(keys::S, false, false);

    // 's' should not apply tone since buffer was cleared
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// METHOD SWITCHING: Telex <-> VNI
// ============================================================

#[test]
fn method_switch_clears_buffer() {
    let mut e = Engine::new();

    // Start in Telex, type 'a'
    e.on_key(keys::A, false, false);

    // Switch to VNI
    e.set_method(1);

    // VNI tone '1' should not work on previous 'a' (buffer cleared)
    let r = e.on_key(keys::N1, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn invalid_method_defaults_to_telex() {
    let mut e = Engine::new();
    e.set_method(99); // Invalid method

    // Should still work as Telex
    let result = type_word(&mut e, "as");
    assert_eq!(result, "á");
}

// ============================================================
// UNKNOWN KEYS: Non-letter keys
// ============================================================

#[test]
fn unknown_key_passes_through() {
    let mut e = Engine::new();

    // Key code 255 is our "unknown" key
    let r = e.on_key(255, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn space_clears_buffer() {
    let mut e = Engine::new();

    // Type 'a', then space, then 's'
    e.on_key(keys::A, false, false);
    e.on_key(keys::SPACE, false, false);
    let r = e.on_key(keys::S, false, false);

    // 's' should not apply tone since space cleared buffer
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// EMPTY BUFFER: Tone/mark without vowel
// ============================================================

#[test]
fn tone_without_vowel_passes_through() {
    let mut e = Engine::new();

    // Telex: 's' alone should pass through
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);

    // 'f' alone
    let r = e.on_key(keys::F, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn mark_without_vowel_passes_through() {
    let mut e = Engine::new();

    // Telex: 'w' alone should pass through
    let r = e.on_key(keys::W, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn vni_tone_without_vowel_passes_through() {
    let mut e = Engine::new();
    e.set_method(1); // VNI

    // '1' alone (sắc tone)
    let r = e.on_key(keys::N1, false, false);
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// CONSECUTIVE BACKSPACE
// ============================================================

#[test]
fn backspace_on_empty_buffer() {
    let mut e = Engine::new();

    // Backspace on empty buffer should pass through
    let r = e.on_key(keys::DELETE, false, false);
    assert_eq!(r.action, Action::None as u8);
}

#[test]
fn multiple_backspace_clears_all() {
    let mut e = Engine::new();

    // Type 'ab', then delete both
    e.on_key(keys::A, false, false);
    e.on_key(keys::B, false, false);
    e.on_key(keys::DELETE, false, false);
    e.on_key(keys::DELETE, false, false);

    // Now 's' should pass through (empty buffer)
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// CONSONANT-ONLY WORDS
// ============================================================

#[test]
fn consonant_only_no_conversion() {
    let mut e = Engine::new();

    let result = type_word(&mut e, "bcd");
    assert_eq!(result, "bcd");
}

#[test]
fn tone_after_consonant_only() {
    let mut e = Engine::new();

    // 'bc' + 's' -> 'bcs' (no vowel to apply tone)
    let result = type_word(&mut e, "bcs");
    assert_eq!(result, "bcs");
}

// ============================================================
// CLEAR BUFFER
// ============================================================

#[test]
fn clear_resets_state() {
    let mut e = Engine::new();

    // Type 'a'
    e.on_key(keys::A, false, false);

    // Clear buffer
    e.clear();

    // 's' should now pass through
    let r = e.on_key(keys::S, false, false);
    assert_eq!(r.action, Action::None as u8);
}

// ============================================================
// MODERN VS CLASSIC ORTHOGRAPHY
// ============================================================

#[test]
fn modern_orthography_hoa() {
    let mut e = Engine::new();
    e.set_modern(true);

    let result = type_word(&mut e, "hoaf");
    // Modern: hoà (tone on last vowel)
    assert_eq!(result, "hoà");
}

#[test]
fn classic_orthography_hoa() {
    let mut e = Engine::new();
    e.set_modern(false);

    let result = type_word(&mut e, "hoaf");
    // Classic: hòa (tone on main vowel)
    assert_eq!(result, "hòa");
}

// ============================================================
// REVERT BEHAVIOR
// ============================================================

#[test]
fn double_tone_reverts() {
    let mut e = Engine::new();

    // á + s -> as (revert)
    let result = type_word(&mut e, "ass");
    assert_eq!(result, "as");
}

#[test]
fn double_mark_reverts() {
    let mut e = Engine::new();

    // â + a -> aa (revert circumflex)
    let result = type_word(&mut e, "aaa");
    assert_eq!(result, "aa");
}

#[test]
fn triple_same_key_behavior() {
    let mut e = Engine::new();

    // aaa -> aa (first adds circumflex, second reverts)
    // Actually: a -> a, aa -> â, aaa -> aa
    let result = type_word(&mut e, "aaaa");
    assert_eq!(result, "aaa"); // â + a -> aa, then +a -> aaa
}
