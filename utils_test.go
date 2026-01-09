package norm

import (
	"database/sql"
	"reflect"
	"strings"
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
		{"camel_to_snake_compound", args{"DevicePolicyMap"}, "`device_policy_map`"},
		{"camel_to_snake_two_words", args{"DevicePolicy"}, "`device_policy`"},
		{"camel_to_snake_single", args{"Device"}, "`device`"},
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
		{"with_postgresql_false", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", false}, []string{"`device`", "`device_policy`", "`device_policy_map`"}},
		{"with_postgresql_true", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", true}, []string{"device", "device_policy", "device_policy_map"}},
		{"ignore_with_pg_false", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", false}, []string{"`device`", "`device_policy`"}},
		{"ignore_with_pg_true", args{struct {
			Device          string `db:"device"`
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", true}, []string{"device", "device_policy"}},
		{"multiple_tag_with_pg_false", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", false}, []string{"`device`", "`device_policy`", "`device_policy_map`"}},
		{"multiple_tag_with_pg_true", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"device_policy_map"`
		}{}, "db", true}, []string{"device", "device_policy", "device_policy_map"}},
		{"multiple_tag_pg_false_ignore", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", false}, []string{"`device`", "`device_policy`"}},
		{"multiple_tag_pg_true_ignore", args{struct {
			Device          string `db:"device, type=char, length=16"`
			DevicePolicy    string `db:"device_policy, type=char"`
			DevicePolicyMap string `db:"-"`
		}{}, "db", true}, []string{"device", "device_policy"}},
		{"empty_tag_pg_false", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:",type=char"`
		}{}, "db", false}, []string{"`Device`", "`device_policy`", "`DevicePolicyMap`"}},
		{"test with empty struct with not pg", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:",type=char"`
		}{}, "db", true}, []string{"Device", "device_policy", "DevicePolicyMap"}},
		{"empty_tag_pg_false_ignore", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-,type=char"`
		}{}, "db", false}, []string{"`Device`", "`device_policy`"}},
		{"empty_tag_pg_true_ignore", args{struct {
			Device          string
			DevicePolicy    string `db:"device_policy"`
			DevicePolicyMap string `db:"-,type=char"`
		}{}, "db", true}, []string{"Device", "device_policy"}},
		{"pointer_struct", args{&struct {
			Device string `db:"device"`
		}{}, "db", false}, []string{"`device`"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rawFieldNames(tt.args.in, tt.args.tag, tt.args.pg)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawFieldNames() failed.\nGot : %+v\nWant: %+v", got, tt.want)
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
		{"empty_slice", args{[]string{}}, map[string]struct{}{}},
		{"distinct_strings", args{[]string{"a", "b", "c"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"duplicate_one_string", args{[]string{"a", "b", "c", "a"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"duplicate_two_strings", args{[]string{"a", "b", "c", "a", "b"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{"duplicate_all_strings", args{[]string{"a", "b", "c", "a", "b", "c"}}, map[string]struct{}{"a": {}, "b": {}, "c": {}}},
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
			TestEmptyTag    string
			TestEmptyTag2   string `db:""`
			TestIgnoreTag   string `db:"-"`
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
			TestEmptyTag:    "test1",
			TestEmptyTag2:   "test2",
			TestIgnoreTag:   "test3",
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
		name           string
		args           args
		want           any
		want1          any
		expectPanic    bool
		expectedErrMsg string
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
		}{}, &[]struct {
			Id   int64  `db:"id"`
			Name string `db:"name"`
		}{}, false, ""},
		{"test_with_nil", args{nil}, nil, nil, true, "model is nil"},
		{"test_with_int", args{1}, nil, nil, true, "model only can be a struct; got int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				got      any
				got1     any
				panicked = false
				panicMsg any
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						panicMsg = r
					}
				}()

				got, got1 = createModelPointerAndSlice(tt.args.input)
			}()

			if tt.expectPanic != panicked {
				t.Errorf("rawFieldNames() failed.\nGot : %+v\nWant: %+v", panicked, tt.expectPanic)
				return
			}

			if tt.expectPanic && panicked {
				errMsg, ok := panicMsg.(error)
				if !ok || !strings.Contains(errMsg.Error(), tt.expectedErrMsg) {
					t.Errorf("panic message mismatch.\nGot : %+v\nWant: %+v", errMsg, tt.expectedErrMsg)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("create model pointer error \nGot : %v\nWant: %v", got, tt.want)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("create model pointer slice error \nGot : %+v : %+v\nWant: %+v : %+v", got, reflect.TypeOf(got), tt.want, reflect.TypeOf(tt.want))
			}
			if reflect.TypeOf(got1) != reflect.TypeOf(tt.want1) {
				t.Errorf("create model pointer slice error \nGot : %+v : %+v\nWant: %+v : %+v", got1, reflect.TypeOf(got1), tt.want1, reflect.TypeOf(tt.want1))
			}

		})
	}
}

func Test_deepCopyModelPtrStructure(t *testing.T) {
	type args struct {
		src any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		changed any
	}{
		// this test is not important
		{"test nil", args{nil}, nil, 1},
		// the following two tests are the only ones about which we care
		{"test pointer to struct", args{&struct{ A int }{}}, &struct{ A int }{}, &struct{ B int }{}},
		{"test pointer to slice struct", args{&[]struct{ A int }{}}, &[]struct{ A int }{}, &[]struct{ B int }{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepCopyModelPtrStructure(tt.args.src)

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("deepCopy() type error\nGet : %T\nWant: %T", got, tt.want)
			}

			tt.args.src = tt.changed

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("deepCopy() changed type error\nGet : %T\nWant: %T", got, tt.want)
			}
		})
	}
}

// Define the Model type and its method at package level
type TestModel struct {
	Name string
}

type TestModelWithTableName struct {
	Name string
}

func (m TestModelWithTableName) TableName() string {
	return "custom_table_name"
}

type TestModelWithTableNamePtr struct {
	Name string
}

func (m *TestModelWithTableNamePtr) TableName() string {
	return "custom_table_name_ptr"
}

func Test_getTableName(t *testing.T) {
	type args struct {
		m any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Do not check whether the model is nil, because it will be checked in the createModelPointerAndSlice
		{"test model without TableName method", args{TestModel{}}, "`test_model`"},
		{"test model with TableName method", args{TestModelWithTableName{}}, "`custom_table_name`"},
		{"test model with TableName method pointer", args{&TestModelWithTableNamePtr{}}, "`custom_table_name_ptr`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTableName(tt.args.m); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}
