#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

if [[ -n $(gofmt -l ./) ]]; then echo "Please run gofmt -w ./ to format code"; exit 1; fi;
