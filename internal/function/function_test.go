package function

import (
	"github.com/leisurelicht/norm/internal/queryset"
	"reflect"
	"testing"
)

func TestToOR(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{"test"}, "| test"},
		{"empty", args{""}, ""},
		{"already", args{"| test"}, "| test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToOR(tt.args.key); got != tt.want {
				t.Errorf("ToOR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEachOR(t *testing.T) {
	type args struct {
		cond any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{"one", args{queryset.Cond{"test": 1}}, queryset.Cond{"| test": 1}},
		{"two", args{queryset.Cond{"test": 1, "test2": 2}}, queryset.Cond{"| test": 1, "| test2": 2}},
		{"three", args{queryset.Cond{"test": 1, "test2": 2, "test3": 3}}, queryset.Cond{"| test": 1, "| test2": 2, "| test3": 3}},
		{"one_string", args{queryset.Cond{"test": "1"}}, queryset.Cond{"| test": "1"}},
		{"two_string", args{queryset.Cond{"test": "1", "test2": "2"}}, queryset.Cond{"| test": "1", "| test2": "2"}},
		{"three_string", args{queryset.Cond{"test": "1", "test2": "2", "test3": "3"}}, queryset.Cond{"| test": "1", "| test2": "2", "| test3": "3"}},
		{"one_and", args{queryset.AND{"test": 1}}, queryset.AND{"| test": 1}},
		{"two_and", args{queryset.AND{"test": 1, "test2": 2}}, queryset.AND{"| test": 1, "| test2": 2}},
		{"three_and", args{queryset.AND{"test": 1, "test2": 2, "test3": 3}}, queryset.AND{"| test": 1, "| test2": 2, "| test3": 3}},
		{"one_or", args{queryset.OR{"test": 1}}, queryset.OR{"| test": 1}},
		{"two_or", args{queryset.OR{"test": 1, "test2": 2}}, queryset.OR{"| test": 1, "| test2": 2}},
		{"three_or", args{queryset.OR{"test": 1, "test2": 2, "test3": 3}}, queryset.OR{"| test": 1, "| test2": 2, "| test3": 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the input condition to the expected type
			// and check if the output matches the expected output
			// Use a type switch to handle different types of conditions
			// and call EachOR accordingly
			var got any
			switch v := tt.args.cond.(type) {
			case queryset.Cond:
				got = EachOR(v)
			case queryset.AND:
				got = EachOR(v)
			case queryset.OR:
				got = EachOR(v)
			default:
				t.Fatalf("unsupported type: %T", v)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get  content: %v, get  type: %v", got, reflect.TypeOf(got).String())
				t.Errorf("want content: %v, want type: %v", tt.want, reflect.TypeOf(tt.want).String())
			}
		})
	}
}
