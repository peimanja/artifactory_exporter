package arti

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestCheckErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		{"simple", args{errors.New("This is a test error")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckErr(tt.args.err)
		})
	}
}

func TestRemoveCommas(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"number", args{"10,101.2"}, 10101.2},
		{"disk_space", args{"10,101.2 GB"}, 10101.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveCommas(tt.args.str); got != tt.want {
				t.Errorf("RemoveCommas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesConverter(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"bytes", args{"10,101.2 bytes"}, 10101.2, false},
		{"KB", args{"10,101.2 KB"}, 10101.2 * 1024, false},
		{"MB", args{"10,101.2 MB"}, 10101.2 * 1024 * 1024, false},
		{"GB", args{"10,101.2 GB"}, 10101.2 * 1024 * 1024 * 1024, false},
		{"TB", args{"10,101.2 TB"}, 10101.2 * 1024 * 1024 * 1024 * 1024, false},
		{"Invalid", args{"10,101.2"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesConverter(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("BytesConverter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BytesConverter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromEpoch(t *testing.T) {
	type args struct {
		t time.Duration
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromEpoch(tt.args.t); got != tt.want {
				t.Errorf("FromEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountUsers(t *testing.T) {
	type args struct {
		users []User
	}
	tests := []struct {
		name string
		args args
		want []UsersCount
	}{
		{
			"no_internal",
			args{
				[]User{{"test1", "saml"}},
			},
			[]UsersCount{{1, "saml"}, {0, "internal"}},
		},
		{
			"no_saml",
			args{
				[]User{{"test2", "internal"}},
			},
			[]UsersCount{{0, "saml"}, {1, "internal"}},
		},
		{
			"internal_saml",
			args{
				[]User{{"test1", "saml"}, {"test2", "internal"}},
			},
			[]UsersCount{{1, "saml"}, {1, "internal"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountUsers(tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CountUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
