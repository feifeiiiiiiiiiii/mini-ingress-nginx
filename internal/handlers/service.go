package handlers

import (
	"log"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// CreateServiceHandlers builds the handler funcs for services
func CreateServiceHandlers(lbc *controller.LoadBalancerController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc := obj.(*api_v1.Service)
			log.Printf("Adding service: %v", svc.Name)
		},
		DeleteFunc: func(obj interface{}) {
		},
		UpdateFunc: func(old, cur interface{}) {

		},
	}
}
