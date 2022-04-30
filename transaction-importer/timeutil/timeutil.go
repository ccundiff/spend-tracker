package timeutil

import "time"

// TODO: the location should be set on the user object, right now I'm just passing
// new york it everywhere
func CurrentDateAsString(location string) string {
	const iso8601TimeFormat = "2006-01-02"
	//startDate := time.Now().Add(-24 * time.Hour).Format(iso8601TimeFormat)
	loc, _ := time.LoadLocation(location)
	currentTime := time.Now().In(loc)
	return currentTime.Format(iso8601TimeFormat)
}
