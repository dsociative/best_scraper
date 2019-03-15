package service_response

import (
	"errors"
	"math/rand"
	"sync"
)

var ErrResponseTimeStoreNoAvailableSite = errors.New("NO_AVAILABLE_SITE")

type ResponseTimeStore struct {
	store map[string]ResponseResult
	lock  *sync.RWMutex
}

func NewResponseTimeStore() *ResponseTimeStore {
	return &ResponseTimeStore{store: map[string]ResponseResult{}, lock: &sync.RWMutex{}}
}

func (r *ResponseTimeStore) Process(result ResponseResult) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if result.Err == nil {
		r.store[result.Site] = result
	} else {
		delete(r.store, result.Site)
	}
}

func (r *ResponseTimeStore) Listen(resultChan chan ResponseResult) {
	for result := range resultChan {
		r.Process(result)
	}
}

func (r *ResponseTimeStore) filterOne(f func(f, s ResponseResult) bool) (ResponseResult, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var result ResponseResult
	for _, rr := range r.store {
		if result.Site == "" || f(result, rr) {
			result = rr
		}
	}
	if result.Site == "" {
		return result, ErrResponseTimeStoreNoAvailableSite
	}
	return result, nil
}

func (r *ResponseTimeStore) Min() (ResponseResult, error) {
	return r.filterOne(func(f, s ResponseResult) bool {
		return f.Duration > s.Duration
	})
}

func (r *ResponseTimeStore) Random() (_ ResponseResult, err error) {
	n := -1
	rFunc := func() int {
		if n < 0 {
			n = rand.Intn(len(r.store))
		}
		return n
	}
	i := -1
	return r.filterOne(func(f, s ResponseResult) bool {
		i++
		return i >= rFunc()
	})
}

func (r *ResponseTimeStore) Max() (ResponseResult, error) {
	return r.filterOne(func(f, s ResponseResult) bool {
		return f.Duration < s.Duration
	})
}
