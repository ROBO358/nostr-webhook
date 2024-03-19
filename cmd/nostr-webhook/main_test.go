package main

import (
	"errors"
	"math/rand/v2"
	"os"
	"testing"
	"time"
)

type argsKV struct {
	key     []byte
	value   []byte
	noValue bool
}

func Test_readSecret(t *testing.T) {
	randString := string(rune(rand.N(10 * time.Minute)))
	type argsFields struct {
		secret string
	}
	type wantFields struct {
		secret string
		err    error
	}
	cases := []struct {
		name string

		args argsFields

		want wantFields
	}{
		{
			name: "Should read secret from environment variable",
			args: argsFields{
				secret: "secret" + randString,
			},
			want: wantFields{
				secret: "secret" + randString,
				err:    nil,
			},
		},
		{
			name: "Should fail to read secret from environment variable",
			args: argsFields{
				secret: "",
			},
			want: wantFields{
				secret: "",
				err:    secretEnvError,
			},
		},
	}

	w := webhook{}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.secret != "" {
				os.Setenv("SECRET", tt.args.secret)
				defer os.Unsetenv("SECRET")
			}

			if err := w.readSecret(); !errors.Is(err, tt.want.err) {
				t.Errorf("readSecret() error = %v, want.err = %v", err, tt.want.err)
			}

			if w.secret != tt.want.secret {
				t.Errorf("readSecret() = %v, want = %v", w.secret, tt.want.secret)
			}
		})
	}
}
