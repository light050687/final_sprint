package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	year               = 1
	dateFormat         = "20060102"
	sundayEU           = 0
	sundayRus          = 7
	lastDayOfMonth     = -1
	predLastDayOfMonth = -2

	january       = 1
	december      = 12
	maxDayOfMonth = 31
)

func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("повторение обязательно")
	}
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return "", nil
	}

	repeatString := strings.Split(repeat, " ")
	switch strings.ToLower(repeatString[0]) {
	case "d":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("повторение должно быть не менее двух символов для дней")
		}
		days, err := nextDayValue(repeatString[1])
		if err != nil {
			return "", err
		}

		date = addDateTask(now, date, 0, 0, days)

	case "y":
		date = addDateTask(now, date, year, 0, 0)
	case "w":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("повторение должно быть не менее двух символов для дней")
		}
		date, err = addDateTaskWeek(now, date, repeatString[1])
		if err != nil {
			return "", err
		}
	case "m":
		if len(repeatString) < 2 {
			return "", fmt.Errorf("повторение должно быть не менее двух символов для дней")
		}
		date, err = addDateTaskMonth(now, date, repeatString)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("недопустимый символ повторения")
	}

	return date.Format(dateFormat), nil
}

func nextDayValue(num string) (int, error) {
	days, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	if days >= 400 || days < 0 {
		return 0, fmt.Errorf("недопустимое значение %d", days)
	}
	return days, nil
}

func addDateTask(now time.Time, dateTask time.Time, year int, month int, day int) time.Time {
	dateTask = dateTask.AddDate(year, month, day)
	for dateTask.Before(now) {
		dateTask = dateTask.AddDate(year, month, day)
	}
	return dateTask
}

func addDateTaskWeek(now, dateTask time.Time, daysString string) (time.Time, error) {
	daysArr := strings.Split(daysString, ",")
	daysOfWeek := regexp.MustCompile("[1-7]")
	daysOfWeekMap := make(map[int]bool)
	for _, day := range daysArr {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}
		if len(daysOfWeek.FindAllString(day, -1)) == 0 {
			return dateTask, fmt.Errorf("недопустимое значение %d день недели", numberOfDay)
		}
		if numberOfDay == sundayRus {
			numberOfDay = sundayEU
		}
		daysOfWeekMap[numberOfDay] = true
	}
	for {
		if daysOfWeekMap[int(dateTask.Weekday())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}
	return dateTask, nil
}

func addDateTaskMonth(now, dateTask time.Time, repeat []string) (time.Time, error) {
	daysString := repeat[1]
	monthsString := ""
	if len(repeat) > 2 {
		monthsString = repeat[2]
	}
	daysSlice := strings.Split(daysString, ",")
	monthsSlice := strings.Split(monthsString, ",")
	daysMap := make(map[int]bool)
	for _, day := range daysSlice {
		numberOfDay, err := strconv.Atoi(day)
		if err != nil {
			return dateTask, err
		}
		if numberOfDay < predLastDayOfMonth || numberOfDay > maxDayOfMonth || numberOfDay == 0 {
			return dateTask, fmt.Errorf("недопустимое значение %d день месяца", numberOfDay)
		}
		daysMap[numberOfDay] = true
	}
	monthsMap := make(map[int]bool)
	for _, month := range monthsSlice {
		if month == "" {
			continue
		}
		numberOfMonth, err := strconv.Atoi(month)
		if err != nil {
			return dateTask, err
		}
		if numberOfMonth < january || numberOfMonth > december {
			return dateTask, fmt.Errorf("недопустимое значение %d месяц", numberOfMonth)
		}
		monthsMap[numberOfMonth] = true
	}
	for {
		if len(monthsMap) == 0 {
			break
		}
		if monthsMap[int(dateTask.Month())] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}
	for {
		lastDay := time.Date(dateTask.Year(), dateTask.Month()+1, 0, 0, 0, 0, 0, dateTask.Location()).Day()
		predLastDay := lastDay - 1
		key := dateTask.Day()
		switch {
		case lastDay == dateTask.Day():
			if _, ok := daysMap[lastDayOfMonth]; ok {
				key = lastDayOfMonth
			}
		case predLastDay == dateTask.Day():
			if _, ok := daysMap[predLastDayOfMonth]; ok {
				key = predLastDayOfMonth
			}
		}
		if daysMap[key] {
			if now.Before(dateTask) {
				break
			}
		}
		dateTask = dateTask.AddDate(0, 0, 1)
	}
	return dateTask, nil
}
