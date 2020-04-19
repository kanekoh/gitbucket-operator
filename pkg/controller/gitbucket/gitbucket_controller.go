package gitbucket

import (
	"context"
	"reflect"

	gitbucketv1alpha1 "github.com/kanekoh/gitbucket-operator/pkg/apis/gitbucket/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gitbucket")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GitBucket Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitBucket{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitbucket-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitBucket
	err = c.Watch(&source.Kind{Type: &gitbucketv1alpha1.GitBucket{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner GitBucket
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitbucketv1alpha1.GitBucket{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGitBucket implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGitBucket{}

// ReconcileGitBucket reconciles a GitBucket object
type ReconcileGitBucket struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitBucket object and makes changes based on the state read
// and what is in the GitBucket.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitBucket")

	// Fetch the GitBucket instance
	gitbucket := &gitbucketv1alpha1.GitBucket{}
	// TODO() は空ではない何らかのオブジェクトを返却する。
	err := r.client.Get(context.TODO(), request.NamespacedName, gitbucket)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: gitbucket.Name, Namespace: gitbucket.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		//Define a new deployment
		dep := r.newDeploymentForGitBucket(gitbucket)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	foundService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: gitbucket.Name, Namespace: gitbucket.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		//Define a service
		svc := r.newServiceForGitBucket(gitbucket)
		reqLogger.Info("Creating a new service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)

		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			reqLogger.Error(err, "Failed to create new service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	// Check if the route already exists, if not create a new one
	if gitbucket.Spec.Enable_public {
		foundRoute := &routev1.Route{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: gitbucket.Name, Namespace: gitbucket.Namespace}, foundRoute)
		if err != nil && errors.IsNotFound(err) {
			// Define a route
			route := r.newRouteForGitBucket(gitbucket)
			reqLogger.Info("Creating a new route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)

			err = r.client.Create(context.TODO(), route)
			if err != nil {
				reqLogger.Error(err, "Faild to create new route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Faild to get route")
			return reconcile.Result{}, err
		}
	}

	// // Ensure the deployment image is the same as the spec
	// image := gitbucket.Spec.Image
	// if found.Spec.Template.Spec.Containers[0].Image != image {
	// 	// Update Image as image.
	// 	found.Spec.Template.Spec.Containers[0].Image = image
	// 	if err != nil {
	// 		reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	// 		return reconcile.Result{}, err
	// 	}
	// 	// Spec updated - return and requeue
	// 	return reconcile.Result{Requeue: true}, nil
	// }

	// Update the Gitbucket status with the pod names
	// List the pods for this gitbucket's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(gitbucket.Namespace),
		client.MatchingLabels(labelsForGitBucket(gitbucket.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Gitbucket.Namespace", gitbucket.Namespace, "Gitbucket.Name", gitbucket.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, gitbucket.Status.Nodes) {
		gitbucket.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), gitbucket)
		if err != nil {
			reqLogger.Error(err, "Failed to update gitbucket status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: true}, nil

}

func (r *ReconcileGitBucket) newServiceForGitBucket(g *gitbucketv1alpha1.GitBucket) *corev1.Service {
	ls := labelsForGitBucket(g.Name)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.Name,
			Namespace: g.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromString("gitbucket"),
				},
			},
		},
	}

	return svc
}

// Refer to the https://github.com/openshift/api/blob/master/route/v1/types.go
func (r *ReconcileGitBucket) newRouteForGitBucket(g *gitbucketv1alpha1.GitBucket) *routev1.Route {
	ls := labelsForGitBucket(g.Name)
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.Name,
			Namespace: g.Namespace,
			Labels:    ls,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: g.Name,
			},
		},
	}
	return route
}

func (r *ReconcileGitBucket) newDeploymentForGitBucket(g *gitbucketv1alpha1.GitBucket) *appsv1.Deployment {
	ls := labelsForGitBucket(g.Name)
	image := g.Spec.Image
	var replicas int32
	replicas = 1

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.Name,
			Namespace: g.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
						Name:  "gitbucket",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "gitbucket",
						}},
					}},
				},
			},
		},
	}

	// Set Gitbucket instance as the owner and controller
	controllerutil.SetControllerReference(g, dep, r.scheme)
	return dep
}

// labelsForGitBucket returns the labels for selecting the resources
// belonging to the given gitbucket CR name.
func labelsForGitBucket(name string) map[string]string {
	return map[string]string{"app": "gitbucket", "gitbucket_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
