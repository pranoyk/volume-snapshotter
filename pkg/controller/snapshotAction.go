package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	snapshotsv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotclientset "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"

	klientset "github.com/pranoyk/volume-snapshotter/pkg/client/clientset/versioned"
	kinf "github.com/pranoyk/volume-snapshotter/pkg/client/informers/externalversions/pranoykundu.dev/v1"
	klister "github.com/pranoyk/volume-snapshotter/pkg/client/listers/pranoykundu.dev/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	clientset      kubernetes.Interface
	snapshotClient snapshotclientset.Clientset
	saClient       klientset.Interface
	saSynced       cache.InformerSynced
	saLister       klister.SnapshotActionLister
	wq             workqueue.RateLimitingInterface
}

var (
	snapshotClassName = "csi-hostpath-snapclass"
	storageClassName  = "csi-hostpath-sc"
)

func NewController(client kubernetes.Interface, snapchotClient snapshotclientset.Clientset, klient klientset.Interface, saInformer kinf.SnapshotActionInformer) *Controller {
	c := &Controller{
		clientset:      client,
		snapshotClient: snapchotClient,
		saClient:       klient,
		saSynced:       saInformer.Informer().HasSynced,
		saLister:       saInformer.Lister(),
		wq:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "snapshotAction"),
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
	item, shutdown := c.wq.Get()
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

	if sa.Spec.Action == "createPVC" {
		apiGroup := "snapshot.storage.k8s.io"
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: sa.Spec.DestinationPVC,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &storageClassName,
				DataSource: &corev1.TypedLocalObjectReference{
					Name:     sa.Spec.SnapshotName,
					Kind:     "VolumeSnapshot",
					APIGroup: &apiGroup,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
			},
		}

		_, err := c.clientset.CoreV1().PersistentVolumeClaims(sa.Namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("error creating pvc for snapshot action %s, error: %s\n", sa.Name, err.Error())
			return false
		}
	}
	if sa.Spec.Action == "takeSnapshot" {
		snapshot := &snapshotsv1.VolumeSnapshot{
			ObjectMeta: metav1.ObjectMeta{
				Name: sa.Spec.SnapshotName,
			},
			Spec: snapshotsv1.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &snapshotClassName,
				Source: snapshotsv1.VolumeSnapshotSource{
					PersistentVolumeClaimName: &sa.Spec.SourcePVC,
				},
			},
		}
		_, err := c.snapshotClient.SnapshotV1().VolumeSnapshots(sa.Namespace).Create(context.TODO(), snapshot, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("error creating snapshot for snapshot action %s, error: %s\n", sa.Name, err.Error())
			return false
		}
	}
	fmt.Printf("snapshot action %s\n", sa.Name)
	return true
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("add event received")
	c.wq.Add(obj)
}
