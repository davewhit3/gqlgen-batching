# gqlgen-batching
GraphQL batch support for gqlgen

## What is gqlgen-batching?

**gqlgen-batching** is an extension of [gqlgen](https://github.com/99designs/gqlgen) to support [GraphQL Batching](https://github.com/graphql/graphql-over-http/blob/main/rfcs/Batching.md).

## Quick start

1. Add package import
    ```go
   import "github.com/davewhit3/gqlgen-batching"
    ```
2. Add batching transport 
   ```go
   srv.AddTransport(batching.POST{})
   ```
3. Prepare your server
   ```go

    schema := generated.NewExecutableSchema(graph.NewResolver())
	srv := handler.New(schema)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(batching.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	http.Handle("/", playground.Handler("Starwars", "/query"))
	http.Handle("/query", srv)

	log.Fatal(http.ListenAndServe(":8080", nil))

   ```

## How to use?

Add header to request `X-Batch: true`

```bash
curl -X POST http://localhost:8080/query \
-H "X-Batch: true" -H "Content-Type: application/json" \
-d '[{"query":"{hero(episode: JEDI) { name }}"},{"query":"{hero(episode: EMPIRE) { name }}"}]'
```

Result:
```json
[
   {
      "data": {
         "hero": {
            "name": "R2-D2"
         }
      }
   },
   {
      "data": {
         "hero": {
            "name": "Luke Skywalker"
         }
      }
   }
]
```