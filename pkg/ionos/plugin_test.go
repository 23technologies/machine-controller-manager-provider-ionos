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

// Package ionos contains the IONOS provider specific implementations to manage machines
package ionos

import (
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin", func() {
	Describe("#NewIonosProvider", func() {
		It("should correctly create a new provider object", func() {
			provider := NewIonosProvider(&spi.PluginSPIImpl{})
			_, ok := provider.(driver.Driver)
			Expect(ok).To(BeTrue())
		})
	})
})
