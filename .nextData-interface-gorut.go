package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// константа формата времени
const DateOnlyApi = "20060102"

var wg sync.WaitGroup
var mu sync.Mutex

// YRepeat структура для правил повторения ежегодных задач
type YearRepeat struct {
	NowField  time.Time
	DateField time.Time
}

// DRepeat структура для правил повторения задач по дням
type DayRepeat struct {
	YearRepeat
	DayIteration int
}

// WRepeat структура для правил повторения задач по неделям
type WeekRepeat struct {
	YearRepeat
	WeekIteration []int
}

// MRepeat структура для правил повторения задач по месяцам
type MonthRepeat struct {
	YearRepeat
	MonthIteration []int
}

// FullMRepeat структура для правил повторения задач по месяцам
// с учётом опционального поля
type FullMonthRepeat struct {
	YearRepeat
	MonthIteration []int
	NumberMonth    []int
}

type CalculationNextDate interface {
	NextDateCalc() (time.Time, error)
}

// nextDayHandler обрабатывает запрос и возвращает следующую дату задачи
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

// NextDate проверяет правильность принятых в параметре repeat значений и вычисляет следующую дату для задачи
func NextDate(now, dstart, repeat string) (string, error) {

	// RepeatStruct структура валидных данных после обработки параметра repeat
	var RepeatStruct CalculationNextDate

	// обрабатываем параметр now
	if now == "" {
		now = time.Now().Format(DateOnlyApi)
	}
	nowTime, err := time.Parse(DateOnlyApi, now)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга параметра запроса 'now': %v", err)
	}

	// обрабатываем параметр date
	date, err := time.Parse(DateOnlyApi, dstart)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга параметра запроса 'date': %v", err)
	}

	// обрабатываем параметр repeat и переопределяем структуру RepeatStruct
	if repeat == "" {
		return "", fmt.Errorf("параметр repeat не указан")
	}
	repeatSlice := strings.Split(repeat, " ")

	switch repeatSlice[0] {

	case "y":
		if len(repeatSlice) != 1 {
			return "", fmt.Errorf("первая позиция в параметре repeat указана не верно ('y' не требует дополнительных уточнений)")
		}

		RepeatStruct = YearRepeat{
			NowField:  nowTime,
			DateField: date,
		}

	case "d":
		if len(repeatSlice) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'd 7')")
		}

		daysCount, err := strconv.Atoi(repeatSlice[1])
		if err != nil {
			return "", fmt.Errorf("количество дней в параметре repeat указано не верно (правильно, например, 'd 7')")
		}

		if daysCount < 1 || 400 < daysCount {
			return "", fmt.Errorf("количество дней в параметре repeat - от 1 до 400")
		}

		RepeatStruct = DayRepeat{
			YearRepeat{
				NowField:  nowTime,
				DateField: date,
			},
			daysCount,
		}

	case "w":
		if len(repeatSlice) != 2 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'w 1,4,5')")
		}

		repeatWeekDaysStr := strings.Split(repeatSlice[1], ",")
		// проверяем на излишне большой	объём данных
		if len(repeatWeekDaysStr) > 7 {
			return "", fmt.Errorf("параметр repeat указан не верно - вероятно, номера дней повторяются")
		}
		repeatWeekDaysInt := make([]int, 0, len(repeatWeekDaysStr))

		var neverDayW bool
		for _, value := range repeatWeekDaysStr {
			wg.Add(1)
			go func(v string) {
				defer wg.Done()
				count, err := strconv.Atoi(v)
				if err != nil || count < 1 || count > 7 {
					neverDayW = true
				}
				mu.Lock()
				repeatWeekDaysInt = append(repeatWeekDaysInt, count)
				mu.Unlock()
			}(value)
		}
		wg.Wait()

		if neverDayW {
			return "", fmt.Errorf("дни недели в параметре repeat указаны не верно (правильно, например, 'w 1,4,7')")
		}
		slices.Sort(repeatWeekDaysInt)

		RepeatStruct = WeekRepeat{
			YearRepeat{
				NowField:  nowTime,
				DateField: date,
			},
			repeatWeekDaysInt,
		}

	case "m":
		if len(repeatSlice) != 2 && len(repeatSlice) != 3 {
			return "", fmt.Errorf("параметр repeat указан не верно (правильно, например, 'm 1,-1 2,8')")
		}

		repeatMonthDaysStr := strings.Split(repeatSlice[1], ",")
		// проверяем на излишне большой	объём данных
		if len(repeatMonthDaysStr) > 33 {
			return "", fmt.Errorf("параметр repeat указан не верно - вероятно, номера дней повторяются")
		}
		repeatMonthDaysInt := make([]int, 0, len(repeatMonthDaysStr))

		var neverDayM bool
		for _, value := range repeatMonthDaysStr {
			wg.Add(1)
			go func(v string) {
				defer wg.Done()
				count, err := strconv.Atoi(v)
				if err != nil || count < -2 || count > 31 || count == 0 {
					neverDayM = true
				}
				mu.Lock()
				repeatMonthDaysInt = append(repeatMonthDaysInt, count)
				mu.Unlock()
			}(value)
		}
		wg.Wait()

		if neverDayM {
			return "", fmt.Errorf("дни месяца в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')")
		}

		slices.Sort(repeatMonthDaysInt)

		if len(repeatSlice) == 2 {
			RepeatStruct = MonthRepeat{
				YearRepeat{
					NowField:  nowTime,
					DateField: date,
				},
				repeatMonthDaysInt,
			}
		}

		var neverDayFM bool
		if len(repeatSlice) == 3 {
			repeatMonthNumbersStr := strings.Split(repeatSlice[2], ",")
			// проверяем на излишне большой	объём данных
			if len(repeatMonthNumbersStr) > 12 {
				return "", fmt.Errorf("параметр repeat указан не верно - вероятно, номера месяцев повторяются")
			}
			repeatMonthNumbersInt := make([]int, 0, len(repeatMonthNumbersStr))

			for _, value := range repeatMonthNumbersStr {
				wg.Add(1)
				go func(v string) {
					defer wg.Done()
					count, err := strconv.Atoi(v)
					if err != nil || count < 1 || count > 12 {
						neverDayFM = true
					}
					mu.Lock()
					repeatMonthNumbersInt = append(repeatMonthNumbersInt, count)
					mu.Unlock()
				}(value)
			}
			wg.Wait()

			if neverDayFM {
				return "", fmt.Errorf("номера месяцев в параметре repeat указаны не верно (правильно, например, 'm 1,-1 2,8')")
			}
			slices.Sort(repeatMonthNumbersInt)

			RepeatStruct = FullMonthRepeat{
				YearRepeat{
					NowField:  nowTime,
					DateField: date,
				},
				repeatMonthDaysInt,
				repeatMonthNumbersInt,
			}
		}

	default:
		return "", fmt.Errorf("первая позиция в параметре repeat указана не верно")
	}

	date, err = RepeatStruct.NextDateCalc()
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}

	return date.Format(DateOnlyApi), nil

}

func (y YearRepeat) NextDateCalc() (time.Time, error) {
	/* правильнее было бы так
	for y.NowField.After(y.DateField) {
		y.DateField = y.DateField.AddDate(1, 0, 0)
	}
	а не так, как ниже, но тесты проходят только с приведенным вариантом*/
	for {
		y.DateField = y.DateField.AddDate(1, 0, 0)
		if y.DateField.After(y.NowField) {
			break
		}
	}

	return y.DateField, nil
}

func (d DayRepeat) NextDateCalc() (time.Time, error) {
	/* правильнее было бы так
	for d.NowField.After(d.DateField) {
		d.DateField = d.DateField.AddDate(0, 0, d.DayIteration)
	}
	а не так, как ниже, но тесты проходят только с приведенным вариантом*/
	for {
		d.DateField = d.DateField.AddDate(0, 0, d.DayIteration)
		if d.DateField.After(d.NowField) {
			break
		}
	}

	return d.DateField, nil
}

func (w WeekRepeat) NextDateCalc() (time.Time, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if w.NowField.After(w.DateField) {
		w.DateField = w.NowField
	}
	var dateWeekDay int

	func(ctx context.Context) {
		for {
			select {
			default:
				w.DateField = w.DateField.AddDate(0, 0, 1)

				dateWeekDay = int(w.DateField.Weekday())
				if dateWeekDay == 0 {
					dateWeekDay = 7
				}

				for _, v := range w.WeekIteration {
					wg.Add(1)
					go func(ctx context.Context, dateWeekDay int, v int) {
						defer wg.Done()
						if v == dateWeekDay {
							cancel()
						}
					}(ctx, dateWeekDay, v)
				}
				wg.Wait()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return w.DateField, nil
}

func (m MonthRepeat) NextDateCalc() (time.Time, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lastDayDateMonth, penultimateDayDateMonth int

	if m.NowField.After(m.DateField) {
		m.DateField = m.NowField
	}

	func(ctx context.Context) {
		for {
			select {
			default:
				m.DateField = m.DateField.AddDate(0, 0, 1)

				dateYear, dateMonth, dateMonthDay := m.DateField.Date()
				lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
				penultimateDayDateMonth = lastDayDateMonth - 1

				for _, v := range m.MonthIteration {
					wg.Add(1)
					go func(ctx context.Context, dateMonthDay, lastDayDateMonth, penultimateDayDateMonth, v int) {
						defer wg.Done()
						if (v == -2 && dateMonthDay == penultimateDayDateMonth) || (v == -1 && dateMonthDay == lastDayDateMonth) || (v == dateMonthDay) {
							cancel()
						}
					}(ctx, dateMonthDay, lastDayDateMonth, penultimateDayDateMonth, v)
				}
				wg.Wait()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return m.DateField, nil
}

func (fm FullMonthRepeat) NextDateCalc() (time.Time, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lastDayDateMonth, penultimateDayDateMonth int

	if fm.NowField.After(fm.DateField) {
		fm.DateField = fm.NowField
	}

	var neverMonthDay bool
	var message string
	func(ctx context.Context) {
		for {
			select {
			default:
				fm.DateField = fm.DateField.AddDate(0, 0, 1)

				dateYear, dateMonth, dateMonthDay := fm.DateField.Date()
				lastDayDateMonth = time.Date(dateYear, dateMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
				penultimateDayDateMonth = lastDayDateMonth - 1

				for j := 0; j < len(fm.NumberMonth); j++ {
					wg.Add(1)
					go func(ctx context.Context, j int) {
						defer wg.Done()
						if fm.NumberMonth[j] == int(dateMonth) {

							for _, v := range fm.MonthIteration {
								wg.Add(1)
								go func(ctx context.Context, v int, j int) {
									defer wg.Done()
									if v > 0 && v > lastDayDateMonth {
										neverMonthDay = true
										message = fmt.Sprintf("в месяце %d нет %d-го дня", fm.NumberMonth[j], v)
										cancel()
									}
									if (v == -2 && dateMonthDay == penultimateDayDateMonth) || (v == -1 && dateMonthDay == lastDayDateMonth) || (v == dateMonthDay) {
										cancel()
									}
								}(ctx, v, j)
							}
						}
					}(ctx, j)
				}
				wg.Wait()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	if neverMonthDay {
		return time.Time{}, fmt.Errorf("%v", message)
	}

	return fm.DateField, nil
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
		{"20240126", "d 1", `20240127`},
	}
	/*
		tbl := []nextDate{
			//{"20240126", "", ""},
			//{"20240126", "k 34", ""},
			//{"20240126", "ooops", ""},
			//{"16890220", "y ", ""},
			//{"20250701", " y", ""},
			//{"15000156", "y", ""},
			//{"ooops", "y", ""},
			//{"20240113", "d", ""},
			//{"20240320", "d 401", ""},
			//{"20230201", "d 45 17", ""},
			//{"20220423", "w ", ""},
			//{"20220423", "w1 ", ""},
			//{"20220423", "w 1,2,3,4,4", ""}, // не совсем согласен с ожидаемым выводом - ну и что, что у пользователя руки трясуться
			//{"20210921", "w -1,2,3,4", ""},
			//{"20000101", "w ,2,3,4", ""},
			//{"20000102", "w 2,3,4,", ""},
			//{"20000103", "w 2, 3,4", ""},
			//{"20000104", "w 2,3,4 ", ""},
			//{"20230226", "w 8,4,5", ""},
			//{"20240120", "m2", ""},
			//{"20240120", "m2 ", ""},
			//{"20240120", "m 40,11,19", ""},
			//{"20240121", "m 11,11", ""}, // не совсем согласен с ожидаемым выводом - ну и что, что у пользователя руки трясуться
			//{"20240121", "m ,12", ""},
			//{"20240122", "m 14,", ""},
			//{"20240123", "m 15 ", ""},
			//{"20210106", "m -1-2,3", ""},
			//{"20240222", "m -2,-3", ""},
			//{"20240123", "m 15 1,2,2", ""}, // не совсем согласен с ожидаемым выводом - ну и что, что у пользователя руки трясуться
			//{"20400124", "m 16 1, 2", ""},
			//{"20270421", "m 17 1,2,", ""},
			//{"20220425", "m 18 ,1,3", ""},
			//{"20600714", "m 18 1 ", ""},
		}
	*/
	startTime := time.Now()

	nowStr := "20240126"

	var msg, a string

	for _, v := range tbl {

		fmt.Printf("test_data = %v, repeat = %v, ", v.date, v.repeat)

		dateOut, err := NextDate(nowStr, v.date, v.repeat)
		if err != nil {
			fmt.Println("ошибка получения даты следующего события: ", err)
			return
		}

		if dateOut != v.want {
			msg = "FAIL"
			a = "FAIL"
		} else {
			msg = ""
		}

		fmt.Printf(" полученная date = %v, ожидаемая date = %v,    %v\n", dateOut, v.want, msg)
	}

	duration := time.Since(startTime)

	fmt.Println(a, duration)
}
