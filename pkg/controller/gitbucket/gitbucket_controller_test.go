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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Gitbucket Controller", func() {
	var (
		name               = "gitbucket-operator"
		namespace          = "gitbucket"
		enable_public bool = false

		routeReplicas int

		dep       *appsv1.Deployment
		routeList *routev1.RouteList
		cl        client.Client
		s         *runtime.Scheme
		objs      []runtime.Object
		gitbucket *operator.GitBucket
		r         *ReconcileGitBucket
		req       reconcile.Request
		res       reconcile.Result
		err       error
	)

	BeforeEach(func() {
		// Create Custom Resource
		gitbucket = &operator.GitBucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: operator.GitBucketSpec{
				Enable_public: enable_public,
			},
		}

		// Set Scheme
		s = scheme.Scheme
		// Add route Openshift scheme
		err := routev1.AddToScheme(s)
		Expect(err).NotTo(HaveOccurred())
		s.AddKnownTypes(operator.SchemeGroupVersion, gitbucket)
	})

	JustBeforeEach(func() {
		r = &ReconcileGitBucket{client: cl, scheme: s}
		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			},
		}

		res, err = r.Reconcile(req)
		Expect(err).NotTo(HaveOccurred())

		
	})

	Describe("Does not define Custom Resource", func() {
		var (
			depList *appsv1.DeploymentList
		)
		BeforeEach(func() {
			objs = []runtime.Object{}
			cl = fake.NewFakeClient(objs...)
		})

		JustBeforeEach(func() {
			// Reconsile Requeue
			Expect(res.Requeue).Should(BeFalse())

			// Get Objects created by Operator
			depList = &appsv1.DeploymentList{}
			err := r.client.List(context.TODO(), depList)
			Expect(err).NotTo(HaveOccurred())

			routeList = &routev1.RouteList{}
			err = r.client.List(context.TODO(), routeList)
			Expect(err).NotTo(HaveOccurred())

		})

		It("should not have Deployment", func() {
			depReplicas := len(depList.Items)
			Expect(depReplicas).Should(BeZero())
		})

		It("should not have route", func() {
			routeReplicas = len(routeList.Items)
			Expect(routeReplicas).Should(BeZero())
		})
	})

	Describe("Define Custom Resource", func() {
		JustBeforeEach(func() {

			// Reconsile Requeue
			Expect(res.Requeue).Should(BeTrue())

			// Get Objects created by Operator
			dep = &appsv1.Deployment{}
			err := r.client.Get(context.TODO(), req.NamespacedName, dep)
			Expect(err).NotTo(HaveOccurred())

			routeList = &routev1.RouteList{}
			err = r.client.List(context.TODO(), routeList)
			Expect(err).NotTo(HaveOccurred())

		})

		Context("When gitbucket is defined without public route", func() {
			BeforeEach(func() {
				gitbucket.Spec.Enable_public = false

				// Reconcile
				objs = []runtime.Object{gitbucket}
				cl = fake.NewFakeClient(objs...)
			})

			JustBeforeEach(func() {
				routeReplicas = len(routeList.Items)
			})

			It("should not have the route", func() {
				Expect(routeReplicas).Should(BeZero())
			})
		})

		Context("When gitbucket is defined with public route", func() {
			BeforeEach(func() {
				gitbucket.Spec.Enable_public = true

				// Reconcile
				objs = []runtime.Object{gitbucket}
				cl = fake.NewFakeClient(objs...)
			})

			JustBeforeEach(func() {
				routeReplicas = len(routeList.Items)
			})

			It("should have a route", func() {
				Expect(routeReplicas).Should(Equal(1))
			})
		})


	})
})
