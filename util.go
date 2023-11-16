package batching

import (
	"encoding/json"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

type Response interface {
	*graphql.Response | []*graphql.Response
}

func writeJson[T Response](w io.Writer, response T) {
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	// nolint:errcheck
	w.Write(b)
}
