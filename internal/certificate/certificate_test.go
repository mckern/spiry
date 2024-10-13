package certificate_test

import (
	"testing"

	"github.com/likexian/gokit/assert"
	"github.com/mckern/spiry/internal/certificate"
)

type args struct {
	name    string
	address string
}

var certTests = []struct {
	name        string
	args        args
	hostName    string
	wantNameErr bool
	wantAddrErr bool
}{
	{name: "a given name is used to request a certificate from a given address",
		args:        args{name: "elpmaxe.com", address: "https://example.com/"},
		hostName:    "example.com",
		wantNameErr: false,
		wantAddrErr: false},
	{name: "a given name may match a given address",
		args:        args{name: "example.com", address: "https://example.com/"},
		hostName:    "example.com",
		wantNameErr: false,
		wantAddrErr: false},
	{name: "an invalid name raises an error",
		args:        args{name: "sanford&son.example.com", address: "https://example.com/"},
		wantNameErr: true,
		wantAddrErr: false},
	{name: "an invalid URL address raises an error",
		args:        args{name: "example.com", address: "https://sanford&son.example.com/"},
		wantNameErr: false,
		wantAddrErr: true},
	{name: "an invalid raw address raises an error",
		args:        args{name: "example.com", address: "sanford&son.example.com"},
		wantNameErr: false,
		wantAddrErr: true},
}

func TestNewWithName(t *testing.T) {
	for _, tt := range certTests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := certificate.NewWithName(tt.args.name, tt.args.address)

			if tt.wantNameErr {
				assert.NotNil(t, err, "an error should be raised for an invalid name")
				// none of the remaining tests will work if an error was returned
				return
			}

			if tt.wantAddrErr {
				assert.NotNil(t, err, "an error should be raised for an invalid address")
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

func TestNew(t *testing.T) {
	for _, tt := range certTests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := certificate.New(tt.args.address)

			if tt.wantAddrErr {
				assert.NotNil(t, err, "an error should be raised for an invalid address")
				return
			}

			assert.Nil(t, err, "no errors should be raised for valid addresses or names")
			assert.NotNil(t, cert, "a complete Certificate instance should be returned")
			assert.NotZero(t, cert.Name(), "returned Certificate instance should have a Name")
		})
	}
}
