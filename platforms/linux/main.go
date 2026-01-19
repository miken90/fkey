package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"fkey-linux/config"
	"fkey-linux/core"
	"fkey-linux/ui"
)

var Version = "0.1.0-dev"

func main() {
	log.Printf("FKey Linux v%s starting...", Version)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Using default config: %v", err)
		cfg = config.Default()
	}

	// Initialize Rust core (FFI)
	bridge, err := core.NewBridge()
	if err != nil {
		log.Fatalf("Failed to init IME core: %v", err)
	}
	defer bridge.Close()

	// Apply settings
	bridge.SetMethod(cfg.InputMethod)
	bridge.SetEnabled(cfg.Enabled)
	bridge.SetModernTone(cfg.ModernTone)

	// Initialize X11 keyboard handler
	kbd, err := core.NewKeyboardHandler(bridge)
	if err != nil {
		log.Fatalf("Failed to init keyboard handler: %v", err)
	}

	// Initialize GTK UI (system tray)
	tray, err := ui.NewTray(cfg, func(enabled bool) {
		bridge.SetEnabled(enabled)
		cfg.Enabled = enabled
		config.Save(cfg)
	})
	if err != nil {
		log.Fatalf("Failed to init tray: %v", err)
	}

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		kbd.Stop()
		tray.Quit()
	}()

	// Start keyboard hook
	go kbd.Start()

	// Run GTK main loop (blocks)
	tray.Run()

	log.Println("FKey Linux stopped")
}
