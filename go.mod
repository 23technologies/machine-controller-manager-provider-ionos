module github.com/23technologies/machine-controller-manager-provider-ionos

go 1.16

require (
	github.com/23technologies/ionos-api-wrapper v0.3.0
	github.com/gardener/machine-controller-manager v0.43.0
	github.com/google/uuid v1.2.0
	github.com/ionos-cloud/sdk-go/v6 v6.0.1
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.17.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/component-base v0.20.6
	k8s.io/klog/v2 v2.4.0
)

replace (
	github.com/gardener/machine-controller-manager => github.com/gardener/machine-controller-manager v0.43.0
	k8s.io/api => k8s.io/api v0.20.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.6
	k8s.io/apiserver => k8s.io/apiserver v0.20.6
	k8s.io/client-go => k8s.io/client-go v0.20.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.6
	k8s.io/code-generator => k8s.io/code-generator v0.20.6
)
