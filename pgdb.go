package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type stickers struct {
	stickerID          string
	stickerDescription string
}

func GetStickerDescriptionPG(userID int64, stickerUniqueID string) (string, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return "", err
	}
	defer db.Close()

	q := `SELECT description from stickers 
	WHERE user_id = $1 AND sticker_unique_id = $2`

	description := stickers{}
	err = db.QueryRow(q, userID, stickerUniqueID).Scan(&description.stickerDescription)

	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", err
	default:
		return description.stickerDescription, nil
	}
}

func DeleteStickerPG(userID int64, col string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	q := `DELETE from stickers
	WHERE user_id = $1 AND sticker_unique_id = $2`

	if col == "~" {
		q = `DELETE from stickers
	WHERE user_id = $1 AND description = $2`
	}

	if _, err := db.Exec(q, userID, col); err != nil {
		return err
	}
	return nil
}

func UpdateDescriptionPG(userID int64, description string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	q := `UPDATE stickers
	SET description = $2
	WHERE user_id = $1 AND description = '~'`

	if _, err = db.Exec(q, userID, description); err != nil {
		return err
	}
	return nil
}

func InsertStickerPG(userID int64, stickerID string, stickerUniqueID string, description string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	q := `INSERT INTO stickers (user_id, sticker_id, sticker_unique_id, description)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT(user_id, sticker_unique_id)
	DO UPDATE SET description = $4`

	if _, err = db.Exec(q, userID, stickerID, stickerUniqueID, description); err != nil {
		return err
	}
	return nil
}

func GetStickerPG(userID int64, searchQuery string, queryOffset int) ([]stickers, error) {
	if searchQuery == "" {
		return nil, nil
	}

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := `SELECT sticker_id from stickers 
	WHERE user_id = $1 AND description LIKE '%' || $2 || '%'
	OFFSET $3 LIMIT 50`

	rows, err := db.Query(q, userID, searchQuery, queryOffset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sticks := []stickers{}
	for rows.Next() {
		s := stickers{}
		err := rows.Scan(&s.stickerID)
		if err != nil {
			log.Println("[Scan]", err)
			continue
		}
		sticks = append(sticks, s)
	}
	return sticks, nil
}

func CreateTablePG() error {
	s, err := os.ReadFile("db.sql")
	if err != nil {
		return err
	}

	fmt.Println(dbInfo)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err = db.Exec(string(s)); err != nil {
		return err
	}
	return nil
}
