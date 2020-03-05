#!/bin/bash

# Copyright 2020 Elotl Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

if [[ $# -lt 1 ]]; then
    echo "usage $0 <tag>"
    echo "where tag should be of the form v1.2.3"
    exit 1
fi

if [[ $(git rev-parse --abbrev-ref HEAD) != "master" ]]; then
    echo "a release must be made from master"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$SCRIPT_DIR/..



version_tag=$1
version_string=${version_tag:1}  # strip off the leading 'v'
if [[ $version_tag =~ ^v[0-9].* ]]; then
    $ROOT_DIR/build

    echo $version_string > $ROOT_DIR/version
    git commit --allow-empty -am "release $version_tag"

    git tag -a $version_tag -m "release $version_tag"
    git push --follow-tags origin master

    aws s3 cp cloud-init s3://itzo-dev-download/itzo-cloud-init-$version_tag --acl public-read
    aws s3 cp cloud-init s3://itzo-dev-download/itzo-cloud-init-latest --acl public-read
else
    echo "tag must be of the form v1.2.3"
fi
