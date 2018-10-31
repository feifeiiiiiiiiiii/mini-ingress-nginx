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
	nginx.NewNginxController("/etc/nginx/", nginxBinaryPath, false)

	lbcInput := controller.NewLoadBalancerControllerInput{
		KubeClient:   kubeClient,
		ResyncPeriod: 30 * time.Second,
		Namespace:    "ingress-mini-nginx",
		IngressClass: "mini-ingress-nginx",
	}

	lbc := controller.NewLoadBalancerController(lbcInput)

	// create handlers for resources we care about
	ingressHandlers := handlers.CreateIngressHandlers(lbc)
	serviceHandlers := handlers.CreateServiceHandlers(lbc)
	endpointHandlers := handlers.CreateEndpointHandlers(lbc)

	lbc.AddResourceHandler("ingresses", ingressHandlers)
	lbc.AddResourceHandler("services", serviceHandlers)
	lbc.AddResourceHandler("endpoints", endpointHandlers)

	lbc.Run()
	lbc.Wait()

	fmt.Printf("End Ingress Nginx")
}
