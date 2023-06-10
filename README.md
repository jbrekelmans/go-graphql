go-graphql
==========

[![Go Reference](https://pkg.go.dev/badge/github.com/jbrekelmans/go-graphql.svg)](https://pkg.go.dev/github.com/jbrekelmans/go-graphql)

Package `graphql` provides a Client for talking to a GraphQL server using HTTP with JSON-encoded requests/responses.

Inspired by [shurcooL/graphQL](https://github.com/shurcooL/graphql) but includes the following enhancements:

1. Maps Golang primitive data types to GraphQL types.
2. Returns structured error types that reflect GraphQL-level errors, and status/headers of HTTP responses, to support error proper handling and backoff (see below).

Installation
------------

```bash
go get github.com/jbrekelmans/go-graphql
```

Usage
-----

Construct a GraphQL client, specifying the GraphQL server URL. Then, you can use it to make GraphQL queries and mutations.

```go
client := graphql.NewClient("https://example.com/graphql", nil)
// Use client...
```

### Authentication

Some GraphQL servers may require authentication. The `graphql` package does not directly handle authentication. Instead, when creating a new client, you're expected to pass an `http.Client` that performs authentication. The easiest and recommended way to do this is to use the [`golang.org/x/oauth2`](https://golang.org/x/oauth2) package. You'll need an OAuth token with the right scopes. Then:

```go
import "golang.org/x/oauth2"

func main() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := graphql.NewClient("https://example.com/graphql", httpClient)
	// Use client...
```

### Simple Query

To make a GraphQL query, you need to define a corresponding Go type.

For example, to make the following GraphQL query:

```GraphQL
query {
	me {
		name
	}
}
```

You can define this variable:

```Go
var query struct {
	Me struct {
		Name string
	}
}
```

Then call `client.Query`, passing a pointer to it:

```Go
err := client.Query(context.Background(), &query, nil)
if err != nil {
	// Handle error.
}
fmt.Println(query.Me.Name)

// Output: Luke Skywalker
```

### Arguments and Variables

Often, you'll want to specify arguments on some fields. You can use the `graphql` struct field tag for this.

For example, to make the following GraphQL query:

```GraphQL
{
	human(id: "1000") {
		name
		height(unit: METER)
	}
}
```

You can define this variable:

```Go
var q struct {
	Human struct {
		Name   string
		Height float64 `graphql:"height(unit: METER)"`
	} `graphql:"human(id: \"1000\")"`
}
```

Then call `client.Query`:

```Go
err := client.Query(context.Background(), &q, nil)
if err != nil {
	// Handle error.
}
fmt.Println(q.Human.Name)
fmt.Println(q.Human.Height)

// Output:
// Luke Skywalker
// 1.72
```

However, that'll only work if the arguments are constant and known in advance. Otherwise, you will need to make use of variables. Replace the constants in the struct field tag with variable names:

```Go
var q struct {
	Human struct {
		Name   string
		Height float64 `graphql:"height(unit: $unit)"`
	} `graphql:"human(id: $id)"`
}
```

Then, define a `variables` map with their values:

```Go
variables := map[string]interface{}{
	"id":   graphql.ID(id),
	"unit": starwars.LengthUnit("METER"),
}
```

Finally, call `client.Query` providing `variables`:

```Go
err := client.Query(context.Background(), &q, variables)
if err != nil {
	// Handle error.
}
```

### Inline Fragments

Some GraphQL queries contain inline fragments. You can use the `graphql` struct field tag to express them.

For example, to make the following GraphQL query:

```GraphQL
{
	hero(episode: "JEDI") {
		name
		... on Droid {
			primaryFunction
		}
		... on Human {
			height
		}
	}
}
```

You can define this variable:

```Go
var q struct {
	Hero struct {
		Name  string
		Droid struct {
			PrimaryFunction string
		} `graphql:"... on Droid"`
		Human struct {
			Height float64
		} `graphql:"... on Human"`
	} `graphql:"hero(episode: \"JEDI\")"`
}
```

Alternatively, you can define the struct types corresponding to inline fragments, and use them as embedded fields in your query:

```Go
type (
	DroidFragment struct {
		PrimaryFunction string
	}
	HumanFragment struct {
		Height float64
	}
)

var q struct {
	Hero struct {
		Name          string
		DroidFragment `graphql:"... on Droid"`
		HumanFragment `graphql:"... on Human"`
	} `graphql:"hero(episode: \"JEDI\")"`
}
```

Then call `client.Query`:

```Go
err := client.Query(context.Background(), &q, nil)
if err != nil {
	// Handle error.
}
fmt.Println(q.Hero.Name)
fmt.Println(q.Hero.PrimaryFunction)
fmt.Println(q.Hero.Height)

// Output:
// R2-D2
// Astromech
// 0
```

### Mutations

Mutations often require information that you can only find out by performing a query first. Let's suppose you've already done that.

For example, to make the following GraphQL mutation:

```GraphQL
mutation($ep: Episode!, $review: ReviewInput!) {
	createReview(episode: $ep, review: $review) {
		stars
		commentary
	}
}
variables {
	"ep": "JEDI",
	"review": {
		"stars": 5,
		"commentary": "This is a great movie!"
	}
}
```

You can define:

```Go
var m struct {
	CreateReview struct {
		Stars      int
		Commentary string
	} `graphql:"createReview(episode: $ep, review: $review)"`
}
variables := map[string]interface{}{
	"ep": starwars.Episode("JEDI"),
	"review": starwars.ReviewInput{
		Stars:      5,
		Commentary: "This is a great movie!",
	},
}
```

Then call `client.Mutate`:

```Go
err := client.Mutate(context.Background(), &m, variables)
if err != nil {
	// Handle error.
}
fmt.Printf("Created a %v star review: %v\n", m.CreateReview.Stars, m.CreateReview.Commentary)

// Output:
// Created a 5 star review: This is a great movie!
```

### Error Handling

Error handling is needed to:

1. Be a good citizen and back off on rate limits.

2. Retry on transient errors when reliability is important.

Error handling should be implemented by handling one (or more) of the following cases.

```go
import "errors"

...

resp, err := client.Query(ctx, &q, nil)
if resp != nil && resp.StatusCode/100 == 5 {
    // 5xx error
}
var gerr *graphql.Error
if errors.As(err, &gerr) && len(gerr.Errors) > 0 && gerr.Errors[0].Message == "invalid value" {
    // we passed an invalid value in the query
}

// For completeness:
if terr := (interface{Timeout() bool})(nil); errors.As(err, &terr) && terr.Timeout() {
    // timeout produced by HTTP client/transport/dial/DNSLookup
}
if terr := (interface{Temporary() bool})(nil); errors.As(err, &terr) && terr.Temporary() {
    // "temporary" error produced by HTTP client/transport/dial/DNSLookup
}
```

Acknowledgements
----------------

- Dmitri Shuralyov for the slick design of [shurcooL/graphQL](https://github.com/shurcooL/graphql).

License
-------

- [Apache License, Version 2.0](LICENSE)

