package nginx

import (
	"os"
	"path"

	"github.com/golang/glog"
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

// Server describes an NGINX server
type Server struct {
	ServerSnippets        []string
	Name                  string
	ServerTokens          string
	Locations             []Location
	SSL                   bool
	SSLCertificate        string
	SSLCertificateKey     string
	SSLCiphers            string
	GRPCOnly              bool
	StatusZone            string
	HTTP2                 bool
	RedirectToHTTPS       bool
	SSLRedirect           bool
	ProxyProtocol         bool
	HSTS                  bool
	HSTSMaxAge            int64
	HSTSIncludeSubdomains bool
	ProxyHideHeaders      []string
	ProxyPassHeaders      []string

	// http://nginx.org/en/docs/http/ngx_http_realip_module.html
	RealIPHeader    string
	SetRealIPFrom   []string
	RealIPRecursive bool

	Ports    []int
	SSLPorts []int
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

// IngressNginxConfig describes an NGINX configuration
type IngressNginxConfig struct {
	Upstreams []Upstream
	Servers   []Server
	Keepalive string
	Ingress   Ingress
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

// NewUpstreamWithDefaultServer creates an upstream with the default server.
// proxy_pass to an upstream with the default server returns 502.
// We use it for services that have no endpoints
func NewUpstreamWithDefaultServer(name string) Upstream {
	return Upstream{
		Name: name,
		UpstreamServers: []UpstreamServer{
			UpstreamServer{
				Address:     "127.0.0.1",
				Port:        "8181",
				MaxFails:    1,
				FailTimeout: "10s",
			},
		},
	}
}

func (nginx *Controller) getIngressNginxConfigFileName(name string) string {
	return path.Join(nginx.nginxConfdPath, name+".conf")
}

// UpdateIngressConfigFile writes the Ingress configuration file to the filesystem
func (nginx *Controller) UpdateIngressConfigFile(name string, cfg []byte) {
	filename := nginx.getIngressNginxConfigFileName(name)
	glog.V(3).Infof("Writing Ingress conf to %v", filename)

	if bool(glog.V(3)) || nginx.local {
		glog.Info(string(cfg))
	}

	if !nginx.local {
		w, err := os.Create(filename)
		if err != nil {
			glog.Fatalf("Failed to open %v: %v", filename, err)
		}
		_, err = w.Write(cfg)
		if err != nil {
			glog.Fatalf("Failed to write to %v: %v", filename, err)
		}
		defer w.Close()
	}
	glog.V(3).Infof("The Ingress config file has been updated")
}

// DeleteIngress deletes the configuration file, which corresponds for the
// specified ingress from NGINX conf directory
func (nginx *Controller) DeleteIngress(name string) {
	filename := nginx.getIngressNginxConfigFileName(name)
	glog.V(3).Infof("deleting %v", filename)

	if !nginx.local {
		if err := os.Remove(filename); err != nil {
			glog.Warningf("Failed to delete %v: %v", filename, err)
		}
	}
}
