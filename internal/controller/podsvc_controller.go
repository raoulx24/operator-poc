package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorpocv1alpha1 "github.com/raoulx24/operator-poc/api/v1alpha1"
)

// PodSvcReconciler reconciles a PodSvc object
type PodSvcReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operatorpoc.my.domain,resources=podsvcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operatorpoc.my.domain,resources=podsvcs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operatorpoc.my.domain,resources=podsvcs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *PodSvcReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PodSvc resource
	var podsvc operatorpocv1alpha1.PodSvc
	if err := r.Get(ctx, req.NamespacedName, &podsvc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List Pods matching the label selector
	var pods corev1.PodList
	if err := r.List(ctx, &pods,
		client.InNamespace(req.Namespace),
		client.MatchingLabels{podsvc.Spec.LabelName: podsvc.Spec.LabelValue},
	); err != nil {
		return ctrl.Result{}, err
	}

	// Prepare status entries
	var statusEntries []operatorpocv1alpha1.PodSvcStatusEntry
	expectedServices := map[string]bool{}

	for _, pod := range pods.Items {
		svcName := fmt.Sprintf("%s-%s", podsvc.Name, pod.Name)
		expectedServices[svcName] = true

		matched, unmatched := matchPorts(podsvc.Spec.Ports, pod)

		// Create or update the Service
		svc := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: req.Namespace,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &svc, func() error {
			svc.Spec.Selector = pod.Labels
			svc.Spec.Ports = matched
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			return controllerutil.SetControllerReference(&podsvc, &svc, r.Scheme)
		})
		if err != nil {
			logger.Error(err, "failed to create/update service", "service", svcName)
			return ctrl.Result{}, err
		}

		statusEntries = append(statusEntries, operatorpocv1alpha1.PodSvcStatusEntry{
			PodName:        pod.Name,
			ServiceName:    svcName,
			MatchedPorts:   matched,
			UnmatchedPorts: unmatched,
		})
	}

	// Cleanup orphaned Services
	var svcList corev1.ServiceList
	if err := r.List(ctx, &svcList, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}

	for _, svc := range svcList.Items {
		if !isOwnedBy(&svc, &podsvc) {
			continue
		}
		if !expectedServices[svc.Name] {
			logger.Info("Deleting orphaned service", "service", svc.Name)
			if err := r.Delete(ctx, &svc); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Update status
	podsvc.Status.Entries = statusEntries
	if err := r.Status().Update(ctx, &podsvc); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func matchPorts(specPorts []corev1.ServicePort, pod corev1.Pod) ([]corev1.ServicePort, []operatorpocv1alpha1.UnmatchedPortStatus) {
	matched := []corev1.ServicePort{}
	unmatched := []operatorpocv1alpha1.UnmatchedPortStatus{}

	containerPorts := map[int32]bool{}
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			containerPorts[p.ContainerPort] = true
		}
	}

	for _, sp := range specPorts {
		if containerPorts[sp.Port] {
			matched = append(matched, sp)
		} else {
			unmatched = append(unmatched, operatorpocv1alpha1.UnmatchedPortStatus{
				Port:     sp.Port,
				Name:     sp.Name,
				Protocol: string(sp.Protocol),
				Reason:   "No matching container port",
			})
		}
	}

	return matched, unmatched
}

func isOwnedBy(svc *corev1.Service, owner *operatorpocv1alpha1.PodSvc) bool {
	for _, ref := range svc.OwnerReferences {
		if ref.UID == owner.UID {
			return true
		}
	}
	return false
}

func (r *PodSvcReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorpocv1alpha1.PodSvc{}).
		Owns(&corev1.Service{}).
		Named("podsvc").
		Complete(r)
}
