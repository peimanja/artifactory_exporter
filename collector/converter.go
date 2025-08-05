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
	// pattFileStoreData matches file store data format like "1.5 TB (75.2%)"
	// Pattern breakdown:
	// - size: [[:digit:]]{1,3}(?:[[:digit:]]|(?:,[[:digit:]]{3})*(?:\.[[:digit:]]{1,2})?)? - matches numbers with optional commas and decimals
	// - [KMGT]B - matches unit (KB, MB, GB, TB)
	// - usage: (?:100|[1-9]?[0-9])(?:\.[0-9]{1,2})?% - matches 0-100% with optional decimals
	pattFileStoreData = `^(?P<size>[[:digit:]]{1,3}(?:[[:digit:]]|(?:,[[:digit:]]{3})*(?:\.[[:digit:]]{1,2})?)?) [KMGT]B \((?P<usage>(?:100|[1-9]?[0-9])(?:\.[0-9]{1,2})?%)\)$`
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
				"The string '%s' not recognisable as known artifactory filestore size: %w",
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
			`The string '%s' does not match '%s' pattern.`,
			artiSize,
			pattFileStoreData,
		)
		return 0, 0, err
	}
	groups := extractNamedGroups(artiSize, reFileStoreData)

	// Extract the unit from the original string
	sizeStr := groups["size"]
	
	// Find the unit (TB, GB, etc.) by looking at what comes after the size in the original string
	sizeIdx := strings.Index(artiSize, sizeStr)
	if sizeIdx == -1 {
		return 0, 0, fmt.Errorf("size string '%s' not found in input '%s'", sizeStr, artiSize)
	}
	
	spaceIdx := sizeIdx + len(sizeStr)
	if spaceIdx >= len(artiSize) || artiSize[spaceIdx] != ' ' {
		return 0, 0, fmt.Errorf("expected space after size string '%s' in input '%s'", sizeStr, artiSize)
	}
	
	unitStart := spaceIdx + 1
	if unitStart >= len(artiSize) {
		return 0, 0, fmt.Errorf("no unit found after size string '%s' in input '%s'", sizeStr, artiSize)
	}
	
	unitEnd := strings.Index(artiSize[unitStart:], " ")
	if unitEnd == -1 {
		return 0, 0, fmt.Errorf("could not find end of unit after size string '%s' in input '%s'", sizeStr, artiSize)
	}
	
	unit := artiSize[unitStart : unitStart+unitEnd]

	// Reconstruct the size with unit for proper conversion
	sizeWithUnit := sizeStr + " " + unit
	size, err := e.convArtiToPromNumber(sizeWithUnit)
	if err != nil {
		return 0, 0, err
	}
	usage, err := e.convArtiToPromNumber(groups["usage"])
	if err != nil {
		return 0, 0, err
	}
	return size, usage, nil
}
