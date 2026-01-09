package queryset

import (
	"fmt"
	"reflect"
	"testing"

	go_zero "github.com/leisurelicht/norm/operator/mysql/go-zero"
)

func Test_isStrList(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test0", args{nil}, false},
		{"test1", args{[]string{"1", "2", "3"}}, true},
		{"test2", args{[]int{1, 2, 3}}, false},
		{"test3", args{[]int64{1, 2, 3}}, false},
		{"test4", args{[]float64{1.0, 2.0, 3.0}}, false},
		{"test5", args{[]float32{1.0, 2.0, 3.0}}, false},
		{"test6", args{[]bool{true, false, true}}, false},
		{"test7", args{[]any{"1", "2", "3"}}, false},
		{"test8", args{[]any{1, 2, 3}}, false},
		{"test9", args{[]any{1.0, 2.0, 3.0}}, false},
		{"test10", args{[]any{true, false, true}}, false},
		{"test1", args{[3]string{"1", "2", "3"}}, true},
		{"test2", args{[3]int{1, 2, 3}}, false},
		{"test3", args{[3]int64{1, 2, 3}}, false},
		{"test4", args{[3]float64{1.0, 2.0, 3.0}}, false},
		{"test5", args{[3]float32{1.0, 2.0, 3.0}}, false},
		{"test6", args{[3]bool{true, false, true}}, false},
		{"test7", args{[3]any{"1", "2", "3"}}, false},
		{"test8", args{[3]any{1, 2, 3}}, false},
		{"test9", args{[3]any{1.0, 2.0, 3.0}}, false},
		{"test10", args{[3]any{true, false, true}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStrList(tt.args.v); got != tt.want {
				t.Errorf("isStrList() error\nGot : %v\nWant %v", got, tt.want)
			}
		})
	}
}

func Test_isNumericKind(t *testing.T) {
	type args struct {
		kind reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{reflect.Int}, true},
		{"test2", args{reflect.Int8}, true},
		{"test3", args{reflect.Int16}, true},
		{"test4", args{reflect.Int32}, true},
		{"test5", args{reflect.Int64}, true},
		{"test6", args{reflect.Uint}, true},
		{"test7", args{reflect.Uint8}, true},
		{"test8", args{reflect.Uint16}, true},
		{"test9", args{reflect.Uint32}, true},
		{"test10", args{reflect.Uint64}, true},
		{"test11", args{reflect.Float32}, true},
		{"test12", args{reflect.Float64}, true},
		{"test13", args{reflect.String}, false},
		{"test14", args{reflect.Bool}, false},
		{"test15", args{reflect.Slice}, false},
		{"test16", args{reflect.Array}, false},
		{"test17", args{reflect.Map}, false},
		{"test18", args{reflect.Chan}, false},
		{"test19", args{reflect.Func}, false},
		{"test20", args{reflect.Interface}, false},
		{"test21", args{reflect.Ptr}, false},
		{"test22", args{reflect.UnsafePointer}, false},
		{"test23", args{reflect.Struct}, false},
		{"test24", args{reflect.Invalid}, false},
		{"test25", args{reflect.Uintptr}, false},
		{"test26", args{reflect.Complex64}, false},
		{"test27", args{reflect.Complex128}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumericKind(tt.args.kind); got != tt.want {
				t.Errorf("isNumericKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isStringKind(t *testing.T) {
	type args struct {
		kind reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{reflect.String}, true},
		{"test2", args{reflect.Int}, false},
		{"test3", args{reflect.Int8}, false},
		{"test4", args{reflect.Int16}, false},
		{"test5", args{reflect.Int32}, false},
		{"test6", args{reflect.Int64}, false},
		{"test7", args{reflect.Uint}, false},
		{"test8", args{reflect.Uint8}, false},
		{"test9", args{reflect.Uint16}, false},
		{"test10", args{reflect.Uint32}, false},
		{"test11", args{reflect.Uint64}, false},
		{"test12", args{reflect.Float32}, false},
		{"test13", args{reflect.Float64}, false},
		{"test14", args{reflect.Bool}, false},
		{"test15", args{reflect.Slice}, false},
		{"test16", args{reflect.Array}, false},
		{"test17", args{reflect.Map}, false},
		{"test18", args{reflect.Chan}, false},
		{"test19", args{reflect.Func}, false},
		{"test20", args{reflect.Interface}, false},
		{"test21", args{reflect.Ptr}, false},
		{"test22", args{reflect.UnsafePointer}, false},
		{"test23", args{reflect.Struct}, false},
		{"test24", args{reflect.Invalid}, false},
		{"test25", args{reflect.Uintptr}, false},
		{"test26", args{reflect.Complex64}, false},
		{"test27", args{reflect.Complex128}, false},
		{"test28", args{reflect.UnsafePointer}, false},
		{"test29", args{reflect.Uintptr}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStringKind(tt.args.kind); got != tt.want {
				t.Errorf("isStringKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isBoolKind(t *testing.T) {
	type args struct {
		kind reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{reflect.Bool}, true},
		{"test2", args{reflect.Int}, false},
		{"test3", args{reflect.Int8}, false},
		{"test4", args{reflect.Int16}, false},
		{"test5", args{reflect.Int32}, false},
		{"test6", args{reflect.Int64}, false},
		{"test7", args{reflect.Uint}, false},
		{"test8", args{reflect.Uint8}, false},
		{"test9", args{reflect.Uint16}, false},
		{"test10", args{reflect.Uint32}, false},
		{"test11", args{reflect.Uint64}, false},
		{"test12", args{reflect.Float32}, false},
		{"test13", args{reflect.Float64}, false},
		{"test14", args{reflect.String}, false},
		{"test15", args{reflect.Slice}, false},
		{"test16", args{reflect.Array}, false},
		{"test17", args{reflect.Map}, false},
		{"test18", args{reflect.Chan}, false},
		{"test19", args{reflect.Func}, false},
		{"test20", args{reflect.Interface}, false},
		{"test21", args{reflect.Ptr}, false},
		{"test22", args{reflect.UnsafePointer}, false},
		{"test23", args{reflect.Struct}, false},
		{"test24", args{reflect.Invalid}, false},
		{"test25", args{reflect.Uintptr}, false},
		{"test26", args{reflect.Complex64}, false},
		{"test27", args{reflect.Complex128}, false},
		{"test28", args{reflect.UnsafePointer}, false},
		{"test29", args{reflect.Uintptr}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBoolKind(tt.args.kind); got != tt.want {
				t.Errorf("isBoolKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isListKind(t *testing.T) {
	type args struct {
		kind reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{reflect.Slice}, true},
		{"test2", args{reflect.Array}, true},
		{"test3", args{reflect.Int}, false},
		{"test4", args{reflect.Int8}, false},
		{"test5", args{reflect.Int16}, false},
		{"test6", args{reflect.Int32}, false},
		{"test7", args{reflect.Int64}, false},
		{"test8", args{reflect.Uint}, false},
		{"test9", args{reflect.Uint8}, false},
		{"test10", args{reflect.Uint16}, false},
		{"test11", args{reflect.Uint32}, false},
		{"test12", args{reflect.Uint64}, false},
		{"test13", args{reflect.Float32}, false},
		{"test14", args{reflect.Float64}, false},
		{"test15", args{reflect.String}, false},
		{"test16", args{reflect.Bool}, false},
		{"test17", args{reflect.Map}, false},
		{"test18", args{reflect.Chan}, false},
		{"test19", args{reflect.Func}, false},
		{"test20", args{reflect.Interface}, false},
		{"test21", args{reflect.Ptr}, false},
		{"test22", args{reflect.UnsafePointer}, false},
		{"test23", args{reflect.Struct}, false},
		{"test24", args{reflect.Invalid}, false},
		{"test25", args{reflect.Uintptr}, false},
		{"test26", args{reflect.Complex64}, false},
		{"test27", args{reflect.Complex128}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isListKind(tt.args.kind); got != tt.want {
				t.Errorf("isListKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genStrListValueLikeSQL(t *testing.T) {
	type args struct {
		p                *QuerySetImpl
		filterConditions map[string]*cond
		fieldName        string
		valueOf          reflect.Value
		notFlag          int
		operator         string
		valueFormat      string
	}
	tests := []struct {
		name      string
		args      args
		wantSQL   string
		wantArgs  []any
		wantError error
	}{
		{
			name: "test contains value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "contains",
				valueFormat:      "%%v%%",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "contains"),
		},
		{
			name: "test icontains value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "icontains",
				valueFormat:      "%%v%%",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "icontains"),
		},
		{
			name: "test startswith value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "startswith",
				valueFormat:      "%v%%",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "startswith"),
		},
		{
			name: "test istartswith value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "istartswith",
				valueFormat:      "%v%%",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "istartswith"),
		},
		{
			name: "test endswitch value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "endswith",
				valueFormat:      "%%%v",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "endswith"),
		},
		{
			name: "test iendswitch value empty error",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil), called: 1, selectColumn: "*", filterConds: [][]cond{}, filterConjTag: []int{0}},
				filterConditions: make(map[string]*cond),
				fieldName:        "name",
				valueOf:          reflect.ValueOf([]string{"", ""}),
				notFlag:          0,
				operator:         "iendswith",
				valueFormat:      "%%%v",
			},
			wantSQL:   "",
			wantArgs:  []any{},
			wantError: fmt.Errorf(operatorValueEmptyError, "iendswith"),
		},
		{
			name: "test with string list not contains",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          1,
				operator:         "contains",
				valueFormat:      "%%%v%%",
			},
			wantSQL:  "`test_field` NOT LIKE BINARY ? AND `test_field` NOT LIKE BINARY ?",
			wantArgs: []any{"%value1%", "%value2%"},
		},
		{
			name: "test with string list startswith",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          0,
				operator:         "startswith",
				valueFormat:      "%v%%",
			},
			wantSQL:  "`test_field` LIKE BINARY ? OR `test_field` LIKE BINARY ?",
			wantArgs: []any{"value1%", "value2%"},
		},
		{
			name: "test with string list endswith",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          0,
				operator:         "endswith",
				valueFormat:      "%%%v",
			},
			wantSQL:  "`test_field` LIKE BINARY ? OR `test_field` LIKE BINARY ?",
			wantArgs: []any{"%value1", "%value2"},
		},
		{
			name: "test with single string",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1"}),
				notFlag:          0,
				operator:         "contains",
				valueFormat:      "%%%v%%",
			},
			wantSQL:  "`test_field` LIKE BINARY ?",
			wantArgs: []any{"%value1%"},
		},
		{
			name: "test with icontains operator",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          0,
				operator:         "icontains",
				valueFormat:      "%%%v%%",
			},
			wantSQL:  "`test_field` LIKE ? OR `test_field` LIKE ?",
			wantArgs: []any{"%value1%", "%value2%"},
		},
		{
			name: "test with istartswith operator",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          0,
				operator:         "istartswith",
				valueFormat:      "%v%%",
			},
			wantSQL:  "`test_field` LIKE ? OR `test_field` LIKE ?",
			wantArgs: []any{"value1%", "value2%"},
		},
		{
			name: "test with iendswith operator",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2"}),
				notFlag:          0,
				operator:         "iendswith",
				valueFormat:      "%%%v",
			},
			wantSQL:  "`test_field` LIKE ? OR `test_field` LIKE ?",
			wantArgs: []any{"%value1", "%value2"},
		},
		{
			name: "test contains more than 2 values",
			args: args{
				p:                &QuerySetImpl{Operator: go_zero.NewOperator(nil)},
				filterConditions: make(map[string]*cond),
				fieldName:        "test_field",
				valueOf:          reflect.ValueOf([]string{"value1", "value2", "value3"}),
				notFlag:          0,
				operator:         "contains",
				valueFormat:      "%%%v%%",
			},
			wantSQL:  "`test_field` LIKE BINARY ? OR `test_field` LIKE BINARY ? OR `test_field` LIKE BINARY ?",
			wantArgs: []any{"%value1%", "%value2%", "%value3%"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genStrListValueLikeSQL(tt.args.p, tt.args.filterConditions, tt.args.fieldName, tt.args.valueOf, tt.args.notFlag, tt.args.operator, tt.args.valueFormat)

			if tt.wantError != nil {
				if tt.args.p.err == nil {
					t.Errorf("Expected error: %v, but got nil", tt.wantError)
				} else if tt.args.p.err.Error() != tt.wantError.Error() {
					t.Errorf("Expected error: %v, but got: %v", tt.wantError, tt.args.p.err)
				}
				return
			}

			// Verify the condition was created correctly
			condition, exists := tt.args.filterConditions[tt.args.fieldName]
			if !exists {
				t.Errorf("No condition created for field %s", tt.args.fieldName)
				return
			}

			// Check SQL
			if condition.SQL != tt.wantSQL {
				t.Errorf("SQL mismatch\nGot : %s\nWant: %s", condition.SQL, tt.wantSQL)
			}

			// Check Args
			if !reflect.DeepEqual(condition.Args, tt.wantArgs) {
				t.Errorf("Args mismatch\nGot : %+v\nWant: %+v", condition.Args, tt.wantArgs)
			}
		})
	}
}

func Test_joinSQL(t *testing.T) {
	type args struct {
		filterSql  *string
		filterArgs *[]any
		index      int
		condition  *cond
	}
	type want struct {
		filterSql  *string
		filterArgs *[]any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"test nil", args{nil, nil, 0, nil}, want{nil, nil}},
		{"test empty condition", args{func() *string { sql := "1"; return &sql }(), &[]any{}, 0, &cond{}}, want{func() *string { sql := "1"; return &sql }(), &[]any{}}},

		// Basic condition tests - first condition (index 0)
		{"test basic condition", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`name` = ?", Args: []any{"test"}},
		}, want{
			func() *string { s := "`name` = ?"; return &s }(),
			&[]any{"test"},
		}},

		// Adding second condition (index 1)
		{"test adding second condition", args{
			func() *string { s := "`name` = ?"; return &s }(),
			&[]any{"test"},
			1,
			&cond{Conj: "AND", SQL: "`id` = ?", Args: []any{1}},
		}, want{
			func() *string { s := "`name` = ? AND `id` = ?"; return &s }(),
			&[]any{"test", 1},
		}},

		// Multiple conditions in single SQL
		{"test multiple conditions in SQL", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` = ? AND `test` = ?", Args: []any{1, 2}},
		}, want{
			func() *string { s := "`test` = ? AND `test` = ?"; return &s }(),
			&[]any{1, 2},
		}},

		// Different operators
		{"test not equal operator", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` != ?", Args: []any{1}},
		}, want{
			func() *string { s := "`test` != ?"; return &s }(),
			&[]any{1},
		}},

		{"test LIKE operator", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` LIKE ?", Args: []any{"%E%"}},
		}, want{
			func() *string { s := "`test` LIKE ?"; return &s }(),
			&[]any{"%E%"},
		}},

		{"test comparison operators", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` > ?", Args: []any{1}},
		}, want{
			func() *string { s := "`test` > ?"; return &s }(),
			&[]any{1},
		}},

		// IN operator
		{"test IN operator", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` IN (?,?)", Args: []any{1, 2}},
		}, want{
			func() *string { s := "`test` IN (?,?)"; return &s }(),
			&[]any{1, 2},
		}},

		// IN operator with static values
		{"test IN operator with static values", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` IN (1,2,3)", Args: []any{}},
		}, want{
			func() *string { s := "`test` IN (1,2,3)"; return &s }(),
			&[]any{},
		}},

		// BETWEEN operator
		{"test BETWEEN operator", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` BETWEEN ? AND ?", Args: []any{1, 2}},
		}, want{
			func() *string { s := "`test` BETWEEN ? AND ?"; return &s }(),
			&[]any{1, 2},
		}},

		// IS NULL condition
		{"test IS NULL condition", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` IS NULL", Args: []any{}},
		}, want{
			func() *string { s := "`test` IS NULL"; return &s }(),
			&[]any{},
		}},

		// Complex chain of conditions
		{"test complex condition chain", args{
			func() *string { s := "`test1` > ?"; return &s }(),
			&[]any{1},
			1,
			&cond{Conj: "AND", SQL: "`test2` = ?", Args: []any{2}},
		}, want{
			func() *string { s := "`test1` > ? AND `test2` = ?"; return &s }(),
			&[]any{1, 2},
		}},

		// Testing conditions with OR conjunction
		{"test OR conjunction", args{
			func() *string { s := "`test` = ?"; return &s }(),
			&[]any{1},
			1,
			&cond{Conj: "OR", SQL: "`test2` = ?", Args: []any{2}},
		}, want{
			func() *string { s := "`test` = ? OR `test2` = ?"; return &s }(),
			&[]any{1, 2},
		}},

		// Testing with grouped conditions
		{"test grouped conditions", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "(`test` = ? AND `test` = ?)", Args: []any{1, 2}},
		}, want{
			func() *string { s := "(`test` = ? AND `test` = ?)"; return &s }(),
			&[]any{1, 2},
		}},

		// Complex scenario with multiple conditions
		{"test complex multi-condition scenario", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test1` = ?", Args: []any{1}},
		}, want{
			func() *string { s := "`test1` = ?"; return &s }(),
			&[]any{1},
		}},

		// Three chained conditions testing
		{"test three chained conditions", args{
			func() *string { s := "`test1` > ? AND `test2` = ?"; return &s }(),
			&[]any{1, 2},
			2,
			&cond{Conj: "AND", SQL: "`test3` = ?", Args: []any{3}},
		}, want{
			func() *string { s := "`test1` > ? AND `test2` = ? AND `test3` = ?"; return &s }(),
			&[]any{1, 2, 3},
		}},

		// Testing with LIKE BINARY operator
		{"test LIKE BINARY operator", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` LIKE BINARY ?", Args: []any{"%e%"}},
		}, want{
			func() *string { s := "`test` LIKE BINARY ?"; return &s }(),
			&[]any{"%e%"},
		}},

		// Test with empty conjunction
		{"test empty conjunction", args{
			func() *string { return new(string) }(),
			&[]any{},
			0,
			&cond{Conj: "", SQL: "`test` LIKE BINARY ? AND `test` LIKE BINARY ?", Args: []any{"%e%", "%s%"}},
		}, want{
			func() *string { s := "`test` LIKE BINARY ? AND `test` LIKE BINARY ?"; return &s }(),
			&[]any{"%e%", "%s%"},
		}},

		// Edge case tests
		{"test empty SQL in condition", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "", Args: []any{1}},
		}, want{
			func() *string { s := ""; return &s }(),
			&[]any{1},
		}},

		{"test empty conjunction", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			1,
			&cond{Conj: "", SQL: "`test` = ?", Args: []any{1}},
		}, want{
			func() *string { s := "  `test` = ?"; return &s }(),
			&[]any{1},
		}},

		{"test empty initial filterSql with index > 0", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			1,
			&cond{Conj: "AND", SQL: "`test` = ?", Args: []any{1}},
		}, want{
			func() *string { s := " AND `test` = ?"; return &s }(),
			&[]any{1},
		}},

		{"test condition with nil Args", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` IS NULL", Args: nil},
		}, want{
			func() *string { s := "`test` IS NULL"; return &s }(),
			&[]any{},
		}},

		{"test empty Args in condition", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			0,
			&cond{Conj: "AND", SQL: "`test` = 1", Args: []any{}},
		}, want{
			func() *string { s := "`test` = 1"; return &s }(),
			&[]any{},
		}},

		{"test adding to existing filterArgs", args{
			func() *string { s := "`field1` = ?"; return &s }(),
			&[]any{"value1"},
			1,
			&cond{Conj: "AND", SQL: "`field2` = ?", Args: []any{"value2"}},
		}, want{
			func() *string { s := "`field1` = ? AND `field2` = ?"; return &s }(),
			&[]any{"value1", "value2"},
		}},

		{"test multiple empty strings", args{
			func() *string { s := ""; return &s }(),
			&[]any{},
			0,
			&cond{Conj: "", SQL: "", Args: []any{}},
		}, want{
			func() *string { s := ""; return &s }(),
			&[]any{},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joinSQL(tt.args.filterSql, tt.args.filterArgs, tt.args.index, tt.args.condition)

			// Check SQL string
			if tt.args.filterSql != nil && tt.want.filterSql != nil {
				if *tt.args.filterSql != *tt.want.filterSql {
					t.Errorf("joinSQL() filterSql error\nGot : %v\nWant: %v", *tt.args.filterSql, *tt.want.filterSql)
				}
			} else if (tt.args.filterSql == nil) != (tt.want.filterSql == nil) {
				t.Errorf("joinSQL() filterSql nil status mismatch\nGot : %v\nWant: %v", tt.args.filterSql, tt.want.filterSql)
			}

			// Check args
			if tt.args.filterArgs != nil && tt.want.filterArgs != nil {
				if !reflect.DeepEqual(*tt.args.filterArgs, *tt.want.filterArgs) {
					t.Errorf("joinSQL() filterArgs\nGot : %v\nWant: %v", *tt.args.filterArgs, *tt.want.filterArgs)
				}
			} else if (tt.args.filterArgs == nil) != (tt.want.filterArgs == nil) {
				t.Errorf("joinSQL() filterArgs nil status mismatch\nGot : %v\nWant: %v", tt.args.filterArgs, tt.want.filterArgs)
			}
		})
	}
}

func Test_wrapWithBackticks(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty_string", args{""}, ""},
		{"unwrapped_string", args{"test"}, "`test`"},
		{"already_wrapped_string", args{"`test`"}, "`test`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapWithBackticks(tt.args.str); got != tt.want {
				t.Errorf("wrapWithBackticks() = %v, want %v", got, tt.want)
			}
		})
	}
}
