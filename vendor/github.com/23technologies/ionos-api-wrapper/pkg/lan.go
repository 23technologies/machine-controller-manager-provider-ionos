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
	"fmt"
	"strconv"

	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

// AttachLANAndFloatingIPToServer attaches the LAN ID given to the server and uses a free floating pool IP from the given IP block ID.
//
// PARAMETERS
// ctx            context.Context     Execution context
// client         *ionossdk.APIClient IONOS client
// datacenterID   string              Datacenter ID
// serverID       string              Server ID
// lanID          string              LAN ID
// floatingPoolID string              Floating pool ID to select IP from
func AttachLANAndFloatingIPToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID, floatingPoolID string) error {
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

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, floatingPoolIP, true)
}

// attachLANToServer attaches the LAN ID given to the server and uses the floating pool IP.
//
// PARAMETERS
// ctx            context.Context     Execution context
// client         *ionossdk.APIClient IONOS client
// datacenterID   string              Datacenter ID
// serverID       string              Server ID
// lanID          string              LAN ID
// floatingPoolIP string              Floating pool IP to use
// enableDHCP     bool                False to disable DHCP for the NIC
func attachLANToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID, floatingPoolIP string, enableDHCP bool) error {
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

	if !enableDHCP {
		nicProperties.Dhcp = &enableDHCP
	}

	nicApiCreateRequest := client.NicApi.DatacentersServersNicsPost(ctx, datacenterID, serverID).Depth(0)
	nic, _, err := nicApiCreateRequest.Nic(ionossdk.Nic{Properties: &nicProperties}).Execute()
	if nil != err {
		return err
	}

	err = WaitForNicModifications(ctx, client, datacenterID, serverID, *nic.Id)
	if nil != err {
		return err
	}

	return nil
}

// AttachLANToServer attaches the LAN ID given to the server.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// lanID        string              LAN ID
func AttachLANToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string) error {
	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, "", true)
}

// AttachLANToServerWithoutDHCP attaches the LAN ID given to the server without DHCP support.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// lanID        string              LAN ID
func AttachLANToServerWithoutDHCP(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string) error {
	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, "", false)
}
