package collector

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	logDbgKeyArtNum = `artifactory.number`
)

// convArtiToPromBool is a very interesting appendix.
// Something needs it, but what and why?
// It would be quite nice if it was written here why such a thing is needed.
func convArtiToPromBool(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

const (
	pattNumber = `^(?P<number>[[:digit:]]{1,3}(?:[[:digit:]]|(?:,[[:digit:]]{3})*(?:\.[[:digit:]]{1,2})?)?) ?(?P<multiplier>%|bytes|[KMGT]B)?$`
)

var (
	reNumber = regexp.MustCompile(pattNumber)
)

func (e *Exporter) convArtiToPromNumber(artiNum string) (float64, error) {
	e.logger.Debug(
		"Attempting to convert a string from artifactory representing a number.",
		logDbgKeyArtNum, artiNum,
	)

	if !reNumber.MatchString(artiNum) {
		e.logger.Debug(
			"The arti number did not match known templates.",
			logDbgKeyArtNum, artiNum,
		)
		err := fmt.Errorf(
			`the string '%s' does not match pattern '%s'.`,
			artiNum,
			pattNumber,
		)

		return 0, err
	}

	groups := extractNamedGroups(artiNum, reNumber)

	// The following `strings.replace` is for those cases that contain a comma
	// thousands separator.  In other cases, unnecessary, but cheaper than if.
	// Sorry.
	f, err := e.convNumber(
		strings.Replace(groups["number"], `,`, ``, -1),
	)
	if err != nil {
		return 0, err
	}

	mAsString, mIsPresent := groups["multiplier"]
	if !mIsPresent {
		return f, nil
	}
	m, err := e.convMultiplier(mAsString)
	if err != nil {
		return 0, err
	}

	return f * m, nil
}

const (
	// Sub-patterns for readability and maintainability
	pattSize  = `[[:digit:]]{1,3}(?:[[:digit:]]|(?:,[[:digit:]]{3})*(?:\.[[:digit:]]{1,2})?)?`
	pattUnit  = `[KMGT]B`
	pattUsage = `(?:100|[1-9]?[0-9])(?:\.[0-9]{1,2})?%`

	// pattFileStoreData matches file store data format like "1.5 TB (75.2%)"
	// Pattern breakdown:
	// - size: matches numbers with optional commas and decimals
	// - unit: matches KB, MB, GB, TB
	// - usage: matches 0-100% with optional decimals
	pattFileStoreData = `^(?P<size>` + pattSize + `) ` + pattUnit + ` \((?P<usage>` + pattUsage + `)\)$`
)

var (
	reFileStoreData = regexp.MustCompile(pattFileStoreData)
)

// convArtiToPromFileStoreData tries to interpret the string from artifactory
// as filestore data.
// Usually the inscription has two parts. Size and percentage of use. However,
// it happens that artifactory only gives the size.
// Please look at the cases in the unit test `TestConvFileStoreData`.
func (e *Exporter) convArtiToPromFileStoreData(artiSize string) (float64, float64, error) {
	e.logger.Debug(
		"Attempting to convert a string from artifactory representing a file store data.",
		logDbgKeyArtNum, artiSize,
	)

	if !strings.Contains(artiSize, `%`) {
		b, err := e.convArtiToPromNumber(artiSize)
		if err != nil {
			return 0, 0, fmt.Errorf(
				"the string '%s' not recognisable as known artifactory filestore size: %w",
				artiSize,
				err,
			)
		}
		return b, 0, nil
	}

	if !reFileStoreData.MatchString(artiSize) {
		e.logger.Debug(
			fmt.Sprintf(
				"The arti number did not match template '%s'.",
				pattFileStoreData,
			),
			logDbgKeyArtNum, artiSize,
		)
		err := fmt.Errorf(
			"the string '%s' does not match '%s' pattern",
			artiSize,
			pattFileStoreData,
		)
		return 0, 0, err
	}

	groups := extractNamedGroups(artiSize, reFileStoreData)

	// Extract size and usage using regex groups instead of manual parsing
	sizeStr := groups["size"]
	usageStr := groups["usage"]

	// Extract unit by finding what comes between size and opening parenthesis
	// This is more reliable than manual string manipulation
	sizeIdx := strings.Index(artiSize, sizeStr)
	if sizeIdx == -1 {
		return 0, 0, fmt.Errorf("size string '%s' not found in input '%s'", sizeStr, artiSize)
	}

	parenIdx := strings.Index(artiSize, "(")
	if parenIdx == -1 {
		return 0, 0, fmt.Errorf("opening parenthesis not found in input '%s'", artiSize)
	}

	// Extract the unit from between size and parenthesis, trimming spaces
	unitPart := strings.TrimSpace(artiSize[sizeIdx+len(sizeStr) : parenIdx])

	// Reconstruct the size with unit for proper conversion
	sizeWithUnit := sizeStr + " " + unitPart
	size, err := e.convArtiToPromNumber(sizeWithUnit)
	if err != nil {
		return 0, 0, err
	}
	usage, err := e.convArtiToPromNumber(usageStr)
	if err != nil {
		return 0, 0, err
	}
	return size, usage, nil
}
