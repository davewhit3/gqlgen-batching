package batching

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const HeaderBatch = "X-Batch"

// POST implements the POST side of the default HTTP transport
// defined in https://github.com/APIs-guru/graphql-over-http#post
type POST struct {
	transport.POST
	// Map of all headers that are added to graphql response. If not
	// set, only one header: Content-Type: application/json will be set.
	ResponseHeaders map[string][]string
}

var _ graphql.Transport = POST{}

func getRequestBody(r *http.Request) (string, error) {
	if r == nil || r.Body == nil {
		return "", nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("unable to get Request Body %w", err)
	}
	return string(body), nil
}

func (h POST) Do(w http.ResponseWriter, r *http.Request, exec graphql.GraphExecutor) {
	var paramsCollection GraphqlRawParamsCollection

	start := graphql.Now()
	ctx := r.Context()
	writeHeaders(w, h.ResponseHeaders)

	if batching := r.Header.Get(HeaderBatch); batching == "true" {
		bodyString, err := getRequestBody(r)
		if err != nil {
			gqlErr := gqlerror.Errorf("could not get json request body: %+v", err)
			resp := exec.DispatchError(graphql.WithOperationContext(ctx, &graphql.OperationContext{}), gqlerror.List{gqlErr})
			writeJson(w, resp)
			return
		}

		bodyReader := io.NopCloser(strings.NewReader(bodyString))
		if err = jsonDecode(bodyReader, &paramsCollection); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			gqlErr := gqlerror.Errorf(
				"json request body could not be decoded: %+v body:%s",
				err,
				bodyString,
			)

			resp := exec.DispatchError(graphql.WithOperationContext(ctx, &graphql.OperationContext{}), gqlerror.List{gqlErr})
			writeJson(w, resp)
			return
		}

		// try to parse raw query (not json)
		if len(paramsCollection) == 1 {
			rootParam := paramsCollection[0]
			q := rootParam.Query

			if IsBatchingRawQuery(q) {
				collection := SplitQuery(q)

				for _, op := range collection {
					op.Variables = rootParam.Variables
					op.Headers = rootParam.Headers
					op.Extensions = rootParam.Extensions
				}

				paramsCollection = collection
			}
		}

		wg := sync.WaitGroup{}
		responses := make([]*graphql.Response, len(paramsCollection))

		for idx, op := range paramsCollection {
			op.Headers = r.Header
			op.ReadTime = graphql.TraceTiming{
				Start: start,
				End:   graphql.Now(),
			}

			wg.Add(1)

			go func(idx int, params *graphql.RawParams) {
				defer wg.Done()

				rc, OpErr := exec.CreateOperationContext(ctx, params)
				if OpErr != nil {
					w.WriteHeader(statusFor(OpErr))
					responses[idx] = exec.DispatchError(graphql.WithOperationContext(ctx, rc), OpErr)
					return
				}

				responseHandler, ctx := exec.DispatchOperation(ctx, rc)
				responses[idx] = responseHandler(ctx)
			}(idx, op)
		}

		wg.Wait()

		writeJson(w, responses)
		return
	}

	h.POST.ResponseHeaders = h.ResponseHeaders
	h.POST.Do(w, r, exec)
}
