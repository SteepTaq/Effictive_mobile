package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rest-service/internal/model"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	q := `INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date)
          VALUES ($1,$2,$3,$4,$5) RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, q, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&id)
	if err != nil {
		return model.Subscription{}, err
	}
	s.ID = id
	return s, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (model.Subscription, error) {
	q := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id=$1`
	var s model.Subscription
	err := r.pool.QueryRow(ctx, q, id).Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate)
	if err != nil {
		return model.Subscription{}, err
	}
	return s, nil
}

func (r *Repository) Update(ctx context.Context, id int64, s model.Subscription) (model.Subscription, error) {
	q := `UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5 WHERE id=$6`
	if _, err := r.pool.Exec(ctx, q, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, id); err != nil {
		return model.Subscription{}, err
	}
	s.ID = id
	return s, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) List(ctx context.Context, limit, offset int, userID uuid.UUID, serviceName string) ([]model.Subscription, error) {
	q := `SELECT id, service_name, price, user_id, start_date, end_date
          FROM subscriptions
          WHERE user_id = $1 AND service_name = $2
          ORDER BY id DESC
          LIMIT $3 OFFSET $4`
	rows, err := r.pool.Query(ctx, q, userID, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *Repository) SumByPeriod(ctx context.Context, from, to time.Time, userID uuid.UUID, serviceName string) (int64, error) {
	q := `SELECT COALESCE(SUM(price), 0)
          FROM subscriptions
          WHERE start_date <= $1 
            AND (end_date IS NULL OR end_date >= $2)
            AND user_id = $3 
            AND service_name = $4`

	var total int64
	if err := r.pool.QueryRow(ctx, q, to, from, userID, serviceName).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
