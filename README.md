# OpenTracing extensions for Go

This package implements extensions to OpenTracing.
Those include middlewares for servers, automatic
span incjection and creation of span in libraries.

## Get package

```sh
go get github.com/foodiefm/opentracing
```

## Extensions

Extension code will be placed under `contrib` folder
and end of path is same as their import path.

## Example

Creating `http.Client` that automatically sends span
from `context.Context`, when http requests is made using
that client.

```go
package main

import (
    "context"
    "net/http"

    wrapping "github.com/foodiefm/opentracing/contrib/net/http"
    "github.com/opentracing/opentracing-go"
)

func main(){
    client := wrapping.WrapClient(&http.Client{},"span.name")

    span := opentracing.GlobalTracer().StartSpan("test_call")
    defer span.Finish()
    ctx := opentracing.ContextWithSpan(context.Background(), span)

    req, _ := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	client.Do(req.WithContext(ctx))
}
```

## Licensing

The  is available as open source under the terms of the [MIT License](./LICENSE.txt).