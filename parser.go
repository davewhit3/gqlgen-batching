package batching

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func IsBatchingRawQuery(query string) bool {
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return false
	}

	return len(doc.Operations) > 1
}

func SplitQuery(query string) []*graphql.RawParams {
	if IsBatchingRawQuery(query) {
		doc, err := parser.ParseQuery(&ast.Source{Input: query})
		if err != nil {
			return nil
		}

		params := make([]*graphql.RawParams, 0, len(doc.Operations))
		for _, op := range doc.Operations {
			params = append(params, &graphql.RawParams{
				Query:         query,
				OperationName: op.Name,
			})
		}

		return params
	}

	return nil
}
