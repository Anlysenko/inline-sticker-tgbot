package main

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrRecordNotFound = errors.New("record not found")

type Sticker struct {
	uniqueID string
	id       string
	tags     string
	userID   int64
}

type StickerModel struct {
	DB *sql.DB
}

func (s StickerModel) GetStickerTags(userID int64, stickerUniqueID string) (string, error) {
	q := `
		SELECT tags FROM stickers 
		WHERE user_id = $1 AND unique_id = $2`

	var sticker Sticker
	err := s.DB.QueryRow(q, userID, stickerUniqueID).Scan(&sticker.tags)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", ErrRecordNotFound
	case err != nil:
		return "", err
	default:
		return sticker.tags, nil
	}
}

func (s StickerModel) DeleteSticker(userID int64, uid string) error {
	q := `
		DELETE FROM stickers
		WHERE user_id = $1 AND unique_id = $2`

	result, err := s.DB.Exec(q, userID, uid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (s StickerModel) UpdateTags(userID int64, tags string) error {
	q := `
		UPDATE stickers
		SET tags = $1
		WHERE user_id = $2`

	if _, err := s.DB.Exec(q, tags, userID); err != nil {
		return err
	}
	return nil
}

func (s StickerModel) InsertSticker(stick *Sticker) error {
	q := `
		INSERT INTO stickers (unique_id, id, tags, user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(unique_id, user_id)
		DO UPDATE SET tags = $3`

	args := []any{stick.uniqueID, stick.id, stick.tags, stick.userID}

	if _, err := s.DB.Exec(q, args...); err != nil {
		return err
	}
	return nil
}

func (s StickerModel) GetSticker(userID int64, query string, queryOffset int) ([]*Sticker, error) {
	q := `
		SELECT id FROM stickers
		WHERE 
			CASE
				WHEN $2 = '' THEN user_id = $1
				WHEN $2 LIKE '//%' AND length($2) = 2 THEN TRUE
				WHEN $2 LIKE '//%' THEN to_tsvector('simple', tags) @@
					plainto_tsquery('simple', substr($2, 3))
				ELSE to_tsvector('simple', tags) @@
					plainto_tsquery('simple', $2) AND user_id = $1
			END
		ORDER BY created_at DESC
		LIMIT 50 OFFSET $3`

	rows, err := s.DB.Query(q, userID, query, queryOffset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stickers := []*Sticker{}

	for rows.Next() {
		var s Sticker

		err := rows.Scan(&s.id)
		if err != nil {
			return nil, err
		}

		stickers = append(stickers, &s)
	}

	return stickers, nil
}

type User struct {
	ID    int64
	Name  string
	State string
	Event string
}

type UserModel struct {
	DB *sql.DB
}

func (u UserModel) GetByUser(user *User) error {
	q := `
		SELECT state, event
		FROM users 
		WHERE id = $1`

	err := u.DB.QueryRow(q, user.ID).Scan(&user.State, &user.Event)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) Upsert(user User) error {
	q := `
		INSERT INTO users (id, name, state, event)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id)
		DO NOTHING`

	args := []any{user.ID, user.Name, user.State, user.Event}

	if _, err := u.DB.Exec(q, args...); err != nil {
		return err
	}
	return nil
}

func (u UserModel) Update(user *User) error {
	q := `
		UPDATE users
		SET state = $1, event = $2
		WHERE id = $3`

	args := []any{user.State, user.Event, user.ID}

	if _, err := u.DB.Exec(q, args...); err != nil {
		return err
	}
	return nil
}
