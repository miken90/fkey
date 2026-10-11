package core

// LaunchArgs holds the command-line flags FKey understands.
type LaunchArgs struct {
	Relaunch bool
	Hub      bool
}

// ParseLaunchArgs scans every arg for the case-sensitive flags.
func ParseLaunchArgs(args []string) LaunchArgs {
	var la LaunchArgs
	for _, arg := range args {
		switch arg {
		case "--relaunch":
			la.Relaunch = true
		case "--hub":
			la.Hub = true
		}
	}
	return la
}

// HubModeEnabled reports whether hub mode applies. An elevated FKey is
// unsupported in hub mode for now: it keeps its tray.
func HubModeEnabled(hub, runAsAdmin, elevated bool) bool {
	return hub && !runAsAdmin && !elevated
}

// ShouldCheckUpdates is false in hub mode: the updater's relaunch would drop
// --hub, and the hub's catalog owns updates.
func ShouldCheckUpdates(hubMode bool) bool {
	return !hubMode
}

// NewEnabledChangedHandler builds the OnEnabledChanged callback that runs on
// the hook thread. It only beeps and posts; update runs inside the posted func.
func NewEnabledChangedHandler(beep func(bool), post func(func()), update func(bool)) func(bool) {
	return func(enabled bool) {
		beep(enabled)
		post(func() { update(enabled) })
	}
}
