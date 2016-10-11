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

# this block will wait check if snapctl and snapd are loaded before the plugins are loaded and the task is started
 for i in `seq 1 5`; do
             if [[ -f /usr/local/bin/snapctl && -f /usr/local/bin/snapd ]];
                then

                    _info "loading plugins"
                    snapctl plugin load "${PLUGIN_PATH}/snap-plugin-collector-disk"

                    _info "creating and starting a task"
                    snapctl task create -t "${__dir}/task-mock-disk.yml" 

                    SNAP_FLAG=1

                    break
             fi 
        
        _info "snapctl and/or snapd are unavailable, sleeping for 3 seconds" 
        sleep 3
done 


# check if snapctl/snapd have loaded
if [ $SNAP_FLAG -eq 0 ]
    then
     echo "Could not load snapctl or snapd"
     exit 1
fi
