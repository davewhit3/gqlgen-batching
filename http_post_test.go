package batching_test

import (
	"github.com/davewhit3/gqlgen-batching"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/99designs/gqlgen/graphql/handler/testserver"
)

func TestPOST(t *testing.T) {
	h := testserver.New()
	h.AddTransport(batching.POST{})

	t.Run("success", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `{"query":"{ name }"}`, "application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, `[{"data":{"name":"test"}}]`, resp.Body.String())
	})

	t.Run("success raw graphQL query", func(t *testing.T) {
		resp := doRequest(
			h,
			"POST",
			"/graphql",
			`{"query":"query { name }"}`,
			"application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, `[{"data":{"name":"test"}}]`, resp.Body.String())
	})

	t.Run("success raw graphQL multiple queries", func(t *testing.T) {
		resp := doRequest(
			h,
			"POST",
			"/graphql",
			`{"query":"query A{ name } query B{ name }"}`,
			"application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, `[{"data":{"name":"test"}},{"data":{"name":"test"}}]`, resp.Body.String())
	})
	t.Run("success multiple queries", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query":"{ name }"}, {"query":"{ name }"}]`, "application/json")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, `[{"data":{"name":"test"}},{"data":{"name":"test"}}]`, resp.Body.String())
	})

	t.Run("decode failure", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", "notjson", "application/json")
		assert.Equal(t, http.StatusBadRequest, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `{"errors":[{"message":"json request body could not be decoded: invalid character 'o' in literal null (expecting 'u') body:notjson"}],"data":null}`, resp.Body.String())
	})

	t.Run("parse failure", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `{"query": "!"}`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("parse failure multiple queries", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query":"{ name }"},{"query": "!"}]`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"data":{"name":"test"}},{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("parse failure unnamed multiple queries", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `{"query":"query { name } query { name }"}`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"This anonymous operation must be the only defined operation.","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}},{"message":"This anonymous operation must be the only defined operation.","locations":[{"line":1,"column":16}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}},{"message":"There can be only one operation named \"\".","locations":[{"line":1,"column":16}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null},{"errors":[{"message":"This anonymous operation must be the only defined operation.","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}},{"message":"This anonymous operation must be the only defined operation.","locations":[{"line":1,"column":16}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}},{"message":"There can be only one operation named \"\".","locations":[{"line":1,"column":16}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("validation failure", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query": "{ title }"}]`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"Cannot query field \"title\" on type \"Query\".","locations":[{"line":1,"column":3}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("validation failure multiple queries", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query": "{ title }"},{"query": "{ name }"}]`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"Cannot query field \"title\" on type \"Query\".","locations":[{"line":1,"column":3}],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null},{"data":{"name":"test"}}]`, resp.Body.String())
	})

	t.Run("invalid variable", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `{"query": "query($id:Int!){find(id:$id)}","variables":{"id":false}}`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("invalid variable multiple queries", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query": "query($id:Int!){find(id:$id)}","variables":{"id":false}}, {"query": "query($id2:Int!){find(id:$id2)}","variables":{"id2":false}}]`, "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null},{"errors":[{"message":"cannot use bool as Int","path":["variable","id2"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}]`, resp.Body.String())
	})

	t.Run("execution failure", func(t *testing.T) {
		resp := doRequest(h, "POST", "/graphql", `[{"query": "mutation { name }"},{"query": "mutation { name }"}]`, "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, resp.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, `[{"errors":[{"message":"mutations are not supported"}],"data":null},{"errors":[{"message":"mutations are not supported"}],"data":null}]`, resp.Body.String())
	})
}

func doRequest(handler http.Handler, method string, target string, body string, contentType string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	r.Header.Set(batching.HeaderBatch, "true")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	return w
}
