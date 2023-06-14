package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAPI(t *testing.T) {

	testCases := []struct {
		name      string
		args      string
		wantAddr  string
		wantToken string
	}{
		{
			name:      "case1",
			args:      "http://localhost:1234",
			wantAddr:  "http://localhost:1234",
			wantToken: "",
		},
		{
			name:      "case2",
			args:      ":http://localhost:1234/",
			wantAddr:  "http://localhost:1234/",
			wantToken: "",
		},
		{
			name:      "case3",
			args:      "token:http://localhost:1234/",
			wantAddr:  "http://localhost:1234/",
			wantToken: "token",
		},
		{
			name:      "case4",
			args:      "token:/ip4/127.0.0.1/tcp/1234",
			wantAddr:  "/ip4/127.0.0.1/tcp/1234",
			wantToken: "token",
		},
		{
			name:      "case5",
			args:      ":/ip4/127.0.0.1/tcp/1234",
			wantAddr:  "/ip4/127.0.0.1/tcp/1234",
			wantToken: "",
		},
		{
			name:      "case6",
			args:      "/ip4/127.0.0.1/tcp/1234",
			wantAddr:  "/ip4/127.0.0.1/tcp/1234",
			wantToken: "",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotAddr, gotToken := ParseAPI(tt.args)
			assert.Equal(t, tt.wantAddr, gotAddr)
			assert.Equal(t, tt.wantToken, gotToken)
		})
	}
}
