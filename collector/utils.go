package collector

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	mulKB = 1024
	mulMB = mulKB * 1024
	mulGB = mulMB * 1024
	mulTB = mulGB * 1024
)

// A one-time regex compilation is far cheaper
// than repeating on each function entry.
// `MustCompile` is used because regex value is hardcoded
// i.e. may have been previously verified by the author.
var regNumber = regexp.MustCompile("[^0-9.]+")

func (e *Exporter) removeCommas(str string) (float64, error) {
	e.logger.Debug("Removing other characters to extract number from string")
	/*
	 * I am very concerned about the “magic” used here.
	 * The code does not in any way explain why this particular
	 * method of extracting content from the text was adopted.
	 * Kacper Perschke
	 */
	strArray := strings.Fields(str)
	strTrimmed := regNumber.ReplaceAllString(strArray[0], "")
	num, err := strconv.ParseFloat(strTrimmed, 64)
	if err != nil {
		e.logger.Debug(
			"Had problem extracting number",
			"string", str,
			"err", err.Error(),
		)
		return 0, err
	}
	e.logger.Debug(
		"Successfully converted string to number",
		"string", str,
		"number", num,
	)
	return num, nil
}

func (e *Exporter) bytesConverter(str string) (float64, error) {
	e.logger.Debug("Converting size to bytes")
	num, err := e.removeCommas(str)
	if err != nil {
		return 0, err
	}
	var bytesValue float64
	if strings.Contains(str, "bytes") {
		bytesValue = num
	} else if strings.Contains(str, "KB") {
		bytesValue = num * mulKB
	} else if strings.Contains(str, "MB") {
		bytesValue = num * mulMB
	} else if strings.Contains(str, "GB") {
		bytesValue = num * mulGB
	} else if strings.Contains(str, "TB") {
		bytesValue = num * mulTB
	} else {
		return 0, fmt.Errorf("Could not convert %s to bytes", str)
	}
	e.logger.Debug(
		"Successfully converted string to bytes",
		"string", str,
		"value", bytesValue,
	)
	return bytesValue, nil
}

// b2f is a very interesting appendix.
// Something needs it, but what and why?
// It would be quite nice if it was written here why such a thing is needed.
func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
