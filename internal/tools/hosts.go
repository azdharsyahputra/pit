package tools

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

const (
	hostsPath   = "/etc/hosts"
	beginMarker = "# BEGIN PIT TOOLS"
	endMarker   = "# END PIT TOOLS"
)

func BuildHostsBlock(domains []string) string {
	var b strings.Builder
	b.WriteString(beginMarker + "\n")
	for _, d := range uniqueNonEmpty(domains) {
		b.WriteString("127.0.0.1 " + d + "\n")
	}
	b.WriteString(endMarker + "\n")
	return b.String()
}

func SyncHosts(domains []string) error {
	raw, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	block := BuildHostsBlock(domains)
	content := string(raw)

	hasBegin := strings.Contains(content, beginMarker)
	hasEnd := strings.Contains(content, endMarker)

	var out string
	if hasBegin && hasEnd {
		out = replaceBetweenMarkers(content, block)
	} else {
		var buf bytes.Buffer
		buf.WriteString(strings.TrimRight(content, "\n"))
		buf.WriteString("\n\n")
		buf.WriteString(block)
		out = buf.String()
	}

	// write atomically
	tmp := hostsPath + ".pit.tmp"
	if err := os.WriteFile(tmp, []byte(out), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, hostsPath)
}

func replaceBetweenMarkers(content, block string) string {
	start := strings.Index(content, beginMarker)
	end := strings.Index(content, endMarker)
	if start == -1 || end == -1 || end < start {
		// fallback append
		return strings.TrimRight(content, "\n") + "\n\n" + block
	}
	end = end + len(endMarker)
	before := strings.TrimRight(content[:start], "\n")
	after := strings.TrimLeft(content[end:], "\n")
	if after != "" {
		after = "\n" + after
	}
	return before + "\n" + block + after
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func DomainsFromManifests(ms []Manifest) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if m.Domain != "" {
			out = append(out, m.Domain)
		}
	}
	return out
}

func ExplainHostsPermHint(err error) error {
	return fmt.Errorf("%w (need root to write /etc/hosts; run pit with sudo or implement privileged helper)", err)
}
