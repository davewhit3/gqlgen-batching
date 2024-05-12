package batching

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"reflect"
	"bytes"
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
	r := bytes.NewReader(b)

	d := json.NewDecoder(r)
	d.UseNumber()

	if err := d.Decode((*[]*graphql.RawParams)(c)); err == nil {
		return nil
	}

	if _, err := r.Seek(0, 0); err != nil {
		return err
	}

	var n graphql.RawParams
	if err := d.Decode(&n); err != nil {
		return nil
	}

	*c = []*graphql.RawParams{&n}
	return nil
}
