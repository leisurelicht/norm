package queryset

import (
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
		{"one", args{Cond{"test": 1}}, Cond{"| test": 1}},
		{"two", args{Cond{"test": 1, "test2": 2}}, Cond{"| test": 1, "| test2": 2}},
		{"three", args{Cond{"test": 1, "test2": 2, "test3": 3}}, Cond{"| test": 1, "| test2": 2, "| test3": 3}},
		{"one_string", args{Cond{"test": "1"}}, Cond{"| test": "1"}},
		{"two_string", args{Cond{"test": "1", "test2": "2"}}, Cond{"| test": "1", "| test2": "2"}},
		{"three_string", args{Cond{"test": "1", "test2": "2", "test3": "3"}}, Cond{"| test": "1", "| test2": "2", "| test3": "3"}},
		{"one_and", args{AND{"test": 1}}, AND{"| test": 1}},
		{"two_and", args{AND{"test": 1, "test2": 2}}, AND{"| test": 1, "| test2": 2}},
		{"three_and", args{AND{"test": 1, "test2": 2, "test3": 3}}, AND{"| test": 1, "| test2": 2, "| test3": 3}},
		{"one_or", args{OR{"test": 1}}, OR{"| test": 1}},
		{"two_or", args{OR{"test": 1, "test2": 2}}, OR{"| test": 1, "| test2": 2}},
		{"three_or", args{OR{"test": 1, "test2": 2, "test3": 3}}, OR{"| test": 1, "| test2": 2, "| test3": 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the input condition to the expected type
			// and check if the output matches the expected output
			// Use a type switch to handle different types of conditions
			// and call EachOR accordingly
			var got any
			switch v := tt.args.cond.(type) {
			case Cond:
				got = EachOR(v)
			case AND:
				got = EachOR(v)
			case OR:
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
