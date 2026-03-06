//go:build linux

package config

import "os/exec"

// Displays a native OS dialog informing the user that a default config was generated at the given path. Tries multiple tools so that maybe it will find one that is installed. If none are found, it does nothing.
func ShowDefaultGeneratedDialog(e *ErrDefaultGenerated) {
	if e == nil {
		return
	}

	const title = "Thug Hunter"
	msg := "A default configuration file has been created at:\n\n" + e.Path + "\n\nPlease edit it and restart the application."

	attempts := [][]string{
		{"kdialog", "--title", title, "--msgbox", msg},
		{"zenity", "--info", "--title=" + title, "--text=" + msg, "--no-wrap"},
		{"yad", "--info", "--title=" + title, "--text=" + msg, "--no-wrap"},
		{"qarma", "--info", "--title=" + title, "--text=" + msg, "--no-wrap"},
		{"matedialog", "--info", "--title=" + title, "--text=" + msg, "--no-wrap"},
		{"gxmessage", "-center", "-title", title, msg},
		{"xmessage", "-center", "-title", title, msg},
		{"Xdialog", "--title", title, "--msgbox", msg, "0", "0"},
		{"gdialog", "--title", title, "--msgbox", msg, "0", "0"},
		{"whiptail", "--title", title, "--msgbox", msg, "0", "0"},
		{"dialog", "--title", title, "--msgbox", msg, "0", "0"},
		{"notify-send", title, msg},
	}

	for _, cmd := range attempts {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
			return
		}
	}
}
