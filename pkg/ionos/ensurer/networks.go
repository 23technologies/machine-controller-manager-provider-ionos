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

func EnsureFloatingPoolIPBlockLANIsCreated(ctx context.Context, client *ionossdk.APIClient, datacenterID, floatingPoolID, lanName string) (string, error) {
	if "" != floatingPoolID {
		floatingPoolIPBlock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, floatingPoolID).Execute()
		if nil != err {
			return "", err
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
			return "", errors.New(fmt.Sprintf("Floating Pool IP Block '%s' given is exhausted", floatingPoolIPBlock))
		}

		public := true

		lanProperties := ionossdk.LanPropertiesPost{
			Name: &lanName,
			IpFailover: &[]ionossdk.IPFailover{ionossdk.IPFailover{Ip: &floatingPoolIP}},
			Public: &public,
		}

		lanApiCreateRequest := client.LanApi.DatacentersLansPost(ctx, datacenterID).Depth(0)
		lan, _, err := lanApiCreateRequest.Lan(ionossdk.LanPost{Properties: &lanProperties}).Execute()
		if nil != err {
			return "", err
		}

		return *lan.Id, nil
	}

	return "", nil
}

func EnsureFloatingPoolIPBlockLANIsDeleted(ctx context.Context, client *ionossdk.APIClient, datacenterID, lanName string) error {
	lans, _, err := client.LanApi.DatacentersLansGet(ctx, datacenterID).Depth(1).Execute()
	if nil != err {
		return err
	}

	for _, lan := range *lans.Items {
		if lanName == *lan.Properties.Name {
			_, _, err := client.LanApi.DatacentersLansDelete(ctx, datacenterID, *lan.Id).Depth(0).Execute()
			if nil != err {
				return err
			}
		}
	}

	return nil
}

func EnsureLANIsAttachedToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, id, lanID string) error {
	numericLANID, err := strconv.Atoi(lanID)
	if nil != err {
		return err
	}

	apiLANID := int32(numericLANID)

	nicProperties := ionossdk.NicProperties{
		Lan: &apiLANID,
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
