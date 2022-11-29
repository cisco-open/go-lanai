package generator

import "testing"

func Test_filenameFromPath(t *testing.T) {
	type args struct {
		pathName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should convert an api path to a filename",
			args: args{pathName: "/my/api/v1/testpath/{scope}"},
			want: "testpath_scope.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filenameFromPath(tt.args.pathName); got != tt.want {
				t.Errorf("filenameFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
