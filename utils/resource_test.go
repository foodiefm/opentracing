package utils

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetResourceName(t *testing.T) {
	type args struct{}
	tests := []struct {
		name   string
		req    *http.Request
		params map[string]string
		want   string
	}{
		{"get without parameters", createRequest(http.MethodGet, "/without/parameters"), map[string]string{}, "GET__without_parameters"},
		{"post with parameters", createRequest(http.MethodGet, "/with/1234/parameters/6395"), map[string]string{"param1": "1234", "param2": "6395"}, "GET__with_param1_parameters_param2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourceName(tt.req, tt.params); got != tt.want {
				t.Errorf("GetResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleGetResourceName_withoutParameters() {
	req, _ := http.NewRequest(http.MethodGet, "/test/without/parameters", nil)

	resource := GetResourceName(req, nil)

	fmt.Println(resource)
}

func ExampleGetResourceName_withParameters() {
	req, _ := http.NewRequest(http.MethodGet, "/test/with/123/parameters", nil)

	resource := GetResourceName(req, map[string]string{"param1": "123"})

	fmt.Println(resource)
}

func createRequest(method, path string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	return req
}
