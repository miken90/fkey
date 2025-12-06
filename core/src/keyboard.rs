use rdev::{listen, Event, EventType};
use std::ffi::CString;
use std::os::raw::c_char;
use std::sync::atomic::{AtomicBool, Ordering};

static RUNNING: AtomicBool = AtomicBool::new(false);

/// Start keyboard event listener
pub fn start(callback: extern "C" fn(*const c_char)) {
    if RUNNING.load(Ordering::SeqCst) {
        return;
    }

    RUNNING.store(true, Ordering::SeqCst);

    std::thread::spawn(move || {
        let _ = listen(move |event: Event| {
            if !RUNNING.load(Ordering::SeqCst) {
                return;
            }

            if let EventType::KeyPress(key) = event.event_type {
                let key_str = format!("{:?}", key);
                if let Ok(c_str) = CString::new(key_str) {
                    callback(c_str.as_ptr());
                }
            }
        });
    });
}

/// Stop keyboard event listener
pub fn stop() {
    RUNNING.store(false, Ordering::SeqCst);
}
