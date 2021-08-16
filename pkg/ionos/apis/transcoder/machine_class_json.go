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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis"
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis/validation"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// DecodeProviderSpecFromMachineClass decodes the given MachineClass to receive the ProviderSpec.
//
// PARAMETERS
// machineClass *v1alpha1.MachineClass MachineClass backing the machine object
// secret       *corev1.Secret         Kubernetes secret that contains any sensitive data/credentials
func DecodeProviderSpecFromMachineClass(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (*apis.ProviderSpec, error) {
	// Extract providerSpec
	var providerSpec *apis.ProviderSpec

	if machineClass == nil {
		return nil, errors.New("MachineClass provided is nil")
	}

	jsonErr := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if jsonErr != nil {
		return nil, fmt.Errorf("Failed to parse JSON data provided as ProviderSpec: %v", jsonErr)
	}

	// Validate the Spec
	validationErr := validation.ValidateIonosProviderSpec(providerSpec, secret)
	if validationErr != nil {
		return nil, fmt.Errorf("Error while validating ProviderSpec %v", validationErr)
	}

	return providerSpec, nil
}
