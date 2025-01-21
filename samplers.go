package morelogger

import (
	"math/rand"

	"github.com/max-chem-eng/morelogger/models"
)

type Sampler interface {
	Allow(record *models.LogRecord) bool
}
type RandomSampler struct {
	N uint32
}

func (s *RandomSampler) Allow(record *models.LogRecord) bool {
	if s.N <= 0 {
		return false
	}
	if rand.Intn(int(s.N)) != 0 {
		return false
	}
	return true
}

type FirstOfSampler struct {
	N       uint32
	counter uint32
}

func (s *FirstOfSampler) Allow(record *models.LogRecord) bool {
	n := s.N
	if n == 1 {
		return true
	}
	c := s.counter
	res := c%n == 1
	s.counter++
	return res
}

// RateLimitSampler allows up to N logs per second
type RateLimitSampler struct {
	N uint32
	// counter uint32
}

func (s *RateLimitSampler) Allow(record *models.LogRecord) bool {
	// TODO: implement rate limiting
	return true
}
