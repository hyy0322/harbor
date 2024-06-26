package project

import "testing"

func TestValidProjectName(t *testing.T) {
	type args struct {
		project string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case1",
			args: args{
				project: "1__test",
			},
			want: true,
		}, {
			name: "case2",
			args: args{
				project: "1_test",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validProjectName.MatchString(tt.args.project); got != tt.want {
				t.Errorf("validProjectName() = %v, want %v", got, tt.want)
			}
		})
	}
}
