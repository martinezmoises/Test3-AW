package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/martinezmoises/Test3/internal/data"
	"github.com/martinezmoises/Test3/internal/validator"
)

func (a *applicationDependencies) listReviewsHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviews, err := a.reviewModel.GetAll(bookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reviews": reviews}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		UserID int64   `json:"user_id"`
		Rating float64 `json:"rating"`
		Review string  `json:"review"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	review := &data.Review{
		BookID: bookID,
		UserID: incomingData.UserID,
		Rating: incomingData.Rating,
		Review: incomingData.Review,
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Insert(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/reviews/%d", review.ID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"review": review}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	review, err := a.reviewModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Rating *float64 `json:"rating"`
		Review *string  `json:"review"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Rating != nil {
		review.Rating = *incomingData.Rating
	}
	if incomingData.Review != nil {
		review.Review = *incomingData.Review
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"review": review}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Read the review ID from the URL
	reviewID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Extract the user ID from the request body
	var incomingData struct {
		UserID int64 `json:"user_id"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Attempt to delete the review using the ID and user ID
	err = a.reviewModel.Delete(reviewID, incomingData.UserID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Respond with a success message
	err = a.writeJSON(w, http.StatusOK, envelope{"message": "review successfully deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
