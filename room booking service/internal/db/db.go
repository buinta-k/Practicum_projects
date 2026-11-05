package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"project/internal/models"

	_ "modernc.org/sqlite"
)

type Base struct {
	Conn *sql.DB
}

func Init(dbFile string) (*Base, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile("init.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(file))
	if err != nil {
		return nil, err
	}

	return &Base{
		Conn: db,
	}, nil
}

func (s *Base) CreateRoom(ctx context.Context, name string) (int64, error) {
	query := `INSERT INTO rooms (name) VALUES (?)`
	res, err := s.Conn.ExecContext(ctx, query, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Base) GetRooms(ctx context.Context) ([]models.Room, error) {
	var rooms []models.Room
	query := `SELECT id, name FROM rooms`
	rows, err := s.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Name)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (s *Base) DeleteBooking(ctx context.Context, id int) (int64, error) {
	query := `UPDATE bookings SET cancelled = ? WHERE id = ? AND cancelled = FALSE`

	res, err := s.Conn.ExecContext(ctx, query, true, id)
	if err != nil {
		log.Println("Ошибка SQL при отмене:", err)
		return 0, err
	}

	n, err := res.RowsAffected()
	log.Printf("Отмена брони: id=%d, изменено строк=%d\n", id, n)

	return n, err
}

func (s *Base) GetBookings(ctx context.Context, roomID int) ([]models.Booking, error) {
	bookings := []models.Booking{}
	query := `SELECT id, room_id, start_at, end_at FROM bookings WHERE room_id = ? AND cancelled = false`
	rows, err := s.Conn.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking models.Booking

		err = rows.Scan(&booking.ID, &booking.RoomID, &booking.StartAt, &booking.EndAt)
		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bookings, err
}

func (s *Base) GetAvailableRooms(ctx context.Context, start string, end string) ([]models.Room, error) {
	rooms := []models.Room{}

	query := `SELECT id, name FROM rooms r
			WHERE NOT EXISTS(
					SELECT 1 FROM bookings b
					WHERE b.room_id = r.id
					AND b.cancelled = FALSE
					AND b.start_at < ?
					AND b.end_at > ?
			)`

	rows, err := s.Conn.QueryContext(ctx, query, end, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Room

		err = rows.Scan(&r.ID, &r.Name)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rooms, nil
}
