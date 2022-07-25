package rcon

import (
	"reflect"
	"testing"
)

func Test_parseServerStatus(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ServerStatus
		wantErr bool
	}{
		{
			"parse-invalid-string",
			args{[]byte("12341-invalid-string")},
			nil,
			true,
		},
		{"parse-response",
			args{[]byte("hostname:  CS 1.6 CLASSIC\nversion :  48/1.1.2.7/Stdio 6153 secure  (10)\ntcp/ip  :  172.21.0.3:27015\nmap     :  de_dust2 at: 0 x, 0 y, 0 z\nplayers :  12 active (32 max)\n\n#      name userid uniqueid frag time ping loss adr\n0 users\n\u0000")},
			&ServerStatus{
				Name:    "CS 1.6 CLASSIC",
				Host:    "172.21.0.3:27015",
				Players: 12,
				Map:     "de_dust2",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseServerStatus(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseServerStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseServerStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}
