package app

import (
	"github.com/gin-gonic/gin"
)

type routeServer struct {
	ping           string
	register       string
	login          string
	logout         string
	getStatus      string
	getLeaderboard string
	taskComplete   string
	referral       string
}

func newRouteServer() *routeServer {
	return &routeServer{
		register:       "/register",          // Путь: /users/register
		login:          "/login",             // Путь: /users/login
		logout:         "/logout",            // Путь: /users/logout
		getStatus:      "/:id/status",        // Путь: /users/:id/status
		getLeaderboard: "/leaderboard",       // Путь: /users/leaderboard
		taskComplete:   "/:id/task/complete", // Путь: /users/:id/task/complete
		referral:       "/:id/referrer",      // Путь: /users/:id/referrer
	}
}

func (app *App) configureApiRoutes(r *gin.Engine) {
	route := newRouteServer()

	// Группа маршрутов /users
	users := r.Group("/users")
	{
		// Публичные маршруты (не требуют аутентификации)
		users.POST(route.register, app.userHandler.RegisterHandler) // Путь: /users/register
		users.POST(route.login, app.userHandler.LoginHandler)       // Путь: /users/login
	}

	// Приватные маршруты (с защитой через middleware)
	privateUsers := users.Group("/")
	privateUsers.Use(app.authMiddleware.AuthMiddleware()) // Применяем middleware аутентификации

	{
		privateUsers.GET(route.getStatus, app.userHandler.UserStatusHandler)            // Путь: /users/:id/status
		privateUsers.GET(route.getLeaderboard, app.userHandler.UsersLeaderboardHandler) // Путь: /users/leaderboard
		privateUsers.POST(route.taskComplete, app.userHandler.TaskCompleteHandler)      // Путь: /users/:id/task/complete
		privateUsers.POST(route.referral, app.userHandler.ReferrerHandler)              // Путь: /users/:id/referrer
		privateUsers.POST(route.logout, app.userHandler.LogoutHandler)                  // Путь: /users/logout
	}
}
