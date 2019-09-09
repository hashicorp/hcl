#!/bin/bash

set -euo pipefail

# All paths from this point on are relative to the directory containing this
# script, for simplicity's sake.
cd "$( dirname "${BASH_SOURCE[0]}" )"

# Read the config file using hcldec and then use jq to extract values in a
# shell-friendly form. jq will ensure that the values are properly quoted and
# escaped for consumption by the shell.
CONFIG_VARS="$(hcldec --spec=spec.hcldec example.conf | jq -r '@sh "NAME=\(.name) GREETING=\(.greeting) FRIENDS=(\(.friends))"')"
if [ $? != 0 ]; then
    # If hcldec or jq failed then it has already printed out some error messages
    # and so we can bail out.
    exit $?
fi

# Import our settings into our environment
eval "$CONFIG_VARS"

# ...and now, some contrived usage of the settings we loaded:
echo "$GREETING $NAME!"
for name in ${FRIENDS[@]}; do
    echo "$GREETING $name, too!"
done
