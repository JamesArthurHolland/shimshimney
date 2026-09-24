package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	shimshimneyv1alpha1 "github.com/shimshimney/operator/api/v1alpha1"
)

// ShimPodReconciler reconciles a ShimPod object
type ShimPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=shimshimney.shimshimney.io,resources=shimpods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=shimshimney.shimshimney.io,resources=shimpods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=shimshimney.shimshimney.io,resources=shimpods/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ShimPod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ShimPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var shimPod shimshimneyv1alpha1.ShimPod
	if err := r.Get(ctx, req.NamespacedName, &shimPod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	serviceName := shimPod.Spec.ServiceName
	if serviceName == "" {
		serviceName = "shim-" + shimPod.Name
	}
	targetPort := shimPod.Spec.TargetPort
	if targetPort == 0 {
		targetPort = shimPod.Spec.Port
	}

	service := &corev1.Service{}
	key := client.ObjectKey{Namespace: shimPod.Namespace, Name: serviceName}
	err := r.Get(ctx, key, service)
	if apierrors.IsNotFound(err) {
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: shimPod.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(&shimPod, shimshimneyv1alpha1.GroupVersion.WithKind("ShimPod")),
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: shimPod.Spec.Selector,
				Ports: []corev1.ServicePort{{
					Name:       "shim",
					Port:       shimPod.Spec.Port,
					TargetPort: intstr.FromInt32(targetPort),
				}},
			},
		}
		return ctrl.Result{}, r.Create(ctx, service)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	service.Spec.Selector = shimPod.Spec.Selector
	service.Spec.Ports = []corev1.ServicePort{{
		Name:       "shim",
		Port:       shimPod.Spec.Port,
		TargetPort: intstr.FromInt32(targetPort),
	}}
	return ctrl.Result{}, r.Update(ctx, service)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ShimPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shimshimneyv1alpha1.ShimPod{}).
		Complete(r)
}
