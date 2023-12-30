package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatchPanic(t *testing.T) {

	panicSource := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			var ptr *int
			*ptr = 5
		}()

	})
	rec := httptest.NewRecorder()

	defused := false
	defer func() {
		if err := recover(); err != nil {
			res := rec.Result()
			if assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "panic didn't result in HTTP status code 500") {
				defused = true
			}
		}
	}()

	catchPanic(panicSource).ServeHTTP(rec, httptest.NewRequest("", "http://bla.bla/bla", nil))
	assert.True(t, defused, "test panic did not land in panic handler")
}
