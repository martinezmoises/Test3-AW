package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/martinezmoises/Test3/internal/data"
	"github.com/martinezmoises/Test3/internal/validator"
)

func (a *applicationDependencies) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var queryParams struct {
		Title   string
		Author  string
		Genre   string
		Filters data.Filters
	}

	query := r.URL.Query()
	queryParams.Title = a.getSingleQueryParameter(query, "title", "")
	queryParams.Author = a.getSingleQueryParameter(query, "author", "")
	queryParams.Genre = a.getSingleQueryParameter(query, "genre", "")
	queryParams.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, nil)
	queryParams.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, nil)
	queryParams.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParams.Filters.SortSafeList = []string{"id", "title", "genre", "-id", "-title", "-genre"}

	v := validator.New()
	data.ValidateFilters(v, queryParams.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := a.bookModel.GetAll(queryParams.Title, queryParams.Author, queryParams.Genre, queryParams.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"books":     books,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) displayBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	book, err := a.bookModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"book": book}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Title           string   `json:"title"`
		Authors         []string `json:"authors"`
		ISBN            string   `json:"isbn"`
		PublicationDate string   `json:"publication_date"`
		Genre           string   `json:"genre"`
		Description     string   `json:"description"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	book := &data.Book{
		Title:           incomingData.Title,
		Authors:         incomingData.Authors,
		ISBN:            incomingData.ISBN,
		PublicationDate: incomingData.PublicationDate,
		Genre:           incomingData.Genre,
		Description:     incomingData.Description,
	}

	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.bookModel.Insert(book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/books/%d", book.ID))

	data := envelope{"book": book}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	book, err := a.bookModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Title           *string   `json:"title"`
		Authors         *[]string `json:"authors"`
		ISBN            *string   `json:"isbn"`
		PublicationDate *string   `json:"publication_date"` // Ensure this matches
		Genre           *string   `json:"genre"`
		Description     *string   `json:"description"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Title != nil {
		book.Title = *incomingData.Title
	}
	if incomingData.Authors != nil {
		book.Authors = *incomingData.Authors
	}

	if incomingData.ISBN != nil {
		book.ISBN = *incomingData.ISBN
	}
	if incomingData.Genre != nil {
		book.Genre = *incomingData.Genre
	}
	if incomingData.Description != nil {
		book.Description = *incomingData.Description
	}

	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.bookModel.Update(book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"book": book}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.bookModel.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"message": "book successfully deleted"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) searchBooksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	title := a.getSingleQueryParameter(query, "title", "")
	author := a.getSingleQueryParameter(query, "author", "")
	genre := a.getSingleQueryParameter(query, "genre", "")

	books, _, err := a.bookModel.GetAll(title, author, genre, data.Filters{Page: 1, PageSize: 100}) // Default pagination
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"books": books}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
