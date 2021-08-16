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

import (
	"context"
	"errors"
	"strings"
	"time"

	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

// Constant defaultMachineOperationInterval is the time to wait between retries
const defaultMachineOperationInterval = 15 * time.Second
// Constant defaultMachineOperationRetries is the maximum number of retries
const defaultMachineOperationRetries = 20

// AddLabelToServer adds a label to the server ID given
func AddLabelToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, key, value string) error {
	labelProperties := ionossdk.LabelResourceProperties{
		Key: &key,
		Value: &value,
	}

	labelApiCreateRequest := client.LabelApi.DatacentersServersLabelsPost(ctx, datacenterID, id).Depth(0)
	_, _, err := labelApiCreateRequest.Label(ionossdk.LabelResource{Properties: &labelProperties}).Execute()
	if nil != err {
		return err
	}

	return nil
}

// AddLabelToVolume adds a label to the server ID given
func AddLabelToVolume(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, key, value string) error {
	labelProperties := ionossdk.LabelResourceProperties{
		Key: &key,
		Value: &value,
	}

	labelApiCreateRequest := client.LabelApi.DatacentersVolumesLabelsPost(ctx, datacenterID, id).Depth(0)
	_, _, err := labelApiCreateRequest.Label(ionossdk.LabelResource{Properties: &labelProperties}).Execute()
	if nil != err {
		return err
	}

	return nil
}

// GetRegionFromZone returns the region for a given zone string
func GetRegionFromZone(zone string) string {
	zoneData := strings.SplitN(zone, "-", 2)
	return zoneData[0]
}

func WaitForNicModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, nicID string) error {
	_, err := WaitForNicModificationsAndGetResult(ctx, client, datacenterID, serverID, nicID)
	return err
}

func WaitForNicModificationsAndGetResult(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, nicID string) (ionossdk.Nic, error) {
	var nic ionossdk.Nic
	repeat := true
	tryCount := 0

	for repeat {
		nicResult, httpResponse, err := client.NicApi.DatacentersServersNicsFindById(ctx, datacenterID, serverID, nicID).Depth(0).Execute()

		if 404 != httpResponse.StatusCode {
			if nil != err {
				return nic, err
			}

			repeat = "BUSY" == *nicResult.Metadata.State
			nic = nicResult
		}

		tryCount += 1

		if repeat {
			if tryCount > defaultMachineOperationRetries {
				return nic, errors.New("Maximum number of retries exceeded waiting for NIC modifications")
			}

			time.Sleep(defaultMachineOperationInterval)
		}
	}

	return nic, nil
}

func WaitForServerModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) error {
	_, err := WaitForServerModificationsAndGetResult(ctx, client, datacenterID, id)
	return err
}

func WaitForServerModificationsAndGetResult(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) (ionossdk.Server, error) {
	var server ionossdk.Server
	repeat := true
	tryCount := 0

	for repeat {
		serverResult, httpResponse, err := client.ServerApi.DatacentersServersFindById(ctx, datacenterID, id).Depth(0).Execute()

		if 404 != httpResponse.StatusCode {
			if nil != err {
				return server, err
			}

			repeat = "BUSY" == *serverResult.Metadata.State
			server = serverResult
		}

		tryCount += 1

		if repeat {
			if tryCount > defaultMachineOperationRetries {
				return server, errors.New("Maximum number of retries exceeded waiting for server modifications")
			}

			time.Sleep(defaultMachineOperationInterval)
		}
	}

	return server, nil
}

func WaitForVolumeModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) error {
	_, err := WaitForVolumeModificationsAndGetResult(ctx, client, datacenterID, id)
	return err
}

func WaitForVolumeModificationsAndGetResult(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) (ionossdk.Volume, error) {
	var volume ionossdk.Volume
	repeat := true
	tryCount := 0

	for repeat {
		volumeResult, httpResponse, err := client.VolumeApi.DatacentersVolumesFindById(ctx, datacenterID, id).Depth(0).Execute()

		if 404 != httpResponse.StatusCode {
			if nil != err {
				return volume, err
			}

			repeat = "BUSY" == *volumeResult.Metadata.State
			volume = volumeResult
		}

		tryCount += 1

		if repeat {
			if tryCount > defaultMachineOperationRetries {
				return volume, errors.New("Maximum number of retries exceeded waiting for volume modifications")
			}

			time.Sleep(defaultMachineOperationInterval)
		}
	}

	return volume, nil
}
