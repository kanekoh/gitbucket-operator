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
	// logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("Gitbucket Controller", func() {
	var (
		name               = "gitbucket-operator"
		namespace          = "gitbucket"
		image              = "https://localhost/testimage"
		enable_public bool = false

		imageURL      string
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
				Image:         image,
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
			// Get Objects created by Operator
			depList = &appsv1.DeploymentList{}
			err := r.client.List(context.TODO(), depList)
			Expect(err).NotTo(HaveOccurred())

			routeList = &routev1.RouteList{}
			err = r.client.List(context.TODO(), routeList)
			Expect(err).NotTo(HaveOccurred())

			// Reconsile Requeue
			Expect(res.Requeue).Should(BeFalse())
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
		BeforeEach(func() {
			// Reconcile
			objs = []runtime.Object{gitbucket}
			cl = fake.NewFakeClient(objs...)
		})

		JustBeforeEach(func() {
			// Get Objects created by Operator
			dep = &appsv1.Deployment{}
			err := r.client.Get(context.TODO(), req.NamespacedName, dep)
			Expect(err).NotTo(HaveOccurred())

			routeList = &routev1.RouteList{}
			err = r.client.List(context.TODO(), routeList)
			Expect(err).NotTo(HaveOccurred())

			// Reconsile Requeue
			Expect(res.Requeue).Should(BeTrue())
		})

		Context("When gitbucket is defined without public route", func() {
			BeforeEach(func() {
				enable_public = false
			})

			JustBeforeEach(func() {
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

		Context("When gitbucket is defined with public route", func() {
			BeforeEach(func() {
				enable_public = true
			})

			JustBeforeEach(func() {
				imageURL = dep.Spec.Template.Spec.Containers[0].Image
				routeReplicas = len(routeList.Items)
			})

			It("should have the image URL was specified", func() {
				Expect(imageURL).Should(Equal(image))
			})

			It("should have a route", func() {
				Expect(routeReplicas).Should(Equal(1))
			})
		})

		Context("When gitbucket is re-defined with modified image URL", func(){
			BeforeEach(func(){
				enable_public = false
			})

			JustBeforeEach(func(){
				dep.Spec.Template.Spec.Containers[0].Image = "fakeImage"
				err := r.client.Update(context.TODO(), dep)
				Expect(err).NotTo(HaveOccurred())

				res, err = r.Reconcile(req)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Requeue).Should(BeTrue())

				// Get Objects created by Operator
				dep = &appsv1.Deployment{}
				err = r.client.Get(context.TODO(), req.NamespacedName, dep)
				Expect(err).NotTo(HaveOccurred())

				routeList = &routev1.RouteList{}
				err = r.client.List(context.TODO(), routeList)
				Expect(err).NotTo(HaveOccurred())
				
				imageURL = dep.Spec.Template.Spec.Containers[0].Image
				routeReplicas = len(routeList.Items)
			})

			It("should have the image URL was specified in CR", func() {
				Expect(imageURL).Should(Equal(image))
			})

			It("should not have the route", func() {
				Expect(routeReplicas).Should(BeZero())
			})
		})
	})
})
