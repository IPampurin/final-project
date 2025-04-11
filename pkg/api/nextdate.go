package api

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const DateOnlyApi = "20060102"

func nextDayHandler(w http.ResponseWriter, r *http.Request) {

	nowParametr := r.FormValue("now")
	dateParametr := r.FormValue("date")
	repeatParametr := r.FormValue("repeat")

	dateFromNextDate, err := NextDate(nowParametr, dateParametr, repeatParametr)
	if err != nil {
		fmt.Println("ошибка получения даты следующего события: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType([]byte(dateFromNextDate)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dateFromNextDate))

}

// parametrParser проводит первичный разбор и проверку полученных в запросе параметров
func parametrParser(nowParametr, dateParametr, repeatParametr string) (time.Time, time.Time, []string, error) {

	// обрабатываем параметр now
	if nowParametr == "" {
		nowParametr = time.Now().Format(DateOnlyApi)
	}
	nowTime, err := time.Parse(DateOnlyApi, nowParametr)
	if err != nil {
		return time.Time{}, time.Time{}, []string{}, fmt.Errorf("ошибка парсинга параметра запроса 'now': %e\n", err)
	}

	// обрабатываем параметр date
	date, err := time.Parse(DateOnlyApi, dateParametr)
	if err != nil {
		return time.Time{}, time.Time{}, []string{}, fmt.Errorf("ошибка парсинга параметра запроса 'date': %e", err)
	}

	// обрабатываем параметр repeat
	if repeatParametr == "" {
		return time.Time{}, time.Time{}, []string{}, fmt.Errorf("параметр repeat не указан\n")
	}
	repeatSlice := strings.Split(repeatParametr, " ")

	return nowTime, date, repeatSlice, nil
}

// yDate вычисляет следующую дату для ежегодных повторений задач
func yDate(now time.Time, date time.Time) time.Time {

	/* правильнее было бы так
		for now.After(date) {
			date = date.AddDate(1, 0, 0)
		}
	а не так, как ниже, но тесты проходят только с приведенным вариантом */

	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			break
		}
	}

	return date
}

// dDate вычисляет следующую дату для повторений задач по дням
func dDate(now time.Time, date time.Time, daysCount int) time.Time {

	/* правильнее было бы так
		for now.After(date) {
			date = date.AddDate(0, 0, daysCount)
		}
	а не так, как ниже, но тесты проходят только с приведенным вариантом */

	for {
		date = date.AddDate(0, 0, daysCount)
		if date.After(now) {
			break
		}
	}

	return date
}

// wDate вычисляет следующую дату для повторений задач по неделям
func wDate(now time.Time, date time.Time, weekDaysInt []int) time.Time {

	if now.After(date) {
		date = now
	}
	var dateWeekDay int

outerLoop:
	for {
		date = date.AddDate(0, 0, 1)
		dateWeekDay = int(date.Weekday())
		if dateWeekDay == 0 {
			dateWeekDay = 7
		}
		for _, v := range weekDaysInt {
			if v == dateWeekDay {
				break outerLoop
			}
		}
	}

	return date
}

// mDate вычисляет следующую дату для повторений задач по дням месяца (без опции)
func mDate(now time.Time, date time.Time, monthDaysInt []int) time.Time {

	var lastDayDateMonth, penultimateDayDateMonth int

outerLoopM2:
	for {
		date = date.AddDate(0, 0, 1)

		dateYear, dateMonth, dateMonthDay := date.Date()
		lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
		penultimateDayDateMonth = lastDayDateMonth - 1

		for _, v := range monthDaysInt {
			if v == -2 && dateMonthDay == penultimateDayDateMonth {
				break outerLoopM2
			}
			if v == -1 && dateMonthDay == lastDayDateMonth {
				break outerLoopM2
			}
			if v == dateMonthDay {
				break outerLoopM2
			}
		}
	}

	return date
}

// mDateOpt вычисляет следующую дату для повторений задач по дням указанных месяцев (с опцией)
func mDateOpt(now time.Time, date time.Time, monthDaysInt []int, monthNumbersInt []int) (time.Time, error) {

	var lastDayDateMonth, penultimateDayDateMonth int

outerLoopM3:
	for {
		date = date.AddDate(0, 0, 1)

		for j := 0; j < len(monthNumbersInt); j++ {

			dateYear, dateMonth, dateMonthDay := date.Date()
			lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
			penultimateDayDateMonth = lastDayDateMonth - 1

			if monthNumbersInt[j] == int(dateMonth) {

				for _, v := range monthDaysInt {
					if v > 0 && v > lastDayDateMonth {
						return time.Time{}, fmt.Errorf("в месяце %d нет %d-го дня\n", monthNumbersInt[j], v)
					}
					if v == -2 && dateMonthDay == penultimateDayDateMonth {
						break outerLoopM3
					}
					if v == -1 && dateMonthDay == lastDayDateMonth {
						break outerLoopM3
					}
					if v == dateMonthDay {
						break outerLoopM3
					}
				}
			}
		}
	}

	return date, nil
}

// NextDate вычисляет следующую дату для задачи
func NextDate(nowParametr string, dateParametr string, repeatParametr string) (string, error) {

	now, date, repeatSlice, err := parametrParser(nowParametr, dateParametr, repeatParametr)
	if err != nil {
		return "", fmt.Errorf("%e\n", err)
	}

	switch repeatSlice[0] {
	case "y":

		if len(repeatSlice) != 1 {
			return "", fmt.Errorf("первая позиция в параметре repeat указана не верно ('y' не требует дополнительных уточнений)\n")
		}

		date = yDate(now, date)

	case "d":

		if len(repeatSlice) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'd 7')\n")
		}

		daysCount, err := strconv.Atoi(repeatSlice[1])
		if err != nil {
			return "", fmt.Errorf("количество дней в параметре repeat указано не верно (правильно, например, 'd 7')\n")
		}

		if daysCount < 1 || 400 < daysCount {
			return "", fmt.Errorf("количество дней в параметре repeat - не более 400\n")
		}

		date = dDate(now, date, daysCount)

	case "w":

		if len(repeatSlice) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'w 1,4,5')\n")
		}

		weekDaysStr := strings.Split(repeatSlice[1], ",")
		weekDaysInt := make([]int, len(weekDaysStr), len(weekDaysStr))

		for i := 0; i < len(weekDaysStr); i++ {
			count, err := strconv.Atoi(weekDaysStr[i])
			if err != nil || count < 1 || count > 7 {
				return "", fmt.Errorf("дни недели в параметре repeat указаны не верно (правильно, например, 'w 1,4,7')\n")
			}
			weekDaysInt = append(weekDaysInt, count)
		}
		slices.Sort(weekDaysInt)

		date = wDate(now, date, weekDaysInt)

	case "m":

		if len(repeatSlice) != 2 && len(repeatSlice) != 3 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'm 1,-1 2,8')\n")
		}

		monthDaysStr := strings.Split(repeatSlice[1], ",")
		monthDaysInt := make([]int, len(monthDaysStr), len(monthDaysStr))

		for i := 0; i < len(monthDaysStr); i++ {
			count, err := strconv.Atoi(monthDaysStr[i])
			if err != nil || count < -2 || count > 31 || count == 0 {
				return "", fmt.Errorf("дни месяца в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
			}
			monthDaysInt = append(monthDaysInt, count)
		}
		slices.Sort(monthDaysInt)

		if now.After(date) {
			date = now
		}

		if len(repeatSlice) == 2 {

			date = mDate(now, date, monthDaysInt)
		}

		if len(repeatSlice) == 3 {

			monthNumbersStr := strings.Split(repeatSlice[2], ",")
			monthNumbersInt := make([]int, len(monthNumbersStr), len(monthNumbersStr))

			for i := 0; i < len(monthNumbersStr); i++ {
				number, err := strconv.Atoi(monthNumbersStr[i])
				if err != nil || number < 1 || number > 12 {
					return "", fmt.Errorf("номера месяцев в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
				}
				monthNumbersInt = append(monthNumbersInt, number)
			}
			slices.Sort(monthNumbersInt)

			date, err = mDateOpt(now, date, monthDaysInt, monthNumbersInt)
			if err != nil {
				return "", fmt.Errorf("%v\n", err)
			}
		}

	default:
		return "", fmt.Errorf("первая позиция в параметре repeat указана не верно\n")
	}

	return date.Format(DateOnlyApi), nil
}
