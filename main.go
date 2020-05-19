package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)
	// handle stop signal
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\r- Ctrl+C pressed in Terminal")
	os.Exit(0)
}
