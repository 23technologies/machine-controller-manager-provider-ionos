# Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BINARY_PATH        := bin/
COVERPROFILE       := test/output/coverprofile.out
PROVIDER_NAME      := ionos
REPO_ROOT          := $(shell dirname $(realpath $(lastword ${MAKEFILE_LIST})))
VERSION            := $(shell cat "${REPO_ROOT}/VERSION")
LD_FLAGS           := "-w $(shell hack/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/VERSION)"
CONTROL_NAMESPACE  := shoot--foobar--ionos
CONTROL_KUBECONFIG := dev/control-kubeconfig.yaml
TARGET_KUBECONFIG  := dev/target-kubeconfig.yaml

#########################################
# Rules for starting machine-controller locally
#########################################

.PHONY: start
start:
	@GO111MODULE=on go run \
			-mod=vendor \
		    -ldflags ${LD_FLAGS} \
			cmd/machine-controller-manager-provider-ionos/main.go \
			--control-kubeconfig=$(CONTROL_KUBECONFIG) \
			--target-kubeconfig=$(TARGET_KUBECONFIG) \
			--namespace=$(CONTROL_NAMESPACE) \
			--machine-creation-timeout=20m \
			--machine-drain-timeout=5m \
			--machine-health-timeout=10m \
			--machine-pv-detach-timeout=2m \
			--machine-safety-apiserver-statuscheck-timeout=30s \
			--machine-safety-apiserver-statuscheck-period=1m \
			--machine-safety-orphan-vms-period=30m \
			--v=5

#########################################
# Rules for re-vendoring
#########################################

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod tidy -compat=1.17
	@GO111MODULE=on go mod vendor

#########################################
# Rules for testing
#########################################

.PHONY: test
test:
	@hack/test.sh

.PHONY: test-cov
test-cov:
	@hack/test.sh --coverage

.PHONY: test-clean
test-clean:
	@hack/test.sh --clean --coverage

#########################################
# Rules for build/release
#########################################

.PHONY: build-local
build-local:
	@env LD_FLAGS=${LD_FLAGS} LOCAL_BUILD=1 hack/build.sh

.PHONY: build
build:
	@env LD_FLAGS=${LD_FLAGS} hack/build.sh

.PHONY: clean
clean:
	@rm -rf bin/
