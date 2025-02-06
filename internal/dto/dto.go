package dto

import "time"

// UserRegLogDTO представляет данные для регистрации и входа пользователя
type UserRegLogDTO struct {
	UserName string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// UserLoginDTO представляет данные для ответа на вход пользователя
type UserLoginDTO struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
}

// UserStatusDTO представляет данные о пользователе
type UserStatusDTO struct {
	ID             int       `json:"id"`
	UserName       string    `json:"username"`
	Balance        int       `json:"balance"`
	UpdatedBalance time.Time `json:"updated_balance"`
	Referrer       *int      `json:"referrer_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserLeaderDTO представляет данные о пользователе для leaderboard
type UserLeaderDTO struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
	Balance  int    `json:"balance"`
}

type TaskDTO struct {
	ID int `json:"task_id"`
}

// ReferrerDTO представляет данные для добавления реферера
type ReferrerDTO struct {
	Referrer int `json:"referrer_id"`
}
