package firefox

import (
	"github.com/adrestea/kooky"
	"github.com/adrestea/kooky/internal/firefox"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &firefox.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `firefox`

	defer s.Close()

	return s.ReadCookies(filters...)
}
