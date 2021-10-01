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

// AddLabelToServer adds a label to the server ID given
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// id           string              Server ID
// key          string              Label key
// value        string              Label value
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

// WaitForServerModifications waits for all pending changes of a server to complete.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// id           string              Server ID
func WaitForServerModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, id string) error {
	_, err := WaitForServerModificationsAndGetResult(ctx, client, datacenterID, id)
	return err
}

// WaitForServerModificationsAndGetResult waits for all pending changes of a server to complete and returns the server result struct.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// id           string              Server ID
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
			if tryCount > defaultOperationRetries {
				return server, errors.New("Maximum number of retries exceeded waiting for server modifications")
			}

			time.Sleep(defaultOperationInterval)
		}
	}

	return server, nil
}
