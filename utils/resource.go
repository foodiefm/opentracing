package utils

import (
	"net/http"
	"strings"
)

// ResourceNameResolver defines function type that can be user to override
// default resource name creation in middlewares
type ResourceNameResolver func(*http.Request, map[string]string) string

// GetResourceName creates informative resource name for http request
func GetResourceName(req *http.Request, params map[string]string) string {
	if req == nil {
		return ""
	}
	// Get resource name
	resource := req.URL.EscapedPath()
	for key, value := range params {
		resource = strings.Replace(resource, value, key, 1)
	}
	resource = strings.Replace(resource, "/", "_", -1)
	resource = strings.Replace(resource, "-", "_", -1)

	return req.Method + "_" + resource
}
