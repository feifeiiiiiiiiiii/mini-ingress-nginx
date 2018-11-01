package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/handlers"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	proxyURL = flag.String("proxy", "",
		`Use a proxy server to connect to Kubernetes API started by "kubectl proxy" command. For testing purposes only.
	The Ingress controller does not start NGINX and does not write any generated NGINX configuration files to disk`)
)

func main() {
	log.Println("Hello Mini Ingress Nginx")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	var kubeconfig *string
	var err error

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nginxBinaryPath := "/usr/sbin/nginx"
	ngxc := nginx.NewNginxController("/Users/qinpengfei/workspace/ingress-nginx", nginxBinaryPath, false)

	nginxConfTemplatePath := "/Users/qinpengfei/workspace/github/nsqbuild/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx/templates/nginx.tmpl"
	nginxIngressTemplatePath := "/Users/qinpengfei/workspace/github/nsqbuild/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx/templates/nginx.ingress.tmpl"

	templateExecutor, err := nginx.NewTemplateExecutor(nginxConfTemplatePath, nginxIngressTemplatePath)
	if err != nil {
		log.Fatalf("Error creating TemplateExecutor: %v", err)
	}

	cnf := nginx.NewNgxConfig(ngxc, templateExecutor)

	lbcInput := controller.NewLoadBalancerControllerInput{
		KubeClient:        kubeClient,
		ResyncPeriod:      30 * time.Second,
		NginxConfigurator: cnf,
		Namespace:         "ingress-mini-nginx",
		IngressClass:      "mini-ingress-nginx",
	}

	lbc := controller.NewLoadBalancerController(lbcInput)

	// create handlers for resources we care about
	ingressHandlers := handlers.CreateIngressHandlers(lbc)
	endpointHandlers := handlers.CreateEndpointHandlers(lbc)
	svcHandlers := handlers.CreateServiceHandlers(lbc)

	lbc.AddIngressHandler(ingressHandlers)
	lbc.AddEndpointHandler(endpointHandlers)
	lbc.AddServiceHandler(svcHandlers)

	lbc.Run()
	lbc.Wait()

	fmt.Printf("End Ingress Nginx")
}
