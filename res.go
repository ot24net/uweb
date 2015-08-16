package uweb

import (
	"log"
	"net/http"
)

// Cookies configure value
var (
	COOKIE_MAX_AGE   = 365 * 24 * 3600
	COOKIE_HTTP_ONLY = false
	COOKIE_PATH      = "/"
)

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
		Path:     COOKIE_PATH,
		HttpOnly: COOKIE_HTTP_ONLY,
		MaxAge:   COOKIE_MAX_AGE,
	}
	http.SetCookie(res, cookie)
}

// Send status and body
func (res *Response) End(req *Request) error {
	// if error, ignore others
	if res.Err != nil {
		if DEBUG {
			http.Error(res, res.Err.Error(), res.Status)
		} else {
			log.Println("[uweb] ERROR", res.Err)
			http.Error(res, "some error happens", res.Status)
		}
		return nil
	}

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

	// fix content-xxx
	if len(res.Body) > 0 {
		if ct := res.Header().Get("Content-Type"); len(ct) == 0 {
			res.Header().Set("Content-Type", http.DetectContentType(res.Body))
		}
	} else {
		res.Status = 204
		res.Header().Del("Content-Type")
		res.Header().Del("Content-Length")
		res.Header().Del("Content-Encoding")
	}

	// write body
	res.WriteHeader(res.Status)
	if _, err := res.Write(res.Body); err != nil {
		return err
	}

	// release if needed
	if res.Close != nil {
		res.Close()
	}

	// ok
	return nil
}
