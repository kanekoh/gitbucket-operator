package e2e

import (
	goctx "context"
	"testing"
	"time"

	apis "github.com/kanekoh/gitbucket-operator/pkg/apis"
	operator "github.com/kanekoh/gitbucket-operator/pkg/apis/gitbucket/v1alpha1"

	"github.com/kanekoh/gitbucket-operator/test/util"

	routev1 "github.com/openshift/api/route/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 90
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestGitbucket(t *testing.T) {
	gitbucketList := &operator.GitBucketList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, gitbucketList)
	util.NotToHaveError(t, err, "failed to add custom resource scheme to framework: %v")

	t.Run("gitbucket-group", func(t *testing.T) {
		t.Run("Operator", GitbucketCluster)
		t.Run("Cluster2", GitbucketCluster)
	})
}

func GitbucketCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	util.NotToHaveError(t, err, "failed to initialize cluster resources: %v")

	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	util.NotToHaveError(t, err, "failed to get a namespace: %v")

	f := framework.Global

	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "gitbucket-operator", 1, retryInterval, timeout)
	util.NotToHaveError(t, err, "failed to wait operator deployment: %v")

	// Main Test
	err = gitbucketWithoutRoute(t, f, ctx)
	util.NotToHaveError(t, err, "Failed test gitbucketWithoutRoute: %v")

}

func gitbucketWithoutRoute(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	// Get current namespace
	namespace, err := ctx.GetNamespace()
	util.NotToHaveError(t, err, "could not get namespace: %v")

	// Get Basic GitBukect Custom Resource
	gitbucket := getBasicGitBucket(namespace)

	// Disable creating route
	gitbucket.Spec.Enable_public = false

	err = f.Client.Create(goctx.TODO(), gitbucket, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	util.NotToHaveError(t, err, "could not create custom resource: %v")

	err = e2eutil.WaitForDeployment(t, f.KubeClient, gitbucket.ObjectMeta.Namespace, gitbucket.ObjectMeta.Name, 1, retryInterval, timeout)
	util.NotToHaveError(t, err, "could not wait for deployment: %v")

	// All test passed
	return nil
}

func gitbucketWitRoute(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	// Get current namespace
	namespace, err := ctx.GetNamespace()
	util.NotToHaveError(t, err, "could not get namespace: %v")

	// Get Basic GitBukect Custom Resource
	gitbucket := getBasicGitBucket(namespace)

	// Disable creating route
	gitbucket.Spec.Enable_public = true

	err = f.Client.Create(goctx.TODO(), gitbucket, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	util.NotToHaveError(t, err, "could not create custom resource: %v")

	err = e2eutil.WaitForDeployment(t, f.KubeClient, gitbucket.ObjectMeta.Namespace, gitbucket.ObjectMeta.Name, 1, retryInterval, timeout)
	util.NotToHaveError(t, err, "could not wait for deployment: %v")

	route := &routev1.Route{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: gitbucket.ObjectMeta.Name, Namespace: gitbucket.ObjectMeta.Name}, route)
	util.NotToHaveError(t, err, "could not get a route: %v")

	// All test passed
	return nil
}


func getBasicGitBucket(namespace string) *operator.GitBucket{
	return &operator.GitBucket{
		ObjectMeta: metav1.ObjectMeta {
			Name: "basic-gitbucket",
			Namespace: namespace + "-instance",
		},
		Spec: operator.GitBucketSpec{
			Image: "test-images:v1",
			Enable_public: false,
		},
	}
}