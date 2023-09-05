package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	klient "github.com/pranoyk/volume-snapshotter/pkg/client/clientset/versioned"
	saInfFac "github.com/pranoyk/volume-snapshotter/pkg/client/informers/externalversions"
	"github.com/pranoyk/volume-snapshotter/pkg/controller"

	snapshotclientset "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}

	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klient set %s\n", err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting std client %s\n", err.Error())
	}

	snapshotClient, err := snapshotclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating snapshot client: %v", err)
	}

	infoFactory := saInfFac.NewSharedInformerFactory(klientset, 10*time.Minute)
	ch := make(chan struct{})
	c := controller.NewController(client, *snapshotClient, klientset, infoFactory.Pranoykundu().V1().SnapshotActions())

	infoFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller %s\n", err.Error())
	}

}
