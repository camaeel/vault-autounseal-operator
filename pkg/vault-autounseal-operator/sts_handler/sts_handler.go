package stsHandler

import "k8s.io/client-go/tools/cache"

func GetPodHandlerFunctions() cache.ResourceEventHandlerFuncs {
	ret := cache.ResourceEventHandlerFuncs{
		AddFunc: stsHandler,
		UpdateFunc: func(oldObj, newObj interface{}) {
			stsHandler(newObj)
		},
		DeleteFunc: nil,
	}
	return ret

}

func stsHandler(obj interface{}) {
	// check if sts is different than pods - can this be done through API
	// initiate somehow patching of nodes -> this should go to podHandler (annotation on a existing pod that it needs drain?)

}
