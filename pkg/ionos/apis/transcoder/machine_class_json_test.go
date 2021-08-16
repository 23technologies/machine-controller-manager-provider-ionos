/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package transcoder is used for API related object transformations
package transcoder

import (
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Transcoder", func() {
	machineClass := mock.NewMachineClass()
	unsupportedMachineClass := mock.NewMachineClassWithProviderSpec([]byte(mock.TestInvalidProviderSpec))

	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"userData": []byte("dummy-user-data"),
		},
	}

	Describe("#DecodeProviderSpecFromMachineClass", func() {
		It("should correctly parse and return a ProviderSpec object", func() {
			providerSpec, err := DecodeProviderSpecFromMachineClass(machineClass, providerSecret)

			Expect(err).NotTo(HaveOccurred())
			Expect(providerSpec.SSHKey).To(Equal(mock.TestProviderSpecSSHKey))
		})

		It("should fail if an invalid machineClass is provided", func() {
			_, err := DecodeProviderSpecFromMachineClass(unsupportedMachineClass, providerSecret)

			Expect(err).To(HaveOccurred())
		})
	})
})
