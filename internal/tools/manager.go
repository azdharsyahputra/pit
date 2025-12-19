package tools

import (
	"fmt"
	"path/filepath"
)

type Manager struct {
	Base        string
	PhpSockAbs  string
	NginxReload func() error
}

func (m *Manager) SyncAll() error {
	manifests, err := Scan(m.Base)
	if err != nil {
		return err
	}

	// 1) hosts
	if err := SyncHosts(DomainsFromManifests(manifests)); err != nil {
		return ExplainHostsPermHint(err)
	}

	// 2) vhosts
	for _, t := range manifests {
		if t.Type != "php" {
			continue
		}
		if t.Index == "" {
			t.Index = "index.php"
		}
		if t.Root == "" {
			t.Root = filepath.Join("tools", t.Name)
		}

		if _, err := WriteToolVhost(m.Base, t, m.PhpSockAbs); err != nil {
			return err
		}
	}

	// 3) reload nginx
	if m.NginxReload != nil {
		if err := m.NginxReload(); err != nil {
			return fmt.Errorf("nginx reload failed: %w", err)
		}
	}

	return nil
}
