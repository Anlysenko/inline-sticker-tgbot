CREATE TABLE IF NOT EXISTS users (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	state text NOT NULL,
	event text NOT NULL
);

CREATE TABLE IF NOT EXISTS stickers (
	unique_id text NOT NULL,
	id text NOT NULL,
	tags text NOT NULL,
	user_id bigserial NOT NULL REFERENCES users ON DELETE CASCADE,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	PRIMARY KEY(unique_id, user_id)
);

CREATE INDEX IF NOT EXISTS stickers_tags_idx ON stickers USING GIN (to_tsvector('simple', tags));