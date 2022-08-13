package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v9"
)

var ctx = context.TODO()

type uidChidRedis struct {
	Event EventType `json:"event"`
	State StateType `json:"state"`
}

func createKeyRedis(uname string, uid int64) string {
	return fmt.Sprintf("%s:%d", uname, uid)
}

func createValRedis(event EventType, state StateType) string {
	return fmt.Sprintf(`{"event":"%s","state":"%s"}`, event, state)
}

func setStateRedis(key string, value string) error {
	if err := rdb.Set(ctx, key, value, 0).Err(); err != nil {
		return err
	}
	return nil
}

func GetStateRedis(uname string, uid int64) (*uidChidRedis, error) {
	var result uidChidRedis

	key := createKeyRedis(uname, uid)
	val, err := rdb.Get(ctx, key).Result()

	if err == redis.Nil {
		err := setStateRedis(key, createValRedis(Nop, Def))
		if err != nil {
			return nil, err
		}
		result = uidChidRedis{Event: Nop, State: Def}
	} else if err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func UpdateDataRedis(uname string, chid int64, data *uidChidRedis) error {
	key := createKeyRedis(uname, chid)
	val := createValRedis(data.Event, data.State)
	if err := setStateRedis(key, val); err != nil {
		return err
	}
	return nil
}
