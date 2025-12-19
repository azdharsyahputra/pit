package util

import (
	"os"
	"path/filepath"
	"strings"
)

func PrepareNginxDirs(base string) {
	folders := []string{
		"client_body_temp",
		"fastcgi_temp",
		"proxy_temp",
		"scgi_temp",
		"uwsgi_temp",
		"logs",
	}
	for _, f := range folders {
		_ = os.MkdirAll(filepath.Join(base, f), 0o755)
	}

	touch(filepath.Join(base, "logs/access.log"))
	touch(filepath.Join(base, "logs/error.log"))
}

func PreparePHPDirs(base string) {
	folders := []string{
		"logs",
		"var/run",
		"var/log",
	}
	for _, f := range folders {
		_ = os.MkdirAll(filepath.Join(base, f), 0o755)
	}

	touch(filepath.Join(base, "logs/php-fpm.log"))
}

func touch(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte{}, 0o644)
	}
}

func GenerateNginxConf(srcTpl, dstConf, rootPath string) error {
	data, err := os.ReadFile(srcTpl)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(string(data), "{{ROOT_PATH}}", rootPath)
	return os.WriteFile(dstConf, []byte(content), 0o644)
}
