//+build darwin,!arm,!arm64

// TODO: fix build tag when/if ios tag is implemented
// https://github.com/golang/go/issues/38485

package safari

import (
	"os"
	"path/filepath"

	"github.com/adrestea/kooky"
	"github.com/adrestea/kooky/internal"
)

type safariFinder struct{}

var _ kooky.CookieStoreFinder = (*safariFinder)(nil)

func init() {
	kooky.RegisterFinder(`safari`, &safariFinder{})
}

func (s *safariFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&safariCookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `safari`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(home, `Library`, `Cookies`, `Cookies.binarycookies`),
			},
		},
	}
	return ret, nil
}

/*
TODO: windows
v5.1.7 last windows version
https://www.heise.de/download/product/safari-44740
*/
