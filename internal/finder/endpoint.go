package finder

import (
	"fmt"
	"time"

	"github.com/llaoj/kube-finder/internal/config"
	"github.com/llaoj/kube-finder/pkg/kube"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type EndpointManager struct {
	Endpoints       map[string]string
	Informer        cache.SharedIndexInformer
	InformerFactory informers.SharedInformerFactory
	Stop            chan struct{}
}

func NewEndpointManager(namespace, labelSelector string) *EndpointManager {
	clientset, err := kube.Client()
	if err != nil {
		log.Fatal(err)
	}

	endpointManager := &EndpointManager{
		Endpoints: make(map[string]string),
		Stop:      make(chan struct{}),
	}

	endpointManager.InformerFactory = informers.NewSharedInformerFactoryWithOptions(clientset, time.Second*0, informers.WithNamespace(namespace), informers.WithTweakListOptions(func(options *metav1.ListOptions) { options.LabelSelector = labelSelector }))
	endpointManager.Informer = endpointManager.InformerFactory.Core().V1().Pods().Informer()
	endpointManager.Informer.AddEventHandler(endpointManager)

	go endpointManager.InformerFactory.Start(endpointManager.Stop)

	return endpointManager
}

func (endpointManager *EndpointManager) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.Status.Phase == v1.PodRunning {
		endpointManager.Endpoints[pod.Status.HostIP] = fmt.Sprintf("%s:%s", pod.Status.PodIP, config.Get().HttpPort)
	}
	log.WithFields(log.Fields{"pod": pod.Name}).Info("pod add")
}

func (endpointManager *EndpointManager) OnUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)
	if oldPod.Status.Phase != v1.PodRunning {
		delete(endpointManager.Endpoints, oldPod.Status.HostIP)
	}
	if newPod.Status.Phase == v1.PodRunning {
		endpointManager.Endpoints[newPod.Status.HostIP] = fmt.Sprintf("%s:%s", newPod.Status.PodIP, config.Get().HttpPort)
	}
	log.WithFields(log.Fields{
		"oldPod": oldPod.Name,
		"newPod": newPod.Name,
	}).Info("pod update")
}

func (endpointManager *EndpointManager) OnDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.Status.Phase != v1.PodRunning {
		delete(endpointManager.Endpoints, pod.Status.HostIP)
	}
	log.WithFields(log.Fields{"pod": pod.Name}).Info("pod delete")
}

func (endpointManager *EndpointManager) StopInformer() {
	if endpointManager.Stop != nil {
		close(endpointManager.Stop)
	}
}
