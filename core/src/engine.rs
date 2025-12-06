/// Vietnamese conversion engine

/// Process input using Telex method
pub fn process_telex(input: &str) -> String {
    let mut result = String::new();
    let mut buffer = String::new();

    for ch in input.chars() {
        buffer.push(ch);

        // Check if buffer forms Vietnamese character
        if let Some(viet_char) = convert_telex(&buffer) {
            result.push(viet_char);
            buffer.clear();
        } else if buffer.len() > 3 {
            // Buffer too long, flush first char
            if let Some(first) = buffer.chars().next() {
                result.push(first);
                buffer = buffer.chars().skip(1).collect();
            }
        }
    }

    // Flush remaining buffer
    result.push_str(&buffer);
    result
}

/// Process input using VNI method
pub fn process_vni(input: &str) -> String {
    let mut result = String::new();
    let mut buffer = String::new();

    for ch in input.chars() {
        buffer.push(ch);

        if let Some(viet_char) = convert_vni(&buffer) {
            result.push(viet_char);
            buffer.clear();
        } else if buffer.len() > 2 {
            if let Some(first) = buffer.chars().next() {
                result.push(first);
                buffer = buffer.chars().skip(1).collect();
            }
        }
    }

    result.push_str(&buffer);
    result
}

/// Convert Telex input to Vietnamese character
fn convert_telex(input: &str) -> Option<char> {
    match input {
        // Vowels
        "aw" => Some('ă'),
        "aa" => Some('â'),
        "ee" => Some('ê'),
        "oo" => Some('ô'),
        "ow" => Some('ơ'),
        "uw" => Some('ư'),
        "dd" => Some('đ'),

        // Tones on 'a'
        "as" => Some('á'),
        "af" => Some('à'),
        "ar" => Some('ả'),
        "ax" => Some('ã'),
        "aj" => Some('ạ'),

        // Tones on 'e'
        "es" => Some('é'),
        "ef" => Some('è'),
        "er" => Some('ẻ'),
        "ex" => Some('ẽ'),
        "ej" => Some('ẹ'),

        // Tones on 'i'
        "is" => Some('í'),
        "if" => Some('ì'),
        "ir" => Some('ỉ'),
        "ix" => Some('ĩ'),
        "ij" => Some('ị'),

        // Tones on 'o'
        "os" => Some('ó'),
        "of" => Some('ò'),
        "or" => Some('ỏ'),
        "ox" => Some('õ'),
        "oj" => Some('ọ'),

        // Tones on 'u'
        "us" => Some('ú'),
        "uf" => Some('ù'),
        "ur" => Some('ủ'),
        "ux" => Some('ũ'),
        "uj" => Some('ụ'),

        // Tones on 'y'
        "ys" => Some('ý'),
        "yf" => Some('ỳ'),
        "yr" => Some('ỷ'),
        "yx" => Some('ỹ'),
        "yj" => Some('ỵ'),

        _ => None,
    }
}

/// Convert VNI input to Vietnamese character
fn convert_vni(input: &str) -> Option<char> {
    match input {
        // Vowels
        "a8" => Some('ă'),
        "a6" => Some('â'),
        "e6" => Some('ê'),
        "o6" => Some('ô'),
        "o7" => Some('ơ'),
        "u7" => Some('ư'),
        "d9" => Some('đ'),

        // Tones
        "a1" => Some('á'),
        "a2" => Some('à'),
        "a3" => Some('ả'),
        "a4" => Some('ã'),
        "a5" => Some('ạ'),

        "e1" => Some('é'),
        "e2" => Some('è'),
        "e3" => Some('ẻ'),
        "e4" => Some('ẽ'),
        "e5" => Some('ẹ'),

        "i1" => Some('í'),
        "i2" => Some('ì'),
        "i3" => Some('ỉ'),
        "i4" => Some('ĩ'),
        "i5" => Some('ị'),

        "o1" => Some('ó'),
        "o2" => Some('ò'),
        "o3" => Some('ỏ'),
        "o4" => Some('õ'),
        "o5" => Some('ọ'),

        "u1" => Some('ú'),
        "u2" => Some('ù'),
        "u3" => Some('ủ'),
        "u4" => Some('ũ'),
        "u5" => Some('ụ'),

        "y1" => Some('ý'),
        "y2" => Some('ỳ'),
        "y3" => Some('ỷ'),
        "y4" => Some('ỹ'),
        "y5" => Some('ỵ'),

        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_telex_basic() {
        assert_eq!(process_telex("aw"), "ă");
        assert_eq!(process_telex("dd"), "đ");
    }

    #[test]
    fn test_vni_basic() {
        assert_eq!(process_vni("a8"), "ă");
        assert_eq!(process_vni("d9"), "đ");
    }
}
