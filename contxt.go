package contxt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"unsafe"

	"github.com/gorilla/schema"
	jsoniter "github.com/json-iterator/go"
)

const tagName = "header"

var schemaDecoder = schema.NewDecoder()

func New(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		request: &Request{Request: r},
		writer:  w,
	}
}

type Ctx struct {
	responseStatus int

	request *Request
	writer  http.ResponseWriter
}

// BodyParser reads values from the header mime type `Content-Type`
// its values is used to inform the function how to read the request
// body. It parses 3 mimes `application/json` using json unmarshalling
// into a struct, `application/x-www-form-urlencoded` using an external
// library which is github.com/gorilla/schema and the last mime being
// `multipart/form-data` using github.com/gorilla/schema
func (ctx *Ctx) BodyParser(v interface{}) error {
	ctype := ctx.request.Header.Get("Content-Type")
	// application/json
	if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		byteBody, err := ioutil.ReadAll(ctx.request.Body)
		if err != nil {
			return err
		}
		return jsoniter.Unmarshal(byteBody, v)
	}
	// application/x-www-form-urlencoded
	if strings.HasPrefix(ctype, MIMEApplicationForm) {
		err := ctx.request.ParseForm()
		if err != nil {
			return err
		}
		return schemaDecoder.Decode(v, ctx.request.PostForm)
	}
	// multipart/form-data
	if strings.HasPrefix(ctype, MIMEMultipartForm) {
		err := ctx.request.ParseMultipartForm(32 << 20)
		if err != nil {
			return err
		}
		return schemaDecoder.Decode(v, ctx.request.Form)
	}
	return fmt.Errorf("cannot parse content-type: %v", ctype)
}

// Get returns value of a key in the header object
func (ctx *Ctx) Get(key string) string {
	return ctx.request.Header.Get(key)
}

// Headers reads values from a request headers and attempts to
// write them to an interface
func (ctx *Ctx) Headers(v interface{}) error {

	// we check if the parameter is a pointer
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("value should be a pointer")
	}

	valPtr := reflect.Indirect(reflect.ValueOf(v))
	if valPtr.Kind() != reflect.Struct {
		return errors.New("type of value should be a struct")
	}

	t := valPtr.Type()
	e := reflect.ValueOf(v).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := e.Type().Field(i)
		prop := e.Field(i)

		tag := field.Tag.Get(tagName)
		prop.SetString(ctx.Get(tag))
	}

	return nil
}

// JSON marshals a struct into json and adds an application json content type
// into the header
func (ctx *Ctx) JSON(v interface{}) error {
	ctx.writer.Header().Set("Content-Type", MIMEApplicationJSON)

	if ctx.responseStatus == 0 {
		ctx.responseStatus = http.StatusOK
	}
	ctx.writer.WriteHeader(ctx.responseStatus)

	raw, err := jsoniter.Marshal(&v)
	if err != nil {
		_, _ = ctx.writer.Write([]byte(""))
		return err
	}
	_, _ = ctx.writer.Write(raw)
	return nil
}

func (ctx *Ctx) Query(key string) string {
	keys := ctx.request.URL.Query()
	return keys.Get(key)
}

func (ctx *Ctx) Redirect(path string, status ...int) {
	code := http.StatusFound
	if len(status) > 0 {
		code = status[0]
	}

	ctx.writer.Header().Set("Location", path)
	http.Redirect(ctx.writer, ctx.request.Request, path, code)
}

// Send takes an interface an writes it as a string into response body
func (ctx *Ctx) Send(args ...interface{}) {
	if ctx.responseStatus == 0 {
		ctx.responseStatus = http.StatusOK
	}
	ctx.writer.WriteHeader(ctx.responseStatus)

	if len(args) == 0 {
		return
	}

	switch body := args[0].(type) {
	case string:
		_, _ = ctx.writer.Write([]byte(body))
	case []byte:
		_, _ = ctx.writer.Write(body)
	default:
		byteBody, _ := jsoniter.Marshal(body)
		_, _ = ctx.writer.Write(byteBody)
	}
}

// Status writes the response http status code
func (ctx *Ctx) Status(code int) *Ctx {
	ctx.responseStatus = code
	return ctx
}

// XML writes an xml string into response body
func (ctx *Ctx) XML(v interface{}) {
	ctx.writer.Header().Set("Content-Type", MIMEApplicationXML)
	ctx.Send(v)
}

func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
