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
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis"
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis/transcoder"
	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/ensurer"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	ionossdk "github.com/ionos-cloud/sdk-go/v5"
	"k8s.io/klog"
)

// Constant ionosVolumeType is the volume type
const ionosVolumeType = "SSD"

// CreateMachine handles a machine creation request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The create request for VM creation
func (p *MachineProvider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	var (
		machine      = req.Machine
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", machine.Name)

	if "" != machine.Spec.ProviderID {
		return nil, status.Error(codes.InvalidArgument, "Machine creation with existing provider ID is not supported")
	}

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userDataBytes, ok := secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, "userData doesn't exist")
	}

	client := apis.GetClientForUser(string(secret.Data["user"]), string(secret.Data["password"]))
	region := apis.GetRegionFromZone(providerSpec.Zone)

	image, _, err := client.ImageApi.ImagesFindById(ctx, providerSpec.ImageID).Depth(1).Execute()
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	} else if (!image.Properties.HasCloudInit() || "NONE" == *image.Properties.CloudInit) {
		return nil, status.Error(codes.InvalidArgument, "imageID given doesn't belong to a cloud-init enabled image")
	}

	sshKeys := []string{fmt.Sprintf("%s\n", providerSpec.SSHKey)}
	userDataBase64Enc := base64.StdEncoding.EncodeToString(userDataBytes)
	volumeSize := providerSpec.VolumeSize
	volumeType := ionosVolumeType

	if 0 == volumeSize {
		volumeSize = *image.Properties.Size
	} else {
		volumeSize = float32(math.Max(math.Ceil(float64(volumeSize) / 1048576), float64(*image.Properties.Size)))
	}

	volumeProperties := ionossdk.VolumeProperties{
		Type: &volumeType,
		Name: &machine.Name,
		Size: &volumeSize,
		Image: &providerSpec.ImageID,
		SshKeys: &sshKeys,
		UserData: &userDataBase64Enc,
	}

	volumeApiCreateRequest := client.VolumeApi.DatacentersVolumesPost(ctx, providerSpec.DatacenterID).Depth(0)
	volume, _, err := volumeApiCreateRequest.Volume(ionossdk.Volume{Properties: &volumeProperties}).Execute()
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	clusterValue := hex.EncodeToString([]byte(providerSpec.Cluster))
	volumeID := *volume.Id

	volume, err = apis.WaitForVolumeModificationsAndGetResult(ctx, client, providerSpec.DatacenterID, volumeID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = apis.AddLabelToVolume(ctx, client, providerSpec.DatacenterID, volumeID, "cluster", clusterValue)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var floatingPoolIPLANID string

	if "" != providerSpec.FloatingPoolIPID {
		name := fmt.Sprintf("%s-floating-pool-ip", machine.Name)

		floatingPoolIPLANID, err = ensurer.EnsureFloatingPoolIPBlockLANIsCreated(ctx, client, providerSpec.DatacenterID, providerSpec.FloatingPoolIPID, name)
		if nil != err {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	cores := int32(providerSpec.Cores)
	memory := int32(providerSpec.Memory)

	serverEntities := ionossdk.ServerEntities{
		Volumes: &ionossdk.AttachedVolumes{Items: &[]ionossdk.Volume{ionossdk.Volume{Id: &volumeID}}},
	}

	serverProperties := ionossdk.ServerProperties{
		Name:       &machine.Name,
		Cores:      &cores,
		Ram:        &memory,
		BootVolume: &ionossdk.ResourceReference{Id: &volumeID},
	}

	serverApiCreateRequest := client.ServerApi.DatacentersServersPost(ctx, providerSpec.DatacenterID).Depth(0)
	server, _, err := serverApiCreateRequest.Server(ionossdk.Server{Entities: &serverEntities, Properties: &serverProperties}).Execute()
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	serverID := *server.Id

	err = apis.WaitForServerModifications(ctx, client, providerSpec.DatacenterID, serverID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, _, err = client.ServerApi.DatacentersServersStopPost(ctx, providerSpec.DatacenterID, serverID).Execute()
	if nil != err {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	err = apis.WaitForServerModifications(ctx, client, providerSpec.DatacenterID, serverID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = apis.AddLabelToServer(ctx, client, providerSpec.DatacenterID, serverID, "cluster", clusterValue)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = apis.AddLabelToServer(ctx, client, providerSpec.DatacenterID, serverID, "role", "node")
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = apis.AddLabelToServer(ctx, client, providerSpec.DatacenterID, serverID, "region", hex.EncodeToString([]byte(region)))
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = apis.AddLabelToServer(ctx, client, providerSpec.DatacenterID, serverID, "zone", hex.EncodeToString([]byte(providerSpec.Zone)))
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if "" != floatingPoolIPLANID {
		err = ensurer.EnsureLANIsAttachedToServer(ctx, client, providerSpec.DatacenterID, serverID, floatingPoolIPLANID)
		if nil != err {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if "" != providerSpec.NetworkID {
		err = ensurer.EnsureLANIsAttachedToServer(ctx, client, providerSpec.DatacenterID, serverID, providerSpec.NetworkID)
		if nil != err {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = apis.WaitForServerModifications(ctx, client, providerSpec.DatacenterID, serverID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, _, err = client.ServerApi.DatacentersServersStartPost(ctx, providerSpec.DatacenterID, serverID).Execute()
	if nil != err {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	server, err = apis.WaitForServerModificationsAndGetResult(ctx, client, providerSpec.DatacenterID, serverID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &driver.CreateMachineResponse{
		ProviderID: transcoder.EncodeProviderID(providerSpec.DatacenterID, *server.Id),
		NodeName:   *server.Properties.Name,
	}

	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The delete request for VM deletion
func (p *MachineProvider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	var (
		machine      = req.Machine
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", machine.Name)

	serverID, err := transcoder.DecodeServerIDFromProviderID(machine.Spec.ProviderID)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := apis.GetClientForUser(string(secret.Data["user"]), string(secret.Data["password"]))

	_, httpResponse, err := client.ServerApi.DatacentersServersStopPost(ctx, providerSpec.DatacenterID, serverID).Execute()
	if nil != err {
		if 404 == httpResponse.StatusCode {
			klog.V(3).Infof("VM %s (%s) does not exist", machine.Name, serverID)
			return &driver.DeleteMachineResponse{}, nil
		} else {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}

	err = apis.WaitForServerModifications(ctx, client, providerSpec.DatacenterID, serverID)
	if nil != err {
		return nil, status.Error(codes.Internal, err.Error())
	}

	server, _, err := client.ServerApi.DatacentersServersFindById(ctx, providerSpec.DatacenterID, serverID).Depth(3).Execute()
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	if "" != providerSpec.FloatingPoolIPID {
		name := fmt.Sprintf("%s-floating-pool-ip", machine.Name)

		err := ensurer.EnsureFloatingPoolIPBlockLANIsDeleted(ctx, client, providerSpec.DatacenterID, name)
		if nil != err {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}

	for _, volume := range *server.Entities.Volumes.Items {
		_, _, err := client.VolumeApi.DatacentersVolumesDelete(ctx, providerSpec.DatacenterID, *volume.Id).Depth(0).Execute()
		if nil != err {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}

	_, _, err = client.ServerApi.DatacentersServersDelete(ctx, providerSpec.DatacenterID, serverID).Depth(0).Execute()
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The get request for VM info
func (p *MachineProvider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	var (
		machine      = req.Machine
		secret       = req.Secret
	)

	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been received for %q", machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", machine.Name)

	// Handle case where machine lookup occurs with empty provider ID
	if machine.Spec.ProviderID == "" {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Provider ID for machine %q is not defined", machine.Name))
	}

	serverData, err := transcoder.DecodeServerDataFromProviderID(machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := apis.GetClientForUser(string(secret.Data["user"]), string(secret.Data["password"]))

	server, _, err := client.ServerApi.DatacentersServersFindById(ctx, serverData.DatacenterID, serverData.ID).Depth(1).Execute()
	if nil != err {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if "INACTIVE" == *server.Metadata.State {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("VM %s (%s) does not exist", *server.Properties.Name, serverData.ID))
	}

	return &driver.GetMachineStatusResponse{ ProviderID: machine.Spec.ProviderID, NodeName: *server.Properties.Name }, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request object to get a list of VMs belonging to a machineClass
func (p *MachineProvider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	var (
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	providerSpec, err := transcoder.DecodeProviderSpecFromMachineClass(machineClass, secret)
	if nil != err {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been received for %q", machineClass.Name)
	defer klog.V(2).Infof("List machines request has been processed for %q", machineClass.Name)

	client := apis.GetClientForUser(string(secret.Data["user"]), string(secret.Data["password"]))

	servers, _, err := client.ServerApi.DatacentersServersGet(ctx, providerSpec.DatacenterID).Depth(1).Execute()
	if nil != err {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	clusterValue := hex.EncodeToString([]byte(providerSpec.Cluster))
	listOfVMs := make(map[string]string)
	zoneValue := hex.EncodeToString([]byte(providerSpec.Zone))

	for _, server := range *servers.Items {
		if "INACTIVE" == *server.Metadata.State {
			continue
		}

		labels, _, err := client.LabelApi.DatacentersServersLabelsGet(ctx, providerSpec.DatacenterID, *server.Id).Depth(1).Execute()
		if nil != err {
			return nil, status.Error(codes.Unavailable, err.Error())
		}

		labelMatches := 0

		for _, label := range *labels.Items {
			switch *label.Properties.Key {
			case "cluster":
				if clusterValue == *label.Properties.Value {
					labelMatches++
				}

				break
			case "role":
				if "node" == *label.Properties.Value {
					labelMatches++
				}

				break
			case "zone":
				if zoneValue == *label.Properties.Value {
					labelMatches++
				}

				break
			}
		}

		if 3 == labelMatches {
			listOfVMs[transcoder.EncodeProviderID(providerSpec.DatacenterID, *server.Id)] = *server.Properties.Name
		}
	}

	return &driver.ListMachinesResponse{ MachineList: listOfVMs }, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request object to get a list of VolumeIDs for a PVSpec
func (p *MachineProvider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "")
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
//
// PARAMETERS
// ctx context.Context              Request context
// req *driver.CreateMachineRequest The request for generating the generic machineClass
func (p *MachineProvider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been received for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}
