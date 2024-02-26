#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

if [[ -n $(gofmt -l ./) ]]; then echo "Please run gofmt -w ./ to format code"; exit 1; fi;
