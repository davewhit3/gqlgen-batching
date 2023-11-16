package batching_test

import (
	batching "github.com/davewhit3/gqlgen-batching"
	"net/http"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/stretchr/testify/assert"
)

func TestHeadersWithPOST(t *testing.T) {
	t.Run("Headers not set", func(t *testing.T) {
		h := testserver.New()
		h.AddTransport(batching.POST{})

		resp := doRequest(h, "POST", "/graphql", `[{"query":"{ name }"}]`, "application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 1, len(resp.Header()))
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	})

	t.Run("Headers set", func(t *testing.T) {
		headers := map[string][]string{
			"Content-Type": {"application/json; charset: utf8"},
			"Other-Header": {"dummy-post", "another-one"},
		}

		h := testserver.New()
		h.AddTransport(batching.POST{ResponseHeaders: headers})

		resp := doRequest(h, "POST", "/graphql", `[{"query":"{ name }"},{"query":"{ name }"}]`, "application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 2, len(resp.Header()))
		assert.Equal(t, "application/json; charset: utf8", resp.Header().Get("Content-Type"))
		assert.Equal(t, "dummy-post", resp.Header().Get("Other-Header"))
		assert.Equal(t, "another-one", resp.Header().Values("Other-Header")[1])
	})
}
