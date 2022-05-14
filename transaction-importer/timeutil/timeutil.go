package timeutil

import (
	"time"

	"github.com/ccundiff/spend-tracker/transaction-importer/constants"
	"github.com/pkg/errors"
)

// TODO: the location should be set on the user object, right now I'm just passing
// new york it everywhere
func YesterdaysDateAsString(location string) string {
	//startDate := time.Now().Add(-24 * time.Hour).Format(iso8601TimeFormat)
	// TODO: need to handle error here
	loc, _ := time.LoadLocation(location)
	currentTime := time.Now().Add(-24 * time.Hour).In(loc)
	return currentTime.Format(constants.DATE_FORMAT)
}

func EastCoastYesterdaysDateAsString() string {
	return YesterdaysDateAsString(constants.EAST_COAST_TIME_LOCATION)
}

func GetMonthFromDateString(date string) (int, error) {
	timeDate, err := time.Parse(constants.DATE_FORMAT, date)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse date when creating dss, err=%v", err)
	}

	return int(timeDate.Month()), nil
}
