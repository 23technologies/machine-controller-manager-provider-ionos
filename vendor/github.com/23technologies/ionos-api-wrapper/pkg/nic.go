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

// WaitForNicModifications waits for all pending changes of a NIC to complete.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// nicID        string              NIC ID
func WaitForNicModifications(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, nicID string) error {
	_, err := WaitForNicModificationsAndGetResult(ctx, client, datacenterID, serverID, nicID)
	return err
}

// WaitForNicModificationsAndGetResult waits for all pending changes of a NIC to complete and returns the NIC result struct.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// nicID        string              NIC ID
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
			if tryCount > defaultOperationRetries {
				return nic, errors.New("Maximum number of retries exceeded waiting for NIC modifications")
			}

			time.Sleep(defaultOperationInterval)
		}
	}

	return nic, nil
}
