package service_response

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	s := NewResponseTimeStore()
	assert.Equal(t, map[string]ResponseResult{}, s.store)

	tc := []struct {
		name   string
		result ResponseResult
		store  map[string]ResponseResult
		min    string
		max    string
	}{
		{
			"add_one",
			ResponseResult{Site: "google.ru", Duration: time.Second},
			map[string]ResponseResult{"google.ru": ResponseResult{Site: "google.ru", Duration: time.Second}},
			"google.ru", "google.ru",
		},
		{
			"add_another",
			ResponseResult{Site: "yandex.ru", Duration: 2 * time.Second},
			map[string]ResponseResult{
				"google.ru": ResponseResult{Site: "google.ru", Duration: time.Second},
				"yandex.ru": ResponseResult{Site: "yandex.ru", Duration: 2 * time.Second},
			},
			"google.ru", "yandex.ru",
		},
		{
			"add_error",
			ResponseResult{Site: "yandex.ru", Duration: 2 * time.Second, Err: errors.New("some")},
			map[string]ResponseResult{
				"google.ru": ResponseResult{Site: "google.ru", Duration: time.Second},
			},
			"google.ru", "google.ru",
		},
		{
			"add_error_twice",
			ResponseResult{Site: "yandex.ru", Duration: 2 * time.Second, Err: errors.New("some")},
			map[string]ResponseResult{
				"google.ru": ResponseResult{Site: "google.ru", Duration: time.Second},
			},
			"google.ru", "google.ru",
		},
		{
			"add_error_rest",
			ResponseResult{Site: "google.ru", Err: errors.New("1")},
			map[string]ResponseResult{},
			"", "",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			s.Process(tt.result)
			assert.Equal(t, tt.store, s.store)

			min, err := s.Min()
			if tt.min != "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.min, min.Site)
			} else {
				assert.Equal(t, ErrResponseTimeStoreNoAvailableSite, err)
			}

			max, err := s.Max()
			if tt.max != "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.max, max.Site)
			} else {
				assert.Equal(t, ErrResponseTimeStoreNoAvailableSite, err)
			}
		})
	}
}
