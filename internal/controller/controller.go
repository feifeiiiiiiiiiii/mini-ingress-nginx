package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/nginx"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/queue"
	"github.com/feifeiiiiiiiiiii/mini-ingress-nginx/internal/utils"
	"github.com/golang/glog"

	api_v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	configurator       *nginx.NgxConfig
}

// NewLoadBalancerControllerInput holds the input needed to call NewLoadBalancerController.
type NewLoadBalancerControllerInput struct {
	KubeClient        kubernetes.Interface
	ResyncPeriod      time.Duration
	NginxConfigurator *nginx.NgxConfig
	Namespace         string
	IngressClass      string
}

// NewLoadBalancerController creates a controller
func NewLoadBalancerController(input NewLoadBalancerControllerInput) *LoadBalancerController {
	lbc := LoadBalancerController{
		namespace:    input.Namespace,
		resync:       input.ResyncPeriod,
		client:       input.KubeClient,
		ingressClass: input.IngressClass,
		stopChan:     make(chan struct{}),
		configurator: input.NginxConfigurator,
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
		lbc.svcLister.Store = store
		lbc.svcController = controller
	default:
		log.Fatalf("unknown resource %v", resource)
	}
}

// Run starts the loadbalancerController controller
func (lbc *LoadBalancerController) Run() {
	go lbc.svcController.Run(lbc.stopChan)
	go lbc.endpointController.Run(lbc.stopChan)
	go lbc.ingressController.Run(lbc.stopChan)
	go lbc.syncQueue.Run(time.Second, lbc.stopChan)
}

// Wait the loadbalancerController stop
func (lbc *LoadBalancerController) Wait() {
	<-lbc.stopChan
}

func (lbc *LoadBalancerController) sync(task queue.Task) {
	switch task.Kind {
	case queue.Ingress:
		lbc.syncIng(task)
		return
	case queue.Endpoints:
		lbc.syncEndpoint(task)
		return
	}
}

// AddSyncQueue enqueues the provided item on the sync queue
func (lbc *LoadBalancerController) AddSyncQueue(item interface{}) {
	lbc.syncQueue.Enqueue(item)
}

func (lbc *LoadBalancerController) syncIng(task queue.Task) {
	key := task.Key
	ing, ingExists, err := lbc.ingressLister.GetByKeySafe(key)
	if err != nil {
		lbc.syncQueue.Requeue(task, err)
		return
	}
	if !ingExists {
		log.Printf("Deleting Ingress: %v %v\n", key, ing)
	} else {
		log.Printf("Adding or Updating Ingress: %v\n", key)
	}

	ingEx, err := lbc.createIngress(ing)
	if err != nil {
		return
	}

	err = lbc.configurator.AddOrUpdateIngress(ingEx)
	if err != nil {
		fmt.Printf("AddedOrUpdatedWithError Configuration for %v was added or updated, but not applied: %v\n", key, err)
	} else {
		fmt.Printf("AddedOrUpdated Configuration for %v was added or updated\n", key)
	}
}

func (lbc *LoadBalancerController) syncEndpoint(task queue.Task) {
	key := task.Key
	obj, endpExists, err := lbc.endpointLister.GetByKey(key)
	if err != nil {
		lbc.syncQueue.Requeue(task, err)
		return
	}

	if !endpExists {
		return
	}

	fmt.Println(obj)

}

func (lbc *LoadBalancerController) createIngress(ing *extensions.Ingress) (*nginx.IngressEx, error) {
	ingEx := &nginx.IngressEx{
		Ingress:      ing,
		Endpoints:    make(map[string][]string),
		HealthChecks: make(map[string]*api_v1.Probe),
	}

	/**
	 * spec:
	 *   backend:
	 *	   serviceName: test
	 *	   servicePort: 80
	 */
	if ing.Spec.Backend != nil {
		endps, err := lbc.getEndpointsForIngressBackend(ing.Spec.Backend, ing.Namespace)
		if err != nil {
			fmt.Printf("Error retrieving endpoints for the service %v: %v\n", ing.Spec.Backend.ServiceName, err)
			ingEx.Endpoints[ing.Spec.Backend.ServiceName+ing.Spec.Backend.ServicePort.String()] = []string{}
		} else {
			ingEx.Endpoints[ing.Spec.Backend.ServiceName+ing.Spec.Backend.ServicePort.String()] = endps
		}
	}

	validRules := 0

	for _, rule := range ing.Spec.Rules {
		if rule.IngressRuleValue.HTTP == nil {
			continue
		}

		if rule.Host == "" {
			return nil, fmt.Errorf("Ingress rule contains empty host")
		}

		for _, path := range rule.HTTP.Paths {
			endps, err := lbc.getEndpointsForIngressBackend(&path.Backend, ing.Namespace)
			if err != nil {
				fmt.Printf("Error retrieving endpoints for the service %v: %v\n", path.Backend.ServiceName, err)
				ingEx.Endpoints[path.Backend.ServiceName+path.Backend.ServicePort.String()] = []string{}
			} else {
				ingEx.Endpoints[path.Backend.ServiceName+path.Backend.ServicePort.String()] = endps
			}
		}
		validRules++
	}
	if validRules == 0 {
		return nil, fmt.Errorf("Ingress contains no valid rules")
	}

	return ingEx, nil
}

func (lbc *LoadBalancerController) getEndpointsForIngressBackend(backend *extensions.IngressBackend, namespace string) ([]string, error) {
	svc, err := lbc.getServiceForIngressBackend(backend, namespace)
	if err != nil {
		glog.V(3).Infof("Error getting service %v: %v", backend.ServiceName, err)
		return nil, err
	}

	endps, err := lbc.endpointLister.GetServiceEndpoints(svc)
	if err != nil {
		glog.V(3).Infof("Error getting endpoints for service %s from the cache: %v", svc.Name, err)
		return nil, err
	}

	result, err := lbc.getEndpointsForPort(endps, backend.ServicePort, svc)
	if err != nil {
		glog.V(3).Infof("Error getting endpoints for service %s port %v: %v", svc.Name, backend.ServicePort, err)
		return nil, err
	}
	return result, nil
}

func (lbc *LoadBalancerController) getServiceForIngressBackend(backend *extensions.IngressBackend, namespace string) (*api_v1.Service, error) {
	svcKey := namespace + "/" + backend.ServiceName
	svcObj, svcExists, err := lbc.svcLister.GetByKey(svcKey)
	if err != nil {
		return nil, err
	}

	if svcExists {
		return svcObj.(*api_v1.Service), nil
	}

	return nil, fmt.Errorf("service %s doesn't exist", svcKey)
}

func (lbc *LoadBalancerController) getEndpointsForPort(endps api_v1.Endpoints, ingSvcPort intstr.IntOrString, svc *api_v1.Service) ([]string, error) {
	var targetPort int32
	var err error
	found := false

	for _, port := range svc.Spec.Ports {
		if (ingSvcPort.Type == intstr.Int && port.Port == int32(ingSvcPort.IntValue())) || (ingSvcPort.Type == intstr.String && port.Name == ingSvcPort.String()) {
			targetPort, err = lbc.getTargetPort(&port, svc)
			if err != nil {
				return nil, fmt.Errorf("Error determining target port for port %v in Ingress: %v", ingSvcPort, err)
			}
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("No port %v in service %s", ingSvcPort, svc.Name)
	}

	for _, subset := range endps.Subsets {
		for _, port := range subset.Ports {
			if port.Port == targetPort {
				var endpoints []string
				for _, address := range subset.Addresses {
					endpoint := fmt.Sprintf("%v:%v", address.IP, port.Port)
					endpoints = append(endpoints, endpoint)
				}
				return endpoints, nil
			}
		}
	}

	return nil, fmt.Errorf("No endpoints for target port %v in service %s", targetPort, svc.Name)
}

func (lbc *LoadBalancerController) getTargetPort(svcPort *api_v1.ServicePort, svc *api_v1.Service) (int32, error) {
	if (svcPort.TargetPort == intstr.IntOrString{}) {
		return svcPort.Port, nil
	}

	if svcPort.TargetPort.Type == intstr.Int {
		return int32(svcPort.TargetPort.IntValue()), nil
	}

	pods, err := lbc.client.Core().Pods(svc.Namespace).List(meta_v1.ListOptions{LabelSelector: labels.Set(svc.Spec.Selector).String()})
	if err != nil {
		return 0, fmt.Errorf("Error getting pod information: %v", err)
	}

	if len(pods.Items) == 0 {
		return 0, fmt.Errorf("No pods of service %s", svc.Name)
	}

	pod := &pods.Items[0]

	portNum, err := utils.FindPort(pod, svcPort)
	if err != nil {
		return 0, fmt.Errorf("Error finding named port %v in pod %s: %v", svcPort, pod.Name, err)
	}

	return portNum, nil
}
