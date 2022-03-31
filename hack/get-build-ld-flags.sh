#!/bin/bash -e
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PACKAGE_PATH="${1:-k8s.io/component-base}"
VERSION_PATH="${2:-$(dirname $0)/../VERSION}"
VERSION="$(cat "$VERSION_PATH")"

PROGRAM_NAME="machine-controller-manager-provider-ionos"
MAJOR_VERSION=""
MINOR_VERSION=""

if [[ "${VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?([+].*)?$ ]]; then
  MAJOR_VERSION=${BASH_REMATCH[1]}
  MINOR_VERSION=${BASH_REMATCH[2]}
  if [[ -n "${BASH_REMATCH[4]}" ]]; then
    MINOR_VERSION+="+"
  fi
fi

echo "-X $PACKAGE_PATH/version.gitMajor=$MAJOR_VERSION
      -X $PACKAGE_PATH/version.gitMinor=$MINOR_VERSION
      -X $PACKAGE_PATH/version.gitVersion=$VERSION
      -X $PACKAGE_PATH/version.buildDate=$(date '+%Y-%m-%dT%H:%M:%S%z' | sed 's/\([0-9][0-9]\)$/:\1/g')
      -X $PACKAGE_PATH/version/verflag.programName=$PROGRAM_NAME
      -X github.com/23technologies/$PROGRAM_NAME/pkg/version.Version=$VERSION"
