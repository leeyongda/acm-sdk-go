package acm

import "testing"

func TestIsValid(t *testing.T) {

	type args struct {
		param string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "dataId", args: args{"test.app.com.test"}, want: true},
		{name: "dataId", args: args{"test.app.com.test_1"}, want: true},
		{name: "dataId", args: args{"test.app.com.test*"}, want: false},
		{name: "dataId", args: args{"test.app.com.test2"}, want: true},
		{name: "group", args: args{"test.app.com.test2"}, want: true},
		{name: "group", args: args{"test-1.app.com.test2"}, want: true},
		{name: "group", args: args{"test-1.app.com.test-1"}, want: true},
		{name: "group", args: args{"test-1.app.com.test-1;"}, want: false},
		{name: "group", args: args{"test-1.app.com.test-1111"}, want: true},
		{name: "group", args: args{"test-1.app.com.test__1111"}, want: true},
		{name: "group", args: args{"test-1.app.com.test__1111@"}, want: false},
		{name: "group", args: args{"test-1.app.com.test——"}, want: false},
		{name: "group", args: args{"test-1.app.com.test__1111!@#$%^&*()-+[]//:',.|?、？ "}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.args.param); got != tt.want {
				t.Errorf("IsValid() = %v, args %s want %v", got, tt.args.param, tt.want)
			}
		})
	}
}
