package controller

import (
	"context"
	"fmt"
	controllerv1 "github.com/arnab-baishnab/sample-controller/pkg/apis/arnabbaishnab.com/v1alpha1"
	clientset "github.com/arnab-baishnab/sample-controller/pkg/generated/clientset/versioned"
	informer "github.com/arnab-baishnab/sample-controller/pkg/generated/informers/externalversions/arnabbaishnab.com/v1alpha1"
	lister "github.com/arnab-baishnab/sample-controller/pkg/generated/listers/arnabbaishnab.com/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"math/rand"
	"time"
)

// For generating unique name for deploy/svc

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func randString(n int) string {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// controller structure for Kluster controller
type Controller struct {
	// kubeclient is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// kuberclientset isclientset for our own API group
	sampleclientset clientset.Interface

	//For deployment resource
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced

	// For Kluster resource
	klusterLister lister.KlusterLister
	klusterSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workQueue workqueue.RateLimitingInterface
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformer.DeploymentInformer,
	klusterInformer informer.KlusterInformer) *Controller {

	ctrl := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		klusterLister:     klusterInformer.Lister(),
		klusterSynced:     klusterInformer.Informer().HasSynced,
		workQueue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	log.Println("Setting up event handlers")

	// Set up an event handler for when Kluster resources change
	_, err := klusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			ctrl.enqueueKluster(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctrl.enqueueKluster(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			ctrl.enqueueKluster(obj)
		},
	})

	if err != nil {
		return nil
	}

	return ctrl
}

// enqueueKluster takes an Kluster resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Kluster.
func (c *Controller) enqueueKluster(obj interface{}) {
	log.Println("Enqueuing Kluster...")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.AddRateLimited(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shut down the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Println("Starting Kluster Controller")

	// Wait for the caches to be synced before starting workers
	log.Println("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.klusterSynced); !ok {
		return fmt.Errorf("failed to wait for cache to sync")
	}

	log.Println("Starting Workers")

	// Launch two workers to process Kluster resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Println("Worker Started")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the workqueue.
func (c *Controller) runWorker() {
	for c.ProcessNextItem() {

	}
}

func (c *Controller) ProcessNextItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func, so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off period.
		defer c.workQueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Kluster resource to be synced.
		if err := c.syncHandler(key); err != nil {
			c.workQueue.Forget(obj)
			return fmt.Errorf("error syncing '%s': %s, requeueing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item, so it does not
		// get queued again until another change happens.
		c.workQueue.Forget(obj)
		log.Printf("Successfully synced '%s'\n", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Kluster resource
// with the current status of the resource.
// implement the business logic here.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	log.Println("SyncHandler called")
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Kluster resource with this namespace/name
	kluster, err := c.klusterLister.Klusters(namespace).Get(name)
	if err != nil {
		// The Kluster resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			// We choose to absorb the error here as the worker would requeue the
			// resource otherwise. Instead, the next time the resource is updated
			// the resource will be queued again.
			//
			utilruntime.HandleError(fmt.Errorf("Kluster '%s' in work queue no longer exists\n", key))
			return nil
		}
		return err
	}

	// Deploment Creating ...
	deploymentName := kluster.Name + "-" + kluster.Spec.Name
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s : deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Kluster.spec
	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(kluster.Namespace).Create(context.TODO(), newDeployment(kluster, deploymentName), metav1.CreateOptions{})
	}

	// If an error occurs during Get/Create, we'll requeue the item, so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If this number of the replicas on the Ishtiaq resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if kluster.Spec.Replicas != nil && *kluster.Spec.Replicas != *deployment.Spec.Replicas {
		log.Println("Kluster %s replicas: %d, deployment replicas: %d\n", name, *kluster.Spec.Replicas)

		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), newDeployment(kluster, deploymentName), metav1.UpdateOptions{})

		// If an error occurs during Update, we'll requeue the item, so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
	}

	// Finally, we update the status block of the Kluster resource to reflect the
	// current state of the world
	err = c.updateKlusterStatus(kluster, deployment)
	if err != nil {
		return err
	}

	//Service Creating ...

	serviceName := deploymentName + "-service"

	// Check if service already exists or not
	service, err := c.kubeclientset.CoreV1().Services(kluster.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		service, err = c.kubeclientset.CoreV1().Services(kluster.Namespace).Create(context.TODO(), newService(kluster, serviceName), metav1.CreateOptions{})
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("service %s created...\n", service.Name)
	} else if err != nil {
		log.Println(err)
		return err
	}

	_, err = c.kubeclientset.CoreV1().Services(kluster.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Controller) updateKlusterStatus(kluster *controllerv1.Kluster, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	klusterCopy := kluster.DeepCopy()
	klusterCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.KlusterV1alpha1().Klusters(kluster.Namespace).Update(context.TODO(), klusterCopy, metav1.UpdateOptions{})

	return err
}

func newDeployment(kluster *controllerv1.Kluster, deploymentName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: kluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(kluster, controllerv1.SchemeGroupVersion.WithKind("Kluster")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: kluster.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  kluster.Name,
					"Kind": "Kluster",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  kluster.Name,
						"Kind": "Kluster",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: kluster.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: kluster.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(kluster *controllerv1.Kluster, serviceName string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(kluster, controllerv1.SchemeGroupVersion.WithKind("Kluster")),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					//Port:       kluster.Spec.Container.Port,
					Port:       3200,
					TargetPort: intstr.FromInt(int(kluster.Spec.Container.Port)),
				},
			},
			Selector: map[string]string{
				"app":  kluster.Name,
				"Kind": "Kluster",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}
