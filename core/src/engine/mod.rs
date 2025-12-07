//! Vietnamese IME Engine

pub mod buffer;

use buffer::{Buffer, Char, MAX};
use crate::data::{chars, keys};
use crate::input;

/// Convert key code to character (letters and numbers)
fn key_to_char(key: u16, caps: bool) -> Option<char> {
    // Letters
    let ch = match key {
        keys::A => 'a', keys::B => 'b', keys::C => 'c', keys::D => 'd',
        keys::E => 'e', keys::F => 'f', keys::G => 'g', keys::H => 'h',
        keys::I => 'i', keys::J => 'j', keys::K => 'k', keys::L => 'l',
        keys::M => 'm', keys::N => 'n', keys::O => 'o', keys::P => 'p',
        keys::Q => 'q', keys::R => 'r', keys::S => 's', keys::T => 't',
        keys::U => 'u', keys::V => 'v', keys::W => 'w', keys::X => 'x',
        keys::Y => 'y', keys::Z => 'z',
        // Numbers (for VNI revert)
        keys::N0 => return Some('0'),
        keys::N1 => return Some('1'),
        keys::N2 => return Some('2'),
        keys::N3 => return Some('3'),
        keys::N4 => return Some('4'),
        keys::N5 => return Some('5'),
        keys::N6 => return Some('6'),
        keys::N7 => return Some('7'),
        keys::N8 => return Some('8'),
        keys::N9 => return Some('9'),
        _ => return None,
    };
    Some(if caps { ch.to_ascii_uppercase() } else { ch })
}

/// Engine action result
#[repr(u8)]
#[derive(Clone, Copy, Debug, PartialEq)]
pub enum Action {
    None = 0,    // Pass through
    Send = 1,    // Delete + send new chars
    Restore = 2, // Invalid, restore original
}

/// Result for FFI - fields ordered for alignment
#[repr(C)]
pub struct Result {
    pub chars: [u32; MAX],  // 128 bytes (32 x 4)
    pub action: u8,
    pub backspace: u8,
    pub count: u8,
    pub _pad: u8,           // padding to align
}

impl Result {
    pub fn none() -> Self {
        Self {
            chars: [0; MAX],
            action: Action::None as u8,
            backspace: 0,
            count: 0,
            _pad: 0,
        }
    }

    pub fn send(backspace: u8, chars: &[char]) -> Self {
        let mut result = Self {
            chars: [0; MAX],
            action: Action::Send as u8,
            backspace,
            count: chars.len() as u8,
            _pad: 0,
        };
        for (i, &c) in chars.iter().enumerate() {
            if i < MAX {
                result.chars[i] = c as u32;
            }
        }
        result
    }
}

/// Main engine
pub struct Engine {
    buf: Buffer,
    method: u8,   // 0=Telex, 1=VNI
    enabled: bool,
    modern: bool, // oà vs òa
    last_transform: Option<(u16, u8, u8)>, // (key, transform_type, value) for revert
    // transform_type: 1=mark, 2=tone
}

impl Default for Engine {
    fn default() -> Self {
        Self::new()
    }
}

impl Engine {
    pub fn new() -> Self {
        Self {
            buf: Buffer::new(),
            method: 0,
            enabled: true,
            modern: true,
            last_transform: None,
        }
    }

    pub fn set_method(&mut self, method: u8) {
        self.method = method;
    }

    pub fn set_enabled(&mut self, enabled: bool) {
        self.enabled = enabled;
        if !enabled {
            self.buf.clear();
        }
    }

    pub fn set_modern(&mut self, modern: bool) {
        self.modern = modern;
    }

    /// Handle key event
    pub fn on_key(&mut self, key: u16, caps: bool, ctrl: bool) -> Result {
        // Disabled or ctrl held
        if !self.enabled || ctrl {
            self.buf.clear();
            return Result::none();
        }

        // Break keys (space, punctuation, etc.)
        if keys::is_break(key) {
            self.buf.clear();
            return Result::none();
        }

        // Backspace
        if key == keys::DELETE {
            self.buf.pop();
            return Result::none();
        }

        // Process Vietnamese
        self.process(key, caps)
    }

    fn process(&mut self, key: u16, caps: bool) -> Result {
        let m = input::get(self.method);
        let prev_key = self.buf.last().map(|c| c.key);

        // Check đ (dd or d9)
        if m.is_d(key, prev_key) {
            self.last_transform = None;
            return self.handle_d();
        }

        // Check tone (aa, aw, a6, a7, etc.)
        // For VNI, use is_tone_for to search entire buffer
        let vowel_keys: Vec<u16> = self.buf.iter()
            .filter(|c| keys::is_vowel(c.key))
            .map(|c| c.key)
            .collect();

        if let Some((tone, target_key)) = m.is_tone_for(key, &vowel_keys) {
            // Check for double-key revert: same key pressed again
            if let Some((last_key, 2, _)) = self.last_transform {
                if last_key == key {
                    return self.revert_tone(key, caps);
                }
            }
            return self.handle_tone_with_key(key, tone, target_key);
        }

        // Check mark (s/f/r/x/j or 1-5)
        if let Some(mark) = m.is_mark(key) {
            // Check for double-key revert: same key pressed again
            if let Some((last_key, 1, _)) = self.last_transform {
                if last_key == key {
                    return self.revert_mark(key, caps);
                }
            }
            return self.handle_mark_with_key(key, mark);
        }

        // Check remove mark (z or 0)
        if m.is_remove(key) {
            self.last_transform = None;
            return self.handle_remove();
        }

        // Normal key - add to buffer, clear last transform
        self.last_transform = None;
        if keys::is_letter(key) {
            self.buf.push(Char::new(key, caps));
        } else {
            // Non-letter breaks word
            self.buf.clear();
        }

        Result::none()
    }

    /// Handle đ (dd/d9)
    fn handle_d(&mut self) -> Result {
        // Get caps from the D in buffer before popping it
        let caps = self.buf.last().map(|c| c.caps).unwrap_or(false);
        self.buf.pop();
        Result::send(1, &[chars::get_d(caps)])
    }

    /// Handle tone with key tracking for revert
    fn handle_tone_with_key(&mut self, key: u16, tone: u8, target_key: u16) -> Result {
        if let Some(pos) = self.buf.find_vowel_by_key(target_key) {
            if let Some(c) = self.buf.get_mut(pos) {
                c.tone = tone;
                self.last_transform = Some((key, 2, tone)); // 2 = tone transform
                return self.rebuild_from(pos);
            }
        }
        Result::none()
    }

    /// Revert tone transform: remove tone and output the key character
    fn revert_tone(&mut self, key: u16, caps: bool) -> Result {
        self.last_transform = None;

        // Find vowel with tone and remove it
        let vowels = self.buf.find_vowels();
        for &i in vowels.iter().rev() {
            if let Some(c) = self.buf.get_mut(i) {
                if c.tone > 0 {
                    c.tone = 0;
                    // Rebuild from this position, then append the key char
                    let mut result = self.rebuild_from(i);
                    // Add the revert key to output
                    if let Some(ch) = key_to_char(key, caps) {
                        if result.count < MAX as u8 {
                            result.chars[result.count as usize] = ch as u32;
                            result.count += 1;
                        }
                    }
                    return result;
                }
            }
        }
        Result::none()
    }

    /// Handle mark with key tracking for revert
    fn handle_mark_with_key(&mut self, key: u16, mark: u8) -> Result {
        let vowels = self.buf.find_vowels();
        if vowels.is_empty() {
            return Result::none();
        }

        // Find position to place mark
        let pos = self.find_mark_pos(&vowels);

        if let Some(c) = self.buf.get_mut(pos) {
            c.mark = mark;
            self.last_transform = Some((key, 1, mark)); // 1 = mark transform
            return self.rebuild_from(pos);
        }

        Result::none()
    }

    /// Revert mark transform: remove mark and output the key character
    fn revert_mark(&mut self, key: u16, caps: bool) -> Result {
        self.last_transform = None;

        // Find vowel with mark and remove it
        let vowels = self.buf.find_vowels();
        for &i in vowels.iter().rev() {
            if let Some(c) = self.buf.get_mut(i) {
                if c.mark > 0 {
                    c.mark = 0;
                    // Rebuild from this position, then append the key char
                    let mut result = self.rebuild_from(i);
                    // Add the revert key to output
                    if let Some(ch) = key_to_char(key, caps) {
                        if result.count < MAX as u8 {
                            result.chars[result.count as usize] = ch as u32;
                            result.count += 1;
                        }
                    }
                    return result;
                }
            }
        }
        Result::none()
    }

    /// Handle mark (sắc, huyền, hỏi, ngã, nặng) - legacy, without key tracking
    #[allow(dead_code)]
    fn handle_mark(&mut self, mark: u8) -> Result {
        let vowels = self.buf.find_vowels();
        if vowels.is_empty() {
            return Result::none();
        }

        let pos = self.find_mark_pos(&vowels);

        if let Some(c) = self.buf.get_mut(pos) {
            c.mark = mark;
            return self.rebuild_from(pos);
        }

        Result::none()
    }

    /// Handle remove mark
    fn handle_remove(&mut self) -> Result {
        let vowels = self.buf.find_vowels();
        if vowels.is_empty() {
            return Result::none();
        }

        // Find vowel with mark and remove it
        for &i in vowels.iter().rev() {
            if let Some(c) = self.buf.get_mut(i) {
                if c.mark > 0 {
                    c.mark = 0;
                    return self.rebuild_from(i);
                }
                if c.tone > 0 {
                    c.tone = 0;
                    return self.rebuild_from(i);
                }
            }
        }

        Result::none()
    }

    /// Find position to place mark based on Vietnamese tone rules
    /// Rules: https://vi.wikipedia.org/wiki/Quy_tắc_đặt_dấu_thanh_của_chữ_Quốc_ngữ
    ///
    /// Vietnamese syllable structure: (C)(w)V(G/C)
    /// - C = consonant, w = medial (âm đệm), V = main vowel (âm chính)
    /// - G = glide ending (i/y, u/o), C = consonant ending
    ///
    /// Mark placement principles:
    /// 1. Single vowel: mark on it
    /// 2. Two vowels with final consonant: mark on 2nd (toàn, muốn)
    /// 3. Medial + main vowel (oa, oe, uy, ươ): mark on main (2nd)
    /// 4. Main vowel + glide (ai, ao, oi): mark on main (1st)
    /// 5. Three vowels: mark on middle or the main vowel with diacritic
    fn find_mark_pos(&self, vowels: &[usize]) -> usize {
        if vowels.is_empty() {
            return 0;
        }

        let n = vowels.len();
        if n == 1 {
            return vowels[0];
        }

        let last_vowel_pos = *vowels.last().unwrap();
        let has_final = self.has_final_consonant(last_vowel_pos);

        // Get vowel info: (buffer_position, key, has_diacritic)
        let vowel_info: Vec<(usize, u16, bool)> = vowels.iter()
            .filter_map(|&i| self.buf.get(i).map(|c| (i, c.key, c.tone > 0)))
            .collect();

        // Extract just keys for pattern matching
        let vowel_keys: Vec<u16> = vowel_info.iter().map(|&(_, k, _)| k).collect();

        if n == 2 {
            let (v1, v2) = (vowel_keys[0], vowel_keys[1]);
            let v1_has_tone = vowel_info.get(0).map(|&(_, _, t)| t).unwrap_or(false);
            let v2_has_tone = vowel_info.get(1).map(|&(_, _, t)| t).unwrap_or(false);

            // Check for compound vowels where 2nd is main vowel:
            // - ươ (người, mười): O is main
            // - uô (muốn, cuộc): Ô is main (when O has diacritic)
            // - iê (việt, tiếng): Ê is main
            // Note: "ưa" (sứa) is different - mark on ư (1st)
            if v1 == keys::U && v2 == keys::O {
                return vowels[1]; // ươ/uô: mark on O
            }
            if v1 == keys::I && v2 == keys::E {
                return vowels[1]; // iê: mark on ê
            }

            // If 1st vowel has diacritic (ư, ơ, etc.) and 2nd doesn't
            // This handles: ưa (sứa), ơi (đời)
            if v1_has_tone && !v2_has_tone {
                return vowels[0]; // Mark on 1st (the modified vowel)
            }

            // If 2nd vowel has diacritic (ê, ô, ơ, ư, â, ă)
            if v2_has_tone {
                return vowels[1];
            }

            if has_final {
                // With final consonant: mark on 2nd vowel
                // toàn, hoàng, tiến, muốn
                return vowels[1];
            }

            // Open syllable - check if it's glide+vowel (oa, oe, uy)
            if self.is_glide_vowel_pair(v1, v2) {
                // oa, oe, uy, ua (qua): Modern=2nd, Old=1st
                if self.modern { vowels[1] } else { vowels[0] }
            } else {
                // ao, ai, au, oi, ui, etc.: mark on 1st (main vowel)
                vowels[0]
            }
        } else {
            // 3+ vowels (ươi, uyê, oai, uoi, ươu, etc.)
            // Find the main vowel position

            // Priority 1: For ươi pattern (U, O, I), mark on O
            if n == 3 && vowel_keys[0] == keys::U && vowel_keys[1] == keys::O {
                return vowels[1]; // người, mười, lưỡi
            }

            // Priority 2: Look for E or A in middle positions (uyê, oai)
            for (idx, &key) in vowel_keys.iter().enumerate() {
                if matches!(key, keys::E | keys::A) && idx > 0 && idx < n - 1 {
                    return vowels[idx];
                }
            }

            // Priority 3: Vowel with diacritic that's not the first position
            for (idx, &(pos, _, has_tone)) in vowel_info.iter().enumerate() {
                if has_tone && idx > 0 {
                    return pos;
                }
            }

            // Priority 4: Any vowel with diacritic
            for &(pos, _, has_tone) in vowel_info.iter() {
                if has_tone {
                    return pos;
                }
            }

            // Default: middle vowel
            vowels[n / 2]
        }
    }

    /// Check if v1+v2 is a glide+vowel pair (âm đệm + âm chính)
    /// oa, oe, uy where first is glide (o before a/e, u before y)
    fn is_glide_vowel_pair(&self, v1: u16, v2: u16) -> bool {
        matches!(
            (v1, v2),
            (keys::O, keys::A) |  // oa
            (keys::O, keys::E) |  // oe
            (keys::U, keys::Y) |  // uy
            (keys::U, keys::A) |  // ua (qua)
            (keys::U, keys::E)    // ue (que)
        )
    }

    /// Check if there's a consonant after the last vowel
    fn has_final_consonant(&self, last_vowel_pos: usize) -> bool {
        for i in (last_vowel_pos + 1)..self.buf.len() {
            if let Some(c) = self.buf.get(i) {
                if keys::is_consonant(c.key) {
                    return true;
                }
            }
        }
        false
    }

    /// Rebuild output from position
    fn rebuild_from(&self, from: usize) -> Result {
        let mut output = Vec::new();
        let mut backspace = 0u8;

        for i in from..self.buf.len() {
            if let Some(c) = self.buf.get(i) {
                backspace += 1;
                // Try to convert with tone/mark (for vowels)
                if let Some(ch) = chars::to_char(c.key, c.caps, c.tone, c.mark) {
                    output.push(ch);
                } else {
                    // Consonant - convert key to char directly
                    if let Some(ch) = key_to_char(c.key, c.caps) {
                        output.push(ch);
                    }
                }
            }
        }

        if output.is_empty() {
            Result::none()
        } else {
            Result::send(backspace, &output)
        }
    }

    /// Clear buffer (new session)
    pub fn clear(&mut self) {
        self.buf.clear();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn type_keys(e: &mut Engine, s: &str) -> Vec<Result> {
        s.chars().map(|c| {
            let key = match c.to_ascii_lowercase() {
                'a' => keys::A, 'b' => keys::B, 'c' => keys::C, 'd' => keys::D,
                'e' => keys::E, 'f' => keys::F, 'g' => keys::G, 'h' => keys::H,
                'i' => keys::I, 'j' => keys::J, 'k' => keys::K, 'l' => keys::L,
                'm' => keys::M, 'n' => keys::N, 'o' => keys::O, 'p' => keys::P,
                'q' => keys::Q, 'r' => keys::R, 's' => keys::S, 't' => keys::T,
                'u' => keys::U, 'v' => keys::V, 'w' => keys::W, 'x' => keys::X,
                'y' => keys::Y, 'z' => keys::Z,
                '0' => keys::N0, '1' => keys::N1, '2' => keys::N2, '3' => keys::N3,
                '4' => keys::N4, '5' => keys::N5, '6' => keys::N6, '7' => keys::N7,
                '8' => keys::N8, '9' => keys::N9,
                _ => 0,
            };
            e.on_key(key, c.is_uppercase(), false)
        }).collect()
    }

    fn last_char(r: &Result) -> Option<char> {
        if r.action == Action::Send as u8 && r.count > 0 {
            char::from_u32(r.chars[0])
        } else {
            None
        }
    }

    #[test]
    fn telex_basic() {
        let mut e = Engine::new();
        let r = type_keys(&mut e, "as");
        assert_eq!(last_char(&r[1]), Some('á'));
    }

    #[test]
    fn vni_basic() {
        let mut e = Engine::new();
        e.set_method(1);
        let r = type_keys(&mut e, "a1");
        assert_eq!(last_char(&r[1]), Some('á'));
    }
}
