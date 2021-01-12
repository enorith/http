package middlware

import (
	"fmt"
	"github.com/enorith/cache"
	gocache "github.com/enorith/cache"
	"github.com/enorith/exception"
	"github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"time"
)

type ThrottleRequests struct {
	limiter *Limiter
	minutes int
	max     int
}

func (t *ThrottleRequests) Handle(r contracts.RequestContract, next http.PipeHandler) contracts.ResponseContract {
	if t.limiter.cache == nil {
		t.limiter.cache = cache.AppCache
	}

	resp := next(r)
	key := t.requestSignature(r)

	if t.limiter.TooManyAttempts(key, t.max) {
		return t.responseTooManyAttempts(r)
	}

	t.limiter.Hit(key, t.minutes)

	resp.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", t.max)).
		SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", t.max-t.limiter.Attempts(key)))

	return resp
}

func (t *ThrottleRequests) requestSignature(r contracts.RequestContract) string {
	return fmt.Sprintf("request:hit:%x", r.GetSignature())
}

func (t *ThrottleRequests) responseTooManyAttempts(r contracts.RequestContract) contracts.ResponseContract {
	e := exception.NewHttpException("too many attempts", 429, 429, nil)
	return content.ErrResponse(e, 429, nil)
}

type Limiter struct {
	cache gocache.Repository
}

func (l *Limiter) Hit(key string, minutes int) int {
	duration := time.Duration(minutes) * time.Minute
	l.cache.Add(key+":timer", minutes*60, duration)

	if l.cache.Has(key) {
		l.cache.Increment(key)
	} else {
		l.cache.Put(key, 1, duration)
	}

	return 1
}
func (l *Limiter) TooManyAttempts(key string, max int) bool {

	if l.Attempts(key) >= max {
		if l.cache.Has(key + ":timer") {
			return true
		}
		l.ResetAttempts(key)
	}

	return false
}

func (l *Limiter) Attempts(key string) int {
	var h int
	if !l.cache.Has(key) {
		return 0
	}

	l.cache.Get(key, &h)
	return h
}

func (l *Limiter) ResetAttempts(key string) bool {
	return l.cache.Remove(key)
}

func Throttle(minutes int, max int) *ThrottleRequests {
	return &ThrottleRequests{
		&Limiter{},
		minutes,
		max,
	}
}
