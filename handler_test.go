package batching_test

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	batching "github.com/davewhit3/gqlgen-batching"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphqlRawParamsCollection(t *testing.T) {
	t.Run("collection size", func(t *testing.T) {
		assert.Equal(t, 0, batching.GraphqlRawParamsCollection(nil).Len())
		assert.Equal(t, 0, batching.GraphqlRawParamsCollection([]*graphql.RawParams{}).Len())
		assert.Equal(
			t,
			3,
			batching.GraphqlRawParamsCollection([]*graphql.RawParams{{}, {}, {}}).Len())
	})

	t.Run("collection unmarshal", func(t *testing.T) {
		collection := batching.GraphqlRawParamsCollection{}
		assert.Error(t, json.Unmarshal([]byte(``), &collection))
		assert.Equal(t, batching.GraphqlRawParamsCollection{}, collection)

		collection = nil
		assert.Error(t, json.Unmarshal([]byte(`notjson`), &collection))
		assert.Equal(t, (batching.GraphqlRawParamsCollection)(nil), collection)
		
		collection = nil
		assert.NoError(t, json.Unmarshal([]byte(`{"query":"{ name }"}`), &collection))
		assert.Equal(t, batching.GraphqlRawParamsCollection{&graphql.RawParams{
			Query: `{ name }`,
		}}, collection)

		collection = nil
		assert.NoError(t, json.Unmarshal([]byte(`[{"query":"{ name }"}]`), &collection))
		assert.Equal(t, batching.GraphqlRawParamsCollection{&graphql.RawParams{
			Query: `{ name }`,
		}}, collection)

		collection = nil
		assert.NoError(t, json.Unmarshal([]byte(`[{"query":"{ name }"}, {"query":"{ name2 }"}]`), &collection))
		assert.Equal(t, batching.GraphqlRawParamsCollection{{
			Query: `{ name }`,
		}, {
			Query: `{ name2 }`,
		}}, collection)

		collection = nil
		assert.NoError(t, json.Unmarshal([]byte(`[{"query":"{ name }"}, {"query":"{ name2 }"}]`), &collection))
		assert.Equal(t, batching.GraphqlRawParamsCollection{{
			Query: `{ name }`,
		}, {
			Query: `{ name2 }`,
		}}, collection)
	})
}
