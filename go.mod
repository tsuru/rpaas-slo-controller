module github.com/tsuru/rpaas-slo-controller

go 1.16

require (
	github.com/elastic/gosigar v0.9.0 // indirect
	github.com/globocom/slo-generator v0.2.2-0.20210922120954-fe6dee4f2f6e
	github.com/go-logr/logr v0.4.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0 // indirect
	github.com/tsuru/rpaas-operator v0.19.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.1
)

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190219214528-cbe11bdc6da8
