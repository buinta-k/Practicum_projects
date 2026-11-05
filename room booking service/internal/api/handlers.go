package api

import (
	"encoding/json"
	"net/http"
	"project/internal/models"
	"strconv"
	"time"
)

// Формат дат: 2026-10-16. Такие строки сортируются и сравниваются
// в SQL правильно, как обычный текст.
const layout = "2006-01-02"

type CreateBookingResponse struct {
	ID int64 `json:"id"`
}

func (s *Api) CreateBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсинг JSON
	var req models.Booking
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// 2. Валидация
	if req.RoomID <= 0 || req.StartAt == "" || req.EndAt == "" {
		writeError(w, http.StatusBadRequest, "room_id, start_at, end_at are required")
		return
	}

	start, err := time.Parse(layout, req.StartAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_at must be YYYY-MM-DD")
		return
	}
	end, err := time.Parse(layout, req.EndAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "end_at must be YYYY-MM-DD")
		return
	}
	if !start.Before(end) {
		writeError(w, http.StatusBadRequest, "start_at must be before end_at")
		return
	}

	// 3. Транзакция: проверки и вставка выполняются как одно целое
	tx, err := s.DB.Conn.BeginTx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	// Если до Commit что-то пошло не так, изменения откатятся.
	// После успешного Commit этот Rollback просто ничего не делает.
	defer tx.Rollback()

	// 3.1 Комната существует?
	var roomExists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM rooms WHERE id = ?)`, req.RoomID,
	).Scan(&roomExists)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !roomExists {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	// 3.2 Пересекается ли с другой бронью этой комнаты?
	// Два интервала пересекаются, если: старая.начало < новая.конец И старая.конец > новая.начало
	var overlap bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE room_id = ? AND cancelled = FALSE AND start_at < ? AND end_at > ?
		)`,
		req.RoomID, end.Format(layout), start.Format(layout),
	).Scan(&overlap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if overlap {
		writeError(w, http.StatusConflict, "room is already booked for these dates")
		return
	}

	// 3.3 Вставка
	res, err := tx.ExecContext(ctx,
		`INSERT INTO bookings (room_id, start_at, end_at) VALUES (?, ?, ?)`,
		req.RoomID, start.Format(layout), end.Format(layout),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// 4. Ответ
	writeJson(w, CreateBookingResponse{ID: id})
}

func (s *Api) DeleteBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "неверный id")
		return
	}

	n, err := s.DB.DeleteBooking(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удаляется бронь")
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "бронь не найдена")
		return
	}
	writeJson(w, map[string]int{"id": id})
}

func (s *Api) GetBookings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID, err := strconv.Atoi(r.URL.Query().Get("room_id"))
	if err != nil || roomID <= 0 {
		writeError(w, http.StatusBadRequest, "неверный room_id")
		return
	}

	bookings, err := s.DB.GetBookings(ctx, roomID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка извлечения брони")
		return
	}

	writeJson(w, bookings)

}
