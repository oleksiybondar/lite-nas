package httpcookie

import (
	"net/http"
	"testing"
	"time"

	"lite-nas/shared/testutil/testcasetest"
)

func TestExpiredReturnsDeletionCookie(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	got := Expired("lite-nas-at", now)

	testCases := []testcasetest.FieldCase[http.Cookie]{
		{Name: "name", Got: func(cookie http.Cookie) any { return cookie.Name }, Want: "lite-nas-at"},
		{Name: "value", Got: func(cookie http.Cookie) any { return cookie.Value }, Want: ""},
		{Name: "path", Got: func(cookie http.Cookie) any { return cookie.Path }, Want: "/"},
		{Name: "http only", Got: func(cookie http.Cookie) any { return cookie.HttpOnly }, Want: true},
		{Name: "secure", Got: func(cookie http.Cookie) any { return cookie.Secure }, Want: true},
		{Name: "same site", Got: func(cookie http.Cookie) any { return cookie.SameSite }, Want: http.SameSiteLaxMode},
		{Name: "max age", Got: func(cookie http.Cookie) any { return cookie.MaxAge }, Want: -1},
	}

	testcasetest.RunFieldCases(t, func(*testing.T) http.Cookie { return got }, testCases)

	if !got.Expires.Before(now) {
		t.Fatalf("Expires = %v, want before %v", got.Expires, now)
	}
}
