CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,                                 -- Идентификатор пользователя
    username VARCHAR(255) NOT NULL,                        -- Имя пользователя
    password VARCHAR(255) NOT NULL,                        -- Пароль
    balance  BIGINT DEFAULT 0,                             -- Баланс поинтов
    updated_balance TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, -- Время последнего обновления баланса
    referrer BIGINT,                                       -- Реферальный код
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP       -- Время создания
    );

-- Добавление индекса на поле 'name' для быстрого поиска пользователей по имени
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);