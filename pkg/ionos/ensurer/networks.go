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

// Package ensurer provides functions used to ensure changes to be applied
package ensurer

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/23technologies/machine-controller-manager-provider-ionos/pkg/ionos/apis"
	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

func attachLANToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, lanID, floatingPoolIP string) error {
	numericLANID, err := strconv.Atoi(lanID)
	if nil != err {
		return err
	}

	apiLANID := int32(numericLANID)

	nicProperties := ionossdk.NicProperties{
		Lan: &apiLANID,
	}

	if "" != floatingPoolIP {
		ips := []string{floatingPoolIP}
		nicProperties.Ips = &ips
	}

	nicApiCreateRequest := client.NicApi.DatacentersServersNicsPost(ctx, datacenterID, id).Depth(0)
	nic, _, err := nicApiCreateRequest.Nic(ionossdk.Nic{Properties: &nicProperties}).Execute()
	if nil != err {
		return err
	}

	err = apis.WaitForNicModifications(ctx, client, datacenterID, id, *nic.Id)
	if nil != err {
		return err
	}

	return nil
}

func EnsureLANAndFloatingIPIsAttachedToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, lanID, floatingPoolID string) error {
	floatingPoolIPBlock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, floatingPoolID).Execute()
	if nil != err {
		return err
	}

	var floatingPoolIP string

	for _, ip := range *floatingPoolIPBlock.Properties.Ips {
		isIPInUse := false

		for _, ipConsumer := range *floatingPoolIPBlock.Properties.IpConsumers {
			isIPInUse = ip == *ipConsumer.Ip

			if isIPInUse {
				break
			}
		}

		if !isIPInUse {
			floatingPoolIP = ip
		}
	}

	if "" == floatingPoolIP {
		return errors.New(fmt.Sprintf("Floating Pool IP Block '%s' given is exhausted", floatingPoolID))
	}

	return attachLANToServer(ctx, client, datacenterID, id, lanID, floatingPoolIP)
}

func EnsureLANIsAttachedToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, lanID string) error {
	return attachLANToServer(ctx, client, datacenterID, id, lanID, "")
}
