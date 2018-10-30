package nginx

import (
	"path"
)

// Controller updates NGINX configuration, starts and reloads NGINX
type Controller struct {
	nginxConfdPath        string
	nginxSecretsPath      string
	local                 bool
	nginxBinaryPath       string
	configVersion         int
}

// NewNginxController creates a NGINX controller
func NewNginxController(nginxConfPath string, nginxBinaryPath string, local bool) *Controller {
	ngxc := Controller{
		nginxConfdPath:        path.Join(nginxConfPath, "conf.d"),
		nginxSecretsPath:      path.Join(nginxConfPath, "secrets"),
		local:                 local,
		nginxBinaryPath:       nginxBinaryPath,
		configVersion:         0,
	}

	return &ngxc
}