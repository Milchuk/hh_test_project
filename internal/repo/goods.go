package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hh_test_project/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("goods not found")
)

type GoodsRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewGoodsRepository(db *sql.DB, redis *redis.Client) *GoodsRepository {
	return &GoodsRepository{db: db, redis: redis}
}

func (r *GoodsRepository) Create(ctx context.Context, projectID int, input models.GoodsCreate) (*models.Goods, error) {
	query := `
		INSERT INTO goods (project_id, name)
		VALUES ($1, $2)
		RETURNING id, project_id, name, priority, removed, created_at
	`

	var goods models.Goods
	err := r.db.QueryRowContext(ctx, query, projectID, input.Name).Scan(
		&goods.ID,
		&goods.ProjectID,
		&goods.Name,
		&goods.Priority,
		&goods.Removed,
		&goods.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create goods failed: %w", err)
	}

	return &goods, nil
}

func (r *GoodsRepository) GetByID(ctx context.Context, projectID, goodsID int) (*models.Goods, error) {
	// Попытка получить из Redis
	cacheKey := fmt.Sprintf("goods:%d:%d", projectID, goodsID)
	cached, redis_err := r.redis.Get(ctx, cacheKey).Result()
	if redis_err == nil {
		var goods models.Goods
		if err := json.Unmarshal([]byte(cached), &goods); err == nil {
			return &goods, nil
		}
	}

	query := `
		SELECT id, project_id, name, description, priority, removed, created_at
		FROM goods
		WHERE id = $1 AND project_id = $2 AND removed = false
	`

	var goods models.Goods
	err := r.db.QueryRowContext(ctx, query, goodsID, projectID).Scan(
		&goods.ID,
		&goods.ProjectID,
		&goods.Name,
		&goods.Description,
		&goods.Priority,
		&goods.Removed,
		&goods.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get goods failed: %w", err)
	}

	// Сохраняем в Redis на 1 минуту
	goodsJSON, _ := json.Marshal(goods)
	r.redis.Set(ctx, cacheKey, goodsJSON, time.Minute)

	return &goods, nil
}

func (r *GoodsRepository) Update(ctx context.Context, projectID, goodsID int, input models.GoodsUpdate) (*models.Goods, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Блокировка записи
	_, err = tx.ExecContext(ctx, "SELECT 1 FROM goods WHERE id = $1 AND project_id = $2 FOR UPDATE", goodsID, projectID)
	if err != nil {
		return nil, fmt.Errorf("lock goods failed: %w", err)
	}

	query := `
		UPDATE goods
		SET 
			name = COALESCE($1, name),
			description = COALESCE($2, description)
		WHERE id = $3 AND project_id = $4
		RETURNING id, project_id, name, description, priority, removed, created_at
	`

	var updated models.Goods
	err = tx.QueryRowContext(ctx, query,
		input.Name,
		input.Description,
		goodsID,
		projectID,
	).Scan(
		&updated.ID,
		&updated.ProjectID,
		&updated.Name,
		&updated.Description,
		&updated.Priority,
		&updated.Removed,
		&updated.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update goods failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	return &updated, nil
}

func (r *GoodsRepository) Delete(ctx context.Context, projectID, goodsID int) error {
	query := `
		UPDATE goods
		SET removed = true
		WHERE id = $1 AND project_id = $2
		RETURNING id
	`

	var deletedID int
	err := r.db.QueryRowContext(ctx, query, goodsID, projectID).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("delete goods failed: %w", err)
	}

	return nil
}

func (r *GoodsRepository) ProjectExists(ctx context.Context, projectID int) (bool, error) {
	if projectID <= 0 {
		return false, nil
	}

	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return exists, nil
}
