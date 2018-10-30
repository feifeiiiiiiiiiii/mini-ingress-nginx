package handlers

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/controller"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/utils"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
)

// CreateIngressHandlers builds the handler funcs fro ingresses
func CreateIngressHandlers(lbc *controller.LoadBalancerController) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress := obj.(*extensions.Ingress)
			if !lbc.IsNginxIngress(ingress) {
				glog.Infof("Ignoring Ingress %v based on Annotation %v", ingress.Name, lbc.GetIngressClassKey())
				return
			}
			glog.V(3).Infof("Adding Ingress: %v", ingress.Name)
		},
		DeleteFunc: func(obj interface{}) {
			ingress, isIng := obj.(*extensions.Ingress)
			if !isIng {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.V(3).Infof("Error received unexpected object: %v", obj)
					return
				}
				ingress, ok = deletedState.Obj.(*extensions.Ingress)
				if !ok {
					glog.V(3).Infof("Error DeletedFinalStateUnknown contained non-Ingress object: %v", deletedState.Obj)
					return
				}
			}
			if !lbc.IsNginxIngress(ingress) {
				return
			}
			glog.V(3).Infof("Removing Ingress: %v", ingress.Name)
		},
		UpdateFunc: func(old, current interface{}) {
			c := current.(*extensions.Ingress)
			o := old.(*extensions.Ingress)
			if !lbc.IsNginxIngress(c) {
				return
			}
			if utils.HasChanges(o, c) {
				glog.V(3).Infof("Ingress %v changed, syncing", c.Name)
			}
		},
	}
}
