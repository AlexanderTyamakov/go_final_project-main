package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	if nowParam == "" || dateParam == "" || repeatParam == "" {
		http.Error(w, "Недостаточно параметров", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		http.Error(w, "Неверный формат даты now", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при вычислении следующей даты: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	switch {
	case repeat == "":
		return "", errors.New("пустое правило повторения")
	case repeat == "y":
		// Годичное повторение
		nextDate := startDate.AddDate(1, 0, 0)
		return nextDate.Format("20060102"), nil
	case len(repeat) > 2 && repeat[:2] == "d ":
		// Повторение через заданное количество дней
		days, err := strconv.Atoi(repeat[2:])
		if err != nil {
			return "", err
		}
		if days <= 0 || days > 400 {
			return "", errors.New("недопустимый интервал дней")
		}
		nextDate := startDate.AddDate(0, 0, days)
		return nextDate.Format("20060102"), nil
	default:
		return "", errors.New("неподдерживаемый формат правила повторения")
	}
}
