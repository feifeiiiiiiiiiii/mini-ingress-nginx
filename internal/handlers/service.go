package handlers

import (
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"k8s.io/client-go/tools/cache"
)

// CreateServiceHandlers builds the handler funcs for services
func CreateServiceHandlers(lbc *controller.LoadBalancerController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
		},
		UpdateFunc: func(old, cur interface{}) {

		},
	}
}
