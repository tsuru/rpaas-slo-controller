module github.com/tsuru/rpaas-slo-controller

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/tsuru/rpaas-operator v0.19.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
)

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190219214528-cbe11bdc6da8
