package api

import fmt


func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case:
	case:
	default:

	}
/*
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
*/
}