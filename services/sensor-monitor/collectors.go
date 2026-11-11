package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lite-nas/shared/metrics"
)

// Collector reads one family of host sources and returns its obtained readings.
// A collector may return both readings and an error when some sources fail.
type Collector interface {
	Collect(context.Context) (metrics.SensorSnapshot, error)
}

// LinuxCollector reads standard hwmon, thermal, and CPU frequency interfaces.
type LinuxCollector struct {
	hwmonRoot   string
	thermalRoot string
	clockRoot   string
	names       NameCatalog
}

// NewLinuxCollector binds standard Linux source roots and a readable-name catalog.
func NewLinuxCollector(names NameCatalog) LinuxCollector {
	return LinuxCollector{
		hwmonRoot: "/sys/class/hwmon", thermalRoot: "/sys/class/thermal",
		clockRoot: "/sys/devices/system/cpu/cpufreq", names: names,
	}
}

// Collect reads all present standard Linux sensors without requiring every
// host to expose every category.
func (collector LinuxCollector) Collect(_ context.Context) (metrics.SensorSnapshot, error) {
	snapshot := emptySnapshot()
	var failures []error
	if err := collector.collectHwmon(&snapshot); err != nil {
		failures = append(failures, err)
	}
	if err := collector.collectThermal(&snapshot); err != nil {
		failures = append(failures, err)
	}
	if err := collector.collectClocks(&snapshot); err != nil {
		failures = append(failures, err)
	}
	return snapshot, errors.Join(failures...)
}

// collectHwmon discovers available input attributes and converts their kernel units.
func (collector LinuxCollector) collectHwmon(snapshot *metrics.SensorSnapshot) error {
	chips, err := os.ReadDir(collector.hwmonRoot)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read hwmon: %w", err)
	}
	var failures []error
	for _, chip := range chips {
		chipPath := filepath.Join(collector.hwmonRoot, chip.Name())
		chipName := readOptionalText(filepath.Join(chipPath, "name"))
		if chipName == "" {
			chipName = "hwmon"
		}
		deviceKey := hwmonDeviceKey(chipPath, chipName)
		entries, readErr := os.ReadDir(chipPath)
		if readErr != nil {
			failures = append(failures, fmt.Errorf("read hwmon chip %s: %w", chipName, readErr))
			continue
		}
		for _, entry := range entries {
			category, unit, divisor := hwmonAttribute(entry.Name())
			if category == "" {
				continue
			}
			value, valueErr := readNumeric(filepath.Join(chipPath, entry.Name()), divisor)
			if valueErr != nil {
				failures = append(failures, fmt.Errorf("read %s/%s: %w", chipName, entry.Name(), valueErr))
				continue
			}
			key := chipName + ":" + deviceKey + ":" + entry.Name()
			labelPath := filepath.Join(chipPath, strings.TrimSuffix(entry.Name(), "_input")+"_label")
			reading := metrics.SensorNumericReading{
				Key: key, Name: collector.names.Resolve(category, key, readOptionalText(labelPath)),
				Value: value, Unit: unit,
			}
			switch category {
			case "temperature":
				snapshot.Temperatures = append(snapshot.Temperatures, reading)
			case "voltage":
				snapshot.Voltages = append(snapshot.Voltages, reading)
			case "fan":
				snapshot.Fans = append(snapshot.Fans, reading)
			}
		}
	}
	return errors.Join(failures...)
}

// hwmonDeviceKey excludes volatile hwmonN indexes when deriving source identity.
func hwmonDeviceKey(chipPath, chipName string) string {
	devicePath, err := filepath.EvalSymlinks(filepath.Join(chipPath, "device"))
	if err == nil {
		return strings.TrimPrefix(devicePath, "/sys/devices/")
	}
	resolved, err := filepath.EvalSymlinks(chipPath)
	if err == nil && filepath.Base(filepath.Dir(resolved)) == "hwmon" {
		return strings.TrimPrefix(filepath.Dir(filepath.Dir(resolved)), "/sys/devices/")
	}
	return chipName
}

// hwmonAttribute recognizes read-only sensor inputs and their conversion units.
func hwmonAttribute(name string) (category, unit string, divisor float64) {
	if !strings.HasSuffix(name, "_input") {
		return "", "", 0
	}
	stem := strings.TrimSuffix(name, "_input")
	for _, sensor := range []struct {
		prefix, category, unit string
		divisor                float64
	}{
		{"temp", "temperature", "°C", 1000},
		{"in", "voltage", "V", 1000},
		{"fan", "fan", "rpm", 1},
	} {
		if !strings.HasPrefix(stem, sensor.prefix) {
			continue
		}
		if _, err := strconv.Atoi(strings.TrimPrefix(stem, sensor.prefix)); err == nil {
			return sensor.category, sensor.unit, sensor.divisor
		}
	}
	return "", "", 0
}

// collectThermal supplies stable thermal-zone readings alongside hwmon sensors.
func (collector LinuxCollector) collectThermal(snapshot *metrics.SensorSnapshot) error {
	zones, err := filepath.Glob(filepath.Join(collector.thermalRoot, "thermal_zone*"))
	if err != nil {
		return fmt.Errorf("discover thermal zones: %w", err)
	}
	var failures []error
	for _, zone := range zones {
		value, readErr := readNumeric(filepath.Join(zone, "temp"), 1000)
		if readErr != nil {
			failures = append(failures, fmt.Errorf("read thermal zone %s: %w", filepath.Base(zone), readErr))
			continue
		}
		kind := readOptionalText(filepath.Join(zone, "type"))
		if kind == "" {
			kind = filepath.Base(zone)
		}
		key := "thermal:" + kind
		snapshot.Temperatures = append(snapshot.Temperatures, metrics.SensorNumericReading{
			Key: key, Name: collector.names.Resolve("temperature", key, ""), Value: value, Unit: "°C",
		})
	}
	return errors.Join(failures...)
}

// collectClocks reads current frequencies exposed by CPU frequency policies.
func (collector LinuxCollector) collectClocks(snapshot *metrics.SensorSnapshot) error {
	policies, err := filepath.Glob(filepath.Join(collector.clockRoot, "policy*"))
	if err != nil {
		return fmt.Errorf("discover CPU clock policies: %w", err)
	}
	var failures []error
	for _, policy := range policies {
		value, readErr := readNumeric(filepath.Join(policy, "scaling_cur_freq"), 1000)
		if readErr != nil {
			failures = append(failures, fmt.Errorf("read %s clock: %w", filepath.Base(policy), readErr))
			continue
		}
		key := "cpu:" + filepath.Base(policy)
		snapshot.Clocks = append(snapshot.Clocks, metrics.SensorNumericReading{
			Key: key, Name: collector.names.Resolve("clock", key, ""), Value: value, Unit: "MHz",
		})
	}
	return errors.Join(failures...)
}

// readNumeric reads one kernel numeric attribute and converts its native unit.
func readNumeric(path string, divisor float64) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		if err == nil {
			err = errors.New("non-finite reading")
		}
		return 0, err
	}
	return value / divisor, nil
}

// readOptionalText returns a source label when the optional attribute exists.
func readOptionalText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// RaspberryPiCollector reads firmware-only supply and throttling information.
type RaspberryPiCollector struct {
	command       string
	names         NameCatalog
	pmicAvailable bool
}

// NewRaspberryPiCollector enables the fixed, read-only firmware operations when
// vcgencmd and its device are available to the service account.
func NewRaspberryPiCollector(names NameCatalog) *RaspberryPiCollector {
	const command = "/usr/bin/vcgencmd"
	info, err := os.Stat(command)
	if err != nil || info.Mode()&0o111 == 0 {
		return nil
	}
	if _, err := os.Stat("/dev/vcio_gencmd"); err != nil {
		if _, err := os.Stat("/dev/vcio"); err != nil {
			return nil
		}
	}
	model := readOptionalText("/sys/firmware/devicetree/base/model")
	return &RaspberryPiCollector{
		command: command, names: names,
		pmicAvailable: strings.Contains(model, "Raspberry Pi 5") || strings.Contains(model, "Compute Module 5"),
	}
}

// Collect obtains the supply voltage and current/latched firmware flags.
func (collector RaspberryPiCollector) Collect(ctx context.Context) (metrics.SensorSnapshot, error) {
	snapshot := emptySnapshot()
	var failures []error
	if collector.pmicAvailable {
		voltageOutput, err := collector.run(ctx, "pmic")
		if err == nil {
			if value, parseErr := parsePMICVoltage(voltageOutput); parseErr == nil {
				snapshot.Voltages = append(snapshot.Voltages, metrics.SensorNumericReading{
					Key: "EXT5V_V", Name: collector.names.Resolve("voltage", "EXT5V_V", ""),
					Value: value, Unit: "V",
				})
			} else {
				failures = append(failures, parseErr)
			}
		} else {
			failures = append(failures, err)
		}
	}
	throttleOutput, err := collector.run(ctx, "throttling")
	if err == nil {
		if flags, parseErr := parseThrottleFlags(throttleOutput, collector.names); parseErr == nil {
			snapshot.Throttling = append(snapshot.Throttling, flags...)
		} else {
			failures = append(failures, parseErr)
		}
	} else {
		failures = append(failures, err)
	}
	return snapshot, errors.Join(failures...)
}

// run accepts only the two read-only firmware commands used by this service.
func (collector RaspberryPiCollector) run(ctx context.Context, operation string) (string, error) {
	var args []string
	switch operation {
	case "pmic":
		args = []string{"pmic_read_adc", "EXT5V_V"}
	case "throttling":
		args = []string{"get_throttled"}
	default:
		return "", fmt.Errorf("unsupported vcgencmd operation %q", operation)
	}
	commandCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, collector.command, args...).Output()
	if err != nil {
		return "", fmt.Errorf("vcgencmd %s: %w", args[0], err)
	}
	return string(output), nil
}

// parsePMICVoltage accepts only the explicitly requested PMIC voltage channel.
func parsePMICVoltage(output string) (float64, error) {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[0] != "EXT5V_V" || !strings.HasPrefix(fields[1], "volt(") {
			continue
		}
		_, amount, ok := strings.Cut(fields[1], "=")
		if !ok {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSuffix(amount, "V"), 64)
		if err == nil && value > 0 {
			return value, nil
		}
	}
	return 0, errors.New("EXT5V_V voltage unavailable or malformed")
}

// parseThrottleFlags decodes the eight documented current and occurred flags.
func parseThrottleFlags(output string, names NameCatalog) ([]metrics.SensorStateReading, error) {
	raw := strings.TrimSpace(output)
	valueText, ok := strings.CutPrefix(raw, "throttled=0x")
	if !ok {
		return nil, errors.New("malformed get_throttled response")
	}
	value, err := strconv.ParseUint(valueText, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("parse throttling flags: %w", err)
	}
	keys := []string{
		"under_voltage_now", "frequency_capped_now", "throttled_now", "soft_temperature_limit_now",
		"under_voltage_occurred", "frequency_capped_occurred", "throttled_occurred", "soft_temperature_limit_occurred",
	}
	bits := []uint{0, 1, 2, 3, 16, 17, 18, 19}
	readings := make([]metrics.SensorStateReading, 0, len(keys))
	for index, key := range keys {
		readings = append(readings, metrics.SensorStateReading{
			Key: key, Name: names.Resolve("throttling", key, ""), Active: value&(1<<bits[index]) != 0,
		})
	}
	return readings, nil
}

// emptySnapshot initializes all sections so absent readings encode as arrays.
func emptySnapshot() metrics.SensorSnapshot {
	return metrics.SensorSnapshot{
		Temperatures: []metrics.SensorNumericReading{}, Voltages: []metrics.SensorNumericReading{},
		Fans: []metrics.SensorNumericReading{}, Clocks: []metrics.SensorNumericReading{},
		Throttling: []metrics.SensorStateReading{},
	}
}
