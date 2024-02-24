package collector

import (
	"fmt"
	"regexp"
	"strings"
)

// convArtiBoolToProm is a very interesting appendix.
// Something needs it, but what and why?
// It would be quite nice if it was written here why such a thing is needed.
func convArtiBoolToProm(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

const (
	pattOneNumber = `^(?P<number>[[:digit:]]{1,3}(?:,[[:digit:]]{3})*)(?:\.[[:digit:]]{1,2})? ?(?P<multiplier>%|bytes|[KMGT]B)?$`
)

var (
	reOneNumber = regexp.MustCompile(pattOneNumber)
)

func (e *Exporter) convNumArtiToProm(artiNum string) (float64, error) {
	e.logger.Debug(
		"Attempting to convert a string from artifactory representing a number.",
		"artifactory.number", artiNum,
	)

	if !reOneNumber.MatchString(artiNum) {
		e.logger.Debug(
			"The arti number did not match known templates.",
			"artifactory.number", artiNum,
		)
		err := fmt.Errorf(
			`The string '%s' does not match pattern '%s'.`,
			artiNum,
			pattOneNumber,
		)
		return 0, err
	}

	groups := extractNamedGroups(artiNum, reOneNumber)

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
	pattTBytesPercent = `^(?P<tbytes>[[:digit:]]+(?:\.[[:digit:]]{1,2})?) TB \((?P<percent>[[:digit:]]{1,2}(?:\.[[:digit:]]{1,2})?)%\)$`
)

var (
	reTBytesPercent = regexp.MustCompile(pattTBytesPercent)
)

func (e *Exporter) convTwoNumsArtiToProm(artiSize string) (float64, float64, error) {
	e.logger.Debug(
		"Attempting to convert a string from artifactory representing a number.",
		"artifactory.number", artiSize,
	)

	if !reTBytesPercent.MatchString(artiSize) {
		e.logger.Debug(
			"The arti number did not match known templates.",
			"artifactory.number", artiSize,
		)
		err := fmt.Errorf(
			`The string '%s' does not match '%s' pattern.`,
			artiSize,
			pattTBytesPercent,
		)
		return 0, 0, err
	}
	groups := extractNamedGroups(artiSize, reTBytesPercent)
	b, err := e.convNumber(groups["tbytes"])
	if err != nil {
		return 0, 0, err
	}
	mulTB, _ := e.convMultiplier(`TB`)
	p, err := e.convNumber(groups["percent"])
	if err != nil {
		return 0, 0, err
	}
	mulPercent, _ := e.convMultiplier(`%`)
	return b * mulTB, p * mulPercent, nil
}
