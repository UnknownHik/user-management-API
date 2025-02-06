# User Management API
## Описание

Простое RESTful API для управления пользователями, реализованное на Go с использованием фреймворка Gin и базы данных PostgreSQL. API позволяет создавать пользователей, управлять их задачами, отслеживать статус и участвовать в реферальной системе.

## Функционал

-   **Авторизация**: Middleware для проверки Access Token (JWT)
-   **Управление пользователями**:
    -   Получение информации о пользователе
    -   Выполнение заданий
    -   Ввод реферального кода
    -   Просмотр топа пользователей по количеству поинтов
-   **Хранение данных**: PostgreSQL + миграции через golang-migrate
-   **Docker**: Запуск через docker-compose

## Установка и запуск

### 1. Настройка переменных окружения

Укажите корректные настройки базы данных и JWT в файле `.env`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=youruser
DB_PASSWORD=yourpassword
DB_NAME=yourdb
JWT_SECRET=your_secret_key
```

### 2. Запуск через Docker

```
docker-compose up --build
```


## API Эндпоинты
### 1. Регистрация пользователя
```
POST /users/register
```

Тело запроса:

```
{
  "username":  "TommyVercetti",
  "password":  "FaNnYmAgNeT"
}
```

Ответ:

```
{
  "status":  "Успешная регистрация",
  "user_id":  1
}
```

### 2. Логин пользователя
```
POST /users/login
```

Тело запроса:

```
{
  "username":  "TommyVercetti",
  "password":  "FaNnYmAgNeT"
}
```

Ответ:

```
{

  "expires_at":  "2024-12-24T22:00:00.000000Z",
  "status":  "Авторизация успешна",
  "token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzg5NzQyNTQsInVzZXJfaWQiOjF9.sqoHWcm3dqVuJaR-YIk25DgPyu1lqLwoosWD8g6EgS0"

}
```

### 3. Получение информации о пользователе

```
GET /users/{id}/status
```

Ответ:

```
{
  "id":  1,
  "username":  "TommyVercetti",
  "balance":  500,
  "updated_balance":  "2024-12-25T07:00:00.000000Z",
  "referrer":  2,
  "created_at":  "2024-11-11T11:11:11.000000Z"
}
```

### 4. Топ пользователей

```
GET /users/leaderboard
```

Ответ:

```
[
  {"id": 1, "username":  "TommyVercetti", "balance": 500},
  {"id": 3, "username":  "NikoBellic", "points": 300}
]
```

### 5. Выполнение задания

```
POST /users/{id}/task/complete
```

Тело запроса:

```
{
  "task":  "Return to Vice City"
}
```

Ответ:

```
{
  "status":  "Задание выполнено",
  "task":  "Return to Vice City"
}
```

### 6. Ввод реферального кода

```
POST /users/{id}/referrer
```

Тело запроса:

```
{
  "referrer": 2
}
```


Ответ:

```
{
  "referrer":  2,
  "status":  "Реферер добавлен"
}
```

### 6. Logout пользователя

```
POST /users/logout
```

Ответ:

```
{
  "status": "Вышел из системы"
}
```