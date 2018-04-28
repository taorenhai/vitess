#!/bin/bash

VTCTL="abe1bbcd84af511e8813b0a5ff78d9c6-2104917781.us-east-2.elb.amazonaws.com"

exec vtctlclient -server $VTCTL:15999 "$@"
