package gitbucket

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	operator "github.com/kanekoh/gitbucket-operator/pkg/apis/gitbucket/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("Gitbucket Controller", func() {
	var (
		name               = "gitbucket-operator"
		namespace          = "gitbucket"
		image              = "https://localhost/testimage"
		enable_public bool = false

		dep       *appsv1.Deployment
		routeList *routev1.RouteList

		imageURL      string
		routeReplicas int
	)

	BeforeEach(func() {
		gitbucket := &operator.GitBucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: operator.GitBucketSpec{
				Image:         image,
				Enable_public: enable_public,
			},
		}

		objs := []runtime.Object{gitbucket}

		s := scheme.Scheme
		s.AddKnownTypes(operator.SchemeGroupVersion, gitbucket)
		// Add route Openshift scheme
		err := routev1.AddToScheme(s)
		Expect(err).NotTo(HaveOccurred())

		cl := fake.NewFakeClient(objs...)

		r := &ReconcileGitBucket{client: cl, scheme: s}

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			},
		}

		res, err := r.Reconcile(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Requeue).Should(BeTrue())

		dep = &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), req.NamespacedName, dep)
		Expect(err).NotTo(HaveOccurred())

		routeList = &routev1.RouteList{}
		err = r.client.List(context.TODO(), routeList)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When gitbucket is define without public route", func() {
		JustBeforeEach(func() {
			enable_public = false
		})

		BeforeEach(func() {
			imageURL = dep.Spec.Template.Spec.Containers[0].Image
			routeReplicas = len(routeList.Items)
		})

		It("should have the image URL was specified", func() {
			Expect(imageURL).Should(Equal(image))
		})

		It("should not have the route", func() {
			Expect(routeReplicas).Should(BeZero())
		})
	})

	Context("When gitbucket is define with public route", func() {
		JustBeforeEach(func() {
			enable_public = true
		})

		BeforeEach(func() {
			imageURL = dep.Spec.Template.Spec.Containers[0].Image
			routeReplicas = len(routeList.Items)
		})

		It("should have the image URL was specified", func() {
			Expect(imageURL).Should(Equal(image))
		})

		It("should have a route", func() {
			Expect(routeReplicas).Should(HaveLen(1))
		})
	})
})
