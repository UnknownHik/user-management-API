package models

import "time"

type User struct {
	ID            int       `db:"id"`
	UserName      string    `db:"username"`
	Password      string    `db:"password"`
	Balance       int       `db:"balance"`
	UpdateBalance time.Time `db:"updated_balance"`
	Referrer      *int      `db:"referrer"`
	CreatedAt     time.Time `db:"created_at"`
}

type Task struct {
	ID          int    `db:"id"`
	Description string `db:"description"`
	Reward      int    `db:"reward"`
}
