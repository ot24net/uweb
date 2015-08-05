package uweb

import (
	"net/http"
)

// TODO: add api to configure
var COOKIE_MAX_AGE = 7 * 24 * 3600

//
// Http response
//
type Response struct {
	http.ResponseWriter

	Status int
	Err    error
	Body   []byte

	Close func()
}

// Create response with response
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{w, 0, nil, nil, nil}
}

// Set response cookie
func (res *Response) SetCookie(name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: false, // make js can read
		MaxAge:   COOKIE_MAX_AGE,
	}
	http.SetCookie(res, cookie)
}

// Send status and body
func (res *Response) End(req *Request) error {
	// fix status
	if res.Status == 0 {
		switch req.Method {
		case "GET":
			res.Status = http.StatusOK
		case "POST", "PUT":
			res.Status = http.StatusCreated
		case "DELETE":
			res.Status = http.StatusNoContent
		default:
			res.Status = http.StatusOK
		}
		if res.Err != nil {
			res.Status = http.StatusInternalServerError
		}
	}

	// content-type
	if ct := res.Header().Get("Content-Type"); len(ct) == 0 {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	// err
	if res.Err != nil {
		http.Error(res, res.Err.Error(), res.Status)
		return nil
	}

	// no content
	if len(res.Body) == 0 {
		res.Status = 204
		res.Header().Del("Content-Type")
		res.Header().Del("Content-Length")
		res.Header().Del("Content-Encoding")
	}

	// send now
	res.WriteHeader(res.Status)
	if _, err := res.Write(res.Body); err != nil {
		return err
	}

	// close
	if res.Close != nil {
		res.Close()
	}

	// ok
	return nil
}
