package norm

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
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
				t.Errorf("shiftName() failed.\nGot : %v\nWant: %v", got, tt.want)
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
				t.Errorf("rawFieldNames() failed.\nGot : %v\nWant: %v", got, tt.want)
			}
		})
	}
}

func Test_strSlice2Map(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]struct{}
	}{
		{"test0", args{[]string{}}, map[string]struct{}{}},
		{"test1", args{[]string{"a", "b", "c"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"test2", args{[]string{"a", "b", "c", "a"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"test3", args{[]string{"a", "b", "c", "a", "b"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"test4", args{[]string{"a", "b", "c", "a", "b", "c"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := strSlice2Map(tt.args.s); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("strSlice2Map() failed.\nGot : %v\nWant: %v", gotRes, tt.wantRes)
			}
		})
	}
}

func Test_modelStruct2Map(t *testing.T) {
	timeNow := time.Now()
	type args struct {
		obj any
		tag string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{"test tempty", args{struct{}{}, "db"}, map[string]any{}},
		{"test more", args{struct {
			Id              int64           `db:"id"`
			TestInt         int             `db:"test_int"`
			TestInt8        int8            `db:"test_int8"`
			TestInt16       int16           `db:"test_int16"`
			TestInt32       int32           `db:"test_int32"`
			TestInt64       int64           `db:"test_int64"`
			TestUint        uint            `db:"test_uint"`
			TestUint8       uint8           `db:"test_uint8"`
			TestUint16      uint16          `db:"test_uint16"`
			TestUint32      uint32          `db:"test_uint32"`
			TestUint64      uint64          `db:"test_uint64"`
			TestFloat32     float32         `db:"test_float32"`
			TestFloat64     float64         `db:"test_float64"`
			TestString      string          `db:"test_string"`
			TestBool        bool            `db:"test_bool"`
			TestTime        time.Time       `db:"test_time"`
			TestTimePtr     *time.Time      `db:"test_time_ptr"`
			TestNullByte    sql.NullByte    `db:"test_null_byte"`
			TestNullInt16   sql.NullInt16   `db:"test_null_int16"`
			TestNullInt32   sql.NullInt32   `db:"test_null_int32"`
			TestNullInt64   sql.NullInt64   `db:"test_null_int64"`
			TestNullFloat64 sql.NullFloat64 `db:"test_null_float64"`
			TestNullString  sql.NullString  `db:"test_null_string"`
			TestNullBool    sql.NullBool    `db:"test_null_bool"`
			TestNullTime    sql.NullTime    `db:"test_null_time"`
		}{
			Id:              1,
			TestInt:         1,
			TestInt8:        2,
			TestInt16:       3,
			TestInt32:       4,
			TestInt64:       5,
			TestUint:        6,
			TestUint8:       7,
			TestUint16:      8,
			TestUint32:      9,
			TestUint64:      10,
			TestFloat32:     11.0,
			TestFloat64:     12.0,
			TestString:      "test",
			TestBool:        true,
			TestTime:        timeNow,
			TestTimePtr:     &timeNow,
			TestNullByte:    sql.NullByte{Byte: 1, Valid: true},
			TestNullInt16:   sql.NullInt16{Int16: 2, Valid: true},
			TestNullInt32:   sql.NullInt32{Int32: 3, Valid: true},
			TestNullInt64:   sql.NullInt64{Int64: 4, Valid: true},
			TestNullFloat64: sql.NullFloat64{Float64: 5.0, Valid: true},
			TestNullString:  sql.NullString{String: "test", Valid: true},
			TestNullBool:    sql.NullBool{Bool: true, Valid: true},
			TestNullTime:    sql.NullTime{Time: timeNow, Valid: true},
		}, "db"}, map[string]any{
			"id":                int64(1),
			"test_int":          int(1),
			"test_int8":         int8(2),
			"test_int16":        int16(3),
			"test_int32":        int32(4),
			"test_int64":        int64(5),
			"test_uint":         uint(6),
			"test_uint8":        uint8(7),
			"test_uint16":       uint16(8),
			"test_uint32":       uint32(9),
			"test_uint64":       uint64(10),
			"test_float32":      float32(11.0),
			"test_float64":      float64(12.0),
			"test_string":       "test",
			"test_bool":         true,
			"test_time":         timeNow,
			"test_time_ptr":     &timeNow,
			"test_null_byte":    byte(1),
			"test_null_int16":   int16(2),
			"test_null_int32":   int32(3),
			"test_null_int64":   int64(4),
			"test_null_float64": float64(5.0),
			"test_null_string":  "test",
			"test_null_bool":    true,
			"test_null_time":    timeNow,
		}},
		{"test valid false", args{struct {
			TestNullByte    sql.NullByte    `db:"test_null_byte"`
			TestNullInt16   sql.NullInt16   `db:"test_null_int16"`
			TestNullInt32   sql.NullInt32   `db:"test_null_int32"`
			TestNullInt64   sql.NullInt64   `db:"test_null_int64"`
			TestNullFloat64 sql.NullFloat64 `db:"test_null_float64"`
			TestNullString  sql.NullString  `db:"test_null_string"`
			TestNullBool    sql.NullBool    `db:"test_null_bool"`
			TestNullTime    sql.NullTime    `db:"test_null_time"`
		}{}, "db"}, map[string]any{
			"test_null_byte":    nil,
			"test_null_int16":   nil,
			"test_null_int32":   nil,
			"test_null_int64":   nil,
			"test_null_float64": nil,
			"test_null_string":  nil,
			"test_null_bool":    nil,
			"test_null_time":    nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modelStruct2Map(tt.args.obj, tt.args.tag)

			// Check map length first
			if len(got) != len(tt.want) {
				t.Errorf("modelStruct2Map() failed: map length mismatch. \nGot : %d \nwant: %d",
					len(got), len(tt.want))
				return
			}

			// Check non-time fields
			for k, wantVal := range tt.want {
				gotVal, exists := got[k]
				if !exists {
					t.Errorf("modelStruct2Map() failed: key %q missing in result", k)
					continue
				}

				// Skip time.Time fields in this loop - we'll check them separately
				if _, isTime := wantVal.(time.Time); isTime {
					continue
				}

				if !reflect.DeepEqual(gotVal, wantVal) {
					t.Errorf("modelStruct2Map() failed for key %q: \nGot : %v ( type: %s ) \nWant: %v ( type: %s )",
						k, gotVal, reflect.TypeOf(gotVal).String(), wantVal, reflect.TypeOf(wantVal).String())
				}
			}

			// Check time fields separately
			for k, wantVal := range tt.want {
				if wantTime, isTime := wantVal.(time.Time); isTime {
					gotVal, exists := got[k]
					if !exists {
						t.Errorf("modelStruct2Map() failed: time key %q missing in result", k)
						continue
					}

					gotTime, ok := gotVal.(time.Time)
					if !ok {
						t.Errorf("modelStruct2Map() failed: value for key %q is not a time.Time", k)
						continue
					}

					// Compare times using Equal
					if !wantTime.Equal(gotTime) {
						t.Errorf("modelStruct2Map() failed for time field %q\nGot : %v \nWant: %v",
							k, gotTime, wantTime)
					}
				}
			}
		})
	}
}

func Test_modelStructSlice2MapSlice(t *testing.T) {
	type args struct {
		obj any
		tag string
	}
	tests := []struct {
		name string
		args args
		want []map[string]any
	}{
		{"test empty", args{[]struct{}{}, "db"}, []map[string]any{}},
		{"test one", args{[]struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{{Id: 1, Name: "test"}}, "db"}, []map[string]any{
			{"id": int64(1), "name": "test"},
		}},
		{"test two", args{[]struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{{Id: 1, Name: "test"}, {Id: 2, Name: "test2"}}, "db"}, []map[string]any{
			{"id": int64(1), "name": "test"},
			{"id": int64(2), "name": "test2"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modelStructSlice2MapSlice(tt.args.obj, tt.args.tag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelStructSlice2MapSlice() failed \nGot : %v\nWant: %v", got, tt.want)
			}
		})
	}
}

func Test_createModelPointerAndSlice(t *testing.T) {
	type args struct {
		input any
	}
	tests := []struct {
		name  string
		args  args
		want  any
		want1 any
	}{
		{"test1", args{struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{
			Id:   1,
			Name: "test",
		}}, &struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{
			Id:   1,
			Name: "test",
		}, &[]struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{
			{
				Id:   1,
				Name: "test",
			},
		}},
		{"test2", args{[]string{"1", "2", "3"}}, &[]string{"1", "2", "3"}, &[][]string{{"1", "2", "3"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := createModelPointerAndSlice(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("create model pointer slice error \nGot : %v\nWant: %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("create model pointer error \nGot1: %v\nWant: %v", got1, tt.want1)
			}
		})
	}
}

func Test_isStrList(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
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
		name string
		args args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genStrListValueLikeSQL(tt.args.p, tt.args.filterConditions, tt.args.fieldName, tt.args.valueOf, tt.args.notFlag, tt.args.operator, tt.args.valueFormat)
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

func Test_deepCopy(t *testing.T) {
	type args struct {
		src any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		changed any
	}{
		{"test pointer to struct", args{&struct{ A int }{A: 5}}, &struct{ A int }{}, &struct{ A int }{A: 10}},
		{"test pointer to slice struct", args{&[]struct{ A int }{{A: 1}, {A: 2}}}, &[]struct{ A int }{}, &[]struct{ B int }{{B: 3}, {B: 4}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepCopy(tt.args.src)

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("deepCopy() type error\nGet : %T\nWant: %T", got, tt.want)
			}

			tt.args.src = tt.changed

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("deepCopy() changed type error\nGet : %T\nWant: %T", got, tt.want)
			}
			if reflect.DeepEqual(got, tt.changed) {
				t.Errorf("deepCopy() changed value error\nGot : %v\nWant: %v", got, tt.changed)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wrapWithBackticks(tt.args.str); got != tt.want {
				t.Errorf("wrapWithBackticks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processSQL(t *testing.T) {
	type args struct {
		sqlParts  []string
		isKeyWord func(word string) bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processSQL(tt.args.sqlParts, tt.args.isKeyWord); got != tt.want {
				t.Errorf("processSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
