# snap collector plugin - disk

This plugin gather disk statistics from /proc/diskstats (Linux 2.6 or higher) or /proc/partitions (Linux 2.4.)
															
The plugin is used in the [snap framework] (http://github.com/intelsdi-x/snap).				

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

### System Requirements

- Linux system

### Installation

#### Download the plugin binary:
You can get the pre-built binaries for your OS and architecture at snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page. Download the plugins package from the latest release, unzip and store in a path you want `snapd` to access.

#### To build the plugin binary:

Fork https://github.com/intelsdi-x/snap-plugin-collector-disk  
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-disk.git
```

Build the snap disk plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `/build/rootfs/`

### Configuration and Usage

* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Load the plugin and create a task, see example in [Examples](https://github.com/intelsdi-x/snap-plugin-collector-disk/blob/master/README.md#examples).

## Documentation

### Collected Metrics
This plugin has the ability to gather the following metrics:
                                                                                                
Metric namespace is `/intel/procfs/disk/<disk_device>/<metric_name>`																					where `<disk_device>` expands to sda, sda1, sdb, sdb1 and so on.

Metric namespace | Description
------------ | -------------
/intel/procfs/disk/\<disk_device\>/merged_read | The number of read operations per second that could be merged with already queued operations.
/intel/procfs/disk/\<disk_device\>/merged_write | The number of write operations per second that could be merged with already queued operations.
/intel/procfs/disk/\<disk_device\>/octets_read | The number of octets (bytes) read per second.
/intel/procfs/disk/\<disk_device\>/octets_write | The number of octets (bytes) written per second.
/intel/procfs/disk/\<disk_device\>/ops_read | The number of read operations per second.
/intel/procfs/disk/\<disk_device\>/ops_write | The number of write operations per second.
/intel/procfs/disk/\<disk_device\>/time_read | The average time for a read operation to complete in the last interval.
/intel/procfs/disk/\<disk_device\>/time_write | the average time for a write operation to complete in the last interval.

                                                                          
Data type of all above metrics is float64.

By default metrics are gathered once per second.

### Examples

Example of running snap disk collector and writing data to file.

Run the snap daemon:
```
$ snapd -l 1 -t 0
```

Load disk plugin for collecting:
```
$ snapctl plugin load $SNAP_DISK_PLUGIN_DIR/build/rootfs/snap-plugin-collector-disk
Plugin loaded
Name: disk
Version: 1
Type: collector
Signed: false
Loaded Time: Wed, 23 Dec 2015 11:14:37 EST
```
See all available metrics:
```
$ snapctl metric list
```

Or see available metrics only for specific disk:
```
$ snapctl metric list | grep sda
```

Load file plugin for publishing:
```
$ snapctl plugin load $SNAP_DIR/build/plugin/snap-publisher-file
Plugin loaded
Name: file
Version: 3
Type: publisher
Signed: false
Loaded Time: Wed, 23 Dec 2015 11:15:02 EST
```

Create a task JSON file (exemplary files in [examples/tasks/] (https://github.com/intelsdi-x/snap-plugin-collector-disk/blob/master/examples/tasks/)):
```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "workflow": {
        "collect": {
            "metrics": {
				"/intel/procfs/disk/sda/ops_read": {},
                "/intel/procfs/disk/sda/ops_write": {},
                "/intel/procfs/disk/sda/merged_read": {},
                "/intel/procfs/disk/sda/merged_write": {},
                "/intel/procfs/disk/sda/octets_read": {},
                "/intel/procfs/disk/sda/octets_write": {},
                "/intel/procfs/disk/sda/io_time": {},
                "/intel/procfs/disk/sda/time_read": {},
                "/intel/procfs/disk/sda/time_write": {},
                "/intel/procfs/disk/sda/weighted_io_time": {},
                "/intel/procfs/disk/sda/pending_ops": {},
				
                "/intel/procfs/disk/sda1/ops_read": {},
                "/intel/procfs/disk/sda1/ops_write": {},
                "/intel/procfs/disk/sda1/merged_read": {},
                "/intel/procfs/disk/sda1/merged_write": {},
                "/intel/procfs/disk/sda1/octets_read": {},
                "/intel/procfs/disk/sda1/octets_write": {},
                "/intel/procfs/disk/sda1/io_time": {},
                "/intel/procfs/disk/sda1/time_read": {},
                "/intel/procfs/disk/sda1/time_write": {},
                "/intel/procfs/disk/sda1/weighted_io_time": {},
                "/intel/procfs/disk/sda1/pending_ops": {},
				
                "/intel/procfs/disk/sdb/ops_read": {},
                "/intel/procfs/disk/sdb/ops_write": {},
                "/intel/procfs/disk/sdb/merged_read": {},
                "/intel/procfs/disk/sdb/merged_write": {},
                "/intel/procfs/disk/sdb/octets_read": {},
                "/intel/procfs/disk/sdb/octets_write": {},
                "/intel/procfs/disk/sdb/io_time": {},
                "/intel/procfs/disk/sdb/time_read": {},
                "/intel/procfs/disk/sdb/time_write": {},
                "/intel/procfs/disk/sdb/weighted_io_time": {},
                "/intel/procfs/disk/sdb/pending_ops": {},
				
				"/intel/procfs/disk/sdb1/ops_read": {},
                "/intel/procfs/disk/sdb1/ops_write": {},
                "/intel/procfs/disk/sdb1/merged_read": {},
                "/intel/procfs/disk/sdb1/merged_write": {},
                "/intel/procfs/disk/sdb1/octets_read": {},
                "/intel/procfs/disk/sdb1/octets_write": {},
                "/intel/procfs/disk/sdb1/io_time": {},
                "/intel/procfs/disk/sdb1/time_read": {},
                "/intel/procfs/disk/sdb1/time_write": {},
                "/intel/procfs/disk/sdb1/weighted_io_time": {},
                "/intel/procfs/disk/sdb1/pending_ops": {},
				
				"/intel/procfs/disk/sdb2/ops_read": {},
                "/intel/procfs/disk/sdb2/ops_write": {},
                "/intel/procfs/disk/sdb2/merged_read": {},
                "/intel/procfs/disk/sdb2/merged_write": {},
                "/intel/procfs/disk/sdb2/octets_read": {},
                "/intel/procfs/disk/sdb2/octets_write": {},
                "/intel/procfs/disk/sdb2/io_time": {},
                "/intel/procfs/disk/sdb2/time_read": {},
                "/intel/procfs/disk/sdb2/time_write": {},
                "/intel/procfs/disk/sdb2/weighted_io_time": {},
                "/intel/procfs/disk/sdb2/pending_ops": {}								
            },
            "config": {
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "file",
                    "config": {
                        "file": "/tmp/published_diskstats"
                    }
                }
            ]
        }
    }
}
```

Create a task:
```
$ snapctl task create -t $SNAP_DISK_PLUGIN_DIR/examples/tasks/diskstats-file.json
Using task manifest to create task
Task created
ID: 480323af-15b0-4af8-a526-eb2ca6d8ae67
Name: Task-480323af-15b0-4af8-a526-eb2ca6d8ae67
State: Running
```

See sample output from `snapctl task watch <task_id>`

```
$ snapctl task watch 480323af-15b0-4af8-a526-eb2ca6d8ae67																									
Watching Task (480323af-15b0-4af8-a526-eb2ca6d8ae67):
NAMESPACE                                  DATA                        TIMESTAMP                                       SOURCE
/intel/procfs/disk/sda/io_time                    0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/merged_read                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/merged_write               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/octets_read                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/octets_write               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/ops_read                   0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/ops_write                  0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/pending_ops                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/time_read                  0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/time_write                 0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda/weighted_io_time           0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/io_time                   0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/merged_read               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/merged_write              0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/octets_read               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/octets_write              0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/ops_read                  0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/ops_write                 0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/pending_ops               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/time_read                 0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/time_write                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sda1/weighted_io_time          0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/io_time                    285.24017494599997          2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/merged_read                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/merged_write               155.4955120365347           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/octets_read                4.799664198069947e+06       2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/octets_write               2.0893468155700576e+08      2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/ops_read                   0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/ops_write                  338.722707748375            2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/pending_ops                0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/time_read                  4.300686205                 2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/time_write                 121.47803869131272          2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb/weighted_io_time           33117.57281195954           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/io_time                   285.24017494599997          2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/merged_read               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/merged_write              155.4955120365347           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/octets_read               4.799664198069947e+06       2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/octets_write              2.0617126235076627e+08      2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/ops_read                  0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/ops_write                 338.722707748375            2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/pending_ops               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/time_read                 4.300686205             	   2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/time_write                119.13865467618942          2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb1/weighted_io_time          33117.57281195954           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/io_time                   0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/merged_read               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/merged_write              0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/octets_read               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/octets_write              2.763419206239478e+06       2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/ops_read                  0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/ops_write                 0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/pending_ops               0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/time_read                 0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/time_write                4.503690537128046           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
/intel/procfs/disk/sdb2/weighted_io_time          0                           2015-12-23 11:18:09.224143712 -0500 EST         gklab-108-166
```
(Keys `ctrl+c` terminate task watcher)


These data are published to file and stored there (in this example in /tmp/published_diskstats).

Stop task:
```
$ snapctl task stop 480323af-15b0-4af8-a526-eb2ca6d8ae67
Task stopped:
ID: 480323af-15b0-4af8-a526-eb2ca6d8ae67
```

### Roadmap

There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-disk/issues) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-disk/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap.

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [snap Gitter channel](https://gitter.im/intelsdi-x/snap).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
Snap, along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: 	[Izabella Raulin](https://github.com/IzabellaRaulin)
