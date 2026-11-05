package api

import (
	"encoding/json"
	"net/http"
	"time"

	"project/internal/models"
)

func (s *Api) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room models.Room

	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ошибка парсинга")
		return
	}

	if room.Name == "" {
		writeError(w, http.StatusBadRequest, "пустая строка")
		return
	}

	id, err := s.DB.CreateRoom(r.Context(), room.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка добавления в бд")
		return
	}

	room.ID = id
	writeJson(w, room)
}

func (s *Api) GetRooms(w http.ResponseWriter, r *http.Request) {
	var rooms []models.Room
	rooms, err := s.DB.GetRooms(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка извлечения комнат")
		return
	}

	writeJson(w, rooms)
}

func (s *Api) GetAvailableRooms(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_at")
	endStr := r.URL.Query().Get("end_at")

	start, err := time.Parse(layout, startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "дата не соответсвует формату")
		return
	}

	end, err := time.Parse(layout, endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "дата не соответствует формату")
		return
	}

	if !start.Before(end) {
		writeError(w, http.StatusBadRequest, "дата start !< end")
		return
	}

	rooms, err := s.DB.GetAvailableRooms(r.Context(), start.Format(layout), end.Format(layout))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка поиска свободных комнат")
		return
	}
	writeJson(w, rooms)
}
