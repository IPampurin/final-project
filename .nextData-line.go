package main

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

	nowStr := r.FormValue("now")
	if nowStr == "" {
		nowStr = time.Now().Format(DateOnlyApi)
	}

	nowTime, err := time.Parse(DateOnlyApi, nowStr)
	if err != nil {
		fmt.Println("ошибка парсинга даты параметра 'now': ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	start := r.FormValue("date")

	repeat := r.FormValue("repeat")

	dateOut, err := NextDate(nowTime, start, repeat)
	if err != nil {
		fmt.Println("ошибка получения даты следующего события: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType([]byte(dateOut)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dateOut))

}

func NextDate(now time.Time, start string, repeat string) (string, error) {

	date, err := time.Parse(DateOnlyApi, start)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга даты параметра 'date': %e", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("параметр repeat не указан\n")
	}

	repeatParametrs := strings.Split(repeat, " ")

	switch repeatParametrs[0] {
	case "y":
		if len(repeatParametrs) != 1 {
			return "", fmt.Errorf("первая позиция в параметре repeat указана не верно ('y' не требует дополнительных уточнений)\n")
		}
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

	case "d":
		if len(repeatParametrs) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'd 7')\n")
		}

		daysCount, err := strconv.Atoi(repeatParametrs[1])
		if err != nil {
			return "", fmt.Errorf("количество дней в параметре repeat указано не верно (правильно, например, 'd 7')\n")
		}

		if daysCount < 1 || 400 < daysCount {
			return "", fmt.Errorf("количество дней в параметре repeat - не более 400\n")
		}

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

	case "w":
		if len(repeatParametrs) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'w 1,4,5')\n")
		}

		repeatWeekDaysStr := strings.Split(repeatParametrs[1], ",")
		repeatWeekDaysInt := make([]int, len(repeatWeekDaysStr), len(repeatWeekDaysStr))

		for i := 0; i < len(repeatWeekDaysStr); i++ {
			repeatWeekDaysInt[i], err = strconv.Atoi(repeatWeekDaysStr[i])
			if err != nil || repeatWeekDaysInt[i] < 1 || repeatWeekDaysInt[i] > 7 {
				return "", fmt.Errorf("дни недели в параметре repeat указаны не верно (правильно, например, 'w 1,4,7')\n")
			}
		}

		slices.Sort(repeatWeekDaysInt)

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
			for _, v := range repeatWeekDaysInt {
				if v == dateWeekDay {
					break outerLoop
				}
			}
		}

	case "m":
		if len(repeatParametrs) != 2 && len(repeatParametrs) != 3 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'm 1,-1 2,8')\n")
		}

		repeatMonthDaysStr := strings.Split(repeatParametrs[1], ",")
		repeatMonthDaysInt := make([]int, len(repeatMonthDaysStr), len(repeatMonthDaysStr))

		for i := 0; i < len(repeatMonthDaysStr); i++ {
			repeatMonthDaysInt[i], err = strconv.Atoi(repeatMonthDaysStr[i])
			if err != nil || repeatMonthDaysInt[i] < -2 || repeatMonthDaysInt[i] > 31 || repeatMonthDaysInt[i] == 0 {
				return "", fmt.Errorf("дни месяца в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
			}
		}

		slices.Sort(repeatMonthDaysInt)

		var lastDayDateMonth, penultimateDayDateMonth int

		if now.After(date) {
			date = now
		}

		if len(repeatParametrs) == 2 {

		outerLoopM2:
			for {
				date = date.AddDate(0, 0, 1)

				dateYear, dateMonth, dateMonthDay := date.Date()
				lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
				penultimateDayDateMonth = lastDayDateMonth - 1

				for _, v := range repeatMonthDaysInt {
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
		}

		if len(repeatParametrs) == 3 {

			repeatMonthNumbersStr := strings.Split(repeatParametrs[2], ",")
			repeatMonthNumbersInt := make([]int, len(repeatMonthNumbersStr), len(repeatMonthNumbersStr))

			for i := 0; i < len(repeatMonthNumbersStr); i++ {
				repeatMonthNumbersInt[i], err = strconv.Atoi(repeatMonthNumbersStr[i])
				if err != nil || repeatMonthNumbersInt[i] < 1 || repeatMonthNumbersInt[i] > 12 {
					return "", fmt.Errorf("номера месяцев в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
				}
			}

			slices.Sort(repeatMonthNumbersInt)

		outerLoopM3:
			for {
				date = date.AddDate(0, 0, 1)

				for j := 0; j < len(repeatMonthNumbersInt); j++ {

					dateYear, dateMonth, dateMonthDay := date.Date()
					lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
					penultimateDayDateMonth = lastDayDateMonth - 1

					if repeatMonthNumbersInt[j] == int(dateMonth) {

						for _, v := range repeatMonthDaysInt {
							if v > 0 && v > lastDayDateMonth {
								return "", fmt.Errorf("в месяце %d нет %d-го дня\n", repeatMonthNumbersInt[j], v)
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
		}

	default:
		return "", fmt.Errorf("первая позиция в параметре repeat указана не верно\n")
	}

	return date.Format(DateOnlyApi), nil
}

type nextDate struct {
	date   string
	repeat string
	want   string
}

func main() {
	tbl := []nextDate{
		//{"20240126", "", ""},
		//{"20240126", "k 34", ""},
		//{"20240126", "ooops", ""},
		//{"15000156", "y", ""},
		//{"ooops", "y", ""},
		{"16890220", "y", `20240220`},
		{"20250701", "y", `20260701`},
		{"20240101", "y", `20250101`},
		{"20231231", "y", `20241231`},
		{"20240229", "y", `20250301`},
		{"20240301", "y", `20250301`},
		//{"20240113", "d", ""},
		{"20240113", "d 7", `20240127`},
		{"20240120", "d 20", `20240209`},
		{"20240202", "d 30", `20240303`},
		//{"20240320", "d 401", ""},
		{"20231225", "d 12", `20240130`},
		{"20240228", "d 1", "20240229"},
		{"20231106", "m 13", "20240213"},
		//{"20240120", "m 40,11,19", ""},
		{"20240116", "m 16,5", "20240205"},
		{"20240126", "m 25,26,7", "20240207"},
		{"20240409", "m 31", "20240531"},
		{"20240329", "m 10,17 12,8,1", "20240810"},
		{"20230311", "m 07,19 05,6", "20240507"},
		{"20230311", "m 1 1,2", "20240201"},
		{"20240127", "m -1", "20240131"},
		{"20240222", "m -2", "20240228"},
		//{"20240222", "m -2,-3", ""},
		{"20240326", "m -1,-2", "20240330"},
		{"20240201", "m -1,18", "20240218"},
		{"20240125", "w 1,2,3", "20240129"},
		{"20240126", "w 7", "20240128"},
		{"20230126", "w 4,5", "20240201"},
		//{"20230226", "w 8,4,5", ""},
		//{"20230126", "m 31 2", "20240201"},
	}

	startTime := time.Now()

	nowStr := "20240126"
	if nowStr == "" {
		nowStr = time.Now().Format(DateOnlyApi)
	}

	nowTime, err := time.Parse(DateOnlyApi, nowStr)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	var msg, a string

	for _, v := range tbl {

		fmt.Printf("test_data = %v, repeat = %v, ", v.date, v.repeat)

		date, err := NextDate(nowTime, v.date, v.repeat)
		if err != nil {
			fmt.Println("ошибка получения даты следующего события: ", err)
			return
		}

		if date != v.want {
			msg = "FAIL"
			a = "FAIL"
		} else {
			msg = ""
		}

		fmt.Printf(" полученная date = %v, ожидаемая date = %v,    %v\n", date, v.want, msg)
	}

	duration := time.Since(startTime)
	fmt.Println(a, duration)
}
