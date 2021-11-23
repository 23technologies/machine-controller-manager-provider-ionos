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
	"net"
	"strconv"

	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

// nicConfiguration is a struct to provide configuration values for new NICs
type nicConfiguration struct {
	// lanIP contains the LAN IP to use
	LanIP string
	// enableDHCP should be false to disable DHCP for the NIC
	EnableDHCP     bool
	// enableFirewall should be true to enable the firewall for the NIC
	EnableFirewall bool
}

// attachLANToServer attaches the LAN ID given to the server and uses the floating pool IP.
//
// PARAMETERS
// ctx              context.Context     Execution context
// client           *ionossdk.APIClient IONOS client
// datacenterID     string              Datacenter ID
// serverID         string              Server ID
// lanID            string              LAN ID
// nicConfiguration *nicConfiguration   Configuration to apply
func attachLANToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string, nicConfiguration *nicConfiguration) error {
	numericLANID, err := strconv.Atoi(lanID)
	if nil != err {
		return err
	}

	apiLANID := int32(numericLANID)

	nicProperties := ionossdk.NicProperties{
		Lan: &apiLANID,
	}

	if "" != nicConfiguration.LanIP {
		ips := []string{nicConfiguration.LanIP}
		nicProperties.Ips = &ips
	}

	if !nicConfiguration.EnableDHCP {
		nicProperties.Dhcp = &nicConfiguration.EnableDHCP
	}

	if nicConfiguration.EnableFirewall {
		nicProperties.FirewallActive = &nicConfiguration.EnableFirewall
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
	nicConfiguration := &nicConfiguration{
		EnableDHCP:     true,
		EnableFirewall: false,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
}

// AttachLANToServerWithIP attaches the LAN ID given to the server with the IP given.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// lanID        string              LAN ID
// lanIP        string              LAN IP to use
func AttachLANToServerWithIP(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string, lanIP *net.IP) error {
	nicConfiguration := &nicConfiguration{
		LanIP:          lanIP.String(),
		EnableDHCP:     true,
		EnableFirewall: false,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
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
	nicConfiguration := &nicConfiguration{
		EnableDHCP:     false,
		EnableFirewall: false,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
}

// AttachWANToServer attaches the LAN ID given to the server.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// lanID        string              LAN ID
func AttachWANToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string) error {
	nicConfiguration := &nicConfiguration{
		EnableDHCP:     true,
		EnableFirewall: true,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
}

// AttachWANToServerWithIP attaches the LAN ID given to the server with the IP given.
//
// PARAMETERS
// ctx          context.Context     Execution context
// client       *ionossdk.APIClient IONOS client
// datacenterID string              Datacenter ID
// serverID     string              Server ID
// lanID        string              LAN ID
// lanIP        string              LAN IP to use
func AttachWANToServerWithIP(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID string, lanIP *net.IP) error {
	nicConfiguration := &nicConfiguration{
		LanIP:          lanIP.String(),
		EnableDHCP:     true,
		EnableFirewall: true,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
}

// AttachWANAndFloatingIPToServer attaches the LAN ID given to the server and uses a free floating pool IP from the given IP block ID.
//
// PARAMETERS
// ctx            context.Context     Execution context
// client         *ionossdk.APIClient IONOS client
// datacenterID   string              Datacenter ID
// serverID       string              Server ID
// lanID          string              LAN ID
// floatingPoolID string              Floating pool ID to select IP from
func AttachWANAndFloatingIPToServer(ctx context.Context, client *ionossdk.APIClient, datacenterID, serverID, lanID, floatingPoolID string) error {
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

	nicConfiguration := &nicConfiguration{
		LanIP:          floatingPoolIP,
		EnableDHCP:     true,
		EnableFirewall: true,
	}

	return attachLANToServer(ctx, client, datacenterID, serverID, lanID, nicConfiguration)
}
