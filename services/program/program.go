// services/program/program.go
package program

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/common/templating"
)

// Service provides methods for opening URLs and launching system handlers
// for known protocols (VNC, RDP, SPICE, etc.).
type Service struct{}

// serviceTemplateData is the data bag for protocol command templates.
type serviceTemplateData struct {
	PROTOCOL string
	IP       string
	PORT     string
}

// commandTemplateForProtocol returns the configured command template for
// the given protocol name, or "" if none is configured.
func commandTemplateForProtocol(protocol string) string {
	t := config.Get().Scanner.Templates
	switch strings.ToLower(protocol) {
	case "vnc":
		return t.VncCommandTemplate
	case "rdp":
		return t.RdpCommandTemplate
	case "spice":
		return t.SpiceCommandTemplate
	case "ssh":
		return t.SshCommandTemplate
	case "http":
		return t.HttpCommandTemplate
	case "https":
		return t.HttpsCommandTemplate
	default:
		return ""
	}
}

// OpenURL opens the given URL in the default system browser/handler.
func (s *Service) OpenURL(url string) error {
	return openSystem(url)
}

// OpenService launches the configured command for the given protocol,
// ip and port. If a command template is set in the config, it is resolved
// and executed via the shell; otherwise it falls back to the system handler.
func (s *Service) OpenService(protocol, ip, port string) error {
	tmpl := commandTemplateForProtocol(protocol)
	if tmpl == "" {
		scheme := strings.ToLower(protocol)
		url := fmt.Sprintf("%s://%s:%s", scheme, ip, port)
		logger.Info().Str("url", url).Msg("opening service via system handler (no template)")
		return openSystem(url)
	}

	data := serviceTemplateData{
		PROTOCOL: strings.ToLower(protocol),
		IP:       ip,
		PORT:     port,
	}

	resolved, err := templating.Resolve(tmpl, data)
	if err != nil {
		return fmt.Errorf("resolve %s command template: %w", protocol, err)
	}

	logger.Info().Str("command", resolved).Str("protocol", protocol).Msg("launching service command")

	cmd := exec.Command("sh", "-c", resolved)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch %s command: %w", protocol, err)
	}
	go cmd.Wait()
	return nil
}

func openSystem(target string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, target)
	return exec.Command(cmd, args...).Start()
}
