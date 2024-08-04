package ruleengine

import (
	"reflect"
	"regexp"
	"testing"
)

func TestParseRuleMatcher(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  RuleMatcher
	}{
		{
			name:  "simple Regexp",
			value: "X",
			want:  regexpMatcher{regexp.MustCompile("X")},
		}, {
			name:  "Regexp",
			value: "=~ X",
			want:  regexpMatcher{regexp.MustCompile("X")},
		}, {
			name:  "numeric equality matcher",
			value: "= 1",
			want: numberMatcher{
				operation: eq,
				number:    1,
			},
		}, {
			name:  "string equality matcher",
			value: "= a",
			want:  equalityMatcher{s: "a"},
		}, {
			name:  "numeric greater or equal matcher",
			value: ">= 1",
			want: numberMatcher{
				operation: ge,
				number:    1,
			},
		}, {
			name:  "numeric less than matcher",
			value: "< 1",
			want: numberMatcher{
				operation: lt,
				number:    1,
			},
		}, {
			name:  "numeric greater than matcher",
			value: "> 1",
			want: numberMatcher{
				operation: gt,
				number:    1,
			},
		}, {
			name:  "floating point greater than matcher",
			value: "> 1.2",
			want: numberMatcher{
				operation: gt,
				number:    1.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRuleMatcher(tt.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRuleMatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleMatching(t *testing.T) {
	tests := []struct {
		name string
		rule string
		str  string
		want bool
	}{
		{
			name: "simple Regexp",
			rule: "X",
			str:  "Xavier",
			want: true,
		}, {
			name: "regexp",
			rule: "=~ X",
			str:  "Xavier",
			want: true,
		}, {
			name: "float equality",
			rule: "= 1.01",
			str:  "1.01",
			want: true,
		}, {
			name: "float equality",
			rule: "= 1.01",
			str:  "1.010000001",
			want: true,
		}, {
			name: "float non-equality",
			rule: "= 1.01",
			str:  "1.001",
			want: false,
		}, {
			name: "string equality",
			rule: "= a",
			str:  "a",
			want: true,
		}, {
			name: "anchored regexp",
			rule: "(^|,)b(,|$)",
			str:  "a,b,c",
			want: true,
		}, {
			name: "anchored regexp",
			rule: "(^|,)c(,|$)",
			str:  "a,b,c",
			want: true,
		}, {
			name: "anchored regexp",
			rule: "(^|,)a(,|$)",
			str:  "a,b,c",
			want: true,
		}, {
			name: "anchored regexp",
			rule: "(^|,)d(,|$)",
			str:  "a,b,c",
			want: false,
		}, {
			name: "STAAAR",
			rule: "modality2star.*deadletter",
			str:  "CPU wait",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRuleMatcher(tt.rule); got.MatchString(tt.str) != tt.want {
				t.Errorf("Matching %v =~ %v, want %v", tt.rule, tt.str, tt.want)
			}
		})
	}
}

func Test_floatEqualEnough(t *testing.T) {
	type args struct {
		a float64
		b float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal",
			args: args{1, 1},
			want: true,
		}, {
			name: "very close",
			args: args{1.0, 1.000000001},
			want: true,
		}, {
			name: "near",
			args: args{1.0, 1.0001},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floatEqualEnough(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("floatEqualEnough() = %v, want %v", got, tt.want)
			}
		})
	}
}
