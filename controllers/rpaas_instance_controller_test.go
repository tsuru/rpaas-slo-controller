package controllers

import (
	"context"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()

	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
)

func init() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

func TestReconcileRpaasInstanceSLOCritical(t *testing.T) {
	ctx := context.TODO()
	rpaasInstance1 := &v1alpha1.RpaasInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "instance1",
			Labels: map[string]string{
				rpaasTeamOwnerAnnotation:    "my-team",
				rpaasInstanceNameAnnotation: "instance1",
				rpaasServiceNameAnnotation:  "rpaasv2",
			},
			Annotations: map[string]string{
				rpaasTagsAnnotation: "slo:critical",
			},
		},
	}

	k8sClient := fake.NewFakeClientWithScheme(scheme, rpaasInstance1)
	reconciler := &RpaasInstanceReconciler{
		Client: k8sClient,
		Log:    ctrl.Log,
	}

	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "instance1",
		},
	})
	assert.NoError(t, err)

	prometheusRule := monitoringv1.PrometheusRule{}
	err = k8sClient.Get(ctx, client.ObjectKey{
		Namespace: rpaasInstance1.Namespace,
		Name:      "slos-alerts-tsuru.default.instance1",
	}, &prometheusRule)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		rpaasTeamOwnerAnnotation:    "my-team",
		rpaasInstanceNameAnnotation: "instance1",
		rpaasServiceNameAnnotation:  "rpaasv2",
	}, prometheusRule.Labels)

	_true := true
	assert.Equal(t, []metav1.OwnerReference{
		{
			APIVersion:         "extensions.tsuru.io/v1alpha1",
			Kind:               "RpaasInstance",
			Name:               "instance1",
			Controller:         &_true,
			BlockOwnerDeletion: &_true,
		},
	}, prometheusRule.OwnerReferences)

	require.Len(t, prometheusRule.Spec.Groups, 1)
	assert.Len(t, prometheusRule.Spec.Groups[0].Rules, 4)
}
