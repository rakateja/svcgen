package page

import (
	"net/http"
	"strconv"
)

type Page[T interface{}] struct {
	Items       []T  `json:"items"`
	Total       int  `json:"total"`
	PerPage     int  `json:"perPage"`
	CurrentPage int  `json:"currentPage"`
	LastPage    int  `json:"lastPage"`
	PrevPage    *int `json:"prevPage"`
	NextPage    *int `json:"nextPage"`
}

func New[T interface{}](items []T, total, pageNumber, pageSize int) Page[T] {
	var prevPage *int
	if pageNumber > 1 {
		val := pageNumber - 1
		prevPage = &val
	}
	offset := GetOffset(pageNumber, pageSize)
	var nextPage *int
	if offset+pageSize < total {
		val := pageNumber + 1
		nextPage = &val
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return Page[T]{
		Items:       items,
		Total:       total,
		PerPage:     pageSize,
		CurrentPage: pageNumber,
		LastPage:    totalPages,
		PrevPage:    prevPage,
		NextPage:    nextPage,
	}
}

func EmptyItems[T interface{}](pageNumber, pageSize int) Page[T] {
	return New([]T{}, 0, pageNumber, pageSize)
}

func GetOffset(pageNumber int, pageSize int) int {
	return (pageNumber - 1) * pageSize
}

func ParamsFromHTTPRequest(req *http.Request) (int, int, error) {
	pageStr := req.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageSizeStr := req.URL.Query().Get("limit")
	if pageSizeStr == "" {
		pageSizeStr = "30"
	}
	pageNumber, err := strconv.Atoi(pageStr)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return 0, 0, err
	}
	return pageNumber, pageSize, nil
}
