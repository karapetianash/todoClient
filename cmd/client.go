package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrConnection      = errors.New("connection error")
	ErrNotFound        = errors.New("not found")
	ErrInvalidResponse = errors.New("invalid server response")
	ErrInvalid         = errors.New("invalid data")
	ErrNotNumber       = errors.New("not a number")
)

// item represents a to-do item
type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

// response represents the API response
type response struct {
	Results      []item `json:"results"`
	Data         int    `json:"data"`
	TotalResults int    `json:"total_results"`
}

func newClient() *http.Client {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	return c
}

// getItems retrieves to-do items from the REST API using the client
func getItems(url string) ([]item, error) {
	r, err := newClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnection, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read body: %w", err)
		}

		err = ErrInvalidResponse
		if r.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}

		return nil, fmt.Errorf("%w: %s", err, msg)
	}

	var resp response

	if err = json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.TotalResults == 0 {
		return nil, fmt.Errorf("%w: No results found", ErrNotFound)
	}

	return resp.Results, nil
}

// getAll wraps getItems
func getAll(apiRoot string) ([]item, error) {
	u := fmt.Sprintf("%s/todo", apiRoot)

	return getItems(u)
}
