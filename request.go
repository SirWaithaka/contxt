package contxt

import (
	"net/http"
)

// Request is a wrapper struct
// for the standard lib http.Request struct
type Request struct {
	*http.Request
}

