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

// Package apis is the main package for provider specific APIs
package apis

// ProviderSpec is the spec to be used while parsing the calls.
type ProviderSpec struct {
	DatacenterID string `json:"datacenterID,omitempty"`
	Cluster      string `json:"cluster"`
	Zone         string `json:"zone"`
	Cores        uint   `json:"cores"`
	Memory       uint   `json:"memory"`
	ImageID      string `json:"imageID"`
	SSHKey       string `json:"sshKey"`

	FloatingPoolID string      `json:"floatingPoolID,omitempty"`
	NetworkIDs     *NetworkIDs `json:"networkIDs,omitempty"`
	// Default: If you're creating the volume from a snapshot and don't specify
	// a volume size, the default is the snapshot size.
	VolumeSize     float32     `json:"volumeSize,omitempty"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type NetworkIDs struct {
	// WAN is the network ID for the public facing network interface.
	WAN string `json:"wan"`
	// Workers is the network ID of a worker subnet.
	Workers string `json:"workers,omitempty"`
}
