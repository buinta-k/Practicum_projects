package models

type Room struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Booking struct {
	ID        int    `json:"id"`
	RoomID    int    `json:"room_id"`
	StartAt   string `json:"start_at"`
	EndAt     string `json:"end_at"`
	Cancelled bool   `json:"cancelled"`
}
