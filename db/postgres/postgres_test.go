package postgres_test

import (
	"reflect"
	"testing"

	"github.com/OZahed/bob/db/postgres"
)

func TestWithHost(t *testing.T) {
	type args struct {
		hostname string
	}
	tests := []struct {
		name string
		args args
		want postgres.DBOption
	}{
		// TODO: Add test cases.
		{
			name: "WithHost not having default values (not having side effects)",
			args: args{
				hostname: "",
			},
			want: func(opts *postgres.Options) *postgres.Options {
				opts.Host = ""
				return opts
			},
		},
		{
			name: "WithHost having value localhost",
			args: args{
				hostname: "localhost",
			},
			want: func(opts *postgres.Options) *postgres.Options {
				opts.Host = "localhost"
				return opts
			},
		},
	}

	opt := &postgres.Options{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := postgres.WithHost(tt.args.hostname); !reflect.DeepEqual(got(opt), tt.want(opt)) {
				t.Errorf("WithHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
