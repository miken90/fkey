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