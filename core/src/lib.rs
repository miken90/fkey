use std::ffi::{CStr, CString};
use std::os::raw::c_char;

mod engine;
mod keyboard;
mod config;

/// FFI: Process Vietnamese input
///
/// # Arguments
/// * `input` - Input string (C string)
/// * `mode` - 0 = Telex, 1 = VNI
///
/// # Returns
/// Processed Vietnamese string (must be freed with free_string)
#[no_mangle]
pub extern "C" fn process_input(
    input: *const c_char,
    mode: u8,
) -> *mut c_char {
    let c_str = unsafe { CStr::from_ptr(input) };
    let input_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let result = match mode {
        0 => engine::process_telex(input_str),
        1 => engine::process_vni(input_str),
        _ => input_str.to_string(),
    };

    match CString::new(result) {
        Ok(s) => s.into_raw(),
        Err(_) => std::ptr::null_mut(),
    }
}

/// FFI: Start keyboard hook
#[no_mangle]
pub extern "C" fn start_hook(callback: extern "C" fn(*const c_char)) {
    keyboard::start(callback);
}

/// FFI: Stop keyboard hook
#[no_mangle]
pub extern "C" fn stop_hook() {
    keyboard::stop();
}

/// FFI: Save configuration
#[no_mangle]
pub extern "C" fn save_config(enabled: bool, mode: u8) {
    let config = config::Config {
        enabled,
        mode,
    };

    if let Err(e) = config.save() {
        eprintln!("Failed to save config: {}", e);
    }
}

/// FFI: Load configuration
#[no_mangle]
pub extern "C" fn load_config() -> *mut config::Config {
    let config = config::Config::load();
    Box::into_raw(Box::new(config))
}

/// FFI: Free string allocated by Rust
#[no_mangle]
pub extern "C" fn free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe { drop(CString::from_raw(s)) };
    }
}

/// FFI: Free config
#[no_mangle]
pub extern "C" fn free_config(config: *mut config::Config) {
    if !config.is_null() {
        unsafe { drop(Box::from_raw(config)) };
    }
}
