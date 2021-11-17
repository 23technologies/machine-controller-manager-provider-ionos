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

// WaitForDatacenterModifications waits for all pending changes of a datacenter to complete.
//
// PARAMETERS
// ctx    context.Context     Execution context
// client *ionossdk.APIClient IONOS client
// id     string              Datacenter ID
func WaitForDatacenterModifications(ctx context.Context, client *ionossdk.APIClient, id string) error {
	_, err := WaitForDatacenterModificationsAndGetResult(ctx, client, id)
	return err
}

// WaitForDatacenterModificationsAndGetResult waits for all pending changes of a datacenter to complete and returns the datacenter result struct.
//
// PARAMETERS
// ctx    context.Context     Execution context
// client *ionossdk.APIClient IONOS client
// id     string              Datacenter ID
func WaitForDatacenterModificationsAndGetResult(ctx context.Context, client *ionossdk.APIClient, id string) (ionossdk.Datacenter, error) {
	var datacenter ionossdk.Datacenter
	repeat := true
	tryCount := 0

	for repeat {
		datacenterResult, httpResponse, err := client.DataCenterApi.DatacentersFindById(ctx, id).Depth(0).Execute()

		if nil == httpResponse || 404 != httpResponse.StatusCode {
			if nil != err {
				return datacenter, err
			}

			repeat = "BUSY" == *datacenterResult.Metadata.State
			datacenter = datacenterResult
		}

		tryCount += 1

		if repeat {
			if tryCount > defaultOperationRetries {
				return datacenter, errors.New("Maximum number of retries exceeded waiting for datacenter modifications")
			}

			time.Sleep(defaultOperationInterval)
		}
	}

	return datacenter, nil
}
