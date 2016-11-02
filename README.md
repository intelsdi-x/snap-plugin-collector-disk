[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-collector-disk.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap-plugin-collector-disk) [![Build Status](https://ci.snap-telemetry.io/buildStatus/icon?job=snap-plugin-collector-disk-large-prb-k8s)](https://ci.snap-telemetry.io/job/snap-plugin-collector-disk-large-prb-k8s/)

# snap collector plugin - disk

This plugin gather disk statistics from /proc/diskstats (Linux 2.6+) or /proc/partitions (Linux 2.4) for the [Snap telemetry framework](http://github.com/intelsdi-x/snap).

## Table of Contents
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

- Linux (kernel 2.4, 2.6+)

### Installation

#### Download the plugin binary:

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-disk/releasess) page. Download the plugin from the latest release and load it into `snapd` (`/opt/snap/plugins` is the default location for snap packages).

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
This builds the plugin in `./build/`

### Configuration and Usage

* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Load the plugin and create a task, see example in [Examples](https://github.com/intelsdi-x/snap-plugin-collector-disk/blob/master/README.md#examples).

Configuration parameters:

- `proc_path`: path to 'diskstats' or 'partitions' file (default: `/proc`)

## Documentation

This collector gathers metrics from `/proc/diskstats` for kernel 2.6+, or from `/proc/partitions` for kernel 2.4. The configuration `proc_path` determines where the plugin obtains these metrics, with a default setting of `/proc`. This setting is only required to obtain data from a docker container that mounts the host `/proc` in an alternative path.

To read more about disk I/O statistics fields, please visit [www.kernel.org/doc/Documentation/iostats.txt](https://www.kernel.org/doc/Documentation/iostats.txt)


### Collected Metrics

This plugin has the ability to gather the following metrics:

Metric namespace is `/intel/procfs/disk/<disk_device>/<metric_name>` where `<disk_device>` expands to sda, sda1, sdb, sdb1 and so on.

Metric namespace | Type | Description
------------ | ------------- | -------------
/intel/procfs/disk/\<disk_device\>/merged_read | derive | The number of read operations per second that could be merged with already queued operations.
/intel/procfs/disk/\<disk_device\>/merged_write | derive | The number of write operations per second that could be merged with already queued operations.
/intel/procfs/disk/\<disk_device\>/octets_read |derive |  The number of octets (bytes) read per second.
/intel/procfs/disk/\<disk_device\>/octets_write | derive | The number of octets (bytes) written per second.
/intel/procfs/disk/\<disk_device\>/ops_read | derive | The number of read operations per second.
/intel/procfs/disk/\<disk_device\>/ops_write | derive | The number of write operations per second.
/intel/procfs/disk/\<disk_device\>/time_read | derive |  The average time for a read operation to complete in the last interval, in miliseconds.
/intel/procfs/disk/\<disk_device\>/time_write | derive |  The average time for a write operation to complete in the last interval, in miliseconds.
/intel/procfs/disk/\<disk_device\>/io_time | derive | The time spent doing I/Os, in miliseconds.
/intel/procfs/disk/\<disk_device\>/weighted_io_time<sup>(1)</sup> | derive | The weighted time spent doing I/Os, in miliseconds.
/intel/procfs/disk/\<disk_device\>/pending_ops</sup> | gauge | The queue size of pending I/O operation.

<sup>1)</sup> The value of metric `weighted_io_time` is incremented at each I/O start, I/O completion, I/O merge, or read of these stats by the number of I/Os in progress times the number of milliseconds spent doing I/O since the
last update of this field.

Data type of all above metrics is float64.

By default metrics are gathered once per second.

### Examples

Example of running snap disk collector and writing data to file.

Ensure [snap daemon is running](https://github.com/intelsdi-x/snap#running-snap):
* initd: `service snap-telemetry start`
* systemd: `sysctl start snap-telemetry`
* command line: `snapd -l 1 -t 0 &`

Download and load snap plugins:
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-disk/latest/linux/x86_64/snap-plugin-collector-disk
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snapctl plugin load snap-plugin-collector-disk
$ snapctl plugin load snap-plugin-publisher-file
```

See all available metrics:
```
$ snapctl metric list
NAMESPACE 				 VERSIONS
/intel/procfs/disk/*/io_time 		 3
/intel/procfs/disk/*/merged_read 	 3
/intel/procfs/disk/*/merged_write 	 3
/intel/procfs/disk/*/octets_read 	 3
/intel/procfs/disk/*/octets_write 	 3
/intel/procfs/disk/*/ops_read 		 3
/intel/procfs/disk/*/ops_write 		 3
/intel/procfs/disk/*/pending_ops 	 3
/intel/procfs/disk/*/time_read 		 3
/intel/procfs/disk/*/time_write 	 3
/intel/procfs/disk/*/weighted_io_time 	 3
```

Download an [example task file](https://github.com/intelsdi-x/snap-plugin-collector-disk/blob/master/examples/tasks/) and load it:
```
$ curl -sfLO https://raw.githubusercontent.com/intelsdi-x/snap-plugin-collector-disk/master/examples/tasks/diskstats-file.json
$ snapctl task create -t diskstats-file.json
Using task manifest to create task
Task created
ID: 480323af-15b0-4af8-a526-eb2ca6d8ae67
Name: Task-480323af-15b0-4af8-a526-eb2ca6d8ae67
State: Running
```

See realtime output from `snapctl task watch <task_id>` (CTRL+C to exit)
```
$ snapctl task watch 480323af-15b0-4af8-a526-eb2ca6d8ae67
Watching Task (480323af-15b0-4af8-a526-eb2ca6d8ae67):
NAMESPACE                                  DATA                        TIMESTAMP                                       SOURCE
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
```

This data is published to a file `/tmp/published_diskstats` per task specification

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

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [Slack](http://slack.snap-telemetry.io).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
Snap, along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: 	[Izabella Raulin](https://github.com/IzabellaRaulin)

**Thank you!** Your contribution is incredibly important to us.
