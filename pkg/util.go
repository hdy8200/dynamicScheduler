package plugins

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	// utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/klog"
	// v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	// "k8s.io/kubernetes/pkg/features"

	"context"

	// corev1 "k8s.io/api/core/v1"
	// "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core
	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 50 * 1024 * 1024 // 200 MB
)

// // GetNonzeroRequestForResource returns the default resource request if none is found or
// // what is provided on the request.
// func GetNonzeroRequestForResource(resource v1.ResourceName, requests *v1.ResourceList) int64 {
// 	switch resource {
// 	case v1.ResourceCPU:
// 		// Override if un-set, but not if explicitly set to zero
// 		if _, found := (*requests)[v1.ResourceCPU]; !found {
// 			return DefaultMilliCPURequest
// 		}
// 		return requests.Cpu().MilliValue()
// 	case v1.ResourceMemory:
// 		// Override if un-set, but not if explicitly set to zero
// 		if _, found := (*requests)[v1.ResourceMemory]; !found {
// 			return DefaultMemoryRequest
// 		}
// 		return requests.Memory().Value()
// 	case v1.ResourceEphemeralStorage:
// 		// if the local storage capacity isolation feature gate is disabled, pods request 0 disk.
// 		if !utilfeature.DefaultFeatureGate.Enabled(features.LocalStorageCapacityIsolation) {
// 			return 0
// 		}

// 		quantity, found := (*requests)[v1.ResourceEphemeralStorage]
// 		if !found {
// 			return 0
// 		}
// 		return quantity.Value()
// 	default:
// 		if v1helper.IsScalarResourceName(resource) {
// 			quantity, found := (*requests)[resource]
// 			if !found {
// 				return 0
// 			}
// 			return quantity.Value()
// 		}
// 	}
// 	return 0
// }

func getResourceUsage(nodeName string) (cpuUsage, memoryUsage int64){
	kubeconfig := "/etc/hdykubernetes/kubeconfig" // change to your kubeconfig path
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	metricsClientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodes.Items {
		if node.Name == nodeName {
			nodeMetrics, err := metricsClientset.MetricsV1beta1().NodeMetricses().Get(context.Background(), node.Name, metav1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}
	
			cpuQuantity := nodeMetrics.Usage[v1.ResourceCPU]
			memorQuantity := nodeMetrics.Usage[v1.ResourceMemory]
			cpuUsage, memoryUsage = cpuQuantity.MilliValue(), memorQuantity.Value()
			klog.V(3).Infof("Node: %s, CPU usage: %v,  Memory usage: %v\n", node.Name, cpuUsage, memoryUsage)
		}
	}
	if cpuUsage == 0 || memoryUsage == 0 {
		panic(fmt.Sprintf("cannot get %s resource usage, get CPU usage: %v  Memory usage: %v\n", nodeName, cpuUsage, memoryUsage))
	}
	return cpuUsage, memoryUsage
}
