CREATE INDEX IF NOT EXISTS games_title_idx ON games USING GIN (to_tsvector('simple', `name`));
CREATE INDEX IF NOT EXISTS games_genres_idx ON games USING GIN (genres);
