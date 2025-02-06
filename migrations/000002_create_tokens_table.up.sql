CREATE TABLE IF NOT EXISTS tokens (
    id BIGSERIAL PRIMARY KEY,                       -- Уникальный идентификатор токена
    user_id BIGINT NOT NULL,                        -- ID пользователя, получившего токен
    token VARCHAR(255) NOT NULL,                    -- Токен
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Дата и время создания токена
    expires_at TIMESTAMP NOT NULL,                  -- Дата и время истечения срока действия токена
    is_revoked BOOLEAN DEFAULT FALSE                -- Флаг, указывающий отозван ли токен
    );

-- Индекс для быстрого поиска по токену
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(token);