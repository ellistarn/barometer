package barometer

import (
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ellistarn/barometer/pkg/apis/v1alpha1"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
)

const (
	CPUPressureFile    = "cpu.pressure"
	MemoryPressureFile = "memory.pressure"
	IOPressureFile     = "io.pressure"
	SomeStall          = "some"
	FullStall          = "full"
)

func GetPSI(threshold *v1alpha1.PSI, pod *v1.Pod, containerStatus *v1.ContainerStatus) (*v1alpha1.PSI, error) {
	if threshold == nil {
		return nil, nil
	}
	cgroupDir := fmt.Sprintf("/sys/fs/cgroup/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod%s.slice/%s.scope",
		strings.ReplaceAll(string(pod.UID), "-", "_"),
		strings.ReplaceAll(containerStatus.ContainerID, "://", "-"),
	)
	psi := &v1alpha1.PSI{}
	var err error
	psi.CPU, err = getStallMetrics(threshold.CPU, path.Join(cgroupDir, CPUPressureFile))
	if err != nil {
		return nil, err
	}
	psi.Memory, err = getStallMetrics(threshold.Memory, path.Join(cgroupDir, MemoryPressureFile))
	if err != nil {
		return nil, err
	}
	psi.IO, err = getStallMetrics(threshold.IO, path.Join(cgroupDir, IOPressureFile))
	if err != nil {
		return nil, err
	}
	return psi, nil
}

// Expected format
//
// some avg10=0.00 avg60=0.00 avg300=0.00 total=0
// full avg10=0.00 avg60=0.00 avg300=0.00 total=0
// newline
func getStallMetrics(threshold *v1alpha1.StallMetrics, file string) (*v1alpha1.StallMetrics, error) {
	if threshold == nil {
		return nil, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading '%s', %w", file, err)
	}
	parts := strings.Split(string(data), "\n")
	if len(parts) != 3 {
		return &v1alpha1.StallMetrics{}, fmt.Errorf("invalid pressure format '%s'", string(data))
	}
	stallMetrics := &v1alpha1.StallMetrics{}
	if stallMetrics.Some, err = parseStallMetric(threshold.Some, SomeStall, parts[0]); err != nil {
		return &v1alpha1.StallMetrics{}, err
	}
	if stallMetrics.Full, err = parseStallMetric(threshold.Full, FullStall, parts[1]); err != nil {
		return &v1alpha1.StallMetrics{}, err
	}
	return stallMetrics, nil
}

func parseStallMetric(theshold *v1alpha1.StallMetric, stallType string, s string) (*v1alpha1.StallMetric, error) {
	if theshold == nil {
		return nil, nil
	}
	parts := strings.Split(s, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid metric format %q", s)
	}
	if parts[0] != stallType {
		return nil, fmt.Errorf("unexpected stall type %q in %q", stallType, s)
	}
	for _, metricValue := range []string{parts[1], parts[2], parts[3], parts[4]} {
		parts := strings.Split(metricValue, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metric format %q", metricValue)
		}
		if _, err := strconv.ParseFloat(parts[1], 64); err != nil {
			return nil, fmt.Errorf("invalid metric value format %q", metricValue)
		}
	}

	stallMetric := &v1alpha1.StallMetric{}

	if avg10, err := parseMetricValue(parts[1]); err != nil {
		return nil, err
	} else if theshold.Avg10 != nil && avg10 > *theshold.Avg10 {
		stallMetric.Avg10 = &avg10
	}

	if avg60, err := parseMetricValue(parts[2]); err != nil {
		return nil, err
	} else if theshold.Avg60 != nil && avg60 > *theshold.Avg60 {
		stallMetric.Avg60 = &avg60
	}

	if avg300, err := parseMetricValue(parts[3]); err != nil {
		return nil, err
	} else if theshold.Avg300 != nil && avg300 > *theshold.Avg300 {
		stallMetric.Avg300 = &avg300
	}
	return stallMetric, nil
}

func parseMetricValue(metricValue string) (int32, error) {
	parts := strings.Split(metricValue, "=")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid metric format %q", metricValue)
	}
	if _, err := strconv.ParseFloat(parts[1], 64); err != nil {
		return 0, fmt.Errorf("invalid metric value format %q", metricValue)
	}
	return int32(math.Round(lo.Must(strconv.ParseFloat(strings.Split(metricValue, "=")[1], 64)))), nil
}
