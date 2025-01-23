package main

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type RedisGrid struct {
	Connection redis.Client
}

func NewRedisGrid() *RedisGrid {
	return &RedisGrid{
		Connection: *redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}),
	}
}

func (r *RedisGrid) Init(ctx context.Context) {
	r.Connection.Del(ctx, "grid_cells")
	r.Connection.Del(ctx, "grid_empty_cell")
	for i := 0; i < 10; i++ {
		r.Connection.SAdd(ctx, "grid_empty_cell", i)
	}
}

func (r *RedisGrid) GetEmptyCell(ctx context.Context) (int, error) {
	v, err := r.Connection.SPop(ctx, "grid_empty_cell").Result()
	if err != nil {
		return -1, err
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func (r *RedisGrid) SetEmptyCell(ctx context.Context, i int) error {
	r.Connection.HSet(ctx, "grid_cells", i, "")
	return r.Connection.SAdd(ctx, "grid_empty_cell", i).Err()

}

func (r *RedisGrid) SetCell(ctx context.Context, i int, username string) error {
	return r.Connection.HSet(ctx, "grid_cells", i, username).Err()
}

type Cell struct {
	Cell     int
	Username string
}

func (r *RedisGrid) GetGrid(ctx context.Context) ([]Cell, error) {
	var cells []Cell
	keys, err := r.Connection.HKeys(ctx, "grid_cells").Result()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		v, err := r.Connection.HGet(ctx, "grid_cells", k).Result()
		if err != nil {
			return nil, err
		}
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		cells = append(cells, Cell{Cell: i, Username: v})
	}
	return cells, nil
}
