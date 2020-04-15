package e2e

import (
	"testing"
	"time"

	apis "github.com/kanekoh/gitbucket-operator/pkg/apis"
	operator "github.com/kanekoh/gitbucket-operator/pkg/apis/gitbucket/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
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
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	t.Run("gitbucket-group", func(t *testing.T) {
		t.Run("Cluster", GitbucketCluster)
		t.Run("Cluster2", GitbucketCluster)
	})
}

func GitbucketCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global

	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "gitbucket-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

}
