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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	jsonImageData = `
{
	"id": "15f67991-0f51-4efc-a8ad-ef1fb31a480c",
	"type": "image",
	"href": "",
	"metadata": {
		"etag": "45480eb3fbfc31f1d916c1eaa4abdcc3",
		"createdDate": "2015-12-04T14:34:09.809Z",
		"createdBy": "user@example.com",
		"createdByUserId": "user@example.com",
		"lastModifiedDate": "2015-12-04T14:34:09.809Z",
		"lastModifiedBy": "user@example.com",
		"lastModifiedByUserId": "63cef532-26fe-4a64-a4e0-de7c8a506c90",
		"state": "AVAILABLE"
	},
	"properties": {
		"name": "Ubuntu 20.04",
		"description": "Proudly copied from the IONOS Cloud API documentation",
		"location": "us/las",
		"size": 100,
		"cpuHotPlug": true,
		"cpuHotUnplug": true,
		"ramHotPlug": true,
		"ramHotUnplug": true,
		"nicHotPlug": true,
		"nicHotUnplug": true,
		"discVirtioHotPlug": true,
		"discVirtioHotUnplug": true,
		"discScsiHotPlug": true,
		"discScsiHotUnplug": true,
		"licenceType": "LINUX",
		"imageType": "HDD",
		"public": true,
		"imageAliases": [],
		"cloudInit": "V1"
	}
}
	`
	jsonNicTemplate = `{
		"id": %q,
		"type": "nic",
		"href": "",
		"metadata": {
			"etag": "45480eb3fbfc31f1d916c1eaa4abdcc3",
			"createdDate": "2015-12-04T14:34:09.809Z",
			"createdBy": "user@example.com",
			"createdByUserId": "user@example.com",
			"lastModifiedDate": "2015-12-04T14:34:09.809Z",
			"lastModifiedBy": "user@example.com",
			"lastModifiedByUserId": "63cef532-26fe-4a64-a4e0-de7c8a506c90",
			"state": "AVAILABLE"
		},
		"properties": {
			"name": "NIC",
			"mac": "00:11:22:33:44:55",
			"ips": [],
			"dhcp": true,
			"lan": 2,
			"firewallActive": false,
			"firewallType": "INGRESS",
			"deviceNumber": 1,
			"pciSlot": 1
		},
		"entities": {
			"flowlogs": {},
			"firewallrules": {}
		}
	}
	`
	jsonServerDataTemplate = `
{
	"id": %q,
	"type": "server",
	"href": "",
	"metadata": {
		"etag": "45480eb3fbfc31f1d916c1eaa4abdcc3",
		"createdDate": "2015-12-04T14:34:09.809Z",
		"createdBy": "user@example.com",
		"createdByUserId": "user@example.com",
		"lastModifiedDate": "2015-12-04T14:34:09.809Z",
		"lastModifiedBy": "user@example.com",
		"lastModifiedByUserId": "63cef532-26fe-4a64-a4e0-de7c8a506c90",
		"state": %q
	},
	"properties": {
		"templateUuid": "15f67991-0f51-4efc-a8ad-ef1fb31a480c",
		"name": %q,
		"cores": 4,
		"ram": 4096,
		"availabilityZone": "AUTO",
		"vmState": %q,
		"bootCdrom": {
			"id": "",
			"type": "resource",
			"href": ""
		},
		"bootVolume": {
			"id": %q,
			"type": "resource",
			"href": ""
		},
		"cpuFamily": "AMD_OPTERON",
		"type": "CUBE"
	},
	"entities": {
		"cdroms": {},
		"volumes": {
			"id": "15f67991-0f51-4efc-a8ad-ef1fb31a480c",
			"type": "collection",
			"href": "",
			"items": [%s],
			"offset": 0,
			"limit": 1000,
			"_links": {}
		},
		"nics": {}
	}
}
	`
	jsonVolumeTemplate = `
{
	"id": %q,
	"type": "volume",
	"href": "",
	"metadata": {
		"etag": "45480eb3fbfc31f1d916c1eaa4abdcc3",
		"createdDate": "2015-12-04T14:34:09.809Z",
		"createdBy": "user@example.com",
		"createdByUserId": "user@example.com",
		"lastModifiedDate": "2015-12-04T14:34:09.809Z",
		"lastModifiedBy": "user@example.com",
		"lastModifiedByUserId": "63cef532-26fe-4a64-a4e0-de7c8a506c90",
		"state": "AVAILABLE"
	},
	"properties": {
		"name": "My resource",
		"type": "HDD",
		"size": 100,
		"availabilityZone": "AUTO",
		"image": "15f67991-0f51-4efc-a8ad-ef1fb31a480c",
		"imagePassword": null,
		"sshKeys": [],
		"bus": "VIRTIO",
		"licenceType": "LINUX",
		"cpuHotPlug": true,
		"ramHotPlug": true,
		"nicHotPlug": true,
		"nicHotUnplug": true,
		"discVirtioHotPlug": true,
		"discVirtioHotUnplug": true,
		"deviceNumber": 3,
		"pciSlot": 7,
		"userData": ""
	}
}
	`
	TestNamespace = "test"
	TestServerNameTemplate = "machine-%s"
	TestServerID = "6789abcd-ef01-4345-6789-abcdef012325"
	TestServerNicID = "23456789-abcd-4f01-23e5-6789abcdef01"
	TestServerVolumeID = "3456789a-bcde-4012-3f56-789abcdef012"
)

// handleLabelEndpointRequest provides support for a generic "/labels" endpoint.
//
// PARAMETERS
// req *http.Request       Request instance
// res http.ResponseWriter Response instance
func handleLabelEndpointRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json; charset=utf-8")

	if (strings.ToLower(req.Method) == "get") {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(fmt.Sprintf(`
{
"id": %q,
"type": "collection",
"href": "",
"items": [ ]
}
		`, uuid.NewString())))
	} else if (strings.ToLower(req.Method) == "post") {
		res.WriteHeader(http.StatusCreated)

		jsonData := make([]byte, req.ContentLength)
		req.Body.Read(jsonData)

		var data map[string]interface{}

		jsonErr := json.Unmarshal(jsonData, &data)
		if jsonErr != nil {
			panic(jsonErr)
		}

		res.Write([]byte(jsonData))
	} else {
		panic("Unsupported HTTP method call")
	}
}

// ManipulateMachine changes given machine data.
//
// PARAMETERS
// machine *v1alpha1.Machine      Machine data
// data    map[string]interface{} Members to change
func ManipulateMachine(machine *v1alpha1.Machine, data map[string]interface{}) *v1alpha1.Machine {
	for key, value := range data {
		if (strings.Index(key, "ObjectMeta") == 0) {
			manipulateStruct(&machine.ObjectMeta, key[11:], value)
		} else if (strings.Index(key, "Spec") == 0) {
			manipulateStruct(&machine.Spec, key[5:], value)
		} else if (strings.Index(key, "Status") == 0) {
			manipulateStruct(&machine.Status, key[7:], value)
		} else if (strings.Index(key, "TypeMeta") == 0) {
			manipulateStruct(&machine.TypeMeta, key[9:], value)
		} else {
			manipulateStruct(&machine, key, value)
		}
	}

	return machine
}

// NewMachine generates new v1alpha1 machine data for testing purposes.
//
// PARAMETERS
// ServerID string Server UUID to use for machine specification
func NewMachine(serverID string) *v1alpha1.Machine {
	machine := &v1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "machine.sapcloud.io",
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(TestServerNameTemplate, serverID),
			Namespace: TestNamespace,
		},
	}

	// Don't initialize providerID and node if ServerID == ""
	if "" != serverID {
		machine.Spec = v1alpha1.MachineSpec{
			ProviderID: fmt.Sprintf("ionos:///%s/%s", TestProviderSpecDatacenterID, serverID),
		}
		machine.Status = v1alpha1.MachineStatus{
			Node: fmt.Sprintf("ip-%s", serverID),
		}
	}

	return machine
}

// NewMachineClass generates new v1alpha1 machine class data for testing purposes.
func NewMachineClass() *v1alpha1.MachineClass {
	return NewMachineClassWithProviderSpec([]byte(TestProviderSpec))
}

// NewMachineClassWithProviderSpec generates new v1alpha1 machine class data based on the given provider specification for testing purposes.
//
// PARAMETERS
// providerSpec []byte ProviderSpec to use
func NewMachineClassWithProviderSpec(providerSpec []byte) *v1alpha1.MachineClass {
	return &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: providerSpec,
		},
	}
}

// newJsonServerData generates a JSON server data object for testing purposes.
//
// PARAMETERS
// serverID    string Server ID to use
// serverState string Server state to use
func newJsonServerData(serverID string, serverState string) string {
	serverBootState := "RUNNING"

	if "AVAILABLE" != serverState {
		serverBootState = "SHUTOFF"
	}

	jsonVolumeData := fmt.Sprintf(jsonVolumeTemplate, TestServerVolumeID)
	testServerName := fmt.Sprintf(TestServerNameTemplate, serverID)
	return fmt.Sprintf(jsonServerDataTemplate, serverID, serverState, testServerName, serverBootState, TestServerVolumeID, jsonVolumeData)
}

// SetupImagesEndpointOnMux configures a "/images" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupImagesEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc(apiBasePath + "/images/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.HasSuffix(req.URL.Path, fmt.Sprintf("/%s", TestProviderSpecImageID))) {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(jsonImageData))
		} else {
			panic("Unsupported image ID requested")
		}
	})
}

// SetupServersEndpointOnMux configures a "/datacenters/<id>/servers" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupServersEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("%s/datacenters/%s/servers", apiBasePath, TestProviderSpecDatacenterID), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "get") {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(fmt.Sprintf(`
{
	"id": %q,
	"type": "collection",
	"href": "",
	"items": [
		%s
	]
}
			`, uuid.NewString(), newJsonServerData(TestServerID, "AVAILABLE"))))

		} else if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusAccepted)

			jsonData := make([]byte, req.ContentLength)
			req.Body.Read(jsonData)

			var data map[string]interface{}

			jsonErr := json.Unmarshal(jsonData, &data)
			if jsonErr != nil {
				panic(jsonErr)
			}

			res.Write([]byte(newJsonServerData(TestServerID, "BUSY")))
		} else {
			panic("Unsupported HTTP method call")
		}
	})
}

// SetupTestServerEndpointOnMux configures "/datacenters/<dcid>/servers/<sid>/*" endpoints on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupTestServerEndpointOnMux(mux *http.ServeMux) {
	baseURL := fmt.Sprintf("%s/datacenters/%s/servers/%s", apiBasePath, TestProviderSpecDatacenterID, TestServerID)

	mux.HandleFunc(baseURL, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "delete") {
			res.WriteHeader(http.StatusAccepted)
		} else if (strings.ToLower(req.Method) == "get") {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(newJsonServerData(TestServerID, "AVAILABLE")))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/labels", baseURL), handleLabelEndpointRequest)

	mux.HandleFunc(fmt.Sprintf("%s/nics", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusAccepted)
			res.Write([]byte(fmt.Sprintf(jsonNicTemplate, TestServerNicID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/nics/%s", baseURL, TestServerNicID), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "get") {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(fmt.Sprintf(jsonNicTemplate, TestServerNicID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/start", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusAccepted)
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/stop", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusAccepted)
		} else {
			panic("Unsupported HTTP method call")
		}
	})
}

// SetupTestVolumeEndpointOnMux configures "/datacenters/<dcid>/volumes/<vid>" endpoints on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupTestVolumeEndpointOnMux(mux *http.ServeMux) {
	baseURL := fmt.Sprintf("%s/datacenters/%s/volumes/%s", apiBasePath, TestProviderSpecDatacenterID, TestServerVolumeID)

	mux.HandleFunc(baseURL, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "delete") {
			res.WriteHeader(http.StatusAccepted)
		} else if (strings.ToLower(req.Method) == "get") {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(fmt.Sprintf(jsonVolumeTemplate, TestServerVolumeID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/labels", baseURL), handleLabelEndpointRequest)
}

// SetupVolumesEndpointOnMux configures a "/datacenters/<id>/volumes" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupVolumesEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("%s/datacenters/%s/volumes", apiBasePath, TestProviderSpecDatacenterID), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if (strings.ToLower(req.Method) == "post") {
			res.WriteHeader(http.StatusAccepted)

			jsonData := make([]byte, req.ContentLength)
			req.Body.Read(jsonData)

			var data map[string]interface{}

			jsonErr := json.Unmarshal(jsonData, &data)
			if jsonErr != nil {
				panic(jsonErr)
			}

			res.Write([]byte(fmt.Sprintf(jsonVolumeTemplate, TestServerVolumeID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})
}
