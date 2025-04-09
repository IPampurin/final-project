package api

import (
	"fmt"
	"net/http"
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

	startTime, err := time.Parse(DateOnlyApi, start)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга даты параметра 'date': %e", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("параметр repeat не указан.")
	}

	repeatParametrs := strings.Split(repeat, " ")

	var dateResult time.Time

	switch repeatParametrs[0] {
	case "y":
		if len(repeatParametrs) != 1 {
			return "", fmt.Errorf("первая позиция в параметре repeat указана не верно ('y' не требует дополнительных уточнений).")
		}

		if startTime.After(now) {
			dateResult = startTime
		} else {
			for {
				dateResult = startTime.AddDate(1, 0, 0)
				if dateResult.After(now) {
					break
				}
			}
		}

	case "d":
		if len(repeatParametrs) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'd 7').")
		}

		daysCount, err := strconv.Atoi(repeatParametrs[1])
		if err != nil {
			return "", fmt.Errorf("количество дней в параметре repeat указано не верно (правильно, например, 'd 7').")
		}

		if daysCount < 1 || 400 < daysCount {
			return "", fmt.Errorf("количество дней в параметре repeat не более 400.")
		}

		if startTime.After(now) {
			dateResult = startTime
		} else {
			for {
				dateResult = startTime.AddDate(0, 0, daysCount)
				if dateResult.After(now) {
					break
				}
			}
		}

	default:
		return "", fmt.Errorf("первая позиция в параметре repeat указана не верно.")

	}

	return dateResult.Format(DateOnlyApi), nil
}
