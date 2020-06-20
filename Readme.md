# Go HTTP Contxt
Naming is difficult I agree. This small lib is my own idea of how I do and handle http requests using the default Golang
http.HandlerFunc `func(w http.ResponseWriter, r *http.Request)`.

This lib focuses on the `*http.Request` request. Also, how to write different response types to `http.ResponseWriter`.
This comes kinda makes the code clean and focused on the business logic only.

## Usage
```go
package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/SirWaithaka/contxt"
)

// Headers describes properties of the request headers
// that we might need to inflate into the struct
type Headers struct {
	UserAgent string `header:"User-Agent"`
}

// Form describes properties of a form request fields
// the fields can come in to the application as different MIME types
// and the struct will define tags to handle all cases
// json tag for Content-type: application/json
// schema tag for Content-type: application/x-www-form-urlencoded
type Form struct {
	FirstName   string `json:"first_name" schema:"first_name"`
	LastName    string `json:"last_name" schema:"last_name"`
	YearOfBirth int    `json:"year_of_birth" schema:"year_of_birth"`
}

func HandleRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we can create a new context object using the new function
		ctx := contxt.New(w, r)

		// we can get one item out of the request headers using Get
		agent := ctx.Get("User-Agent")
		log.Println(agent)

		// if we want to initialize a struct with header values
		// we can pass a pointer to the struct to the Headers func
		var headers Headers
		err := ctx.Headers(&headers)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Headers parsed %+v", headers)

		// if you have a request with query parameters e.g.
		// localhost:9090?name=hello&age=20
		// we can retrieve the query params name and hello using
		// the Query func
		country := ctx.Query("country")
		age := ctx.Query("age")

		// when we are expecting POST requests and forms are being submitted
		// to the application, we can use the BodyParser func to inflate a struct
		// with fields we are expecting. We can use the same struct to handle form fields
		// coming in different MIME types
		var form Form
		err = ctx.BodyParser(&form)
		if err != nil {
			log.Println(err)
		}

		// we can respond to requests using different formats
		// we can send just string back, we can send a json or xml
		// we also have a hook func that adds a status code to the
		// response before we send back the content
		ageInt, err := strconv.Atoi(age)
		if err != nil {
			ctx.Status(http.StatusBadRequest).Send("Age should be a number")
			return
		}

		person := map[string]interface{}{
			"country": country,
			"place": "world",
			"age": ageInt,
			"name": form.FirstName+ " " +form.LastName,
			"birthYear": form.YearOfBirth,
		}
		_ = ctx.Status(http.StatusOK).JSON(person)

	})
}

func main() {

	srv := http.Server{
		Addr:    ":9090",
		Handler: HandleRequest(),
	}
	log.Fatal(srv.ListenAndServe())
}
```

## cURL examples
Send a form `POST` request with html form values
```bash
curl --location --request POST 'localhost:9090?age=20&country=kenya' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'first_name=john' \
--data-urlencode 'last_name=smith' \
--data-urlencode 'year_of_birth=2000'
```

Response
```json
{
    "name": "john smith",
    "birthYear": 2000,
    "country": "kenya",
    "place": "world",
    "age": 20
}
```

Send a `POST` with `application/json` mime type
```bash
curl --location --request POST 'localhost:9090?age=20&country=kenya' \
--header 'Content-Type: application/json' \
--data-raw '{
	"first_name": "john",
	"last_name": "smith",
	"year_of_birth": 2000
}'
```

Response
```json
{
    "place": "world",
    "age": 20,
    "name": "john smith",
    "birthYear": 2000,
    "country": "kenya"
}
```
