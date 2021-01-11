package core

import (
	"path/filepath"
	"testing"
)

func Test_normalizeFsPath(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Slash",
			args: args{v: "asd/dsa/qwe"},
			want: filepath.Join("asd", "dsa", "qwe"),
		},
		{
			name: "Without Slash",
			args: args{v: "asd"},
			want: "asd",
		},
		{
			name: "Empty string",
			args: args{v: ""},
			want: "",
		},
		{
			name: "Double slash",
			args: args{v: "asd//dsa/qwe"},
			want: filepath.Join("asd", "dsa", "qwe"),
		},
		{
			name: "Slash prefix",
			args: args{v: "/asd/dsa/qwe"},
			want: filepath.Join("asd", "dsa", "qwe"),
		},
		{
			name: "Slash suffix",
			args: args{v: "asd/dsa/qwe/"},
			want: filepath.Join("asd", "dsa", "qwe"),
		},
		{
			name: "Parent dir1",
			args: args{v: "../../dsa/qwe/"},
			want: filepath.Join("dsa", "qwe"),
		},
		{
			name: "Parent dir2",
			args: args{v: "dsa/qwe/../../../../asd"},
			want: filepath.Join("dsa", "qwe", "asd"),
		},
		{
			name: "Dot1",
			args: args{v: "dsa/qwe/./asd"},
			want: filepath.Join("dsa", "qwe", "asd"),
		},
		{
			name: "Dot2",
			args: args{v: "dsa/qwe/asd/."},
			want: filepath.Join("dsa", "qwe", "asd"),
		},
		{
			name: "Dot3",
			args: args{v: "./dsa/qwe/asd"},
			want: filepath.Join("dsa", "qwe", "asd"),
		},
		{
			name: "Dot4",
			args: args{v: ".../dsa/./../qwe/.../asd/..."},
			want: filepath.Join("dsa", "qwe", "asd"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFsPath(tt.args.v); got != tt.want {
				t.Errorf("normalizeFsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizeUrlPath(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Slash",
			args: args{v: "asd/dsa/qwe"},
			want: "asd/dsa/qwe",
		},
		{
			name: "Without Slash",
			args: args{v: "asd"},
			want: "asd",
		},
		{
			name: "Empty string",
			args: args{v: ""},
			want: "",
		},
		{
			name: "Double slash",
			args: args{v: "asd//dsa/qwe"},
			want: "asd/dsa/qwe",
		},
		{
			name: "Slash prefix",
			args: args{v: "/asd/dsa/qwe"},
			want: "asd/dsa/qwe",
		},
		{
			name: "Slash suffix",
			args: args{v: "asd/dsa/qwe/"},
			want: "asd/dsa/qwe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeUrlPath(tt.args.v); got != tt.want {
				t.Errorf("normalizeUrlPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
