package go_api

import (
	"fmt"
	"net/http"
	"strconv"
)

type PaginatedResponse struct {
	Data     []interface{} `json:"data"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Previous *int          `json:"previous"`
	Next     *int          `json:"next"`
}

func NewPaginatedResponse(data []interface{}, total int, limit int, page int) PaginatedResponse {
	var previous *int
	var next *int

	if page > 1 {
		previousPage := page - 1
		previous = &previousPage
	}

	if (page * limit) < total {
		nextPage := page + 1
		next = &nextPage
	}

	return PaginatedResponse{
		Data:     data,
		Total:    total,
		Limit:    limit,
		Previous: previous,
		Next:     next,
	}
}

func getPaginationParameters(r *http.Request) (int, int, error) {
	const DEFAULT_LIMIT = 50
	const DEFAULT_PAGE = 1

	const MAX_LIMIT = 100

	limit := DEFAULT_LIMIT
	page := DEFAULT_PAGE

	// Get the page and limit from the query params
	limitFromQuery := r.URL.Query().Get("limit")
	if limitFromQuery != "" {
		// Convert the limit to an integer
		limitFromQueryAsInteger, strConvErr := strconv.Atoi(limitFromQuery)
		if strConvErr == nil && limitFromQueryAsInteger > 0 && limitFromQueryAsInteger <= MAX_LIMIT {
			limit = limitFromQueryAsInteger
		} else {
			return 0, 0, fmt.Errorf("invalid limit")
		}
	}
	pageFromQuery := r.URL.Query().Get("page")
	if pageFromQuery != "" {
		// Convert the page to an integer
		pageFromQueryAsInteger, strConvErr := strconv.Atoi(pageFromQuery)
		if strConvErr == nil && pageFromQueryAsInteger > 0 {
			page = pageFromQueryAsInteger
		} else {
			return 0, 0, fmt.Errorf("invalid page")
		}
	}

	return page, limit, nil
}
