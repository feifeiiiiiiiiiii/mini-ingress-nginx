package main

import (
	"flag"
	"time"
	"os"
	"path/filepath"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/handlers"
)

func main() {
	glog.Infof("Hello Mini Ingress Nginx")

	var kubeconfig *string
	var err error

	home := os.Getenv("HOME")
	if home != "" {
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
		glog.Fatalf("Failed to create client: %v.", err)
	}

	nginxBinaryPath := "/usr/sbin/nginx"
	nginx.NewNginxController("/etc/nginx/", nginxBinaryPath, false)

	lbcInput := controller.NewLoadBalancerControllerInput{
		KubeClient:              kubeClient,
		ResyncPeriod:            30 * time.Second,
		Namespace:               "mini-ingress-nginx",
		IngressClass:            "kubernetes.io/mini.nginx.ingress.class",
	}

	lbc := controller.NewLoadBalancerController(lbcInput)

	// create handlers for resources we care about
	ingressHandlers := handlers.CreateIngressHandlers(lbc)	
	serviceHandlers := handlers.CreateServiceHandlers(lbc)
	endpointHandlers := handlers.CreateEndpointHandlers(lbc)

	lbc.AddResourceHandler("ingresses", ingressHandlers)
	lbc.AddResourceHandler("services", serviceHandlers)
	lbc.AddResourceHandler("endpoints", endpointHandlers)

	glog.Infof("End Ingress Nginx")
}
