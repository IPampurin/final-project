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

// YRepeat структура для правил повторения ежегодных задач
type YearRepeat struct {
	nowField  time.Time
	dateField time.Time
}

// DRepeat структура для правил повторения задач по дням
type DayRepeat struct {
	YearRepeat
	dayIteration int
}

// WRepeat структура для правил повторения задач по неделям
type WeekRepeat struct {
	YearRepeat
	weekIteration []int
}

// MRepeat структура для правил повторения задач по месяцам
type MonthRepeat struct {
	YearRepeat
	monthIteration []int
}

// FullMRepeat структура для правил повторения задач по месяцам
// с учётом опционального поля
type FullMonthRepeat struct {
	YearRepeat
	monthIteration []int
	numberMonth    []int
}

type CalculationNextDate interface {
	NextDateCalc() (time.Time, error)
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {

	nowParametr := r.FormValue("now")
	dateParametr := r.FormValue("date")
	repeatParametr := r.FormValue("repeat")

	RepeatStruct, err := parametrParser(nowParametr, dateParametr, repeatParametr)
	if err != nil {
		fmt.Println("ошибка ввода параметров повториения задачи: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dateFromNextDate, err := NextDate(RepeatStruct)
	if err != nil {
		fmt.Println("ошибка получения даты следующего события: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType([]byte(dateFromNextDate)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dateFromNextDate))

}

// parametrParser проводит разбор и проверку полученных в запросе параметров
func parametrParser(now, dstart, repeat string) (CalculationNextDate, error) {

	var RepeatStruct CalculationNextDate

	// обрабатываем параметр now
	if now == "" {
		now = time.Now().Format(DateOnlyApi)
	}
	nowTime, err := time.Parse(DateOnlyApi, now)
	if err != nil {
		return RepeatStruct, fmt.Errorf("ошибка парсинга параметра запроса 'now': %v\n", err)
	}

	// обрабатываем параметр date
	date, err := time.Parse(DateOnlyApi, dstart)
	if err != nil {
		return RepeatStruct, fmt.Errorf("ошибка парсинга параметра запроса 'date': %v", err)
	}

	// обрабатываем параметр repeat и переопределяем структуру RepeatStruct
	if repeat == "" {
		return RepeatStruct, fmt.Errorf("параметр repeat не указан\n")
	}
	repeatSlice := strings.Split(repeat, " ")

	switch repeatSlice[0] {

	case "y":
		if len(repeatSlice) != 1 {
			return YearRepeat{}, fmt.Errorf("первая позиция в параметре repeat указана не верно ('y' не требует дополнительных уточнений)\n")
		}

		RepeatStruct = YearRepeat{
			nowField:  nowTime,
			dateField: date,
		}

	case "d":
		if len(repeatSlice) != 2 {
			return DayRepeat{}, fmt.Errorf("параметр repeat указан не верно (правильно, например, 'd 7')\n")
		}

		daysCount, err := strconv.Atoi(repeatSlice[1])
		if err != nil {
			return DayRepeat{}, fmt.Errorf("количество дней в параметре repeat указано не верно (правильно, например, 'd 7')\n")
		}

		if daysCount < 1 || 400 < daysCount {
			return DayRepeat{}, fmt.Errorf("количество дней в параметре repeat - от 1 до 400\n")
		}

		RepeatStruct = DayRepeat{
			YearRepeat{
				nowField:  nowTime,
				dateField: date,
			},
			daysCount,
		}

	case "w":
		if len(repeatSlice) != 2 {
			return WeekRepeat{}, fmt.Errorf("параметр repeat указан не верно (правильно, например, 'w 1,4,5')\n")
		}

		repeatWeekDaysStr := strings.Split(repeatSlice[1], ",")
		repeatWeekDaysInt := make([]int, 0, len(repeatWeekDaysStr))

		for _, value := range repeatWeekDaysStr {
			count, err := strconv.Atoi(value)
			if err != nil || count < 1 || count > 7 {
				return WeekRepeat{}, fmt.Errorf("дни недели в параметре repeat указаны не верно (правильно, например, 'w 1,4,7')\n")
			}
			repeatWeekDaysInt = append(repeatWeekDaysInt, count)
		}
		slices.Sort(repeatWeekDaysInt)

		RepeatStruct = WeekRepeat{
			YearRepeat{
				nowField:  nowTime,
				dateField: date,
			},
			repeatWeekDaysInt,
		}

	case "m":
		if len(repeatSlice) != 2 && len(repeatSlice) != 3 {
			return MonthRepeat{}, fmt.Errorf("параметр repeat указан не верно (правильно, например, 'm 1,-1 2,8')\n")
		}

		repeatMonthDaysStr := strings.Split(repeatSlice[1], ",")
		repeatMonthDaysInt := make([]int, 0, len(repeatMonthDaysStr))

		for _, value := range repeatMonthDaysStr {
			count, err := strconv.Atoi(value)
			if err != nil || count < -2 || count > 31 || count == 0 {
				return MonthRepeat{}, fmt.Errorf("дни месяца в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
			}
			repeatMonthDaysInt = append(repeatMonthDaysInt, count)
		}
		slices.Sort(repeatMonthDaysInt)

		if len(repeatSlice) == 2 {
			RepeatStruct = MonthRepeat{
				YearRepeat{
					nowField:  nowTime,
					dateField: date,
				},
				repeatMonthDaysInt,
			}
		}

		if len(repeatSlice) == 3 {
			repeatMonthNumbersStr := strings.Split(repeatSlice[2], ",")
			repeatMonthNumbersInt := make([]int, 0, len(repeatMonthNumbersStr))

			for _, value := range repeatMonthNumbersStr {
				count, err := strconv.Atoi(value)
				if err != nil || count < 1 || count > 12 {
					return FullMonthRepeat{}, fmt.Errorf("номера месяцев в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')\n")
				}
				repeatMonthNumbersInt = append(repeatMonthNumbersInt, count)
			}
			slices.Sort(repeatMonthNumbersInt)

			RepeatStruct = FullMonthRepeat{
				YearRepeat{
					nowField:  nowTime,
					dateField: date,
				},
				repeatMonthDaysInt,
				repeatMonthNumbersInt,
			}
		}

	default:
		return RepeatStruct, fmt.Errorf("первая позиция в параметре repeat указана не верно\n")
	}

	return RepeatStruct, nil
}

func (y YearRepeat) NextDateCalc() (time.Time, error) {
	/* правильнее было бы так
		for now.After(date) {
			date = date.AddDate(1, 0, 0)
		}
	а не так, как ниже, но тесты проходят только с приведенным вариантом */

	for {
		y.dateField = y.dateField.AddDate(1, 0, 0)
		if y.dateField.After(y.nowField) {
			break
		}
	}

	return y.dateField, nil
}

func (d DayRepeat) NextDateCalc() (time.Time, error) {
	/* правильнее было бы так
		for now.After(date) {
			date = date.AddDate(0, 0, daysCount)
		}
	а не так, как ниже, но тесты проходят только с приведенным вариантом */

	for {
		d.dateField = d.dateField.AddDate(0, 0, d.dayIteration)
		if d.dateField.After(d.nowField) {
			break
		}
	}
	return d.dateField, nil
}

func (w WeekRepeat) NextDateCalc() (time.Time, error) {

	if w.nowField.After(w.dateField) {
		w.dateField = w.nowField
	}
	var dateWeekDay int

outerLoop:
	for {
		w.dateField = w.dateField.AddDate(0, 0, 1)
		dateWeekDay = int(w.dateField.Weekday())
		if dateWeekDay == 0 {
			dateWeekDay = 7
		}
		for _, v := range w.weekIteration {
			if v == dateWeekDay {
				break outerLoop
			}
		}
	}

	return w.dateField, nil
}

func (m MonthRepeat) NextDateCalc() (time.Time, error) {

	var lastDayDateMonth, penultimateDayDateMonth int

	if m.nowField.After(m.dateField) {
		m.dateField = m.nowField
	}

outerLoopM2:
	for {
		m.dateField = m.dateField.AddDate(0, 0, 1)

		dateYear, dateMonth, dateMonthDay := m.dateField.Date()
		lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
		penultimateDayDateMonth = lastDayDateMonth - 1

		for _, v := range m.monthIteration {
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

	return m.dateField, nil
}

func (fm FullMonthRepeat) NextDateCalc() (time.Time, error) {

	var lastDayDateMonth, penultimateDayDateMonth int

	if fm.nowField.After(fm.dateField) {
		fm.dateField = fm.nowField
	}

outerLoopM3:
	for {
		fm.dateField = fm.dateField.AddDate(0, 0, 1)

		for j := 0; j < len(fm.numberMonth); j++ {

			dateYear, dateMonth, dateMonthDay := fm.dateField.Date()
			lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
			penultimateDayDateMonth = lastDayDateMonth - 1

			if fm.numberMonth[j] == int(dateMonth) {

				for _, v := range fm.monthIteration {
					if v > 0 && v > lastDayDateMonth {
						return time.Time{}, fmt.Errorf("в месяце %d нет %d-го дня\n", fm.numberMonth[j], v)
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

	return fm.dateField, nil
}

// NextDate проверяет правильность принятых в параметре repeat значений и возвращает следующую дату для задачи
func NextDate(RepeatStruct CalculationNextDate) (string, error) {

	date, err := RepeatStruct.NextDateCalc()
	if err != nil {
		return "", fmt.Errorf("%v\n", err)
	}

	return date.Format(DateOnlyApi), nil
}
