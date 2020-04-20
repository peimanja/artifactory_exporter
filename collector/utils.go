package collector

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log/level"
)

func (e *Exporter) removeCommas(str string) (float64, error) {
	level.Debug(e.logger).Log("msg", "Removing other characters to extract number from string")
	reg, err := regexp.Compile("[^0-9.]+")
	if err != nil {
		return 0, err
	}
	strArray := strings.Fields(str)
	convertedStr, err := strconv.ParseFloat(reg.ReplaceAllString(strArray[0], ""), 64)
	if err != nil {
		return 0, err
	}
	level.Debug(e.logger).Log("msg", "Successfully converted string to number", "string", str, "number", convertedStr)
	return convertedStr, nil
}

func (e *Exporter) bytesConverter(str string) (float64, error) {
	var bytesValue float64
	level.Debug(e.logger).Log("msg", "Converting size to bytes")
	num, err := e.removeCommas(str)
	if err != nil {
		return 0, err
	}

	if strings.Contains(str, "bytes") {
		bytesValue = num
	} else if strings.Contains(str, "KB") {
		bytesValue = num * 1024
	} else if strings.Contains(str, "MB") {
		bytesValue = num * 1024 * 1024
	} else if strings.Contains(str, "GB") {
		bytesValue = num * 1024 * 1024 * 1024
	} else if strings.Contains(str, "TB") {
		bytesValue = num * 1024 * 1024 * 1024 * 1024
	} else {
		return 0, fmt.Errorf("Could not convert %s to bytes", str)
	}
	level.Debug(e.logger).Log("msg", "Successfully converted string to bytes", "string", str, "value", bytesValue)
	return bytesValue, nil
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
