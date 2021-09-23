package webhook

import (
	"context"

	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhwebhook "github.com/slok/kubewebhook/v2/pkg/webhook"
	kwhvalidating "github.com/slok/kubewebhook/v2/pkg/webhook/validating"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
	"github.com/tsuru/rpaas-slo-controller/definition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type rpaasV1Validator struct{}

func (d *rpaasV1Validator) Validate(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhvalidating.ValidatorResult, error) {
	rpaasInstance, ok := obj.(*v1alpha1.RpaasInstance)
	if !ok {
		// If not a rpaasInstance just continue the validation chain(if there is one) and don't do nothing.
		return &kwhvalidating.ValidatorResult{Valid: true}, nil
	}

	_, err := definition.SLOClass(rpaasInstance)
	if err != nil {
		return &kwhvalidating.ValidatorResult{
			Valid:   false,
			Message: "Invalid SLO class",
		}, nil
	}

	return &kwhvalidating.ValidatorResult{Valid: true}, nil
}

func NewRpaasInstancesWebhook(logger kwhlog.Logger) (kwhwebhook.Webhook, error) {
	return kwhvalidating.NewWebhook(
		kwhvalidating.WebhookConfig{
			ID:        "webhook-rpaasInstanceValidator",
			Obj:       &v1alpha1.RpaasInstance{},
			Validator: &rpaasV1Validator{},
			Logger:    logger,
		})
}
