# Automatic Opentracing to github.com/go-pg/pg/v9

## Usage

Initialize tracing by adding `TracingQueryHook` to query hooks. That will automatically
instruments all database queries made by using that database instance.

```go
	db = pg.Connect(&pg.Options{})
	defer db.Close()

	db.AddQueryHook(TracingQueryHook{})
```