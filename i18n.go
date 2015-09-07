package uweb

import (
	"github.com/robfig/config"
	"log"
	"os"
	"path/filepath"
)

var (
	LOCALE_KEY = "_locale"
)

//
// Create i18n middleware
//
// detect - if true, will detect locale from query, cookie, session
// locale - fallback locale
// root   - locale files directory
//
func MdI18n(detect bool, locale, root string) Middleware {
	// locale
	if len(locale) == 0 {
		panic("I18n: locale is empty")
	}

	// cfgs
	cfgs := make(map[string]*config.Config)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		cfg, err := config.ReadDefault(path)
		if err != nil {
			return err
		}
		if DEBUG {
			log.Printf("I18n: Walk path-%s, name-%s\n", path, info.Name())
		}
		cfgs[info.Name()] = cfg
		return nil
	})
	if len(cfgs) == 0 {
		panic("I18n: no locale files, at least one fallback file is needed")
	}

	// i18n
	return &I18n{
		detect: detect,
		locale: locale,
		cfgs:   cfgs,
	}
}

//
// I18n
//
type I18n struct {
	detect bool                      // detect locale from query, cookie, session
	locale string                    // fallback locale
	cfgs   map[string]*config.Config // all locales, need reboot if changed
}

// @impl Middleware
func (i *I18n) Handle(c *Context) int {
	code := ""

	// detect in order
	if i.detect {
		// 1. from query
		if q := c.Req.FormValue(LOCALE_KEY); len(q) > 0 {
			if DEBUG {
				log.Println("I18n: found locale in form")
			}
			code = q
		} else {
			// 2. from cookie
			if k, err := c.Req.Cookie(LOCALE_KEY); err == nil && k != nil && len(k.Value) > 0 {
				if DEBUG {
					log.Println("I18n: found locale in cookie")
				}
				code = k.Value
			} else {
				// 3. from session
				if v := c.Sess.Get(LOCALE_KEY); len(v) > 0 {
					if DEBUG {
						log.Println("I18n: found locale in session")
					}
					code = v
				}
			}
		}
	}

	// fallback
	if len(code) == 0 {
		if DEBUG {
			log.Println("I18n: use fallback locale")
		}
		code = i.locale
	}

	// c
	c.Locale = &Locale{code: code, i18n: i}
	return NEXT_CONTINUE
}

//
// Locale
//
type Locale struct {
	code string
	i18n *I18n
}

// Get locale code
func (l *Locale) Code() string {
	return l.code
}

// Get string value
func (l *Locale) Str(section, key string) string {
	// data
	data, ok := l.i18n.cfgs[l.i18n.locale]
	if !ok {
		if DEBUG {
			log.Println("I18n: not found value in locale files, check section and key")
		}
		return ""
	}

	// value
	value, err := data.String(section, key)
	if err != nil {
		panic(err)
	}
	return value
}
