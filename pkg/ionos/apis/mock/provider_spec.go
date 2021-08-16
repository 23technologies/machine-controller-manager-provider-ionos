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

// Package mock provides all methods required to simulate a driver
package mock

import (
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis"
)

const (
	TestProviderSpec = "{\"datacenterID\":\"01234567-89ab-4def-0123-c56789abcdef\",\"cluster\":\"xyz\",\"zone\":\"de-fra\",\"cores\":1,\"memory\":1024,\"imageID\":\"15f67991-0f51-4efc-a8ad-ef1fb31a480c\",\"SSHKey\":\"23456789-abcd-4f01-23e5-6789abcdef01\"}"
	TestProviderSpecCluster = "xyz"
	TestProviderSpecDatacenterID = "01234567-89ab-4def-0123-c56789abcdef"
	TestProviderSpecNetworkID = "789abcde-f012-4456-789a-bcdef0123356"
	TestProviderSpecSSHKey = "23456789-abcd-4f01-23e5-6789abcdef01"
	TestProviderSpecImageID = "15f67991-0f51-4efc-a8ad-ef1fb31a480c"
	TestProviderSpecZone = "de-fra"
	TestInvalidProviderSpec = "{\"test\":\"invalid\"}"
)

// ManipulateProviderSpec changes given provider specification.
//
// PARAMETERS
// providerSpec *apis.ProviderSpec      Provider specification
// data         map[string]interface{} Members to change
func ManipulateProviderSpec(providerSpec *apis.ProviderSpec, data map[string]interface{}) *apis.ProviderSpec {
	for key, value := range data {
		manipulateStruct(&providerSpec, key, value)
	}

	return providerSpec
}

// NewProviderSpec generates a new provider specification for testing purposes.
func NewProviderSpec() *apis.ProviderSpec {
	return &apis.ProviderSpec{
		DatacenterID: TestProviderSpecDatacenterID,
		Cluster: TestProviderSpecCluster,
		Zone: TestProviderSpecZone,
		Cores: 1,
		Memory: 1024,
		ImageID: TestProviderSpecImageID,
		SSHKey: TestProviderSpecSSHKey,
		NetworkID: TestProviderSpecNetworkID,
	}
}
