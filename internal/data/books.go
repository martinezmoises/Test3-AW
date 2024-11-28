package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/martinezmoises/Test3/internal/validator"
)

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Authors         []string  `json:"authors"` // Change to slice of strings
	ISBN            string    `json:"isbn"`
	PublicationDate string    `json:"publication_date"`
	Genre           string    `json:"genre"`
	Description     string    `json:"description"`
	AverageRating   float64   `json:"average_rating"`
	CreatedAt       time.Time `json:"created_at"`
	Version         int       `json:"version"`
}

type BookModel struct {
	DB *sql.DB
}

func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Title) <= 200, "title", "must not be more than 200 bytes long")
	v.Check(len(book.Authors) > 0, "authors", "at least one author must be provided")
	v.Check(book.ISBN != "", "isbn", "must be provided")
	v.Check(len(book.ISBN) <= 13, "isbn", "must not be more than 13 bytes")
	v.Check(book.Genre != "", "genre", "must be provided")
	v.Check(len(book.Genre) <= 50, "genre", "must not be more than 50 bytes")
	v.Check(book.Description != "", "description", "must be provided")
	v.Check(len(book.Description) <= 500, "description", "must not be more than 500 bytes")
}

func (m BookModel) Insert(book *Book) error {
	query := `
        INSERT INTO books (title, authors, isbn, publication_date, genre, description, average_rating)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, version
    `

	args := []any{
		book.Title,
		pq.Array(book.Authors), // Pass authors as array
		book.ISBN,
		book.PublicationDate,
		book.Genre,
		book.Description,
		book.AverageRating,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.CreatedAt, &book.Version)
}
func (m BookModel) Get(id int64) (*Book, error) {
	query := `
        SELECT id, created_at, title, authors, isbn, publication_date, genre, description, average_rating, version
        FROM books
        WHERE id = $1
    `

	var book Book
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.CreatedAt,
		&book.Title,
		pq.Array(&book.Authors), // Scan authors as array
		&book.ISBN,
		&book.PublicationDate,
		&book.Genre,
		&book.Description,
		&book.AverageRating,
		&book.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &book, nil
}

func (m BookModel) Update(book *Book) error {
	query := `
        UPDATE books
        SET title = $1, authors = $2, isbn = $3, publication_date = $4, genre = $5, description = $6, 
            average_rating = $7, version = version + 1
        WHERE id = $8
        RETURNING version
    `

	args := []any{
		book.Title,
		pq.Array(book.Authors), // Use pq.Array to handle the slice of strings
		book.ISBN,
		book.PublicationDate,
		book.Genre,
		book.Description,
		book.AverageRating,
		book.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)
}

func (m BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM books WHERE id = $1`
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

func (m BookModel) GetAll(title string, author string, genre string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), id, created_at, title, authors, isbn, publication_date, genre, description, average_rating, version
        FROM books
        WHERE (title ILIKE '%%' || $1 || '%%' OR $1 = '')
        AND ($2 = '' OR EXISTS (
            SELECT 1
            FROM unnest(authors) AS a
            WHERE a ILIKE '%%' || $2 || '%%'
        ))
        AND (genre ILIKE '%%' || $3 || '%%' OR $3 = '')
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, author, genre, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	books := []*Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(
			&totalRecords,
			&book.ID,
			&book.CreatedAt,
			&book.Title,
			pq.Array(&book.Authors), // Use pq.Array to handle TEXT[]
			&book.ISBN,
			&book.PublicationDate,
			&book.Genre,
			&book.Description,
			&book.AverageRating,
			&book.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return books, metadata, nil
}
