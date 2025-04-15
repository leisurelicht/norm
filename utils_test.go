package norm

import (
	"reflect"
	"testing"
)

func Test_shiftName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{"DevicePolicyMap"}, "`device_policy_map`"},
		{"test2", args{"DevicePolicy"}, "`device_policy`"},
		{"test3", args{"Device"}, "`device`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shiftName(tt.args.s); got != tt.want {
				t.Errorf("shiftName -> get :%v", got)
				t.Errorf("shiftName -> want: %v", tt.want)
			}
		})
	}
}

func Test_rawFieldNames(t *testing.T) {
	type args struct {
		in  any
		tag string
		pg  bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"test pg", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", false}, []string{"`device`", "`device_policy`", "`device_policy_map`"}},
		{"test not pg", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", true}, []string{"device", "device_policy", "device_policy_map"}},
		{"test ignore with pg", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", false}, []string{"`device`", "`device_policy`"}},
		{"test ignore with not pg", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", true}, []string{"device", "device_policy"}},
		{"test with multiple tag with pg", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", false}, []string{"`device`", "`device_policy`", "`device_policy_map`"}},
		{"test with multiple tag with not pg", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", true}, []string{"device", "device_policy", "device_policy_map"}},
		{"test with multiple tag with pg and ignore", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", false}, []string{"`device`", "`device_policy`"}},
		{"test with multiple tag with not pg and ignore", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", true}, []string{"device", "device_policy"}},
		{"test with empty tag", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:",type=char"`
		}{}, "db", false}, []string{"`Device`", "`device_policy`", "`DevicePolicyMap`"}},
		{"test with empty struct with not pg", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:",type=char"`
		}{}, "db", true}, []string{"Device", "device_policy", "DevicePolicyMap"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rawFieldNames(tt.args.in, DefaultModelTag, tt.args.pg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawFieldNames -> get : %v", got)
				t.Errorf("rawFieldNames -> want: %v", tt.want)
			}
		})
	}
}
