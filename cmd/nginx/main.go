package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/handlers"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	proxyURL = flag.String("proxy", "",
		`Use a proxy server to connect to Kubernetes API started by "kubectl proxy" command. For testing purposes only.
	The Ingress controller does not start NGINX and does not write any generated NGINX configuration files to disk`)

	namespace = flag.String("namespace", "mini-nginx-ingress", "ingress namespace")

	ingressClass = flag.String("ingressClass", "mini-ingress-nginx", "ingress class")

	mainTemplatePath = flag.String("main-template-path", "",
		`Path to the main NGINX configuration template. (default for NGINX "nginx.tmpl"; default for NGINX Plus "nginx-plus.tmpl")`)

	ingressTemplatePath = flag.String("ingress-template-path", "",
		`Path to the ingress NGINX configuration template for an ingress resource.
	(default for NGINX "nginx.ingress.tmpl"; default for NGINX Plus "nginx-plus.ingress.tmpl")`)
)

func main() {
	log.Println("Hello Mini Ingress Nginx")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	var config *rest.Config
	var err error

	if *proxyURL != "" {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{},
			&clientcmd.ConfigOverrides{
				ClusterInfo: clientcmdapi.Cluster{
					Server: *proxyURL,
				},
			}).ClientConfig()
		if err != nil {
			log.Fatalf("error creating client configuration: %v", err)
		}
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			log.Fatalf("error creating client configuration: %v", err)
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	}

	nginxConfTemplatePath := "nginx.tmpl"
	nginxIngressTemplatePath := "nginx.ingress.tmpl"

	if *mainTemplatePath != "" {
		nginxConfTemplatePath = *mainTemplatePath
	}
	if *ingressTemplatePath != "" {
		nginxIngressTemplatePath = *ingressTemplatePath
	}

	nginxBinaryPath := "/usr/sbin/nginx"
	ngxc := nginx.NewNginxController("/etc/nginx/", nginxBinaryPath, false)

	templateExecutor, err := nginx.NewTemplateExecutor(nginxConfTemplatePath, nginxIngressTemplatePath)
	if err != nil {
		log.Fatalf("Error creating TemplateExecutor: %v", err)
	}

	content, err := templateExecutor.ExecuteMainConfigTemplate()
	if err != nil {
		glog.Fatalf("Error generating NGINX main config: %v", err)
	}
	ngxc.UpdateMainConfigFile(content)

	cnf := nginx.NewNgxConfig(ngxc, templateExecutor)

	nginxDone := make(chan error, 1)
	ngxc.Start(nginxDone)

	lbcInput := controller.NewLoadBalancerControllerInput{
		KubeClient:        kubeClient,
		ResyncPeriod:      30 * time.Second,
		NginxConfigurator: cnf,
		Namespace:         *namespace,
		IngressClass:      *ingressClass,
	}

	lbc := controller.NewLoadBalancerController(lbcInput)

	// create handlers for resources we care about
	ingressHandlers := handlers.CreateIngressHandlers(lbc)
	endpointHandlers := handlers.CreateEndpointHandlers(lbc)
	svcHandlers := handlers.CreateServiceHandlers(lbc)

	lbc.AddIngressHandler(ingressHandlers)
	lbc.AddEndpointHandler(endpointHandlers)
	lbc.AddServiceHandler(svcHandlers)

	go handleTermination(lbc, ngxc, nginxDone)

	lbc.Run()

	fmt.Printf("End Ingress Nginx")
}

func handleTermination(lbc *controller.LoadBalancerController, ngxc *nginx.Controller, nginxDone chan error) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	exitStatus := 0
	exited := false

	select {
	case err := <-nginxDone:
		if err != nil {
			log.Printf("nginx command exited with an error: %v", err)
			exitStatus = 1
		} else {
			log.Printf("nginx command exited successfully")
		}
		exited = true
	case <-signalChan:
		log.Printf("Received SIGTERM, shutting down")
	}

	glog.Infof("Shutting down the controller")
	lbc.Stop()

	if !exited {
		glog.Infof("Shutting down NGINX")
		ngxc.Quit()
		<-nginxDone
	}

	glog.Infof("Exiting with a status: %v", exitStatus)
	os.Exit(exitStatus)
}
