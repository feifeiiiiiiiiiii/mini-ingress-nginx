package nginx

import (
	"fmt"
	"sort"
	"strings"

	"github.com/golang/glog"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NgxConfig transforms ingress to nginx config
type NgxConfig struct {
	nginx            *Controller
	ingresses        map[string]*IngressEx
	templateExecutor *TemplateExecutor
}

// NewNgxConfig create new NgxConfig
func NewNgxConfig(nginx *Controller, templateExecutor *TemplateExecutor) *NgxConfig {
	cnf := NgxConfig{
		nginx:            nginx,
		templateExecutor: templateExecutor,
		ingresses:        make(map[string]*IngressEx),
	}
	return &cnf
}

// AddOrUpdateIngress add or update ingress
func (cnf *NgxConfig) AddOrUpdateIngress(ingEx *IngressEx) error {
	nginxCfg := cnf.generateNginxCfg(ingEx)
	name := objectMetaToFileName(&ingEx.Ingress.ObjectMeta)
	content, err := cnf.templateExecutor.ExecuteIngressConfigTemplate(&nginxCfg)
	if err != nil {
		return fmt.Errorf("Error generating Ingress Config %v: %v", name, err)
	}
	cnf.nginx.UpdateIngressConfigFile(name, content)
	cnf.ingresses[name] = ingEx
	if err := cnf.nginx.Reload(); err != nil {
		return fmt.Errorf("Error reloading NGINX for %v/%v: %v", ingEx.Ingress.Namespace, ingEx.Ingress.Name, err)
	}
	return nil
}

func objectMetaToFileName(meta *meta_v1.ObjectMeta) string {
	return meta.Namespace + "-" + meta.Name
}

func getNameForUpstream(ing *extensions.Ingress, host string, backend *extensions.IngressBackend) string {
	return fmt.Sprintf("%v-%v-%v-%v-%v", ing.Namespace, ing.Name, host, backend.ServiceName, backend.ServicePort.String())
}

func (cnf *NgxConfig) generateNginxCfg(ingEx *IngressEx) IngressNginxConfig {
	upstreams := make(map[string]Upstream)
	rewrites := getRewrites(ingEx)

	if ingEx.Ingress.Spec.Backend != nil {
		name := getNameForUpstream(ingEx.Ingress, "", ingEx.Ingress.Spec.Backend)
		upstream := cnf.createUpstream(ingEx, name, ingEx.Ingress.Spec.Backend, ingEx.Ingress.Namespace)
		upstreams[name] = upstream
	}

	var servers []Server

	for _, rule := range ingEx.Ingress.Spec.Rules {
		if rule.IngressRuleValue.HTTP == nil {
			continue
		}

		serverName := rule.Host

		statuzZone := rule.Host

		server := Server{
			Name:       serverName,
			StatusZone: statuzZone,
		}

		var locations []Location
		rootLocation := false

		for _, path := range rule.HTTP.Paths {
			upsName := getNameForUpstream(ingEx.Ingress, rule.Host, &path.Backend)

			if _, exists := upstreams[upsName]; !exists {
				upstream := cnf.createUpstream(ingEx, upsName, &path.Backend, ingEx.Ingress.Namespace)
				upstreams[upsName] = upstream
			}

			loc := createLocation(pathOrDefault(path.Path), upstreams[upsName], rewrites[path.Backend.ServiceName])

			locations = append(locations, loc)

			if loc.Path == "/" {
				rootLocation = true
			}
		}

		if rootLocation == false && ingEx.Ingress.Spec.Backend != nil {
			upsName := getNameForUpstream(ingEx.Ingress, "", ingEx.Ingress.Spec.Backend)

			loc := createLocation(pathOrDefault("/"), upstreams[upsName], rewrites[ingEx.Ingress.Spec.Backend.ServiceName])
			locations = append(locations, loc)
		}

		server.Locations = locations

		servers = append(servers, server)
	}
	return IngressNginxConfig{
		Upstreams: upstreamMapToSlice(upstreams),
		Servers:   servers,
		Ingress: Ingress{
			Name:        ingEx.Ingress.Name,
			Namespace:   ingEx.Ingress.Namespace,
			Annotations: ingEx.Ingress.Annotations,
		},
	}
}

func (cnf *NgxConfig) createUpstream(ingEx *IngressEx, name string, backend *extensions.IngressBackend, namespace string) Upstream {
	ups := NewUpstreamWithDefaultServer(name)

	endps, exists := ingEx.Endpoints[backend.ServiceName+backend.ServicePort.String()]
	if exists {
		var upsServers []UpstreamServer
		for _, endp := range endps {
			addressport := strings.Split(endp, ":")
			upsServers = append(upsServers, UpstreamServer{
				Address: addressport[0],
				Port:    addressport[1],
			})
		}
		if len(upsServers) > 0 {
			ups.UpstreamServers = upsServers
		}
	}
	return ups
}

// DeleteIngress deletes NGINX configuration for the Ingress resource
func (cnf *NgxConfig) DeleteIngress(key string) error {
	name := strings.Replace(key, "/", "-", -1)
	cnf.nginx.DeleteIngress(name)
	delete(cnf.ingresses, name)
	return nil
}

func createLocation(path string, upstream Upstream, rewrite string) Location {
	loc := Location{
		Path:     path,
		Upstream: upstream,
		Rewrite:  rewrite,
	}

	return loc
}

func pathOrDefault(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func upstreamMapToSlice(upstreams map[string]Upstream) []Upstream {
	keys := make([]string, 0, len(upstreams))
	for k := range upstreams {
		keys = append(keys, k)
	}

	// this ensures that the slice 'result' is sorted, which preserves the order of upstream servers
	// in the generated configuration file from one version to another and is also required for repeatable
	// Unit test results
	sort.Strings(keys)

	result := make([]Upstream, 0, len(upstreams))

	for _, k := range keys {
		result = append(result, upstreams[k])
	}

	return result
}

// UpdateEndpoints updates endpoints in NGINX configuration for the Ingress resources
func (cnf *NgxConfig) UpdateEndpoints(ingExes []*IngressEx) error {
	for _, ingEx := range ingExes {
		err := cnf.AddOrUpdateIngress(ingEx)
		if err != nil {
			return fmt.Errorf("Error adding or updating ingress %v/%v: %v", ingEx.Ingress.Namespace, ingEx.Ingress.Name, err)
		}
	}

	if err := cnf.nginx.Reload(); err != nil {
		return fmt.Errorf("Error reloading NGINX when updating endpoints: %v", err)
	}

	return nil
}

// HasIngress checks if the Ingress resource is present in NGINX configuration
func (cnf *NgxConfig) HasIngress(ing *extensions.Ingress) bool {
	name := objectMetaToFileName(&ing.ObjectMeta)
	_, exists := cnf.ingresses[name]
	return exists
}

func getRewrites(ingEx *IngressEx) map[string]string {
	rewrites := make(map[string]string)

	if services, exists := ingEx.Ingress.Annotations["nginx.org/rewrites"]; exists {
		for _, svc := range strings.Split(services, ";") {
			if serviceName, rewrite, err := parseRewrites(svc); err != nil {
				glog.Errorf("In %v nginx.org/rewrites contains invalid declaration: %v, ignoring", ingEx.Ingress.Name, err)
			} else {
				rewrites[serviceName] = rewrite
			}
		}
	}

	return rewrites
}

func parseRewrites(service string) (serviceName string, rewrite string, err error) {
	parts := strings.SplitN(strings.TrimSpace(service), " ", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid rewrite format: %s", service)
	}

	svcNameParts := strings.Split(parts[0], "=")
	if len(svcNameParts) != 2 {
		return "", "", fmt.Errorf("Invalid rewrite format: %s", svcNameParts)
	}

	rwPathParts := strings.Split(parts[1], "=")
	if len(rwPathParts) != 2 {
		return "", "", fmt.Errorf("Invalid rewrite format: %s", rwPathParts)
	}

	return svcNameParts[1], rwPathParts[1], nil
}
