package uweb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// Send status and body
func (res *Response) End(req *Request) error {
	// if error, ignore others
	if res.Err != nil {
		if DEBUG {
			http.Error(res, res.Err.Error(), res.Status)
		} else {
			log.Println(LOG_TAG, "Response: End err", res.Err)
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

// empty
func (res *Response) Empty() {
	res.Status = 204
}

// Plain text
func (res *Response) Plain(status int, data string) {
	// w
	w := res

	// body
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Body = []byte(data)

	// status
	w.Status = status
	if w.Status == 0 {
		w.Status = 200
	}
}

// about jsonp see:
// http://www.cnblogs.com/dowinning/archive/2012/04/19/json-jsonp-jquery.html
func (res *Response) Jsonp(status int, padding string, v interface{}) error {
	// w
	w := res

	// body
	result, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if len(padding) > 0 {
		result = []byte(fmt.Sprintf("%s(%s);", padding, string(result)))
	}
	w.Body = result

	// header
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// status
	w.Status = status
	if w.Status == 0 {
		w.Status = 200
	}

	// ok
	return nil
}

// json
func (res *Response) Json(status int, v interface{}) error {
	return res.Jsonp(status, "", v)
}

// Html
func (res *Response) Html(status int, body []byte) {
	// w
	w := res

	// set body header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//  body
	w.Body = body

	// status
	w.Status = status
	if w.Status == 0 {
		w.Status = 200
	}
}
