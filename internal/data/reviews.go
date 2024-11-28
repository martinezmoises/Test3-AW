package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/martinezmoises/Test3/internal/validator"
)

type Review struct {
	ID         int64     `json:"id"`
	BookID     int64     `json:"book_id"`
	UserID     int64     `json:"user_id"`
	Rating     float64   `json:"rating"`
	Review     string    `json:"review"`
	ReviewDate time.Time `json:"review_date"`
}

type ReviewModel struct {
	DB *sql.DB
}

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
	v.Check(review.Review != "", "review", "must not be empty")
}

func (m ReviewModel) GetAll(bookID int64) ([]*Review, error) {
	query := `
        SELECT id, book_id, user_id, rating, review, review_date
        FROM reviews
        WHERE book_id = $1
        ORDER BY review_date DESC`

	rows, err := m.DB.Query(query, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.BookID, &review.UserID, &review.Rating, &review.Review, &review.ReviewDate); err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}
	return reviews, nil
}

func (m ReviewModel) Insert(review *Review) error {
	query := `
        INSERT INTO reviews (book_id, user_id, rating, review)
        VALUES ($1, $2, $3, $4)
        RETURNING id, review_date`

	args := []any{review.BookID, review.UserID, review.Rating, review.Review}
	return m.DB.QueryRow(query, args...).Scan(&review.ID, &review.ReviewDate)
}

func (m ReviewModel) Update(review *Review) error {
	query := `
		UPDATE reviews
		SET rating = $1, review = $2, review_date = NOW()
		WHERE id = $3
		RETURNING review_date`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query,
		review.Rating,
		review.Review,
		review.ID,
	).Scan(&review.ReviewDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (m ReviewModel) Get(id int64) (*Review, error) {
	query := `
		SELECT id, book_id, user_id, rating, review, review_date
		FROM reviews
		WHERE id = $1`

	var review Review
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.BookID,
		&review.UserID,
		&review.Rating,
		&review.Review,
		&review.ReviewDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &review, nil
}

func (m ReviewModel) Delete(id int64, userID int64) error {
	query := `
        DELETE FROM reviews
        WHERE id = $1 AND user_id = $2`

	result, err := m.DB.Exec(query, id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ReviewModel) GetByUserID(userID int64) ([]*Review, error) {
	query := `
        SELECT id, book_id, user_id, rating, review, review_date
        FROM reviews
        WHERE user_id = $1
    `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		err := rows.Scan(
			&review.ID,
			&review.BookID,
			&review.UserID,
			&review.Rating,
			&review.Review,
			&review.ReviewDate,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}
