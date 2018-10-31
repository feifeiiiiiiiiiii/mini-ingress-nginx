package nginx

import (
	"path"
)

// Controller updates NGINX configuration, starts and reloads NGINX
type Controller struct {
	nginxConfdPath   string
	nginxSecretsPath string
	local            bool
	nginxBinaryPath  string
	configVersion    int
}

// Location describes an NGINX location
type Location struct {
	Path     string
	Upstream Upstream
}

// Upstream describes an NGINX upstream
type Upstream struct {
	Name            string
	UpstreamServers []UpstreamServer
}

// UpstreamServer describes a server in an NGINX upstream
type UpstreamServer struct {
	Address     string
	Port        string
	MaxFails    int64
	FailTimeout string
	SlowStart   string
}

// Ingress holds information about an Ingress resource
type Ingress struct {
	Name        string
	Namespace   string
	Annotations map[string]string
}

// NewNginxController creates a NGINX controller
func NewNginxController(nginxConfPath string, nginxBinaryPath string, local bool) *Controller {
	ngxc := Controller{
		nginxConfdPath:  path.Join(nginxConfPath, "conf.d"),
		local:           local,
		nginxBinaryPath: nginxBinaryPath,
		configVersion:   0,
	}

	return &ngxc
}

// Reload nginx
func (nginx *Controller) Reload() error {
	return nil
}
