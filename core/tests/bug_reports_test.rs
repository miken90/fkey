//! Bug reports test cases
//! These tests document expected behavior from user bug reports.

mod common;
use common::telex;
use gonhanh_core::engine::Engine;
use gonhanh_core::utils::type_word;

// =============================================================================
// BUG 1: "did" -> expect "đi"
// Current: ?
// Expected: "đi"
// =============================================================================

#[test]
fn bug1_did_to_di() {
    let mut e = Engine::new();
    let result = type_word(&mut e, "did");
    println!("'did' -> '{}' (expected: 'đi')", result);
    // TODO: Verify expected behavior
    // telex(&[("did", "đi")]);
}

// =============================================================================
// BUG 2: "thowii" -> "thơìi", expected "thờii"
// Current: thơìi (horn on o, huyền on second i)
// Expected: thờii (horn+huyền on o, plain ii)
// =============================================================================

#[test]
fn bug2_thowii() {
    let mut e = Engine::new();
    let result = type_word(&mut e, "thowii");
    println!("'thowii' -> '{}' (expected: 'thờii')", result);
    // TODO: Verify expected behavior
    // telex(&[("thowii", "thờii")]);
}

// =============================================================================
// BUG 3: "uawf"
// GoNhanh: uằ (w applies breve to a)
// OS built-in: ừa (w applies horn to u, creating ưa pattern)
// =============================================================================

#[test]
fn bug3_uawf() {
    let mut e = Engine::new();
    let result = type_word(&mut e, "uawf");
    println!("'uawf' -> '{}' (OS built-in gives: 'ừa')", result);
    // TODO: Decide on expected behavior
    // If following OS built-in: telex(&[("uawf", "ừa")]);
}

// =============================================================================
// BUG 4: "ddd" -> "đd", expected "dd"
// Current: đd (đ + d because third d is just added)
// Expected: dd (third d reverts stroke, returning to raw)
// =============================================================================

#[test]
fn bug4_ddd_revert() {
    let mut e = Engine::new();
    let result = type_word(&mut e, "ddd");
    println!("'ddd' -> '{}' (expected: 'dd')", result);
    // TODO: Change behavior
    // telex(&[("ddd", "dd")]);
}

// =============================================================================
// Additional test: Current expected behaviors
// =============================================================================

#[test]
fn current_dd_makes_stroke() {
    // dd → đ (correct, should not change)
    telex(&[("dd", "đ")]);
}

#[test]
fn current_thowi() {
    // Check what thowi produces
    let mut e = Engine::new();
    let result = type_word(&mut e, "thowi");
    println!("'thowi' -> '{}'", result);
}

#[test]
fn current_uaw() {
    // Check what uaw produces (without f)
    let mut e = Engine::new();
    let result = type_word(&mut e, "uaw");
    println!("'uaw' -> '{}'", result);
}
