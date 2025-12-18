package nginx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateBannedIPs generates the banned_ips.conf file with all blocked IPs
func (m *Manager) UpdateBannedIPs(ctx context.Context, bannedIPs []string) error {
	includesPath := filepath.Join(m.configPath, "includes")

	// Ensure includes directory exists
	if err := os.MkdirAll(includesPath, 0755); err != nil {
		return fmt.Errorf("failed to create includes directory: %w", err)
	}

	// Generate banned_ips.conf content
	var content strings.Builder
	content.WriteString("# Auto-generated banned IPs - DO NOT EDIT\n")
	content.WriteString("# This file is managed by Nginx Proxy Guard\n\n")

	if len(bannedIPs) == 0 {
		content.WriteString("# No banned IPs\n")
	} else {
		for _, ip := range bannedIPs {
			content.WriteString(fmt.Sprintf("deny %s;\n", ip))
		}
	}

	// Write to file
	configFile := filepath.Join(includesPath, "banned_ips.conf")
	if err := os.WriteFile(configFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write banned_ips.conf: %w", err)
	}

	// Reload nginx to apply changes
	if !m.skipTest {
		if err := m.ReloadNginx(ctx); err != nil {
			return fmt.Errorf("failed to reload nginx after updating banned IPs: %w", err)
		}
	}

	return nil
}
