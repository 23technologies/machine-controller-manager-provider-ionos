/*
Copyright (c) 2021 23 Technologies GmbH. All rights reserved.

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

// Package pkg is the main package for IONOS specific APIs
package pkg

import (
	"context"
	"errors"
	"time"

	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

// AddLabelToVolume adds a label to the server ID given
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// id           string              Volume ID
// key          string              Label key
// value        string              Label value
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

// WaitForVolumeModifications waits for all pending changes of a volume to complete.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// volumeID     string              Volume ID
func WaitForVolumeModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) error {
	_, err := WaitForVolumeModificationsAndGetResult(ctx, client, datacenterID, id)
	return err
}

// WaitForVolumeModificationsAndGetResult waits for all pending changes of a NIC to complete and returns the NIC result struct.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// volumeID     string              Volume ID
func WaitForVolumeModificationsAndGetResult(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) (ionossdk.Volume, error) {
	var volume ionossdk.Volume
	repeat := true
	tryCount := 0

	for repeat {
		volumeResult, httpResponse, err := client.VolumeApi.DatacentersVolumesFindById(ctx, datacenterID, id).Depth(0).Execute()

		if nil == httpResponse || 404 != httpResponse.StatusCode {
			if nil != err {
				return volume, err
			}

			repeat = "BUSY" == *volumeResult.Metadata.State
			volume = volumeResult
		}

		tryCount += 1

		if repeat {
			if tryCount > defaultOperationRetries {
				return volume, errors.New("Maximum number of retries exceeded waiting for volume modifications")
			}

			time.Sleep(defaultOperationInterval)
		}
	}

	return volume, nil
}
