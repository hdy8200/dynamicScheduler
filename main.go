package main

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, node := range nodes.Items {
			// 获取节点的内存利用率，并根据您定义的规则识别要移除的 Pod
			totalMemory := node.Status.Capacity[corev1.ResourceMemory]
			usedMemory := node.Status.Allocatable[corev1.ResourceMemory]
			usedMemoryPercentage := (1 - (usedMemory.AsDec().Cmp(totalMemory.AsDec()))) * 100

			if usedMemoryPercentage >= 80 {
				podList, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
					FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
				})
				if err != nil {
					panic(err.Error())
				}

				var lowestPriorityPod *corev1.Pod
				for i := range podList.Items {
					pod := &podList.Items[i]
					if pod.Namespace != "kube-system" {
						if lowestPriorityPod == nil || getPodPriority(pod) < getPodPriority(lowestPriorityPod) {
							lowestPriorityPod = pod
						}
					}
				}

				if lowestPriorityPod != nil && (lowestPriorityPod.Status.QOSClass == corev1.PodQOSBurstable || lowestPriorityPod.Status.QOSClass == corev1.PodQOSBestEffort) {
					err := clientset.CoreV1().Pods(lowestPriorityPod.Namespace).Delete(context.Background(), lowestPriorityPod.Name, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("Failed to delete pod %s: %v\n", lowestPriorityPod.Name, err)
					} else {
						fmt.Printf("Deleted pod %s on node %s\n", lowestPriorityPod.Name, node.Name)
					}
				}
			}
		}

		// 休眠一段时间（例如 10 分钟）后再次检查节点状态
		time.Sleep(10 * time.Minute)
	}
}

func getPodPriority(pod *corev1.Pod) int32 {
	// Todo update the getPodPriority logic
	if pod.Spec.Priority != nil {
		return *pod.Spec.Priority
	}
	return 0
}