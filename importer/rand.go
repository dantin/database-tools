package importer

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	yearFormat     = "2006"
	dateFormat     = "2006-01-02"
	timeFormat     = "12:00:00"
	dateTimeFormat = "2006-01-02 12:00:00"

	// Used by randString
	alphabet      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	letterIdxBits = 6                  // 6 bits to represent a letter index
	letterIdxMax  = 63 / letterIdxBits // # of letter indices fitting in 63 bits
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}

func randInt64(min int64, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func randDate(min string, max string) string {
	if len(min) == 0 {
		year := time.Now().Year()
		month := randInt(1, 12)
		day := randInt(1, 28)
		return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	}

	minTime, _ := time.Parse(dateFormat, min)
	if len(max) == 0 {
		t := minTime.Add(time.Duration(randInt(0, 365)) * 24 * time.Hour)
		return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
	}

	maxTime, _ := time.Parse(dateFormat, max)
	days := int(maxTime.Sub(minTime).Hours() / 24)
	t := minTime.Add(time.Duration(randInt(0, days)) * 24 * time.Hour)
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

func randTimestamp(min string, max string) string {
	if len(min) == 0 {
		year := time.Now().Year()
		month := randInt(1, 12)
		day := randInt(1, 28)
		hour := randInt(0, 23)
		min := randInt(0, 59)
		sec := randInt(0, 59)
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, min, sec)
	}

	minTime, _ := time.Parse(dateTimeFormat, min)
	if len(max) == 0 {
		t := minTime.Add(time.Duration(randInt(0, 365)) * 24 * time.Hour)
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	}

	maxTime, _ := time.Parse(dateTimeFormat, max)
	seconds := int(maxTime.Sub(minTime).Seconds())
	t := minTime.Add(time.Duration(randInt(0, seconds)) * time.Second)
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func randTime(min string, max string) string {
	if len(min) == 0 || len(max) == 0 {
		hour := randInt(0, 23)
		min := randInt(0, 59)
		sec := randInt(0, 59)
		return fmt.Sprintf("%02d:%02d:%02d", hour, min, sec)
	}

	minTime, _ := time.Parse(timeFormat, min)
	maxTime, _ := time.Parse(timeFormat, max)
	seconds := int(maxTime.Sub(minTime).Seconds())
	t := minTime.Add(time.Duration(randInt(0, seconds)) * time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}

func randYear(min string, max string) string {
	if len(min) == 0 || len(max) == 0 {
		return fmt.Sprintf("%04d", time.Now().Year()-randInt(0, 10))
	}

	minTime, _ := time.Parse(yearFormat, min)
	maxTime, _ := time.Parse(yearFormat, max)
	seconds := int(maxTime.Sub(minTime).Seconds())
	t := minTime.Add(time.Duration(randInt(0, seconds)) * time.Second)
	return fmt.Sprintf("%04d", t.Year())
}
