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
	ionossdk "github.com/ionos-cloud/sdk-go/v5"
)

var singletons = make(map[string]*ionossdk.APIClient)

// GetClientForUser returns an underlying IONOS client for the given user name.
//
// PARAMETERS
// user  string User name to look up client instance for
// password string Password for the user name. Please note that the password
//                 will not be replaced if an client is already cached.
func GetClientForUser(user, password string) *ionossdk.APIClient {
	client, ok := singletons[user]

	if !ok {
		config := ionossdk.NewConfiguration(user, password, "")
		client = ionossdk.NewAPIClient(config)
	}

    return client
}

// SetClientForUser sets a preconfigured IONOS client for the given user name.
//
// PARAMETERS
// user   string           User name to look up client instance for
// client *ionossdk.APIClient Preconfigured IONOS client
func SetClientForUser(user string, client *ionossdk.APIClient) {
	if client == nil {
		delete(singletons, user)
	} else {
		singletons[user] = client
	}
}
