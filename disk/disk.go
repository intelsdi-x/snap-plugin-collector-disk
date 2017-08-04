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
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	// Name of plugin
	PluginName = "disk"
	// Version of plugin
	PluginVersion = 5

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
	defaultSrcFile = "/proc/diskstats"

	// Source for older kernel versions
	defaultSrcFileOld = "/proc/partitions"
)

// DiskCollector holds disk statistics
type DiskCollector struct {
	data     diskStats             // holds current raw data
	dataPrev diskStats             // previous data, to calculate derivatives
	output   map[metricKey]float64 // contains exposed metrics and their value (calculated based on data & dataPrev)
	first    bool                  // is true for first collecting (do not calculate derivatives), after that set false
}

type diskStats struct {
	stats     map[metricKey]uint64
	timestamp time.Time
}

type diffStats struct {
	diffWriteTime uint64
	diffReadTime  uint64
	diffWriteOps  uint64
	diffReadOps   uint64
}

type metricKey [2]string

// prefix in metric namespace
var prefix = []string{nsVendor, nsClass, nsType}

// New returns snap-plugin-collector-disk instance
func New() (*DiskCollector, error) {
	dc := &DiskCollector{data: diskStats{stats: map[metricKey]uint64{}, timestamp: time.Now()},
		dataPrev: diskStats{stats: map[metricKey]uint64{}, timestamp: time.Now()},
		output:   map[metricKey]float64{},
		first:    true}
	return dc, nil
}

// Meta returns plugin meta data
func Meta() []plugin.MetaOpt {
	return []plugin.MetaOpt{
		plugin.ConcurrencyCount(1),
	}
}

// GetConfigPolicy returns a ConfigPolicy
func (dc *DiskCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule(prefix, "proc_path", false, plugin.SetDefaultString("/proc"))
	return *policy, nil
}

// GetMetricTypes returns list of exposed disk stats metrics
func (dc *DiskCollector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	mts := []plugin.Metric{}

	procFilePath, err := resolveSrcFile(cfg)
	if err != nil {
		return nil, err
	}

	if err := dc.getDiskStats(procFilePath); err != nil {
		return nil, err
	}

	// List of terminal metric names
	mList := make(map[string]bool)
	for key := range dc.data.stats {
		_, metricName := parseMetricKey(key)
		// Keep it if not already seen before
		if !mList[metricName] {
			mList[metricName] = true
			mts = append(mts, plugin.Metric{
				Namespace: plugin.NewNamespace(prefix...).
					AddDynamicElement("disk", "name of disk").
					AddStaticElement(metricName),
				Description: "dynamic disk metric: " + metricName,
			})
		}
	}
	return mts, nil
}

// CollectMetrics retrieves disk stats values for given metrics
func (dc *DiskCollector) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	procFilePath, err := resolveSrcFile(mts[0].Config)
	if err != nil {
		return nil, err
	}

	// dc.first equals true if collection is processed for the first time
	first := dc.first
	if first {
		// set dc.first to false for the next interval
		dc.first = false
	}

	// for the first collection skip stashing the previous data
	if !first {
		// stash disk data (dst, src)
		stashData(&dc.dataPrev, &dc.data)
	}

	// get presence disk stats
	if err := dc.getDiskStats(procFilePath); err != nil {
		return nil, err
	}

	//  for first collecting skip derivatives calculation
	if !first {
		// calculate derivatives based on data (presence) and previous one; results stored in dc.output
		if err := dc.calcDerivatives(); err != nil {
			return nil, err
		}
	}

	dc.calcGauge()

	for _, m := range mts {

		requestedDiskID, requestedMetric, err := parseNamespace(m.Namespace)
		if err != nil {
			return nil, err
		}

		if requestedDiskID == "*" {
			for key, value := range dc.output {
				diskID, metricName := parseMetricKey(key)

				if metricName == requestedMetric {
					// create a copy of incoming namespace and specify disk name
					ns := plugin.CopyNamespace(m.Namespace)
					ns[len(prefix)].Value = diskID

					metric := plugin.Metric{
						Namespace: ns,
						Data:      value,
						Timestamp: dc.data.timestamp,
						Version:   PluginVersion,
						Tags:      m.Tags,
					}
					metrics = append(metrics, metric)
				}

			}
		} else {
			// get this metric for specified disk (given explicitly)
			metricKey := createMetricKey(requestedDiskID, requestedMetric)
			if value, ok := dc.output[metricKey]; ok {
				metric := plugin.Metric{
					Namespace: m.Namespace,
					Data:      value,
					Timestamp: dc.data.timestamp,
					Version:   PluginVersion,
					Tags:      m.Tags,
				}
				metrics = append(metrics, metric)

			} else {
				log.Warning(fmt.Sprintf("Can not find metric value for %s", m.Namespace.Strings()))
			}
		}
	}

	return metrics, nil

}

// getDiskStats gets disk stats from file (/proc/{diskstats|partitions}) and stores them in the DiskCollector structure
func (dc *DiskCollector) getDiskStats(srcFile string) error {

	fieldshift := 0
	if path.Base(srcFile) == "partitions" {
		/* Kernel 2.4, Partition */
		fieldshift = 1
	}

	fh, err := os.Open(srcFile)
	defer fh.Close()

	if err != nil {
		return fmt.Errorf("Error opening /proc/{diskstats|partitions}, error = %s", err)
	}
	scanner := bufio.NewScanner(fh)
	dc.data.timestamp = time.Now()

	//map disk statistics keys (names) to scanned fields
	data := make(map[string]string)

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

		for metricName, val := range data {
			if value, err := strconv.ParseUint(val, 10, 64); err == nil {
				dc.data.stats[createMetricKey(diskName, metricName)] = value
			} else {
				// parse failure
				return fmt.Errorf("Error %v, invalid value of metric %s for disk %s: cannot convert `%v` to uint64", err, metricName, diskName, val)
			}
		}
	} // end of scanner.Scan()

	return nil
}

func (dc *DiskCollector) calcGauge() {
	for key, val := range dc.data.stats {
		if _, metric := parseMetricKey(key); strings.HasSuffix(metric, nPendingOps) {
			// for 'pending_ops' output value is simply stored as-is
			dc.output[key] = float64(val)
		}
	}
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
		case nPendingOps:
			// its a gauge metric - see calGauge()

		default:
			dc.output[key] = deriveVal
		}
	}

	// calculate disk time
	for disk, values := range avgDiskTime {
		dc.output[createMetricKey(disk, nTimeRead)] = calcTimeIncrement(values.diffReadTime, values.diffReadOps, interval)
		dc.output[createMetricKey(disk, nTimeWrite)] = calcTimeIncrement(values.diffWriteTime, values.diffWriteOps, interval)
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

// stashData copies DiskStats struct variables items with their values from 'src' to 'dst'
func stashData(dst *diskStats, src *diskStats) {
	dst.timestamp = src.timestamp

	// copy map, deep copy is needed
	for key, value := range src.stats {
		dst.stats[key] = value
	}
}

// parseNamespace returns extracted disk ID and metric key from a given namespace and true if raw metric is requested
func parseNamespace(ns plugin.Namespace) (string, string, error) {
	if len(ns.Strings()) <= len(prefix)+1 {
		return "", "", fmt.Errorf("Cannot parse a given namespace %s, it's too short (expected length > %d)", ns.Strings(), len(prefix))
	}

	// get the next element after prefix which is disk ID
	diskID := ns.Strings()[len(prefix)]

	// get the last element which is metric's name
	metricName := ns.Strings()[len(ns)-1]

	return diskID, metricName, nil
}

func isRawMetrics(ns plugin.Namespace) bool {
	if ns.Strings()[len(prefix)+1] == "raw" {
		return true
	}
	return false
}

// parseMetricKey extracts information about disk and metric name from metric key (exemplary metric key is `sda/time_write`)
func parseMetricKey(k metricKey) (disk, metric string) {
	return k[0], k[1]
}

// createMetricKey returns metric key which includes disk name and name of metric, exemplary metric key is `sda/time_write`
func createMetricKey(diskName string, metricName string) metricKey {
	return metricKey{diskName, metricName}
}

func resolveSrcFile(config plugin.Config) (string, error) {
	// first configuration
	if srcFile, err := config.GetString("proc_path"); err == nil {
		// diskstats
		diskstats := path.Join(srcFile, "diskstats")
		fh, err := os.Open(diskstats)
		if err == nil {
			fh.Close()
			return diskstats, nil
		}

		// partitions old kernel
		partitions := path.Join(srcFile, "partitions")
		fh, err = os.Open(partitions)
		if err == nil {
			fh.Close()
			return partitions, nil
		} else {
			return "", fmt.Errorf("Provided path to procfs diskstats/partitions is not correct {%s}", srcFile)
		}

	}
	// second default standard procfs
	if fh, err := os.Open(defaultSrcFile); err == nil {
		fh.Close()
		return defaultSrcFile, nil
	}

	// for default old kernel
	if fh, err := os.Open(defaultSrcFileOld); err == nil {
		fh.Close()
		return defaultSrcFileOld, nil
	}

	return "", fmt.Errorf("Could not find procfs disk stats file")
}
