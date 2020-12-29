package firefox

import (
	"errors"
	"fmt"
	"time"

	"github.com/adrestea/kooky"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/go-sqlite/sqlite3"
)

func (s *CookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.Database == nil {
		return nil, errors.New(`database is nil`)
	}

	var cookies []*kooky.Cookie

	var baseDomainRemoved bool = true
	var columnIDs = map[string]int{
		// fallback values
		`baseDomain`:   1, // old
		`name`:         2,
		`value`:        3,
		`host`:         4,
		`path`:         5,
		`expiry`:       6,
		`creationTime`: 8,
		`isSecure`:     9,
		`isHttpOnly`:   10,
	}
	cookiesTableName := `moz_cookies`
	var highestIndex int
	for _, table := range s.Database.Tables() {
		if table.Name() == cookiesTableName {
			for id, column := range table.Columns() {
				name := column.Name()
				if name == `CONSTRAINT` {
					// github.com/go-sqlite/sqlite3.Table.Columns() reports pseudo-columns for host, path, originAttributes
					break
				}
				if name == `baseDomain` {
					baseDomainRemoved = false
				}
				if id > highestIndex {
					highestIndex = id
				}
				columnIDs[name] = id
			}
		}
	}

	err := s.Database.VisitTableRecords(cookiesTableName, func(rowId *int64, rec sqlite3.Record) error {
		/*
		   known column counts for firefox from past/current versions
		   ???:     13 columns
		   v78 LTS: 14 columns
		   v82:     15 columns
		*/

		if highestIndex >= len(rec.Values) {
			return errors.New(`column index out of bound`)
		}

		cookie := kooky.Cookie{}
		var ok bool

		/*
			-- Firefox 78 ESR - copied from sqlitebrowser
			CREATE TABLE moz_cookies(
				id INTEGER PRIMARY KEY,
				originAttributes TEXT NOT NULL DEFAULT '',
				name TEXT,
				value TEXT,
				host TEXT,
				path TEXT,
				expiry INTEGER,
				lastAccessed INTEGER,
				creationTime INTEGER,
				isSecure INTEGER,
				isHttpOnly INTEGER,
				inBrowserElement INTEGER DEFAULT 0,
				sameSite INTEGER DEFAULT 0,
				rawSameSite INTEGER DEFAULT 0,
				CONSTRAINT moz_uniqueid UNIQUE (name, host, path, originAttributes)
			)
		*/

		// Name
		cookie.Name, ok = rec.Values[columnIDs[`name`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Name %v (type %[1]T)", rec.Values[columnIDs[`name`]])
		}

		// Value
		cookie.Value, ok = rec.Values[columnIDs[`value`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Value %v (type %[1]T)", rec.Values[columnIDs[`value`]])
		}

		// Domain
		if baseDomainRemoved {
			if host, ok := rec.Values[columnIDs[`host`]].(string); ok {
				cookie.Domain = domainutil.Domain(host)
			} else {
				return fmt.Errorf("got unexpected value for Host %v (type %[1]T)", rec.Values[columnIDs[`host`]])
			}
		} else {
			// handle databases prior v78 ESR
			cookie.Domain, ok = rec.Values[columnIDs[`baseDomain`]].(string)
			if !ok {
				return fmt.Errorf("got unexpected value for Domain %v (type %[1]T)", rec.Values[columnIDs[`baseDomain`]])
			}
		}

		// Path
		cookie.Path, ok = rec.Values[columnIDs[`path`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Path %v (type %[1]T)", rec.Values[columnIDs[`path`]])
		}

		// Expires
		{
			expiry := rec.Values[columnIDs[`expiry`]]
			if int32Value, ok := expiry.(int32); ok {
				cookie.Expires = time.Unix(int64(int32Value), 0)
			} else if uint64Value, ok := expiry.(uint64); ok {
				cookie.Expires = time.Unix(int64(uint64Value), 0)
			} else {
				return fmt.Errorf("got unexpected value for Expires %v (type %[1]T)", expiry)
			}
		}

		// Creation
		int64Value, ok := rec.Values[columnIDs[`creationTime`]].(int64)
		if !ok {
			return fmt.Errorf("got unexpected value for Creation %v (type %[1]T)", rec.Values[columnIDs[`creationTime`]])
		}
		cookie.Creation = time.Unix(int64Value/1e6, 0) // drop nanoseconds

		// Secure
		intValue, ok := rec.Values[columnIDs[`isSecure`]].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for Secure %v (type %[1]T)", rec.Values[columnIDs[`isSecure`]])
		}
		cookie.Secure = intValue > 0

		// HttpOnly
		intValue, ok = rec.Values[columnIDs[`isHttpOnly`]].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for HttpOnly %v (type %[1]T)", rec.Values[columnIDs[`isHttpOnly`]])
		}
		cookie.HttpOnly = intValue > 0

		if !kooky.FilterCookie(&cookie, filters...) {
			return nil
		}

		cookies = append(cookies, &cookie)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cookies, nil
}
