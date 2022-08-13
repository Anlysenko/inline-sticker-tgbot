BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "stickers" (
	"user_id"			INTEGER NOT NULL,
	"sticker_id"		TEXT NOT NULL,
	"sticker_unique_id" TEXT NOT NULL,
	"description"		TEXT NOT NULL,
	PRIMARY KEY("user_id","sticker_unique_id")
);
COMMIT;