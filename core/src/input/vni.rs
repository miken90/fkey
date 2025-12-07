//! VNI input method
//!
//! Marks: 1=sắc, 2=huyền, 3=hỏi, 4=ngã, 5=nặng
//! Tones: 6=^ (â,ê,ô), 7=horn (ơ,ư), 8=breve (ă), 9=đ
//! Remove: 0
//!
//! Reference: https://en.wikipedia.org/wiki/VNI

use super::Method;
use crate::data::keys;

pub struct Vni;

impl Method for Vni {
    fn is_mark(&self, key: u16) -> Option<u8> {
        match key {
            keys::N1 => Some(1), // sắc
            keys::N2 => Some(2), // huyền
            keys::N3 => Some(3), // hỏi
            keys::N4 => Some(4), // ngã
            keys::N5 => Some(5), // nặng
            _ => None,
        }
    }

    fn is_tone(&self, key: u16, prev: Option<u16>) -> Option<u8> {
        let prev = prev?;

        match key {
            // 6 -> circumflex (^) for a, e, o → â, ê, ô
            keys::N6 if matches!(prev, keys::A | keys::E | keys::O) => Some(1),

            // 7 -> horn for o, u → ơ, ư
            keys::N7 if matches!(prev, keys::O | keys::U) => Some(2),

            // 8 -> breve for a only → ă
            keys::N8 if prev == keys::A => Some(2),

            _ => None,
        }
    }

    /// VNI: Find any valid vowel in buffer for tone
    /// Example: "toi6" -> find 'o' (not 'i') for circumflex tone
    fn is_tone_for(&self, key: u16, vowels: &[u16]) -> Option<(u8, u16)> {
        match key {
            // 6 -> circumflex (^) for a, e, o - find first matching
            keys::N6 => {
                for &v in vowels.iter().rev() {
                    if matches!(v, keys::A | keys::E | keys::O) {
                        return Some((1, v));
                    }
                }
                None
            }

            // 7 -> horn for o, u → ơ, ư
            keys::N7 => {
                for &v in vowels.iter().rev() {
                    if matches!(v, keys::O | keys::U) {
                        return Some((2, v));
                    }
                }
                None
            }

            // 8 -> breve for a only → ă
            keys::N8 => {
                for &v in vowels.iter().rev() {
                    if v == keys::A {
                        return Some((2, v));
                    }
                }
                None
            }

            _ => None,
        }
    }

    fn is_d(&self, key: u16, prev: Option<u16>) -> bool {
        // d9 -> đ (immediate mode)
        key == keys::N9 && prev == Some(keys::D)
    }

    /// VNI delayed đ: find 'd' anywhere in buffer
    /// Example: "dung9" -> find 'd' at start, convert to 'đ'
    fn is_d_for(&self, key: u16, buffer_keys: &[u16]) -> bool {
        key == keys::N9 && buffer_keys.contains(&keys::D)
    }

    fn is_remove(&self, key: u16) -> bool {
        key == keys::N0
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_marks() {
        let v = Vni;
        assert_eq!(v.is_mark(keys::N1), Some(1));
        assert_eq!(v.is_mark(keys::N5), Some(5));
        assert_eq!(v.is_mark(keys::A), None);
    }

    #[test]
    fn test_tones() {
        let v = Vni;
        assert_eq!(v.is_tone(keys::N6, Some(keys::A)), Some(1)); // a6 -> â
        assert_eq!(v.is_tone(keys::N7, Some(keys::O)), Some(2)); // o7 -> ơ
        assert_eq!(v.is_tone(keys::N7, Some(keys::U)), Some(2)); // u7 -> ư
        assert_eq!(v.is_tone(keys::N8, Some(keys::A)), Some(2)); // a8 -> ă
    }

    #[test]
    fn test_d() {
        let v = Vni;
        assert!(v.is_d(keys::N9, Some(keys::D)));
        assert!(!v.is_d(keys::N9, Some(keys::A)));
    }
}
