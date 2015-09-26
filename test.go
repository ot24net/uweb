package uweb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
)

//
// We may run several tests, but only setup once
//
var (
	testOnce sync.Once
	testApp  *Application
)

func setupTest() {
	testOnce.Do(func() {
		testApp = NewApp()
		testApp.Use(MdRouter())
	})
}

//
// Test Get handler
//
func TestGet(path, query string) (*httptest.ResponseRecorder, error) {
	// init
	setupTest()

	// req
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s?%s", path, query), nil)
	if err != nil {
		return nil, err
	}

	// w
	w := httptest.NewRecorder()

	// handle
	testApp.ServeHTTP(w, req)

	// ok
	return w, nil
}

//
// Test Post Handler
//
func TestPost(path string, data url.Values) (*httptest.ResponseRecorder, error) {
	// init
	setupTest()

	// req
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s", path), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// w
	w := httptest.NewRecorder()

	// handle
	testApp.ServeHTTP(w, req)

	// ok
	return w, nil
}

//
// Test Put Handler
//
func TestPut(path string, data url.Values) (*httptest.ResponseRecorder, error) {
	// init
	setupTest()

	// req
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost%s", path), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// w
	w := httptest.NewRecorder()

	// handle
	testApp.ServeHTTP(w, req)

	// ok
	return w, nil
}

//
// Test Del Handler
//
func TestDel(path string) (*httptest.ResponseRecorder, error) {
	// init
	setupTest()

	// req
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost%s", path), nil)
	if err != nil {
		return nil, err
	}

	// w
	w := httptest.NewRecorder()

	// handle
	testApp.ServeHTTP(w, req)

	// ok
	return w, nil
}
