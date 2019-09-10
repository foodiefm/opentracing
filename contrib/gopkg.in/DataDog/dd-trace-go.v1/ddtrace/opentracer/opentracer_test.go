package opentracer

import (
	"reflect"
	"testing"

	"github.com/opentracing/opentracing-go"
)

func TestComponentApply(t *testing.T) {
	tests := []struct {
		name      string
		component string
		opts      *opentracing.StartSpanOptions
	}{
		{"nil opts and empty name", "", nil},
		{"nil opts and name", "test", nil},
		{"opts and empty name", "", &opentracing.StartSpanOptions{Tags: map[string]interface{}{}}},
		{"opts and name", "test", &opentracing.StartSpanOptions{Tags: map[string]interface{}{}}},
		{"component in opts and empty name", "", &opentracing.StartSpanOptions{Tags: map[string]interface{}{"component": "c"}}},
		{"component in opts and name", "test", &opentracing.StartSpanOptions{Tags: map[string]interface{}{"component": "c"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := component{
				component: tt.component,
			}
			c.Apply(tt.opts)
			if tt.opts != nil && tt.component != "" {
				if v, exists := tt.opts.Tags["component"]; exists {
					if c, ok := v.(string); ok {
						if c != tt.component {
							t.Errorf("incorrect component name, expected: '%s', got: '%s'", c, tt.component)
						}
					} else {
						t.Errorf("Incorrect component type, should be string")
					}
				} else {
					t.Errorf("Component name does not exists even it should be: '%s'", tt.component)
				}

			}
			if tt.opts != nil && tt.component == "" {
				if v, exists := tt.opts.Tags["component"]; exists {
					if c, ok := v.(string); ok {
						if c == tt.component {
							t.Errorf("incorrect component name, expected not to be '%s'", tt.component)
						}
					} else {
						t.Errorf("Incorrect component type, should be string")
					}
				}
			}
		})
	}
}

func TestComponentOption(t *testing.T) {
	tests := []struct {
		name string
		c    string
		want component
	}{
		{"empty name", "", component{component: ""}},
		{"name", "test", component{component: "test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComponentOption(tt.c).(component)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComponentOption() = %v, want %v", got, tt.want)
			}

		})
	}
}
