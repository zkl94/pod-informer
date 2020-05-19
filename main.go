package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"time"
)

func main() {

	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		kubeconfig := filepath.Join("/root", ".kube", "config")
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Printf("The kubeconfig cannot be loaded: %v\n", err)
			os.Exit(1)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	podInformer := informerFactory.Core().V1().Pods()

	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	klog.Info("Setting up event handlers")
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			var object metav1.Object
			var ok bool
			if object, ok = obj.(metav1.Object); ok {
				// using default timestamp format
				klog.Infof("Pod %s created in %s at %s", object.GetName(), object.GetNamespace(), object.GetCreationTimestamp())
			} else {
				klog.Info("error processing object of pod AddFunc event")
			}
		},
		DeleteFunc: func(obj interface{}) {
			var object metav1.Object
			var ok bool
			if object, ok = obj.(metav1.Object); ok {
				// using default timestamp format
				klog.Infof("Pod %s deleted in %s at %s", object.GetName(), object.GetNamespace(), object.GetCreationTimestamp())
			} else {
				klog.Info("error processing object of pod DeleteFunc event")
			}
		},
	})

	go podInformer.Informer().Run(stopper)
	if !cache.WaitForCacheSync(stopper, podInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	<-stopper
}
