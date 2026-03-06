//go:build windows

package config

import "os/exec"

// Displays a native OS dialog informing the user that a default config was generated at the given path.
func ShowDefaultGeneratedDialog(e *ErrDefaultGenerated) {
	msg := "A default configuration file has been created at:\\n\\n" + e.Path + "\\n\\nPlease edit it and restart the application."
	_ = exec.Command("powershell", "-Command",
		`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('`+msg+`','Thug Hunter','OK','Information')`).Run()
}
