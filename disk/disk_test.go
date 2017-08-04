// +build small

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
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	invalidValue = iota
	invalidEntry = iota
	invalidMinor = iota
)

var (
	mockMts = []plugin.Metric{
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda", "ops_read"),
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda", "ops_write"),
		},

		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda", "octets_read"),
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda", "octets_write"),
		},

		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda1", "ops_read"),
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda1", "ops_write"),
		},

		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda1", "octets_read"),
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "procfs", "disk", "test_sda1", "octets_write"),
		},
	}

	srcMockFile           = "/tmp/diskstats_mock"
	srcMockFileNext       = "/tmp/diskstats_mock_next"
	srcMockFileOldVer     = "/tmp/partitions_mock"
	srcMockFileInv        = "/tmp/invalid_mock"
	bkupDefaultSrcFile    string
	bkupDefaultSrcFileOld string
)

func TestMain(m *testing.M) {
	PrepareTests()
	ret := m.Run()
	TeardownTests()
	os.Exit(ret)
}

func PrepareTests() {
	bkupDefaultSrcFile = defaultSrcFile
	bkupDefaultSrcFileOld = defaultSrcFileOld
}

func TeardownTests() {
	defaultSrcFile = bkupDefaultSrcFile
	defaultSrcFileOld = bkupDefaultSrcFileOld
}

func TestGetConfigPolicy(t *testing.T) {
	Convey("normal case", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})
		So(func() { diskPlugin.GetConfigPolicy() }, ShouldNotPanic)
		_, err = diskPlugin.GetConfigPolicy()
		So(err, ShouldBeNil)
	})
}

func TestGetMetricTypes(t *testing.T) {
	defaultSrcFile = srcMockFile
	defaultSrcFileOld = srcMockFileOldVer
	var cfg plugin.Config

	createMockFiles()

	Convey("source file available, kernel 2.6+", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		So(func() { diskPlugin.GetMetricTypes(cfg) }, ShouldNotPanic)
		results, err := diskPlugin.GetMetricTypes(cfg)

		So(err, ShouldBeNil)
		So(results, ShouldNotBeEmpty)
		// 8 devices/partitions (sda, sda1, sda2, sda3, sdb, sdb1, sdb2) and for each 11 extended stats gives 88 metrics
		// but has these are now dynamic only 11 entries are returned
		So(len(results), ShouldEqual, 11)

	})

	Convey("source file available, kernel older than 2.6", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		os.Remove(defaultSrcFile) // remove srcFile for kernel 2.6+
		results, err := diskPlugin.GetMetricTypes(cfg)

		So(err, ShouldBeNil)
		So(results, ShouldNotBeEmpty)
		// 8 devices/partitions (sda, sda1, sda2, sda3, sdb, sdb1, sdb2) and for each 4 stats gives 32 metrics
		// but has these are now dynamic only 4 entries are returned
		So(len(results), ShouldEqual, 4)
	})

	Convey("source files not available", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		deleteMockFiles()
		results, err := diskPlugin.GetMetricTypes(cfg)
		So(err, ShouldNotBeNil)
		So(results, ShouldBeNil)
	})

	Convey("invalid syntax of source file", t, func() {
		defaultSrcFile = srcMockFileInv

		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		Convey("invalid value, cannot convert to uint64", func() {
			createInvalidMockFile(invalidValue)
			results, err := diskPlugin.GetMetricTypes(cfg)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeNil)
		})

		Convey("unknown entry, ignore it", func() {
			createInvalidMockFile(invalidEntry)
			results, err := diskPlugin.GetMetricTypes(cfg)

			So(results, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("invalid device minor number", func() {
			createInvalidMockFile(invalidMinor)
			results, err := diskPlugin.GetMetricTypes(cfg)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeNil)
		})
	})
}

func TestCollectMetrics(t *testing.T) {
	defaultSrcFile = srcMockFile
	defaultSrcFileOld = srcMockFileOldVer
	createMockFiles()

	Convey("source file available", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		results, err := diskPlugin.CollectMetrics(mockMts)
		Convey("first data collecting", func() {
			So(err, ShouldBeNil)
			So(results, ShouldBeEmpty)
		})

		Convey("next data collecting", func() {

			Convey("no data change", func() {
				results, err := diskPlugin.CollectMetrics(mockMts)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
				So(len(results), ShouldEqual, len(mockMts))
				for _, r := range results {
					So(r.Data, ShouldEqual, 0)
				}
			})

			Convey("change values of data", func() {
				defaultSrcFile = srcMockFileNext
				results, err := diskPlugin.CollectMetrics(mockMts)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
				So(len(results), ShouldEqual, len(mockMts))

				for _, r := range results {
					So(r.Data, ShouldNotEqual, 0)
				}
			})
		})
	})

	Convey("invalid calculation of derivatives", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		diskPlugin.first = false
		Convey("no previous data", func() {
			results, err := diskPlugin.CollectMetrics(mockMts)

			So(err, ShouldNotBeNil)
			So(results, ShouldBeNil)
		})

	})

	deleteMockFiles()
	Convey("source files not available", t, func() {
		diskPlugin, err := New()
		Convey("new disk collector", func() {
			So(err, ShouldBeNil)
		})

		results, err := diskPlugin.CollectMetrics(mockMts)
		So(err, ShouldNotBeNil)
		So(results, ShouldBeNil)
	})

}

func createMockFiles() {
	deleteMockFiles()
	// 	mocked content of srcMockFile (kernel 2.6+)
	srcMockFileCont := []byte(`    8       0 test_sda 1546 2303 12644 114 0 0 0 0 0 68 114
								   8       1 test_sda1 333 8 2728 28 0 0 0 0 0 28 28
								   8       2 test_sda2 330 2264 2604 20 0 0 0 0 0 20 20
								   8       3 test_sda3 328 0 2624 29 0 0 0 0 0 29 29
								   8       4 test_sda4 325 0 2600 25 0 0 0 0 0 25 25
								   8      16 test_sdb 197890 10231 6405324 76885 11345881 15065577 592035256 7904803 0 1237705 7973111
								   8      17 test_sdb1 79104 984 3301786 23605 8288060 13802816 533998912 6032422 0 1162267 6047374
								   8      18 test_sdb2 118610 9212 3101850 53254 2979961 1262761 58036344 1859439 0 106811 1918865`)

	// 	mocked content of srcMockFileNext (kernel 2.6+)	in next collection
	srcMockFileNextCont := []byte(`8       0 test_sda 1541 2304 12645 115 1 1 1 1 1 69 115
								   8       1 test_sda1 433 0 0 38 10 10 10 5 5 40 49
								   8       2 test_sda2 335 2265 2605 21 1 1 1 1 1 21 21
								   8       3 test_sda3 325 1 2621 30 1 1 1 1 1 35 35
								   8       4 test_sda4 327 8 2667 26 2 2 2 2 2 66 67
								   8      16 test_sdb 197892 10232 6405354 76685 11645881 15066577 592065256 7906803 0 1237805 7973811
								   8      17 test_sdb1 79194 994 3301796 23665 8288070 13802818 533998915 6032426 6 1162287 6047378
								   8      18 test_sdb2 118690 9912 3101860 53256 29710061 1262766 58036346 1859478 4 106881 1918868`)

	// 	mocked content of srcMockFileOldVer (kernel < 2.6)
	srcMockFileOldVerCont := []byte(`   8       0 test_sda 1546 12644 0 0 	
								   8       1 test_sda1 333 2728 0 0								
								   8       2 test_sda2 330 2604 0 0								
								   8       3 test_sda3 328 2624 0 0								
								   8       4 test_sda4 325 2600 0 0								
								   8      16 test_sdb 197890 6405324 11345881 592035256								
								   8      17 test_sdb1 79104 3301786 8288060  533998912								
								   8      18 test_sdb2 118610 3101850 2979961 58036344`)

	f, _ := os.Create(srcMockFile)
	f.Write(srcMockFileCont)

	f, _ = os.Create(srcMockFileNext)
	f.Write(srcMockFileNextCont)

	f, _ = os.Create(srcMockFileOldVer)
	f.Write(srcMockFileOldVerCont)
}

func createInvalidMockFile(kind int) {
	os.Remove(srcMockFileInv)

	var srcMockFileContInv []byte

	switch kind {
	case invalidValue:
		srcMockFileContInv = []byte(`    8       0 test_sda 0 100 ccc 200`)
		break

	case invalidEntry:
		srcMockFileContInv = []byte(`    1       2 unknown entry`)
		break

	case invalidMinor:
		srcMockFileContInv = []byte(`    1       A test_sda 0 100 200 300`)
		break

	default:
		srcMockFileContInv = []byte(``)
		break

	}

	f, _ := os.Create(srcMockFileInv)
	f.Write(srcMockFileContInv)

}

func deleteMockFiles() {
	os.Remove(srcMockFile)
	os.Remove(srcMockFileOldVer)
	os.Remove(srcMockFileInv)
	os.Remove(srcMockFileNext)
}
