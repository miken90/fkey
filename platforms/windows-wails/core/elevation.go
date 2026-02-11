package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

const (
	mutexName          = "Global\\FKeyIME"
	errorAlreadyExists = 183
	maxRelaunchWait    = 10 * time.Second
	relaunchRetryDelay = 500 * time.Millisecond
)

var (
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
	procReleaseMutex = kernel32.NewProc("ReleaseMutex")
	shell32          = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32.NewProc("ShellExecuteW")
)

var (
	mutexHandle syscall.Handle
	QuitApp     func()
	RevertRunAsAdmin func()
)

func AcquireMutex(relaunch bool) error {
	namePtr, _ := syscall.UTF16PtrFromString(mutexName)
	handle, _, err := procCreateMutexW.Call(
		0,
		1,
		uintptr(unsafe.Pointer(namePtr)),
	)
	if handle == 0 {
		return fmt.Errorf("CreateMutexW failed: %v", err)
	}

	if err.(syscall.Errno) == errorAlreadyExists {
		syscall.CloseHandle(syscall.Handle(handle))

		if !relaunch {
			return fmt.Errorf("FKey đã đang chạy")
		}

		deadline := time.Now().Add(maxRelaunchWait)
		for time.Now().Before(deadline) {
			time.Sleep(relaunchRetryDelay)
			handle, _, err = procCreateMutexW.Call(
				0,
				1,
				uintptr(unsafe.Pointer(namePtr)),
			)
			if handle == 0 {
				continue
			}
			if err.(syscall.Errno) != errorAlreadyExists {
				mutexHandle = syscall.Handle(handle)
				return nil
			}
			syscall.CloseHandle(syscall.Handle(handle))
		}
		return fmt.Errorf("FKey đã đang chạy (timeout waiting for previous instance)")
	}

	mutexHandle = syscall.Handle(handle)
	return nil
}

func ReleaseMutex() {
	if mutexHandle != 0 {
		procReleaseMutex.Call(uintptr(mutexHandle))
		syscall.CloseHandle(mutexHandle)
		mutexHandle = 0
	}
}

func ElevateAndRelaunch() {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("ElevateAndRelaunch: failed to get exe path: %v", err)
		return
	}

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	filePtr, _ := syscall.UTF16PtrFromString(exePath)
	paramsPtr, _ := syscall.UTF16PtrFromString("--relaunch")

	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(filePtr)),
		uintptr(unsafe.Pointer(paramsPtr)),
		0,
		1, // SW_SHOWNORMAL
	)

	if ret <= 32 {
		log.Printf("ElevateAndRelaunch: ShellExecuteW failed (ret=%d), UAC cancelled?", ret)
		if RevertRunAsAdmin != nil {
			RevertRunAsAdmin()
		}
		return
	}

	log.Printf("ElevateAndRelaunch: elevated process launched, quitting current instance")
	if QuitApp != nil {
		QuitApp()
	}
}

func DeElevateAndRelaunch() {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("DeElevateAndRelaunch: failed to get exe path: %v", err)
		return
	}

	cmd := exec.Command(exePath, "--relaunch")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		log.Printf("DeElevateAndRelaunch: failed to start process: %v", err)
		return
	}

	log.Printf("DeElevateAndRelaunch: de-elevated process launched, quitting current instance")
	if QuitApp != nil {
		QuitApp()
	}
}
