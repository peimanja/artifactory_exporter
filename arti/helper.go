package arti

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type UsersCount struct {
	Count float64
	Realm string
}

func CheckErr(err error) {
	if err != nil {
		log.Errorln(err)
	}
}

func RemoveCommas(str string) float64 {

	reg, err := regexp.Compile("[^0-9.]+")
	CheckErr(err)
	convertedStr, err := strconv.ParseFloat(reg.ReplaceAllString(str, ""), 64)
	CheckErr(err)

	return convertedStr
}

func BytesConverter(str string) (float64, error) {
	type errorString struct {
		s string
	}
	num := RemoveCommas(str)

	if strings.Contains(str, "bytes") {
		return num, nil
	} else if strings.Contains(str, "KB") {
		return num * 1024, nil
	} else if strings.Contains(str, "MB") {
		return num * 1024 * 1024, nil
	} else if strings.Contains(str, "GB") {
		return num * 1024 * 1024 * 1024, nil
	} else if strings.Contains(str, "TB") {
		return num * 1024 * 1024 * 1024 * 1024, nil
	}
	return 0, fmt.Errorf("Could not convert %s to bytes", str)
}

func FromEpoch(t time.Duration) int64 {

	now := time.Now()
	after := now.Add(-t * time.Minute)
	nanos := after.UnixNano()
	millis := nanos / 1000000

	return millis
}

func CountUsers(users []User) []UsersCount {
	userCount := make([]UsersCount, 2)

	userCount[0].Count = 0
	userCount[0].Realm = "saml"
	userCount[1].Count = 0
	userCount[1].Realm = "internal"

	for _, user := range users {
		if user.Realm == "saml" {
			userCount[0].Count += 1
		} else if user.Realm == "internal" {
			userCount[1].Count += 1
		}
	}
	return userCount
}
