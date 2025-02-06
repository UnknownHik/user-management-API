CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,              -- Уникальный идентификатор задания
    description VARCHAR(255) NOT NULL,  -- Описание задания
    reward INT NOT NULL                 -- Награда за задание
);

-- Добавление заданий в таблицу
INSERT INTO tasks (description, reward) VALUES ('Subscribe to Telegram', 50);
INSERT INTO tasks (description, reward) VALUES ('Subscribe to Twitter', 30);
INSERT INTO tasks (description, reward) VALUES ('Subscribe to YouTube', 40);

-- Добавление индекса на поле 'description' для быстрого поиска задания по описанию
CREATE INDEX IF NOT EXISTS idx_tasks_description ON tasks(description);