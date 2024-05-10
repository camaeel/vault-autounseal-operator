package podhandler

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func GetPodHandlerFunctions() cache.ResourceEventHandlerFuncs {
	ret := cache.ResourceEventHandlerFuncs{
		AddFunc: podHandler,
		UpdateFunc: func(oldObj, newObj interface{}) {
			podHandler(newObj)
		},
		DeleteFunc: nil,
	}
	return ret

}

func podHandler(obj interface{}) {
	// check if certificate served by vault doesn't match one in secret
	//// drain pod (so the API will keep minimum pods according to PDB)
	// check if cluster is initialized
	//// if there is no vault secret
	////// initialize
	// check if cluster is sealed
	//// unseal

}

func isInitialized(pod corev1.Pod) bool {
	return pod.Annotations["vault-initialized"] == "true"
}

func isSealed(pod corev1.Pod) bool {
	return pod.Annotations["vault-sealed"] == "true"
}

func isLeader(pod corev1.Pod) bool {
	return pod.Annotations["vault-active"] == "true"
}
