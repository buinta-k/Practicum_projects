package api

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func IsBefore(t1, t2 time.Time) bool {
	return t1.Format(dateLayout) < t2.Format(dateLayout)
}

func Validate(repeat []string) (string, error) {
	if len(repeat) == 0 {
		return "", fmt.Errorf("Пустая строка правила")
	}

	rule := repeat[0]

	if rule != "d" && rule != "y" && rule != "w" && rule != "m" {
		return "", fmt.Errorf("Не соответствует формату правила повторения")
	}

	if rule == "y" {
		if len(repeat) != 1 {
			return "", fmt.Errorf("Некорректное правило для года")
		}
		return "y", nil
	}

	if rule == "d" {
		if len(repeat) != 2 {
			return "", fmt.Errorf("Не указан интервал в днях")
		}

		num, err := strconv.Atoi(repeat[1])
		if err != nil || num <= 0 || num > 400 {
			return "", fmt.Errorf("Некорректный формат дня")
		}

		return "d", nil
	}

	if rule == "w" {
		if len(repeat) < 2 {
			return "", fmt.Errorf("Не указаны дни недели")
		}

		days := strings.Split(repeat[1], ",")

		for _, day := range days {
			num, err := strconv.Atoi(day)
			if err != nil || num < 1 || num > 7 {
				return "", fmt.Errorf("Недопустимое значение дня недели")
			}
		}

		return "w", nil
	}

	if rule == "m" {
		if len(repeat) < 2 || len(repeat) > 3 {
			return "", fmt.Errorf("Не указаны дни месяца")
		}

		days := strings.Split(repeat[1], ",")

		for _, day := range days {
			num, err := strconv.Atoi(day)

			if err != nil || num == 0 || num < -2 || num > 31 {
				return "", fmt.Errorf("Недопустимый день месяца")
			}
		}

		if len(repeat) == 3 {
			months := strings.Split(repeat[2], ",")

			for _, month := range months {
				num, err := strconv.Atoi(month)

				if err != nil || num < 1 || num > 12 {
					return "", fmt.Errorf("Недопустимый месяц")
				}
			}
		}

		return "m", nil
	}

	return "", fmt.Errorf("Неподдерживаемый тип правила")
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if dstart == "" {
		return "", fmt.Errorf("Пустая дата")
	}

	if repeat == "" {
		return "", fmt.Errorf("Пустое правило повторения")
	}

	repeatParts := strings.Split(repeat, " ")

	ruleType, err := Validate(repeatParts)
	if err != nil {
		return "", err
	}

	start, err := time.Parse(dateLayout, dstart)
	if err != nil {
		return "", err
	}

	switch ruleType {
	case "y":
		return handleYearly(now, start)

	case "d":
		num, err := strconv.Atoi(repeatParts[1])
		if err != nil {
			return "", fmt.Errorf("Ошибка парсинга дня для параметра d")
		}
		return handleDaily(now, start, num)

	case "w":
		return handleWeekly(now, start, repeatParts[1])

	case "m":
		return handleMonthly(now, start, repeatParts)

	default:
		return "", fmt.Errorf("Тип правила не соответствует формату")
	}
}

func handleYearly(now time.Time, start time.Time) (string, error) {
	data := start

	for {
		data = data.AddDate(1, 0, 0)

		if IsBefore(now, data) {
			break
		}
	}

	return data.Format(dateLayout), nil
}

func handleDaily(now time.Time, start time.Time, num int) (string, error) {
	data := start

	for {
		data = data.AddDate(0, 0, num)

		if IsBefore(now, data) {
			break
		}
	}

	return data.Format(dateLayout), nil
}

func handleWeekly(now time.Time, start time.Time, daysStr string) (string, error) {
	tokens := strings.Split(daysStr, ",")

	weekdayMap := make(map[time.Weekday]bool)

	for _, token := range tokens {
		day, err := strconv.Atoi(token)
		if err != nil || day < 1 || day > 7 {
			return "", fmt.Errorf("Недопустимое значение дня недели")
		}

		weekdayMap[time.Weekday(day%7)] = true
	}

	data := start

	for {
		data = data.AddDate(0, 0, 1)

		if IsBefore(now, data) && weekdayMap[data.Weekday()] {
			break
		}
	}

	return data.Format(dateLayout), nil
}

func handleMonthly(now time.Time, start time.Time, parts []string) (string, error) {
	days := strings.Split(parts[1], ",")

	var targetDays []int

	for _, val := range days {
		day, err := strconv.Atoi(val)

		if err != nil || day == 0 || day < -2 || day > 31 {
			return "", fmt.Errorf("Недопустимый день месяца")
		}

		targetDays = append(targetDays, day)
	}

	var targetMonths map[time.Month]bool

	if len(parts) == 3 {
		targetMonths = make(map[time.Month]bool)

		monthTokens := strings.Split(parts[2], ",")

		for _, token := range monthTokens {
			month, err := strconv.Atoi(token)

			if err != nil || month < 1 || month > 12 {
				return "", fmt.Errorf("Недопустимый месяц")
			}

			targetMonths[time.Month(month)] = true
		}
	}

	searchMonth := time.Date(
		start.Year(),
		start.Month(),
		1,
		0, 0, 0, 0,
		start.Location(),
	)

	var result time.Time

	for {
		month := searchMonth.Month()

		if targetMonths != nil && !targetMonths[month] {
			searchMonth = searchMonth.AddDate(0, 1, 0)
			continue
		}

		lastDay := searchMonth.AddDate(0, 1, -1).Day()

		var validDates []time.Time

		for _, day := range targetDays {
			actualDay := day

			if day < 0 {
				actualDay = lastDay + day + 1
			}

			if actualDay > 0 && actualDay <= lastDay {
				date := time.Date(
					searchMonth.Year(),
					month,
					actualDay,
					0, 0, 0, 0,
					searchMonth.Location(),
				)

				if !IsBefore(date, start) && IsBefore(now, date) {
					validDates = append(validDates, date)
				}
			}
		}

		if len(validDates) > 0 {
			sort.Slice(validDates, func(i, j int) bool {
				return IsBefore(validDates[i], validDates[j])
			})

			result = validDates[0]
			break
		}

		searchMonth = searchMonth.AddDate(0, 1, 0)
	}

	return result.Format(dateLayout), nil
}
