package fingerprint

import (
	"reflect"
	"testing"
)

func TestParsePunchTimes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "two punches", raw: "08:03\n16:27", want: []string{"08:03", "16:27"}},
		{name: "one punch with spaces", raw: "08:43\n     ", want: []string{"08:43"}},
		{name: "empty", raw: "", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePunchTimes(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parsePunchTimes(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseDayNumber(t *testing.T) {
	for _, value := range []string{"1", "1.0", "31"} {
		if _, ok := parseDayNumber(value); !ok {
			t.Fatalf("expected %q to be accepted as a day", value)
		}
	}
	if _, ok := parseDayNumber("1.5"); ok {
		t.Fatal("fractional day must be rejected")
	}
}
