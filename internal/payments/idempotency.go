package payments

import "sync"

type idempotencyStore struct {
	mu      sync.Mutex
	results map[string]paymentResult
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{results: make(map[string]paymentResult)}
}

func (s *idempotencyStore) Do(
	key string,
	create func() (paymentResult, error),
) (paymentResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if result, ok := s.results[key]; ok {
		return result, false, nil
	}

	result, err := create()
	if err != nil {
		return paymentResult{}, false, err
	}
	s.results[key] = result
	return result, true, nil
}