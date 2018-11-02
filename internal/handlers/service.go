package handlers

import (
	"fmt"
	"log"
	"reflect"
	"sort"

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
			lbc.EnqueueIngressForService(svc)
		},
		DeleteFunc: func(obj interface{}) {
			svc, isSvc := obj.(*api_v1.Service)
			if !isSvc {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					fmt.Printf("Error received unexpected object: %v", obj)
					return
				}
				svc, ok = deletedState.Obj.(*api_v1.Service)
				if !ok {
					fmt.Printf("Error DeletedFinalStateUnknown contained non-Service object: %v", deletedState.Obj)
					return
				}
			}

			fmt.Printf("Removing service: %v", svc.Name)
			lbc.EnqueueIngressForService(svc)
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				curSvc := cur.(*api_v1.Service)
				oldSvc := old.(*api_v1.Service)
				if hasServicePortChanges(oldSvc.Spec.Ports, curSvc.Spec.Ports) {
					fmt.Printf("Service %v changed, syncing", curSvc.Name)
					lbc.EnqueueIngressForService(curSvc)
				}
			}
		},
	}
}

type portSort []api_v1.ServicePort

func (a portSort) Len() int {
	return len(a)
}

func (a portSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a portSort) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		return a[i].Port < a[j].Port
	}
	return a[i].Name < a[j].Name
}

// hasServicePortChanges only compares ServicePort.Name and .Port.
func hasServicePortChanges(oldServicePorts []api_v1.ServicePort, curServicePorts []api_v1.ServicePort) bool {
	if len(oldServicePorts) != len(curServicePorts) {
		return true
	}

	sort.Sort(portSort(oldServicePorts))
	sort.Sort(portSort(curServicePorts))

	for i := range oldServicePorts {
		if oldServicePorts[i].Port != curServicePorts[i].Port ||
			oldServicePorts[i].Name != curServicePorts[i].Name {
			return true
		}
	}
	return false
}
