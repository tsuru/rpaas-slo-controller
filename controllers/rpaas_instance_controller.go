package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const rpaasTagsAnnotation = "rpaas.extensions.tsuru.io/tags"

var _ reconcile.Reconciler = &RpaasInstanceReconciler{}

// RpaasInstanceReconciler reconciles a RpaasInstance object
type RpaasInstanceReconciler struct {
	client.Client
	Log logr.Logger
}

func (r *RpaasInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rpaasInstance := &v1alpha1.RpaasInstance{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, rpaasInstance)
	if err != nil {
		return ctrl.Result{}, err
	}
	tagsRaw := rpaasInstance.ObjectMeta.Annotations[rpaasTagsAnnotation]
	var tags []string
	if tagsRaw != "" {
		tags = strings.Split(tagsRaw, ",")
	}
	sloTags := extractTagValues([]string{"slo:", "SLO:", "slo=", "SLO:"}, tags)
	if len(sloTags) == 0 {
		return ctrl.Result{}, nil
	}
	fmt.Println("TODO reconcile: ", rpaasInstance.Name, "SLO class", sloTags[0])
	return ctrl.Result{}, nil
}

func (r *RpaasInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RpaasInstance{}).
		Complete(r)
}

func extractTagValues(prefixes, tags []string) []string {
	for _, t := range tags {
		for _, p := range prefixes {
			if !strings.HasPrefix(t, p) {
				continue
			}

			separator := string(p[len(p)-1])
			parts := strings.SplitN(t, separator, 2)
			if len(parts) == 1 {
				return nil
			}

			return parts[1:]
		}
	}

	return nil
}
