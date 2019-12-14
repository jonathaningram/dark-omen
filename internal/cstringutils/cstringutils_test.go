package cstringutils

import "testing"

func TestToGo(t *testing.T) {
	tests := []struct {
		name string
		n    []byte
		want string
	}{
		{
			name: "C string",
			n:    []byte("Morgan Bernhardt\x00"),
			want: "Morgan Bernhardt",
		},
		{
			name: "Non-C string",
			n:    []byte("Morgan Bernhardt"),
			want: "Morgan Bernhardt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToGo(tt.n); got != tt.want {
				t.Errorf("ToGo() = %v, want %v", got, tt.want)
			}
		})
	}
}
