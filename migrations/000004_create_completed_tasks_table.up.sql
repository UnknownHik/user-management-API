CREATE TABLE IF NOT EXISTS completed_tasks (
    id BIGSERIAL PRIMARY KEY,                                       -- Уникальный идентификатор токена
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,    -- Внешний ключ, ссылающийся на id из таблицы users
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,    -- Внешний ключ, ссылающийся на id из таблицы tasks
    completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,               -- Дата и время выполнения задания
    UNIQUE(user_id, task_id)                                        -- Ограничение на повторное выполнения задания
    );