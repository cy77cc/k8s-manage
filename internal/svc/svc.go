package svc

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceContext struct {
	Clientset *kubernetes.Clientset
}

func NewServiceContext() *ServiceContext {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &ServiceContext{
		Clientset: clientset,
	}
}
