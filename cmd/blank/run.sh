#!/bin/bash

host_data='{"lattice_rpc_url": "0.0.0.0:4222", "lattice_rpc_prefix": "default", "provider_key": "blank", "link_name": "default"}' 
echo $host_data | go run ./cmd/blank
