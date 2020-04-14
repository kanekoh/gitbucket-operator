package gitbucket

import (
	"context"
	"testing"

	operator "github.com/kanekoh/gitbucket-operator/pkg/apis/gitbucket/v1alpha1"

    routev1 "github.com/openshift/api/route/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    // logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestGitbucketControllerDeploymentCreate(t *testing.T) {
	var (
		name               = "gitbucket-operator"
		namespace          = "gitbucket"
		image 			   = "https://localhost/testimage"
		enable_public bool = false
	)

	gitbucket := &operator.GitBucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Spec: operator.GitBucketSpec{
			Image: image,
			Enable_public: enable_public,
		},
	}

	objs := []runtime.Object{ gitbucket }

	s := scheme.Scheme
	s.AddKnownTypes(operator.SchemeGroupVersion, gitbucket)
    // Add route Openshift scheme
    if err := routev1.AddToScheme(s); err != nil {
        t.Fatalf("Unable to add route scheme: (%v)", err)
	}
	
	cl := fake.NewFakeClient(objs...)

	r := &ReconcileGitBucket{client: cl, scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: name,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue {
        t.Error("reconcile did not requeue request as expected")
    }
 
	dep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	imageURL := dep.Spec.Template.Spec.Containers[0].Image
	if imageURL != image {
		t.Errorf("Image URL (%s) is not the expected image (%s)", imageURL, image)
	}

	routeList := &routev1.RouteList{}
	err = r.client.List(context.TODO(), routeList)
	if err != nil {
		t.Fatalf("list routes: (%v)", err)
	}
	if len(routeList.Items) != 0 {
		t.Fatalf("Routes should be 0 but %d", len(routeList.Items))
	}
}

