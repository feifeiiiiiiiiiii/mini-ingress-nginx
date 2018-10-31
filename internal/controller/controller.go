package controller

import (
	"log"
	"time"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/queue"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/utils"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	ingressClassKey = "kubernetes.io/ingress.class"
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
	svcController      cache.Controller
	svcLister          utils.StoreToIngressLister
	endpointLister     utils.StoreToEndpointLister
	endpointController cache.Controller
	stopChan           chan struct{}
	syncQueue          *queue.TaskQueue
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
		stopChan:     make(chan struct{}),
	}
	lbc.syncQueue = queue.NewTaskQueue(lbc.sync)
	return &lbc
}

// GetIngressClassKey returns the ingress class key
func (lbc *LoadBalancerController) GetIngressClassKey() string {
	return ingressClassKey
}

// IsNginxIngress checks if resource ingress class annotation is matching with ingress controller class
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
	case "services":
		lbc.svcController = controller
		lbc.svcLister.Store = store
	default:
		log.Fatalf("unknown resource %v", resource)
	}
}

// Run starts the loadbalancerController controller
func (lbc *LoadBalancerController) Run() {
	go lbc.svcController.Run(lbc.stopChan)
	go lbc.endpointController.Run(lbc.stopChan)
	go lbc.ingressController.Run(lbc.stopChan)
}

// Wait the loadbalancerController stop
func (lbc *LoadBalancerController) Wait() {
	<-lbc.stopChan
}

func (lbc *LoadBalancerController) sync(task queue.Task) {
	log.Printf("Syncing %v", task.Key)
}

// AddSyncQueue enqueues the provided item on the sync queue
func (lbc *LoadBalancerController) AddSyncQueue(item interface{}) {
	lbc.syncQueue.Enqueue(item)
}
