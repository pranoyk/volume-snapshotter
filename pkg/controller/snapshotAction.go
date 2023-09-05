package controller

import (
	"fmt"
	"log"
	"time"

	klientset "github.com/pranoyk/volume-snapshotter/pkg/client/clientset/versioned"
	kinf "github.com/pranoyk/volume-snapshotter/pkg/client/informers/externalversions/pranoykundu.dev/v1"
	klister "github.com/pranoyk/volume-snapshotter/pkg/client/listers/pranoykundu.dev/v1"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client kubernetes.Interface
	saClient klientset.Interface
	saSynced cache.InformerSynced
	saLister klister.SnapshotActionLister
	wq workqueue.RateLimitingInterface
	recorder record.EventRecorder
}

func NewController(client kubernetes.Interface, klient klientset.Interface, saInformer kinf.SnapshotActionInformer) *Controller {
	c := &Controller{
		client:   client,
		saClient: klient,
		saSynced: saInformer.Informer().HasSynced,
		saLister: saInformer.Lister(),
		wq:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "snapshotAction"),
	}

	saInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleAdd,
	})

	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.saSynced); !ok {
		log.Println("cache was not sycned")
	}

	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutdown :=c.wq.Get()
	if shutdown {
		return false
	}

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("error getting key from cache %s\n", err.Error())
		return false
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("error splitting key into namespace and name %s\n", err.Error())
		return false
	}

	sa, err := c.saLister.SnapshotActions(ns).Get(name)
	if err != nil {
		fmt.Printf("error getting snapshot action %s\n", err.Error())
		return false
	}

	fmt.Printf("snapshot action %s\n", sa.Name)
	return true
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("add event received")
	c.wq.Add(obj)
}
