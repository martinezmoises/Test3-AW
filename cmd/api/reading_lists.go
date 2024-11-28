package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/martinezmoises/Test3/internal/data"
	"github.com/martinezmoises/Test3/internal/validator"
)

func (a *applicationDependencies) listReadingListsHandler(w http.ResponseWriter, r *http.Request) {
	lists, err := a.readingListModel.GetAll()
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reading_lists": lists}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	list, err := a.readingListModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reading_list": list}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) createReadingListHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		CreatedBy   int64   `json:"created_by"`
		Books       []int64 `json:"books"`
		Status      string  `json:"status"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	list := &data.ReadingList{
		Name:        incomingData.Name,
		Description: incomingData.Description,
		CreatedBy:   incomingData.CreatedBy,
		Books:       incomingData.Books,
		Status:      incomingData.Status,
	}

	v := validator.New()
	data.ValidateReadingList(v, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.readingListModel.Insert(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/lists/%d", list.ID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"reading_list": list}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	list, err := a.readingListModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Books       *[]int64 `json:"books"`
		Status      *string  `json:"status"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Name != nil {
		list.Name = *incomingData.Name
	}
	if incomingData.Description != nil {
		list.Description = *incomingData.Description
	}
	if incomingData.Books != nil {
		list.Books = *incomingData.Books
	}
	if incomingData.Status != nil {
		list.Status = *incomingData.Status
	}

	v := validator.New()
	data.ValidateReadingList(v, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.readingListModel.Update(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reading_list": list}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.readingListModel.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "reading list successfully deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) addBookToListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListModel.AddBook(id, incomingData.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "book successfully added to reading list"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) removeBookFromListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListModel.RemoveBook(id, incomingData.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "book successfully removed from reading list"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
