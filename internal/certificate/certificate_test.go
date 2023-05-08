package certificate_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/likexian/gokit/assert"
	"github.com/mckern/spiry/internal/certificate"
)

func TestNewWithName(t *testing.T) {
	type args struct {
		name    string
		address string
	}

	tests := []struct {
		name     string
		args     args
		hostName string
		wantErr  bool
	}{
		{name: "a given name is used",
			args:     args{name: "elpmaxe.com", address: "https://example.com/"},
			hostName: "example.com",
			wantErr:  false},
		{name: "a given name is may match a given address",
			args:     args{name: "example.com", address: "https://example.com/"},
			hostName: "example.com",
			wantErr:  false},
		{name: "an invalid given name raises an error",
			args:    args{name: "sanford&son.example.com", address: "https://example.com/"},
			wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := certificate.NewWithName(tt.args.name, tt.args.address)
			fmt.Fprintf(os.Stderr, "\n-----\nerr: %+v\n-----\n", err)

			if tt.wantErr {
				assert.NotNil(t, err, "errors should be raised for invalid addresses or names")
				// none of the remaining tests will work if an error was returned
				return
			}

			if tt.args.name != tt.hostName {
				assert.NotEqual(t, cert.Name(), tt.hostName, "should not use the same name of the provided address")
			}

			assert.Nil(t, err, "no errors should be raised for valid addresses or names")
			assert.NotNil(t, cert, "a complete Certificate instance should be returned")
			assert.Equal(t, cert.Name(), tt.args.name, "should use the name provided")

		})
	}
}
