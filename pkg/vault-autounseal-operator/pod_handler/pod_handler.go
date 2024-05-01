package podhandler

import "k8s.io/client-go/tools/cache"

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
