CREATE TABLE IF NOT EXISTS mood (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    icon VARCHAR(50) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT ux_user_date UNIQUE (user_id, date)
);

-- Індекс для швидкого пошуку за датою
CREATE INDEX IF NOT EXISTS idx_mood_user_date ON mood(user_id, date);
