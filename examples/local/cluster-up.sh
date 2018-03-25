#!/bin/bash

# Copyright 2018 Google Inc.
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

script_root=`dirname "${BASH_SOURCE}"`

$script_root/zk-up.sh
$script_root/vtctld-up.sh

$script_root/vttablet-up.sh
sleep 10s
$script_root/lvtctl.sh InitShardMaster -force test_keyspace/0 test-100

$script_root/lvtctl.sh ApplySchema -sql "$(cat create_test_table.sql)" test_keyspace

$script_root/vtgate-up.sh
sleep 5s

# insert some rows
$script_root/client.sh
