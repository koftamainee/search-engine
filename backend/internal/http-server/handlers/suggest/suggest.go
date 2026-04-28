package suggest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	"github.com/koftamainee/search-engine/backend/internal/service/suggest"
)

func New(service *suggest.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := r.URL.Query()

		req := domain.SuggestRequest{
			Query: query.Get("q"),
			Limit: 10,
		}

		if limitStr := query.Get("num"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err == nil {
				req.Limit = limit
			}
		}

		res, err := service.Suggest(ctx, req)
		if err != nil {
			switch {

			case errors.Is(err, suggest.ErrInvalidLimit):
				response.BadRequest(w, "invalid limit")

			case errors.Is(err, suggest.ErrSuggestFailed):
				response.Internal(w)

			default:
				response.Internal(w)
			}
			return
		}

		response.OK(w, res)
	}
}
