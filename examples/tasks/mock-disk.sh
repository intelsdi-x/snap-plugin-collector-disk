#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# source the common bash script 
. "${__dir}/../../scripts/common.sh"

# ensure PLUGIN_PATH is set
TMPDIR=${TMPDIR:-"/tmp"}
PLUGIN_PATH=${PLUGIN_PATH:-"${TMPDIR}/snap/plugins"}
mkdir -p $PLUGIN_PATH

_info "Get latest plugins"
(cd $PLUGIN_PATH && curl -sfLSO http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-disk/latest/linux/x86_64/snap-plugin-collector-disk && chmod 755 snap-plugin-collector-disk)

SNAP_FLAG=0

# this block will wait check if snaptel and snapteld are loaded before the plugins are loaded and the task is started
 for i in `seq 1 5`; do
             if [[ -f /usr/local/bin/snaptel && -f /usr/local/sbin/snapteld ]];
                then

                    _info "loading plugins"
                    snaptel plugin load "${PLUGIN_PATH}/snap-plugin-collector-disk"

                    _info "creating and starting a task"
                    snaptel task create -t "${__dir}/task-mock-disk.yml"

                    SNAP_FLAG=1

                    break
             fi 
        
        _info "snaptel and/or snapteld are unavailable, sleeping for 3 seconds"
        sleep 3
done 


# check if snaptel/snapteld have loaded
if [ $SNAP_FLAG -eq 0 ]
    then
     echo "Could not load snaptel or snapteld"
     exit 1
fi
