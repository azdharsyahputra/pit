package tools

import (
	"os"
	"path/filepath"
	"text/template"
)

type NginxToolVhost struct {
	ServerName string
	RootAbs    string
	Index      string
	PhpSock    string
}

var toolVhostTpl = template.Must(template.New("vhost").Parse(`
server {
    listen 80;
    server_name {{.ServerName}};

    root {{.RootAbs}};
    index {{.Index}};

    location / {
        try_files $uri $uri/ /{{.Index}}?$query_string;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_pass unix:{{.PhpSock}};
    }
}
`))

func WriteToolVhost(base string, m Manifest, phpSockAbs string) (string, error) {
	outDir := filepath.Join(base, "nginx", "conf.d", "tools")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	rootAbs := filepath.Join(base, m.Root)
	outPath := filepath.Join(outDir, m.Name+".conf")

	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data := NginxToolVhost{
		ServerName: m.Domain,
		RootAbs:    rootAbs,
		Index:      m.Index,
		PhpSock:    phpSockAbs,
	}
	if err := toolVhostTpl.Execute(f, data); err != nil {
		return "", err
	}
	return outPath, nil
}
