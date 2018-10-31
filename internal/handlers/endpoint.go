package handlers

import (
	"log"
	"reflect"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// CreateEndpointHandlers builds the handler funcs for endpoints
func CreateEndpointHandlers(lbc *controller.LoadBalancerController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			endpoint := obj.(*api_v1.Endpoints)
			log.Printf("Adding endpoints: %v", endpoint.Name)
		},
		DeleteFunc: func(obj interface{}) {
			endpoint, isEndpoint := obj.(*api_v1.Endpoints)
			if !isEndpoint {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Printf("Error received unexpected object: %v", obj)
					return
				}
				endpoint, ok = deletedState.Obj.(*api_v1.Endpoints)
				if !ok {
					log.Printf("Error DeletedFinalStateUnknown contained non-Endpoints object: %v", deletedState.Obj)
					return
				}
			}
			log.Printf("Removing endpoints: %v", endpoint.Name)
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				log.Printf("Endpoints %v changed, syncing", cur.(*api_v1.Endpoints).Name)
			}
		},
	}
}
