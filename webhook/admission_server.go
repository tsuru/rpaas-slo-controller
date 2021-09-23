package webhook

import (
	"log"
	"net/http"

	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
)

func Run() {
	logger := kwhlog.Noop
	webhook, err := NewRpaasInstancesWebhook(logger)
	if err != nil {
		log.Fatal("Could not instantiate webhook", err)
		return
	}
	webhookHandler, err := kwhhttp.HandlerFor(kwhhttp.HandlerConfig{Webhook: webhook, Logger: logger})
	if err != nil {
		log.Fatal("Could not instantiate webhook handler", err)
		return
	}
	log.Fatal(http.ListenAndServe(":8888", webhookHandler))
}
