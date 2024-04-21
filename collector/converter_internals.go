package collector

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var mulConvDriver = map[string]float64{
	`%`:     0.01,
	`bytes`: 1,
	`KB`:    math.Exp2(10),
	`MB`:    math.Exp2(20),
	`GB`:    math.Exp2(30),
	`TB`:    math.Exp2(40),
}

func (e *Exporter) convMultiplier(m string) (float64, error) {
	mul, present := mulConvDriver[m]
	if present {
		return mul, nil
	}
	e.logger.Error(
		"The string was not recognized as a known multiplier.",
		"artifactory.number.multiplier", m,
	)
	return 0, fmt.Errorf(`Could not recognise '%s' as multiplier`, m)
}

func (e *Exporter) convNumber(n string) (float64, error) {
	f, err := strconv.ParseFloat(n, 64)
	if err != nil {
		e.logger.Error(
			"String not convertible to float64",
			"string", n,
		)
		return 0, err
	}
	return f, nil
}

type reCaptureGroups map[string]string

func extractNamedGroups(artiNum string, re *regexp.Regexp) reCaptureGroups {
	match := re.FindStringSubmatch(artiNum)
	groupsFound := make(reCaptureGroups)
	for i, name := range re.SubexpNames() {
		groupCaptured := match[i]
		if i != 0 && name != `` && groupCaptured != `` {
			groupsFound[name] = groupCaptured
		}
	}
	return groupsFound
}
