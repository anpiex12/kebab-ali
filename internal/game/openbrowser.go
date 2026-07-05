package game

import (
	"os/exec"
	"runtime"
)

// openBrowser tries to open url in the user's default browser. Failure is
// ignored — it is a convenience on the credits screen, not a core feature.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, bsd, …
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
