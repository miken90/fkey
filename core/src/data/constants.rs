//! Vietnamese Phonological Constants
//!
//! Centralized constants for valid initials, finals, and spelling rules.

use crate::data::keys;

/// Valid single initial consonants
pub const VALID_INITIALS_1: &[u16] = &[
    keys::B,
    keys::C,
    keys::D,
    keys::G,
    keys::H,
    keys::K,
    keys::L,
    keys::M,
    keys::N,
    keys::P,
    keys::Q,
    keys::R,
    keys::S,
    keys::T,
    keys::V,
    keys::X,
];

/// Valid double initial consonants
pub const VALID_INITIALS_2: &[[u16; 2]] = &[
    [keys::C, keys::H], // ch
    [keys::G, keys::H], // gh
    [keys::G, keys::I], // gi
    [keys::K, keys::H], // kh
    [keys::N, keys::G], // ng
    [keys::N, keys::H], // nh
    [keys::P, keys::H], // ph
    [keys::Q, keys::U], // qu
    [keys::T, keys::H], // th
    [keys::T, keys::R], // tr
];

/// Valid single final consonants
pub const VALID_FINALS_1: &[u16] = &[
    keys::C,
    keys::M,
    keys::N,
    keys::P,
    keys::T,
    keys::I,
    keys::Y,
    keys::O,
    keys::U, // semi-vowels
];

/// Valid double final consonants
pub const VALID_FINALS_2: &[[u16; 2]] = &[
    [keys::C, keys::H], // ch
    [keys::N, keys::G], // ng
    [keys::N, keys::H], // nh
];

/// Valid vowel pairs in Vietnamese (V1 → V2)
/// Based on Vietnamese phonology matrix - any pair NOT in this list is invalid.
/// This is more comprehensive than listing invalid patterns individually.
///
/// Matrix source: docs/vietnamese-language-system.md section 3.4.2
///
/// NOTE: Includes Telex intermediate states (same-vowel pairs, e→i) to support
/// delayed transformations like "eie" → "êi" and "aaaa" → "aâ".
pub const VALID_VOWEL_PAIRS: &[[u16; 2]] = &[
    // === Standard Vietnamese diphthongs ===
    // a → i, o, u, y
    [keys::A, keys::I],
    [keys::A, keys::O],
    [keys::A, keys::U],
    [keys::A, keys::Y],
    // â → u, y
    [keys::A, keys::U], // Note: 'â' uses same key as 'a' in raw input
    [keys::A, keys::Y],
    // e → i, o
    [keys::E, keys::I], // Telex: "eie" → "êi" (delayed circumflex)
    [keys::E, keys::O],
    // ê → u
    [keys::E, keys::U], // Note: 'ê' uses same key as 'e' in raw input
    // i → a, ê, u
    [keys::I, keys::A],
    [keys::I, keys::E], // iê (tiên, kiên)
    [keys::I, keys::U],
    // o → a, ă, e, i
    [keys::O, keys::A],
    [keys::O, keys::A], // oă uses same key
    [keys::O, keys::E],
    [keys::O, keys::I],
    // ô → i
    [keys::O, keys::I], // ôi
    // ơ → i
    [keys::O, keys::I], // ơi
    // u → a, â, ê, i, o, ô, y
    [keys::U, keys::A],
    [keys::U, keys::A], // uâ
    [keys::U, keys::E], // uê
    [keys::U, keys::I],
    [keys::U, keys::O],
    [keys::U, keys::O], // uô
    [keys::U, keys::Y],
    // ư → a, i, ơ, u
    [keys::U, keys::A], // ưa - Note: 'ư' uses same key as 'u' in raw input
    [keys::U, keys::I], // ưi
    [keys::U, keys::O], // ươ
    [keys::U, keys::U], // ưu
    // y → ê, u
    [keys::Y, keys::E], // yê (yêu, yến)
    [keys::Y, keys::U], // ỷu (khuỷu - elbow)
    // === Telex intermediate states (same-vowel pairs for doubling) ===
    // These support Telex sequences like "aaaa" → "aâ" where buffer
    // temporarily holds consecutive same vowels during transformation.
    [keys::A, keys::A], // aa → â toggle
    [keys::E, keys::E], // ee → ê toggle
    [keys::O, keys::O], // oo → ô toggle
];

/// Legacy: Invalid vowel patterns (for reference/backward compatibility)
/// Use VALID_VOWEL_PAIRS for comprehensive checking instead.
#[allow(dead_code)]
pub const INVALID_VOWEL_PATTERNS: &[[u16; 2]] = &[
    [keys::E, keys::A], // ea - English: sea, beach, teacher, search
    [keys::O, keys::U], // ou - English: you, our, out, house
    [keys::Y, keys::O], // yo - English: yoke, York, your
];

/// Spelling rules: (consonant, invalid_vowels, description)
/// If consonant + vowel matches, it's INVALID
pub const SPELLING_RULES: &[(&[u16], &[u16], &str)] = &[
    // c before e, i, y → invalid (should use k)
    (&[keys::C], &[keys::E, keys::I, keys::Y], "c before e/i/y"),
    // k before a, o, u → invalid (should use c)
    (&[keys::K], &[keys::A, keys::O, keys::U], "k before a/o/u"),
    // g before e → invalid (should use gh)
    (&[keys::G], &[keys::E], "g before e"),
    // ng before e, i → invalid (should use ngh)
    (&[keys::N, keys::G], &[keys::E, keys::I], "ng before e/i"),
    // gh before a, o, u → invalid (should use g)
    (
        &[keys::G, keys::H],
        &[keys::A, keys::O, keys::U],
        "gh before a/o/u",
    ),
    // ngh before a, o, u → invalid (should use ng)
    (
        &[keys::N, keys::G, keys::H],
        &[keys::A, keys::O, keys::U],
        "ngh before a/o/u",
    ),
];
