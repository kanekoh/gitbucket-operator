package e2e_test

// https://itnext.io/testing-kubernetes-operators-with-ginkgo-gomega-and-the-operator-runtime-6ad4c2492379
import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// "k8s.io/client-go/rest"

	// "sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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

// var (
// 	cfg        *rest.Config
// 	k8sClient  client.Client
// 	k8sManager ctrl.Manager
// 	testEnv    *envtest.Environment
// )

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	gitbucketList := &operator.GitBucketList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, gitbucketList)
	Expect(err).NotTo(HaveOccurred())

	ctx := framework.NewContext()
	defer ctx.Cleanup()

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	Expect(err).NotTo(HaveOccurred())

	By("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	Expect(err).NotTo(HaveOccurred())

	f := framework.Global
	err = e2eutil.WaitForOperatorDeployment(f.KubeClient, namespace, "gitbucket-operator", 1, retryInterval, timeout)
	Expect(err).NotTo(HaveOccurred())

})
