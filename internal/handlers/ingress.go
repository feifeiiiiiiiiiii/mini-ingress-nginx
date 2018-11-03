package handlers

import (
	"log"
	"reflect"

	api_v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
)

// CreateIngressHandlers builds the handler funcs fro ingresses
func CreateIngressHandlers(lbc *controller.LoadBalancerController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress := obj.(*extensions.Ingress)
			if !lbc.IsNginxIngress(ingress) {
				log.Printf("Ignoring Ingress %v based on Annotation %v\n", ingress.Name, lbc.GetIngressClassKey())
				return
			}
			log.Printf("Adding Ingress: %v", ingress.Name)
			lbc.AddSyncQueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			ingress, isIng := obj.(*extensions.Ingress)
			if !isIng {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Printf("Error received unexpected object: %v", obj)
					return
				}
				ingress, ok = deletedState.Obj.(*extensions.Ingress)
				if !ok {
					log.Printf("Error DeletedFinalStateUnknown contained non-Ingress object: %v", deletedState.Obj)
					return
				}
			}
			if !lbc.IsNginxIngress(ingress) {
				return
			}
			log.Printf("Removing Ingress: %v", ingress.Name)
			lbc.AddSyncQueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				glog.V(3).Infof("Endpoints %v changed, syncing", cur.(*api_v1.Endpoints).Name)
				lbc.AddSyncQueue(cur)
			}
		},
	}
}
