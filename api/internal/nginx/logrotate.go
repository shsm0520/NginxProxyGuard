package nginx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nginx-proxy-guard/internal/config"
)

// LogrotateConfigPath is the logrotate config the API generates for the raw
// nginx logs. The api and nginx containers mount the same nginx volume at
// /etc/nginx, so this path is identical on both sides: the api writes it, the
// nginx container reads it.
const LogrotateConfigPath = "/etc/nginx/conf.d/.logrotate.conf"

// logrotateInstallPath is where the generated config is installed inside the
// nginx container before `logrotate -f` is pointed at it.
const logrotateInstallPath = "/etc/logrotate.d/nginx-guard"

// ErrLogrotateConfigMissing is returned by RotateLogs when the generated config
// has not been written yet (raw log files never enabled). Callers that treat
// this as "nothing to do" — the daily scheduler — match it with errors.Is.
var ErrLogrotateConfigMissing = errors.New("logrotate config not found")

// RotateLogs installs the generated logrotate config into the nginx container
// and forces a rotation there. The logrotate binary only exists in the nginx
// image, so this has to go through docker exec like every other nginx-side
// command — running it in the api container fails at PATH lookup (#301).
//
// The returned error carries both the exec error and the command's combined
// output: a failing docker exec often produces no output at all (unknown
// container, missing binary), so reporting only the output leaves the caller
// with an empty string and nothing to diagnose.
func (m *Manager) RotateLogs(ctx context.Context) error {
	if _, err := os.Stat(LogrotateConfigPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrLogrotateConfigMissing, LogrotateConfigPath)
		}
		return fmt.Errorf("logrotate config %s: %w", LogrotateConfigPath, err)
	}

	ctx, cancel := context.WithTimeout(ctx, config.NginxLogrotateTimeout)
	defer cancel()

	script := fmt.Sprintf("cp %s %s && logrotate -f %s",
		LogrotateConfigPath, logrotateInstallPath, logrotateInstallPath)
	cmd := exec.CommandContext(ctx, "docker", "exec", m.nginxContainer, "sh", "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out == "" {
			out = "(no output)"
		}
		return fmt.Errorf("logrotate in container %q failed: %w: %s", m.nginxContainer, err, out)
	}
	return nil
}
