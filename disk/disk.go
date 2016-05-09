/*
http://www.apache.org/licenses/LICENSE-2.0.txt
Copyright 2016 Intel Corporation
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package disk

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

const (
	// Name of plugin
	Name = "disk"
	// Version of plugin
	Version = 2
	// Type of plugin
	Type = plugin.CollectorPluginType

	nsVendor = "intel"
	nsClass  = "procfs"
	nsType   = "disk"

	uintMax = ^uint64(0)
)

const (
	// name of available metrics
	nOpsRead        = "ops_read"
	nOctetsRead     = "octets_read"
	nOpsWrite       = "ops_write"
	nOctetsWrite    = "octets_write"
	nMergedRead     = "merged_read"
	nTimeRead       = "time_read"
	nMergedWrite    = "merged_write"
	nTimeWrite      = "time_write"
	nPendingOps     = "pending_ops"
	nIoTime         = "io_time"
	nWeightedIoTime = "weighted_io_time"
)

var (
	// Disk statistics source for kernel 2.6+
	srcFile = "/proc/diskstats"

	// Source for older kernel versions
	srcFileOldVer = "/proc/partitions"
)

// DiskCollector holds disk statistics
type DiskCollector struct {
	data     diskStats
	dataPrev diskStats          // previous data, to calculate derivatives
	output   map[string]float64 // contains exposed metrics and their value (calculated based on data & dataPrev)
	first    bool               // is true for first collecting (do not calculate derivatives), after that set false
}

type diskStats struct {
	stats     map[string]uint64
	timestamp time.Time
}

type diffStats struct {
	diffWriteTime uint64
	diffReadTime  uint64
	diffWriteOps  uint64
	diffReadOps   uint64
}

// prefix in metric namespace
var prefix = []string{nsVendor, nsClass, nsType}

// New returns snap-plugin-collector-disk instance
func New() (*DiskCollector, error) {
	dc := &DiskCollector{data: diskStats{stats: map[string]uint64{}, timestamp: time.Now()},
		dataPrev: diskStats{stats: map[string]uint64{}, timestamp: time.Now()},
		output:   map[string]float64{},
		first:    true}
	return dc, nil
}

// GetConfigPolicy returns a ConfigPolicy
func (dc *DiskCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}

// GetMetricTypes returns list of exposed disk stats metrics
func (dc *DiskCollector) GetMetricTypes(_ plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}

	if err := dc.getDiskStats(); err != nil {
		return nil, err
	}
	for stat := range dc.data.stats {
		metric := plugin.MetricType{Namespace_: core.NewNamespace(createNamespace(stat)...)}
		mts = append(mts, metric)
	}
	return mts, nil
}

// CollectMetrics retrieves disk stats values for given metrics
func (dc *DiskCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	metrics := []plugin.MetricType{}

	first := dc.first // true if collecting for the first time
	if first {
		dc.first = false
	}

	// for first collecting skip stash previous data
	if !first {
		// stash disk data (dst, src)
		stashData(&dc.dataPrev, &dc.data)
	}

	// get presence disk stats
	if err := dc.getDiskStats(); err != nil {
		return nil, err
	}

	//  for first collecting skip derivatives calculation
	if !first {
		// calculate derivatives based on data (presence) and previous one; results stored in dc.output
		if err := dc.calcDerivatives(); err != nil {
			return nil, err
		}
	}

	for _, m := range mts {
		if v, ok := dc.output[parseNamespace(m.Namespace().Strings())]; ok {
			metric := plugin.MetricType{
				Namespace_: m.Namespace(),
				Data_:      v,
				Timestamp_: dc.data.timestamp,
			}
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// getDiskStats gets disk stats from file (/proc/{diskstats|partitions}) and stores them in the DiskCollector structure
func (dc *DiskCollector) getDiskStats() error {

	var scanner *bufio.Scanner
	fieldshift := 0

	//map disk statistics keys (names) to scanned fields
	data := make(map[string]string)

	fh, err := os.Open(srcFile)
	defer fh.Close()

	if err == nil {
		scanner = bufio.NewScanner(fh)
	} else {
		// unable to open /proc/diskstats, try to open /proc/partitions
		/* Kernel 2.4, Partition */
		fh2, err2 := os.Open(srcFileOldVer)
		defer fh2.Close()

		if err2 != nil {
			return fmt.Errorf("Error open /proc/{diskstats|partitions}, errors = %s; %s", err, err2)
		}

		/* Kernel is 2.4.* */
		fieldshift = 1
		scanner = bufio.NewScanner(fh2)
	}

	dc.data.timestamp = time.Now()

	// scan file content
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		numfields := len(fields)

		if numfields != 14+fieldshift && numfields != 7 {
			// unknown entry, ignore it
			continue
		}

		// get minor device number
		minor, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			// invalid format of file
			return err
		}

		diskName := fields[2+fieldshift]

		if numfields == 7 {
			/* Kernel < 2.6, proc/partitions */
			data[nOpsRead] = fields[3]
			data[nOctetsRead] = fields[4] // read sectors
			data[nOpsWrite] = fields[5]
			data[nOctetsWrite] = fields[6] // write sectors

		} else if numfields == (14 + fieldshift) {
			/* Kernel 2.6 or higher, proc/diskstats */
			data[nOpsRead] = fields[3+fieldshift]
			data[nOctetsRead] = fields[5+fieldshift]
			data[nOpsWrite] = fields[7+fieldshift]
			data[nOctetsWrite] = fields[9+fieldshift]

			if fieldshift == 0 || minor == 0 {
				// extended statistics
				data[nMergedRead] = fields[4+fieldshift]
				data[nTimeRead] = fields[6+fieldshift]
				data[nMergedWrite] = fields[8+fieldshift]
				data[nTimeWrite] = fields[10+fieldshift]
				data[nPendingOps] = fields[11+fieldshift] // ops currently in progress
				data[nIoTime] = fields[12+fieldshift]
				data[nWeightedIoTime] = fields[13+fieldshift]
			} // end of extended stats

		}

		for key, val := range data {
			// parse disk stats value
			if value, err := strconv.ParseUint(val, 10, 64); err == nil {
				dc.data.stats[diskName+"/"+key] = value
			} else {
				// parse failure
				return fmt.Errorf("Error %+v, cannot convert value of `%+v` equals %+v to uint64", err, diskName+"/"+key, val)
			}
		}
	} // end of scanner.Scan()

	return nil
}

// calcDerivatives calculates derivatives of metrics values and stored them in DiskCollector structure as a 'output'
func (dc *DiskCollector) calcDerivatives() error {
	interval := dc.data.timestamp.Sub(dc.dataPrev.timestamp).Seconds()

	if interval <= 0 {
		return errors.New("Invalid interval value")
	}

	if len(dc.data.stats) == 0 || len(dc.dataPrev.stats) == 0 {
		return errors.New("No data for calculation")
	}

	var diffVal uint64

	// (for each disk) keep values of the stats differences which are needed to calculate avg disk time
	avgDiskTime := make(map[string]*diffStats)

	for key, val := range dc.data.stats {

		if strings.HasSuffix(key, nPendingOps) {
			// for 'pending_ops' output value is simply stored as-is
			dc.output[key] = float64(val)
			continue
		}

		/** Calculate the change of the value in interval time **/

		valPrev := dc.dataPrev.stats[key]

		// if the counter wraps around
		if val < valPrev {
			diffVal = 1 + val + (uintMax - valPrev)
		} else {
			diffVal = val - valPrev
		}

		deriveVal := float64(diffVal) / interval

		disk, nMetric := parseMetricKey(key)

		if _, exists := avgDiskTime[disk]; exists == false {
			avgDiskTime[disk] = new(diffStats)
		}

		// switch special case for some metrics based on the last part of metric name
		switch nMetric {
		// switch case based on the last part of metric name
		case nOctetsWrite:
			dc.output[key] = 512 * deriveVal

		case nOctetsRead:
			dc.output[key] = 512 * deriveVal

		case nTimeWrite:
			avgDiskTime[disk].diffWriteTime = diffVal

		case nTimeRead:
			avgDiskTime[disk].diffReadTime = diffVal

		case nOpsRead:
			avgDiskTime[disk].diffReadOps = diffVal
			dc.output[key] = deriveVal

		case nOpsWrite:
			avgDiskTime[disk].diffWriteOps = diffVal
			dc.output[key] = deriveVal

		default:
			dc.output[key] = deriveVal
		}
	}

	// calculate disk time
	for disk, values := range avgDiskTime {
		dc.output[disk+"/"+nTimeRead] = calcTimeIncrement(values.diffReadTime, values.diffReadOps, interval)
		dc.output[disk+"/"+nTimeWrite] = calcTimeIncrement(values.diffWriteTime, values.diffWriteOps, interval)
	}

	return nil
}

// calcTimeIncrement returns average time of operation based on
func calcTimeIncrement(deltaTime uint64, deltaOps uint64, interval float64) float64 {
	if deltaOps == 0 {
		//not divide by zero
		return 0
	}
	avgTime := float64(deltaTime) / float64(deltaOps)
	avgTimeIncr := interval * avgTime

	// add 0.5 as it's done in collectd:disk
	return avgTimeIncr + .5
}

// stashData copies DiskStats struct variables items with teir values from 'src' to 'dst'
func stashData(dst *diskStats, src *diskStats) {
	dst.timestamp = src.timestamp

	// copy map, deep copy is needed
	for key, value := range src.stats {
		dst.stats[key] = value
	}
}

// createNamespace returns namespace slice of strings composed from: vendor, class, type and components of metric name
func createNamespace(name string) []string {
	return append(prefix, strings.Split(name, "/")...)
}

// parseNamespace performs reverse operation to createNamespace, extracts metric key from namespace
func parseNamespace(ns []string) string {
	// skip prefix in namespace
	metric := ns[len(prefix):]
	return strings.Join(metric, "/")
}

// parseMetricKey extracts information about disk and metric name from metric key (exemplary metric key is `sda/time_write`)
func parseMetricKey(k string) (disk, metric string) {
	result := strings.Split(k, "/")
	if len(result) < 2 {
		// invalid key format, return empty strings
		return
	}
	return result[0], result[1]
}
