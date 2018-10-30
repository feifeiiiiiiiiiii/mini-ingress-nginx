package controller

import (
	"time"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/utils"
	"github.com/golang/glog"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	ingressClassKey = "kubernetes.io/mini.nginx.ingress.class"
)

// LoadBalancerController watches Kubernetes API and
// reconfigures NGINX via NginxController when needed
type LoadBalancerController struct {
	client             kubernetes.Interface
	ingressController  cache.Controller
	namespace          string
	resync             time.Duration
	ingressClass       string
	ingressLister      utils.StoreToIngressLister
	svcLister          cache.Store
	endpointLister     utils.StoreToEndpointLister
	endpointController cache.Controller
}

// NewLoadBalancerControllerInput holds the input needed to call NewLoadBalancerController.
type NewLoadBalancerControllerInput struct {
	KubeClient   kubernetes.Interface
	ResyncPeriod time.Duration
	Namespace    string
	IngressClass string
}

// NewLoadBalancerController creates a controller
func NewLoadBalancerController(input NewLoadBalancerControllerInput) *LoadBalancerController {
	lbc := LoadBalancerController{
		namespace:    input.Namespace,
		resync:       input.ResyncPeriod,
		client:       input.KubeClient,
		ingressClass: input.IngressClass,
	}
	return &lbc
}

// GetIngressClassKey returns the ingress class key
func (lbc *LoadBalancerController) GetIngressClassKey() string {
	return ingressClassKey
}

// IsNginxIngress checks if resource ingress class annotation (if exists) is matching with ingress controller class
func (lbc *LoadBalancerController) IsNginxIngress(ing *extensions.Ingress) bool {
	if class, exists := ing.Annotations[ingressClassKey]; exists {
		return class == lbc.ingressClass || class == ""
	}
	return false
}

// AddResourceHandler adds the handlers for ingress\services
func (lbc *LoadBalancerController) AddResourceHandler(resource string, handlers cache.ResourceEventHandlerFuncs) {
	store, controller := cache.NewInformer(
		cache.NewListWatchFromClient(
			lbc.client.Extensions().RESTClient(),
			resource,
			lbc.namespace,
			fields.Everything()),
		&extensions.Ingress{},
		lbc.resync,
		handlers,
	)
	switch resource {
	case "ingresses":
		lbc.ingressLister.Store = store
		lbc.ingressController = controller
	case "endpoints":
		lbc.endpointLister.Store = store
		lbc.endpointController = controller
	default:
		glog.Fatalf("unknown resource %v", resource)
	}
}
