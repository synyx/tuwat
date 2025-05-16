package aggregation

import (
	"reflect"
	"testing"
)

func TestExtractUrls(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "single",
			text: "foo http://localhost bar",
			want: []string{"http://localhost"},
		},
		{
			name: "multiple",
			text: "foo http://localhost https://www.example.com bar",
			want: []string{"http://localhost", "https://www.example.com"},
		},
		{
			name: "complex",
			text: "foo https://www.example.com/x/1234?foo=bar bar",
			want: []string{"https://www.example.com/x/1234?foo=bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractUrls(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractUrls() = %v, want %v", got, tt.want)
			}
		})
	}
}
