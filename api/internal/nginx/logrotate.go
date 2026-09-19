package nginx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

// The three outcomes below are not failures. Each one means the rotation the
// caller asked for is already accounted for, so handlers answer 200 and the
// scheduler logs a skip — only a genuine logrotate fault stays an error.
var (
	// ErrLogrotateNothingToRotate: every raw log is empty. logrotate's
	// notifempty is not overridden by -f, so it would leave them alone and
	// still exit 0 — reporting success there would claim a cut that never
	// happened.
	ErrLogrotateNothingToRotate = errors.New("no raw log has anything to rotate")

	// ErrLogrotateAlreadyRotated: the archive logrotate would create is
	// already on disk. Rotated names carry the time to the second (#301), so
	// this can only mean a rotation ran inside this same second and the
	// current log really is freshly cut. Before the timestamp went in, this
	// same refusal was every "Rotate now" for the rest of the day.
	ErrLogrotateAlreadyRotated = errors.New("logs were already rotated")

	// ErrLogrotateBusy: another logrotate holds the state file. Our own calls
	// are serialized below, so this comes from outside the process.
	ErrLogrotateBusy = errors.New("another log rotation is in progress")
)

// logrotate prints these; it exits 1 and 3 respectively and gives nothing else
// machine-readable, so its wording is the only signal available.
const (
	logrotateCollisionMarker = "already exists, skipping rotation"
	logrotateLockedMarker    = "is already locked"
)

// rawLogDir holds the files the generated config rotates. The api and nginx
// containers mount the same nginx volume, so the api can read their sizes
// without another docker exec.
const rawLogDir = "/etc/nginx/logs"

var rawLogNames = []string{"access_raw.log", "error_raw.log"}

// logrotateMutex serializes our own rotations. logrotate locks its state file
// and refuses to run twice at once ("logrotate does not support parallel
// execution on the same set of logfiles", exit 3), which a double-clicked
// "Rotate now" — or a manual click landing on the daily scheduler — reaches
// easily. Serializing here is also what makes the emptiness check below
// truthful: it is evaluated inside the lock, so it cannot describe a log that
// a concurrent rotation is about to cut.
var logrotateMutex sync.Mutex

// rawLogsHaveContent reports whether there is anything for logrotate to cut.
// Both files sit in one stanza but notifempty is evaluated per file, so a
// single non-empty log is enough for a rotation to happen.
func rawLogsHaveContent() bool {
	for _, name := range rawLogNames {
		if info, err := os.Stat(filepath.Join(rawLogDir, name)); err == nil && info.Size() > 0 {
			return true
		}
	}
	return false
}

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
	logrotateMutex.Lock()
	defer logrotateMutex.Unlock()

	if _, err := os.Stat(LogrotateConfigPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrLogrotateConfigMissing, LogrotateConfigPath)
		}
		return fmt.Errorf("logrotate config %s: %w", LogrotateConfigPath, err)
	}

	if !rawLogsHaveContent() {
		return ErrLogrotateNothingToRotate
	}

	ctx, cancel := context.WithTimeout(ctx, config.NginxLogrotateTimeout)
	defer cancel()

	script := fmt.Sprintf("cp %s %s && logrotate -f %s",
		LogrotateConfigPath, logrotateInstallPath, logrotateInstallPath)
	cmd := exec.CommandContext(ctx, "docker", "exec", m.nginxContainer, "sh", "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if strings.Contains(out, logrotateCollisionMarker) {
			return ErrLogrotateAlreadyRotated
		}
		if strings.Contains(out, logrotateLockedMarker) {
			return ErrLogrotateBusy
		}
		if out == "" {
			out = "(no output)"
		}
		return fmt.Errorf("logrotate in container %q failed: %w: %s", m.nginxContainer, err, out)
	}
	return nil
}
