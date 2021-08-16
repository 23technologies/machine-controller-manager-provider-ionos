module github.com/23technologies/machine-controller-manager-provider-ionos

go 1.15

require (
	github.com/gardener/machine-controller-manager v0.39.0
	github.com/google/uuid v1.2.0
	github.com/ionos-cloud/sdk-go/v5 v5.1.4
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/component-base v0.20.6
	k8s.io/klog v1.0.0
)

replace (
	github.com/gardener/machine-controller-manager => github.com/gardener/machine-controller-manager v0.39.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.2.0
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f // kubernetes-1.16.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
