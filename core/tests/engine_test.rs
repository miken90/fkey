use gonhanh_core::*;

#[test]
fn test_telex_vowels() {
    // Test basic vowels
    let input = std::ffi::CString::new("aw").unwrap();
    let result = unsafe { process_input(input.as_ptr(), 0) };
    let output = unsafe { std::ffi::CStr::from_ptr(result) };
    assert_eq!(output.to_str().unwrap(), "ă");
    unsafe { free_string(result) };

    let input = std::ffi::CString::new("dd").unwrap();
    let result = unsafe { process_input(input.as_ptr(), 0) };
    let output = unsafe { std::ffi::CStr::from_ptr(result) };
    assert_eq!(output.to_str().unwrap(), "đ");
    unsafe { free_string(result) };
}

#[test]
fn test_vni_vowels() {
    let input = std::ffi::CString::new("a8").unwrap();
    let result = unsafe { process_input(input.as_ptr(), 1) };
    let output = unsafe { std::ffi::CStr::from_ptr(result) };
    assert_eq!(output.to_str().unwrap(), "ă");
    unsafe { free_string(result) };
}
