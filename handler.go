package batching

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"reflect"
)

type GraphqlRawParamsCollection []*graphql.RawParams

func (c GraphqlRawParamsCollection) Len() int {
	s := reflect.ValueOf(c)
	if s.Kind() != reflect.Slice || s.IsNil() {
		return 0
	}

	return s.Len()
}

// UnmarshalJSON could unmarshal slice or single query
func (c *GraphqlRawParamsCollection) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*[]*graphql.RawParams)(c))
	if err == nil {
		return nil
	}

	var n graphql.RawParams
	err = json.Unmarshal(b, &n)
	if err == nil {
		*c = []*graphql.RawParams{&n}
		return nil
	}

	return nil
}
