package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/martinezmoises/Test3/internal/validator"
)

type ReadingList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   int64     `json:"created_by"`
	Books       []int64   `json:"books"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReadingListModel struct {
	DB *sql.DB
}

func ValidateReadingList(v *validator.Validator, rl *ReadingList) {
	v.Check(rl.Name != "", "name", "must be provided")
	v.Check(len(rl.Name) <= 200, "name", "must not be more than 200 characters long")
	v.Check(rl.Description != "", "description", "must be provided")
	v.Check(len(rl.Description) <= 500, "description", "must not be more than 500 characters long")
	v.Check(rl.Status == "currently reading" || rl.Status == "completed", "status", "must be 'currently reading' or 'completed'")
}

func (m ReadingListModel) Insert(rl *ReadingList) error {
	query := `
        INSERT INTO reading_lists (name, description, created_by, books, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at`
	args := []any{rl.Name, rl.Description, rl.CreatedBy, pq.Array(rl.Books), rl.Status}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&rl.ID, &rl.CreatedAt)
}

func (m ReadingListModel) Get(id int64) (*ReadingList, error) {
	query := `
        SELECT id, name, description, created_by, books, status, created_at
        FROM reading_lists
        WHERE id = $1`

	var rl ReadingList
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&rl.ID, &rl.Name, &rl.Description, &rl.CreatedBy,
		pq.Array(&rl.Books), &rl.Status, &rl.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &rl, nil
}

// GetAll retrieves all reading lists from the database.
func (m ReadingListModel) GetAll() ([]*ReadingList, error) {
	query := `
        SELECT id, name, description, created_by, books, status, created_at
        FROM reading_lists
        ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*ReadingList
	for rows.Next() {
		var list ReadingList
		var books []int64
		err := rows.Scan(
			&list.ID,
			&list.Name,
			&list.Description,
			&list.CreatedBy,
			pq.Array(&books), // Handles the array of book IDs
			&list.Status,
			&list.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list.Books = books
		lists = append(lists, &list)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lists, nil
}

func (m ReadingListModel) Update(rl *ReadingList) error {
	query := `
        UPDATE reading_lists
        SET name = $1, description = $2, books = $3, status = $4
        WHERE id = $5`

	args := []any{rl.Name, rl.Description, pq.Array(rl.Books), rl.Status, rl.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m ReadingListModel) Delete(id int64) error {
	query := `DELETE FROM reading_lists WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ReadingListModel) AddBook(listID int64, bookID int64) error {
	query := `
        UPDATE reading_lists
        SET books = array_append(books, $1)
        WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, bookID, listID)
	return err
}

func (m ReadingListModel) RemoveBook(listID int64, bookID int64) error {
	query := `
        UPDATE reading_lists
        SET books = array_remove(books, $1)
        WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, bookID, listID)
	return err
}
func (m ReadingListModel) GetByUserID(userID int64) ([]*ReadingList, error) {
	query := `
        SELECT id, name, description, created_by, books, status, created_at
        FROM reading_lists
        WHERE created_by = $1
    `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*ReadingList
	for rows.Next() {
		var list ReadingList
		err := rows.Scan(
			&list.ID,
			&list.Name,
			&list.Description,
			&list.CreatedBy,
			pq.Array(&list.Books),
			&list.Status,
			&list.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		lists = append(lists, &list)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lists, nil
}
