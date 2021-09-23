package controllers

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	sloKubernetes "github.com/globocom/slo-generator/kubernetes"
	"github.com/globocom/slo-generator/slo"
	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
	"github.com/tsuru/rpaas-slo-controller/definition"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	tsuruPoolLabel              = "tsuru.io/pool"
	rpaasTagsAnnotation         = "rpaas.extensions.tsuru.io/tags"
	rpaasTeamOwnerAnnotation    = "rpaas.extensions.tsuru.io/team-owner"
	rpaasInstanceNameAnnotation = "rpaas.extensions.tsuru.io/instance-name"
	rpaasServiceNameAnnotation  = "rpaas.extensions.tsuru.io/service-name"
)

var _ reconcile.Reconciler = &RpaasInstanceReconciler{}

// RpaasInstanceReconciler reconciles a RpaasInstance object
type RpaasInstanceReconciler struct {
	AlertLinkTemplate *template.Template
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
		if k8sErrors.IsNotFound(err) {
			err = r.reconcileRemovePrometheusRules(ctx, rpaasInstance)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	sloClass, _ := definition.SLOClass(rpaasInstance)
	if sloClass == nil {
		r.Log.Info("could not find a SLO classs",
			"name", req.Name,
			"namespace", req.Namespace,
		)
		err = r.reconcileRemovePrometheusRules(ctx, rpaasInstance)
		return ctrl.Result{}, err
	}

	sloAnnotations := map[string]string{}
	if r.AlertLinkTemplate != nil {
		var buf bytes.Buffer
		err = r.AlertLinkTemplate.Execute(&buf, rpaasInstance)
		if err != nil {
			r.Log.Error(err, "could not generate alert link",
				"name", req.Name,
				"namespace", req.Namespace,
			)
		}
		sloAnnotations["link"] = buf.String()
	}

	prometheusRules := sloKubernetes.GenerateManifests(sloKubernetes.Opts{
		SLO: slo.SLO{
			Name:  "tsuru." + req.Namespace + "." + req.Name,
			Class: sloClass.Name,
			Labels: map[string]string{
				"tsuru_team_owner": rpaasInstance.ObjectMeta.Annotations[rpaasTeamOwnerAnnotation],
			},
			Annotations: sloAnnotations,
			LatencyRecord: slo.ExprBlock{
				AlertMethod: "multi-window",
			},
			ErrorRateRecord: slo.ExprBlock{
				AlertMethod: "multi-window",
			},
		},
		Class: sloClass,
	})

	rulesNamespace := implicitNamespace(rpaasInstance.Namespace)
	instancePool := implicitPool(rpaasInstance.Namespace)

	existingPrometheusRules, err := r.existingPrometheusRules(ctx, rpaasInstance)
	if err != nil {
		r.Log.Error(err, "could not get PrometheusRules",
			"name", req.Name,
			"namespace", req.Namespace,
		)
		return ctrl.Result{}, err
	}

	existingPrometheusRulesSet := map[string]*monitoringv1.PrometheusRule{}
	for _, existingPrometheusRule := range existingPrometheusRules {
		existingPrometheusRulesSet[existingPrometheusRule.Name] = existingPrometheusRule
	}

	for _, prometheusRule := range prometheusRules {
		prometheusRule.Namespace = rulesNamespace

		if prometheusRule.Labels == nil {
			prometheusRule.Labels = map[string]string{}
		}
		if prometheusRule.Annotations == nil {
			prometheusRule.Annotations = map[string]string{}
		}
		if instancePool != "" {
			prometheusRule.Labels[tsuruPoolLabel] = instancePool
		}
		prometheusRule.Labels[rpaasTeamOwnerAnnotation] = rpaasInstance.Labels[rpaasTeamOwnerAnnotation]
		prometheusRule.Labels[rpaasInstanceNameAnnotation] = rpaasInstance.Labels[rpaasInstanceNameAnnotation]
		prometheusRule.Labels[rpaasServiceNameAnnotation] = rpaasInstance.Labels[rpaasServiceNameAnnotation]

		if rulesNamespace == rpaasInstance.Namespace {
			prometheusRule.OwnerReferences = append(prometheusRule.OwnerReferences, *metav1.NewControllerRef(rpaasInstance, schema.GroupVersionKind{
				Group:   v1alpha1.GroupVersion.Group,
				Version: v1alpha1.GroupVersion.Version,
				Kind:    "RpaasInstance",
			}))
		}

		if existingPrometheusRulesSet[prometheusRule.Name] == nil {
			err := r.Client.Create(ctx, &prometheusRule)
			if err != nil {
				r.Log.Error(err, "could not create PrometheusRule",
					"name", prometheusRule.Name,
					"namespace", prometheusRule.Namespace,
				)
				return ctrl.Result{}, err
			}

			r.Log.Info("created PrometheusRule",
				"name", prometheusRule.Name,
				"namespace", prometheusRule.Namespace)
		} else {
			prometheusRule.ResourceVersion = existingPrometheusRulesSet[prometheusRule.Name].ResourceVersion
			delete(existingPrometheusRulesSet, prometheusRule.Name)
			err := r.Client.Update(ctx, &prometheusRule)
			if err != nil {
				r.Log.Error(err, "could not update PrometheusRule",
					"name", prometheusRule.Name,
					"namespace", prometheusRule.Namespace,
				)
				return ctrl.Result{}, err
			}

			r.Log.Info("updated PrometheusRule",
				"name", prometheusRule.Name,
				"namespace", prometheusRule.Namespace)
		}
	}

	for _, existingPrometheusRule := range existingPrometheusRulesSet {
		err = r.Client.Delete(ctx, existingPrometheusRule)
		if err != nil {
			r.Log.Error(err, "could not remove unused PrometheusRule",
				"name", existingPrometheusRule.Name,
				"namespace", existingPrometheusRule.Namespace,
			)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RpaasInstanceReconciler) reconcileRemovePrometheusRules(ctx context.Context, rpaasInstance *v1alpha1.RpaasInstance) error {
	existingPrometheusRules, err := r.existingPrometheusRules(ctx, rpaasInstance)
	if err != nil {
		r.Log.Error(err, "could not get PrometheusRules",
			"name", rpaasInstance.Name,
			"namespace", rpaasInstance.Namespace,
		)
		return err
	}

	for _, rule := range existingPrometheusRules {
		err = r.Client.Delete(ctx, rule)

		r.Log.Error(err, "could not remove unused PrometheusRule",
			"name", rule.Name,
			"namespace", rule.Namespace,
		)

		return err
	}

	return nil
}

func (r *RpaasInstanceReconciler) existingPrometheusRules(ctx context.Context, rpaasInstance *v1alpha1.RpaasInstance) ([]*monitoringv1.PrometheusRule, error) {
	rulesNamespace := implicitNamespace(rpaasInstance.Namespace)
	list := monitoringv1.PrometheusRuleList{}
	err := r.Client.List(ctx, &list, &client.ListOptions{
		Namespace: rulesNamespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			rpaasInstanceNameAnnotation: rpaasInstance.Labels[rpaasInstanceNameAnnotation],
			rpaasServiceNameAnnotation:  rpaasInstance.Labels[rpaasServiceNameAnnotation],
		}),
	})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func implicitNamespace(ns string) string {
	pool := implicitPool(ns)
	if pool != "" {
		return "tsuru-" + pool
	}

	return ns
}

func implicitPool(ns string) string {
	if strings.HasPrefix(ns, "rpaasv2-be-") || strings.HasPrefix(ns, "rpaasv2-fe-") {
		return ns[11:]
	}

	return ""
}

func (r *RpaasInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RpaasInstance{}).
		Complete(r)
}
