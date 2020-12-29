package chrome

import (
	"runtime"

	"github.com/adrestea/kooky"
	"github.com/adrestea/kooky/internal/chrome"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &chrome.CookieStore{}
	s.FileNameStr = filename

	// TODO
	switch runtime.GOOS {
	case `windows`, `darwin` , `linux`:
		s.BrowserStr = `chrome`
	default:
		s.BrowserStr = `chromium`
	}

	defer s.Close()

	return s.ReadCookies(filters...)
}
