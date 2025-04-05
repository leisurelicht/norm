package norm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	mysqlOp "github.com/leisurelicht/norm/operator/mysql"
)

func TestFilter(t *testing.T) {
	type args struct {
		isNot  int
		filter []any
	}
	type want struct {
		sql  string
		args []any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"default_cond", args{0, []any{Cond{}}}, want{"", []any{}}},
		{"default_cond", args{0, []any{AND{}}}, want{"", []any{}}},
		{"default_cond", args{0, []any{OR{}}}, want{"", []any{}}},

		{"default_cond", args{0, []any{Cond{"test": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"default_list_cond", args{0, []any{Cond{"test": []any{1, 2}}}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"exact_cond", args{0, []any{Cond{"test__exact": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"exact_list_cond", args{0, []any{Cond{"test__exact": []any{1, 2}}}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"exclude_cond", args{0, []any{Cond{"test__exclude": 1}}}, want{" WHERE (`test` != ?)", []any{1}}},
		{"exclude_list_cond", args{0, []any{Cond{"test__exclude": []any{1, 2}}}}, want{" WHERE (`test` != ? AND `test` != ?)", []any{1, 2}}},
		{"iexact_cond", args{0, []any{Cond{"test__iexact": 1}}}, want{" WHERE (`test` LIKE ?)", []any{1}}},
		{"iexact_list_cond", args{0, []any{Cond{"test__iexact": []any{1, 2}}}}, want{" WHERE (`test` LIKE ? AND `test` LIKE ?)", []any{1, 2}}},
		{"gt_cond", args{0, []any{Cond{"test__gt": 1}}}, want{" WHERE (`test` > ?)", []any{1}}},
		{"gte_cond", args{0, []any{Cond{"test__gte": 1}}}, want{" WHERE (`test` >= ?)", []any{1}}},
		{"lt_cond", args{0, []any{Cond{"test__lt": 1}}}, want{" WHERE (`test` < ?)", []any{1}}},
		{"lte_cond", args{0, []any{Cond{"test__lte": 1}}}, want{" WHERE (`test` <= ?)", []any{1}}},
		{"len_cond", args{0, []any{Cond{"test__len": 1}}}, want{" WHERE (LENGTH(`test`) = ?)", []any{1}}},
		{"in_string_cond", args{0, []any{Cond{"test__in": "1,2,3"}}}, want{" WHERE (`test` IN (1,2,3))", []any{}}},
		{"in_list_cond", args{0, []any{Cond{"test__in": []int{1, 2}}}}, want{" WHERE (`test` IN (?,?))", []any{1, 2}}},
		{"not_in_string_cond", args{0, []any{Cond{"test__not_in": "1,2,3"}}}, want{" WHERE (`test` NOT IN (1,2,3))", []any{}}},
		{"not_list_in_cond", args{0, []any{Cond{"test__not_in": []int{1, 2}}}}, want{" WHERE (`test` NOT IN (?,?))", []any{1, 2}}},
		{"between_cond", args{0, []any{Cond{"test__between": []int{1, 2}}}}, want{" WHERE (`test` BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_between_cond", args{0, []any{Cond{"test__not_between": []int{1, 2}}}}, want{" WHERE (`test` NOT BETWEEN ? AND ?)", []any{1, 2}}},
		{"contains_cond", args{0, []any{Cond{"test__contains": "e"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"%e%"}}},
		{"list_contains cond", args{0, []any{Cond{"test__contains": []string{"e", "s"}}}}, want{" WHERE (`test` LIKE BINARY ? AND `test` LIKE BINARY ?)", []any{"%e%", "%s%"}}},
		{"not_contains cond", args{0, []any{Cond{"test__not_contains": "e"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"%e%"}}},
		{"icontains_cond", args{0, []any{Cond{"test__icontains": "E"}}}, want{" WHERE (`test` LIKE ?)", []any{"%E%"}}},
		{"not_icontains cond", args{0, []any{Cond{"test__not_icontains": "E"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"%E%"}}},
		{"startswith_cond", args{0, []any{Cond{"test__startswith": "te"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"te%"}}},
		{"not_startswith cond", args{0, []any{Cond{"test__not_startswith": "te"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"te%"}}},
		{"istartswith_cond", args{0, []any{Cond{"test__istartswith": "tE"}}}, want{" WHERE (`test` LIKE ?)", []any{"tE%"}}},
		{"not_istartswith cond", args{0, []any{Cond{"test__not_istartswith": "tE"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"tE%"}}},
		{"endswith_cond", args{0, []any{Cond{"test__endswith": "st"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"%st"}}},
		{"not_endswith_cond", args{0, []any{Cond{"test__not_endswith": "st"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"%st"}}},
		{"iendswith_cond", args{0, []any{Cond{"test__iendswith": "sT"}}}, want{" WHERE (`test` LIKE ?)", []any{"%sT"}}},
		{"not_iendswith_cond", args{0, []any{Cond{"test__not_iendswith": "sT"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"%sT"}}},

		{"not_default_cond", args{1, []any{Cond{"test": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
		{"not_exact_cond", args{1, []any{Cond{"test__exact": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
		{"not_exclude_cond", args{1, []any{Cond{"test__exclude": 1}}}, want{" WHERE NOT (`test` != ?)", []any{1}}},
		{"not_iexact_cond", args{1, []any{Cond{"test__iexact": 1}}}, want{" WHERE NOT (`test` LIKE ?)", []any{1}}},
		{"not_gt_cond", args{1, []any{Cond{"test__gt": 1}}}, want{" WHERE NOT (`test` > ?)", []any{1}}},
		{"not_gte_cond", args{1, []any{Cond{"test__gte": 1}}}, want{" WHERE NOT (`test` >= ?)", []any{1}}},
		{"not_lt_cond", args{1, []any{Cond{"test__lt": 1}}}, want{" WHERE NOT (`test` < ?)", []any{1}}},
		{"not_lte_cond", args{1, []any{Cond{"test__lte": 1}}}, want{" WHERE NOT (`test` <= ?)", []any{1}}},
		{"not_len_cond", args{1, []any{Cond{"test__len": 1}}}, want{" WHERE NOT (LENGTH(`test`) = ?)", []any{1}}},
		{"not_in_cond", args{1, []any{Cond{"test__in": []int{1, 2}}}}, want{" WHERE NOT (`test` IN (?,?))", []any{1, 2}}},
		{"not_not_in_cond", args{1, []any{Cond{"test__not_in": []int{1, 2}}}}, want{" WHERE NOT (`test` NOT IN (?,?))", []any{1, 2}}},
		{"not_between_cond", args{1, []any{Cond{"test__between": []int{1, 2}}}}, want{" WHERE NOT (`test` BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_not_between_cond", args{1, []any{Cond{"test__not_between": []int{1, 2}}}}, want{" WHERE NOT (`test` NOT BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_contains_cond", args{1, []any{Cond{"test__contains": "e"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"%e%"}}},
		{"not_not_contains_cond", args{1, []any{Cond{"test__not_contains": "e"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"%e%"}}},
		{"not_icontains_cond", args{1, []any{Cond{"test__icontains": "E"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"%E%"}}},
		{"not_not_icontains_cond", args{1, []any{Cond{"test__not_icontains": "E"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"%E%"}}},
		{"not_startswith_cond", args{1, []any{Cond{"test__startswith": "te"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"te%"}}},
		{"not_not_startswith_cond", args{1, []any{Cond{"test__not_startswith": "te"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"te%"}}},
		{"not_istartswith_cond", args{1, []any{Cond{"test__istartswith": "tE"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"tE%"}}},
		{"not_not_istartswith_cond", args{1, []any{Cond{"test__not_istartswith": "tE"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"tE%"}}},
		{"not_endswith_cond", args{1, []any{Cond{"test__endswith": "st"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"%st"}}},
		{"not_not_endswith_cond", args{1, []any{Cond{"test__not_endswith": "st"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"%st"}}},
		{"not_not_iendswith_cond", args{1, []any{Cond{"test__iendswith": "sT"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"%sT"}}},
		{"not_not_iendswith_cond", args{1, []any{Cond{"test__not_iendswith": "sT"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"%sT"}}},

		{"two_default_column", args{0, []any{Cond{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2}}}, want{" WHERE (`test1` = ? AND `test2` = ?)", []any{1, 2}}},
		{"reverse_default_column", args{0, []any{Cond{SortKey: []string{"test2", "test1"}, "test1": 1, "test2": 2}}}, want{" WHERE (`test2` = ? AND `test1` = ?)", []any{2, 1}}},
		{"three_default_column", args{0, []any{Cond{SortKey: []string{"test1", "test2", "test3"}, "test1__gt": 1, "test2": 2, "test3": 3}}}, want{" WHERE (`test1` > ? AND `test2` = ? AND `test3` = ?)", []any{1, 2, 3}}},
		{"reverse_three_default_column", args{0, []any{Cond{SortKey: []string{"test3", "test2", "test1"}, "test1__lt": 1, "test2": 2, "test3": 3}}}, want{" WHERE (`test3` = ? AND `test2` = ? AND `test1` < ?)", []any{3, 2, 1}}},
		{"out_order_three_default_column", args{0, []any{Cond{SortKey: []string{"test1", "test3", "test2"}, "test3__in": []int{1, 4, 6}, "test2": 2, "test1": 3}}}, want{" WHERE (`test1` = ? AND `test3` IN (?,?,?) AND `test2` = ?)", []any{3, 1, 4, 6, 2}}},

		{"default_conj", args{0, []any{Cond{"test": 1}, Cond{"test2": 2}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?))", []any{1, 2}}},
		{"and_conj", args{0, []any{Cond{"test": 1}, AND{"test2": 2}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?))", []any{1, 2}}},
		{"or_conj", args{0, []any{Cond{"test": 1}, OR{"test2": 2}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?))", []any{1, 2}}},
		{"3_default_conj", args{0, []any{Cond{"test": 1}, Cond{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3_and_conj", args{0, []any{AND{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3_or_conj", args{0, []any{OR{"test": 1}, OR{"test2": 2}, OR{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) OR (`test3` = ?))", []any{1, 2, "3"}}},
		{"d_and_d_conj", args{0, []any{Cond{"test": 1}, AND{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&2and_conj", args{0, []any{Cond{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d_or_d_conj", args{0, []any{Cond{"test": 1}, OR{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&2or_conj", args{0, []any{Cond{"test": 1}, OR{"test2": 2}, OR{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) OR (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&or&and_conj", args{0, []any{Cond{"test": 1}, OR{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3and_conj", args{0, []any{AND{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},

		{"default_cond", args{0, []any{Cond{"test": []any{1, 2}}, AND{"test2": 3}, OR{"test3": []any{4, 5}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ?) OR (`test3` = ? AND `test3` = ?))", []any{1, 2, 3, 4, 5}}},
		{"exact_cond", args{0, []any{Cond{"test__exact": []any{1, 2}}, AND{"test2__exact": 3}, OR{"test3__exact": []any{4, 5}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ?) OR (`test3` = ? AND `test3` = ?))", []any{1, 2, 3, 4, 5}}},
		{"exclude_cond", args{0, []any{Cond{"test__exclude": []any{1, 2}}, AND{"test2__exclude": 3}, OR{"test3__exclude": []any{4, 5}}}}, want{" WHERE ((`test` != ? AND `test` != ?) AND (`test2` != ?) OR (`test3` != ? AND `test3` != ?))", []any{1, 2, 3, 4, 5}}},
		{"iexact_cond", args{0, []any{Cond{"test__iexact": []any{1, 2}}, AND{"test2__iexact": 3}, OR{"test3__iexact": []any{4, 5}}}}, want{" WHERE ((`test` LIKE ? AND `test` LIKE ?) AND (`test2` LIKE ?) OR (`test3` LIKE ? AND `test3` LIKE ?))", []any{1, 2, 3, 4, 5}}},
		{"gt_cond", args{0, []any{Cond{"test__gt": 1}, AND{"test2__gt": 2}, OR{"test3__gt": 3}}}, want{" WHERE ((`test` > ?) AND (`test2` > ?) OR (`test3` > ?))", []any{1, 2, 3}}},
		{"in_cond", args{0, []any{Cond{"test__in": "1,2,3"}, AND{"test2__in": []any{4, 5}}, OR{"test3__in": []any{6, 7}}}}, want{" WHERE ((`test` IN (1,2,3)) AND (`test2` IN (?,?)) OR (`test3` IN (?,?)))", []any{4, 5, 6, 7}}},

		{"default_mix_contains_conj", args{0, []any{Cond{"test": 1}, Cond{"test2__contains": []string{"e", "s"}}}}, want{" WHERE ((`test` = ?) AND (`test2` LIKE BINARY ? AND `test2` LIKE BINARY ?))", []any{1, "%e%", "%s%"}}},

		{"exact_one_and_one_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, "test2": 2}}}, want{" WHERE (`test` = ? AND `test2` = ?)", []any{1, 2}}},
		{"exact_one_and_list_and_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, "test2": []any{3, 4}}}}, want{" WHERE (`test` = ? AND (`test2` = ? AND `test2` = ?))", []any{1, 3, 4}}},
		{"exact_list_and_list_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": []any{1, 2}, "test2": []any{3, 4}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ? AND `test2` = ?))", []any{1, 2, 3, 4}}},

		// ToOR
		{"exact_one_or_one_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, ToOR("test2"): 2}}}, want{" WHERE (`test` = ? OR `test2` = ?)", []any{1, 2}}},
		{"exact_one_or_list_and_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, ToOR("test2"): []any{3, 4}}}}, want{" WHERE (`test` = ? OR (`test2` = ? AND `test2` = ?))", []any{1, 3, 4}}},
		{"exact_list_or_list_cond", args{0, []any{Cond{SortKey: []string{"test", "test2"}, "test": []any{1, 2}, ToOR("test2"): []any{3, 4}}}}, want{" WHERE ((`test` = ? AND `test` = ?) OR (`test2` = ? AND `test2` = ?))", []any{1, 2, 3, 4}}},

		// EachOR
		{"each_or", args{0, []any{EachOR(Cond{"test": 1})}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"each_or_list", args{0, []any{EachOR(Cond{"test": []any{1, 2}})}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"each_or_and", args{0, []any{Cond{"test": 1}, EachOR(AND{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2})}}, want{" WHERE ((`test` = ?) AND (`test1` = ? OR `test2` = ?))", []any{1, 1, 2}}},
		{"each_or_or", args{0, []any{Cond{"test": 1}, EachOR(OR{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2})}}, want{" WHERE ((`test` = ?) OR (`test1` = ? OR `test2` = ?))", []any{1, 1, 2}}},

		// meet by accident
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p = p.FilterToSQL(tt.args.isNot, tt.args.filter...)
			sql, sqlArgs := p.GetQuerySet()

			if p.Error() != nil {
				t.Errorf("TestFilter SQL Occur Error -> error:%+v", p.Error())
			}

			wantSQL := tt.want.sql
			if sql != wantSQL {
				t.Errorf("TestFilter SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestFilter SQL Gen Error -> want:%v", wantSQL)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestFilter Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestFilter Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestFilter Args Length Error -> want:%+v", tt.want.args)
			}

			for i, a := range sqlArgs {
				if !reflect.DeepEqual(a, tt.want.args[i]) {
					t.Errorf("TestFilter Arg Value Error -> args:%+v", sqlArgs)
					t.Errorf("TestFilter Arg Value Error -> want:%+v", tt.want.args)
					break
				}
			}
		})
	}
}

func TestMultipleCallFilter(t *testing.T) {
	type args struct {
		isNot  int
		filter []any
	}
	type want struct {
		sql  string
		args []any
	}
	tests := []struct {
		name string
		args []args
		want want
	}{
		// multiple call
		{"double call", []args{{0, []any{Cond{"test1": 1}}}, {0, []any{Cond{"test2": 1}}}}, want{" WHERE (`test1` = ?) AND (`test2` = ?)", []any{1, 1}}},

		// meet by accident
		{"meet1", []args{{0, []any{Cond{SortKey: []string{"delete_flag", "devise_sn"}, "delete_flag": 0, "devise_sn__len": 22}}}, {0, []any{Cond{SortKey: []string{"device_name", "devise_sn", "belong_to_company"}, "device_name__icontains": "test", "devise_sn__icontains": "test", "belong_to_company__icontains": "test"}}}}, want{" WHERE (`delete_flag` = ? AND LENGTH(`devise_sn`) = ?) AND (`device_name` LIKE ? AND `devise_sn` LIKE ? AND `belong_to_company` LIKE ?)", []any{0, 22, "%test%", "%test%", "%test%"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			for _, f := range tt.args {
				p = p.FilterToSQL(f.isNot, f.filter...)
			}
			sql, sqlArgs := p.GetQuerySet()

			if p.Error() != nil {
				t.Errorf("TestFilter SQL Occur Error -> error:%+v", p.Error())
			}

			wantSQL := tt.want.sql
			if sql != wantSQL {
				t.Errorf("TestFilter SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestFilter SQL Gen Error -> want:%v", wantSQL)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestFilter Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestFilter Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestFilter Args Length Error -> want:%+v", tt.want.args)
			}

			for i, a := range sqlArgs {
				if !reflect.DeepEqual(a, tt.want.args[i]) {
					t.Errorf("TestFilter Arg Value Error -> args:%+v", sqlArgs)
					t.Errorf("TestFilter Arg Value Error -> want:%+v", tt.want.args)
					break
				}
			}
		})
	}
}

func TestFilterError(t *testing.T) {
	type args struct {
		isNot  int
		filter []any
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"empty", args{0, []any{}}, want{nil}},
		{"order_key_type", args{0, []any{Cond{SortKey: []int{1, 2}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyTypeError)}},
		{"order_key_type", args{0, []any{Cond{SortKey: []string{"1"}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyLenError)}},
		{"order_key_type", args{0, []any{Cond{"1__not__contains": "b"}}}, want{fmt.Errorf(fieldLookupError, "1__not__contains")}},
		{"order_key_type", args{0, []any{Cond{"1__contain": "b"}}}, want{fmt.Errorf(unknownOperatorError, "contain")}},
		{"order_key_type", args{0, []any{Cond{"test": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exact", 0)}},
		{"order_key_type", args{0, []any{Cond{"test": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exact", "map")}},
		{"order_key_type", args{0, []any{Cond{"test__exclude": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exclude", 0)}},
		{"order_key_type", args{0, []any{Cond{"test__exclude": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exclude", "map")}},
		{"order_key_type", args{0, []any{Cond{"test__iexact": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "iexact", 0)}},
		{"order_key_type", args{0, []any{Cond{"test__iexact": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "iexact", "map")}},
		{"order_key_type", args{0, []any{Cond{"test__gt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gt", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__gte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gte", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__lt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lt", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__lte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lte", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__len": "10"}}}, want{fmt.Errorf(unsupportedValueError, "len", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__gt": true}}}, want{fmt.Errorf(unsupportedValueError, "gt", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__gte": true}}}, want{fmt.Errorf(unsupportedValueError, "gte", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__lt": true}}}, want{fmt.Errorf(unsupportedValueError, "lt", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__lte": true}}}, want{fmt.Errorf(unsupportedValueError, "lte", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__len": true}}}, want{fmt.Errorf(unsupportedValueError, "len", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__gt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "slice")}},
		{"order_key_type", args{0, []any{Cond{"test__gte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "slice")}},
		{"order_key_type", args{0, []any{Cond{"test__lt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "slice")}},
		{"order_key_type", args{0, []any{Cond{"test__lte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "slice")}},
		{"order_key_type", args{0, []any{Cond{"test__len": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "slice")}},
		{"order_key_type", args{0, []any{Cond{"test__gt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "array")}},
		{"order_key_type", args{0, []any{Cond{"test__gte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "array")}},
		{"order_key_type", args{0, []any{Cond{"test__lt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "array")}},
		{"order_key_type", args{0, []any{Cond{"test__lte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "array")}},
		{"order_key_type", args{0, []any{Cond{"test__len": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "array")}},
		{"order_key_type", args{0, []any{Cond{"test__in": 1}}}, want{fmt.Errorf(unsupportedValueError, "in", "int")}},
		{"order_key_type", args{0, []any{Cond{"test__in": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "in", 0)}},
		{"order_key_type", args{0, []any{Cond{"test__between": "test"}}}, want{fmt.Errorf(unsupportedValueError, "between", "string")}},
		{"order_key_type", args{0, []any{Cond{"test__between": []int{}}}}, want{fmt.Errorf(operatorValueLenError, "between", 2)}},
		{"order_key_type", args{0, []any{Cond{"test__contains": ""}}}, want{fmt.Errorf(unsupportedValueError, "contains", "blank string")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"order_key_type", args{0, []any{Cond{"test__contains": [0]int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"order_key_type", args{0, []any{Cond{"test__contains": []int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": [2]int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": true}}}, want{fmt.Errorf(unsupportedValueError, "contains", "bool")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": 0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": 1}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"order_key_type", args{0, []any{Cond{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "float64")}},
		{"order_key_type", args{0, []any{Cond{"test__unimplemented": 1}}}, want{fmt.Errorf(notImplementedOperatorError, "UNIMPLEMENTED")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p = p.FilterToSQL(tt.args.isNot, tt.args.filter...)
			p.GetQuerySet()

			if p.Error() == nil && !errors.Is(p.Error(), tt.want.err) {
				t.Errorf("TestFilterError not works as expected.")
				t.Errorf("want: %s", tt.want.err)
				t.Errorf("get : %s", p.Error())
			}

			if p.Error() != nil && p.Error().Error() != tt.want.err.Error() {
				t.Errorf("TestFilterError not works as expected.")
				t.Errorf("want: %s", tt.want.err)
				t.Errorf("get : %s", p.Error())
			}
		})
	}

}

func TestWhere(t *testing.T) {
	type filterArgs struct {
		isNot  int
		filter []any
	}
	type args struct {
		cond    string
		args    []any
		filters *filterArgs
	}
	type want struct {
		sql  string
		args []any
		err  error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"one", args{"`test` = ?", []any{1}, nil}, want{"`test` = ?", []any{1}, nil}},
		{"two", args{"`test` = ? AND test2 = ?", []any{1, 2}, nil}, want{"`test` = ? AND test2 = ?", []any{1, 2}, nil}},
		{"three", args{"test = ? AND `test2` = ? AND test3 = ?", []any{1, 2, 3}, nil}, want{"test = ? AND `test2` = ? AND test3 = ?", []any{1, 2, 3}, nil}},
		{"one_with_filter", args{"`test` = ?", []any{1}, &filterArgs{0, []any{Cond{"test": 1}}}}, want{"`test` = ?", []any{1}, fmt.Errorf(filterOrWhereError, "Filter")}},
		{"one_with_exclude", args{"`test` = ?", []any{1}, &filterArgs{1, []any{Cond{"test": 1}}}}, want{"`test` = ?", []any{1}, fmt.Errorf(filterOrWhereError, "Exclude")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.WhereToSQL(tt.args.cond, tt.args.args...)
			if tt.args.filters != nil {
				p.FilterToSQL(tt.args.filters.isNot, tt.args.filters.filter...)
			}

			sql, sqlArgs := p.GetQuerySet()

			if !errors.Is(p.Error(), tt.want.err) && p.Error().Error() != tt.want.err.Error() {
				t.Errorf("TestFilter SQL Occur Error -> error:%+v", p.Error())
				return
			}

			wantSQL := " WHERE " + tt.want.sql
			if sql != wantSQL {
				t.Errorf("TestFilter SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestFilter SQL Gen Error -> want:%v", wantSQL)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestFilter Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestFilter Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestFilter Args Length Error -> want:%+v", tt.want.args)
			}

			for i, a := range sqlArgs {
				if !reflect.DeepEqual(a, tt.want.args[i]) {
					t.Errorf("TestFilter Arg Value Error -> args:%+v", sqlArgs)
					t.Errorf("TestFilter Arg Value Error -> want:%+v", tt.want.args)
				}
			}
		})
	}
}

func TestSelect(t *testing.T) {
	type args struct {
		selects any
	}
	type want struct {
		sql string
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"blank string", args{""}, want{sql: "", err: nil}},
		{"string", args{"test, test2 as test3"}, want{sql: "test, test2 as test3", err: nil}},
		{"zero slice", args{[]string{}}, want{sql: "*", err: nil}},
		{"one slice", args{[]string{"test"}}, want{sql: "`test`", err: nil}},
		{"two slice", args{[]string{"test", "test2"}}, want{sql: "`test`, `test2`", err: nil}},
		{"three slice", args{[]string{"test", "test2", "test3"}}, want{sql: "`test`, `test2`, `test3`", err: nil}},
		{"four slice", args{[]string{"test", "test2", "test3", "test4"}}, want{sql: "`test`, `test2`, `test3`, `test4`", err: nil}},
		//{"as in slice", args{[]string{"test as test1"}}, want{sql: "`test` as `test1`", err: nil}},
		//{"two as in slice", args{[]string{"test as test1", "testa as test2"}}, want{sql: "`test` as `test1`, `testa` as `test2`", err: nil}},
		//{"mix in slice", args{[]string{"test as test1", "test2"}}, want{sql: "`test` as `test1`, `test2`", err: nil}},
		//{"excess space in slice", args{[]string{"test as test1", " test2"}}, want{sql: "`test` as `test1`, `test2`", err: nil}},
		//{"excess space in slice 2", args{[]string{"test as  test1", " test2"}}, want{sql: "`test` as `test1`, `test2`", err: nil}},
		//{"DISTINCT in slice 2", args{[]string{"DISTINCT test1", " test2"}}, want{sql: "DISTINCT `test1`, `test2`", err: nil}},
		{"array", args{[1]string{"test"}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())

			sql := p.GetSelectSQL()
			if sql != "*" {
				t.Errorf("TestSelect SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestSelect SQL Gen Error -> want: %v", "*")
			}

			p.SelectToSQL(tt.args.selects)
			sql = p.GetSelectSQL()

			if p.Error() != nil {
				if p.Error().Error() != tt.want.err.Error() {
					t.Errorf("TestSelect SQL Occur Error -> error: %+v", p.Error())
				}

				return
			}

			if sql != tt.want.sql {
				t.Errorf("TestSelect SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestSelect SQL Gen Error -> want: %v", tt.want.sql)
			}
		})
	}
}

func TestLimit(t *testing.T) {
	type args struct {
		PageSize int64
		PageNum  int64
	}
	type want struct {
		sql string
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"zero1", args{0, 0}, want{"", nil}},
		{"zero2", args{10, 0}, want{"", nil}},
		{"zero3", args{0, 10}, want{"", nil}},
		{"negative1", args{-1, -1}, want{"", errors.New(pageSizeORNumberError)}},
		{"negative2", args{10, -1}, want{"", errors.New(pageSizeORNumberError)}},
		{"negative3", args{-1, 0}, want{"", nil}},

		{"page one size ten", args{10, 1}, want{" LIMIT 10 OFFSET 0", nil}},
		{"page two size ten", args{10, 2}, want{" LIMIT 10 OFFSET 10", nil}},
		{"page three size ten", args{10, 3}, want{" LIMIT 10 OFFSET 20", nil}},
		{"page four size ten", args{10, 4}, want{" LIMIT 10 OFFSET 30", nil}},
		{"page five size ten", args{10, 5}, want{" LIMIT 10 OFFSET 40", nil}},
		{"page six size ten", args{10, 6}, want{" LIMIT 10 OFFSET 50", nil}},
		{"page seven size ten", args{10, 7}, want{" LIMIT 10 OFFSET 60", nil}},
		{"page eight size ten", args{10, 8}, want{" LIMIT 10 OFFSET 70", nil}},
		{"page nine size ten", args{10, 9}, want{" LIMIT 10 OFFSET 80", nil}},
		{"page ten size ten", args{10, 10}, want{" LIMIT 10 OFFSET 90", nil}},
		{"page eleven size ten", args{10, 11}, want{" LIMIT 10 OFFSET 100", nil}},
		{"page twelve size ten", args{10, 12}, want{" LIMIT 10 OFFSET 110", nil}},
		{"page thirteen size ten", args{10, 13}, want{" LIMIT 10 OFFSET 120", nil}},
		{"page fourteen size eleven", args{11, 14}, want{" LIMIT 11 OFFSET 143", nil}},
		{"page fifteen size twelve", args{12, 15}, want{" LIMIT 12 OFFSET 168", nil}},
		{"page sixteen size thirteen", args{13, 16}, want{" LIMIT 13 OFFSET 195", nil}},
		{"page seventeen size fourteen", args{14, 17}, want{" LIMIT 14 OFFSET 224", nil}},
		{"page eighteen size fifteen", args{15, 18}, want{" LIMIT 15 OFFSET 255", nil}},
		{"page nineteen size sixteen", args{16, 19}, want{" LIMIT 16 OFFSET 288", nil}},
		{"page twenty size seventeen", args{17, 20}, want{" LIMIT 17 OFFSET 323", nil}},
		{"page twenty one size ten", args{10, 21}, want{" LIMIT 10 OFFSET 200", nil}},
		{"page thirty size ten", args{10, 30}, want{" LIMIT 10 OFFSET 290", nil}},
		{"page forty size ten", args{10, 40}, want{" LIMIT 10 OFFSET 390", nil}},
		{"page fifty size ten", args{10, 50}, want{" LIMIT 10 OFFSET 490", nil}},
		{"page sixty size ten", args{10, 60}, want{" LIMIT 10 OFFSET 590", nil}},
		{"page seventy size ten", args{10, 70}, want{" LIMIT 10 OFFSET 690", nil}},
		{"page eighty size ten", args{10, 80}, want{" LIMIT 10 OFFSET 790", nil}},
		{"page ninety size ten", args{10, 90}, want{" LIMIT 10 OFFSET 890", nil}},
		{"page one hundred size ten", args{10, 100}, want{" LIMIT 10 OFFSET 990", nil}},
		{"page one hundred one size ten", args{10, 101}, want{" LIMIT 10 OFFSET 1000", nil}},

		{"page one size thirty", args{30, 1}, want{" LIMIT 30 OFFSET 0", nil}},
		{"page two size thirty", args{30, 2}, want{" LIMIT 30 OFFSET 30", nil}},
		{"page three size thirty", args{30, 3}, want{" LIMIT 30 OFFSET 60", nil}},
		{"page four size thirty", args{30, 4}, want{" LIMIT 30 OFFSET 90", nil}},
		{"page five size thirty", args{30, 5}, want{" LIMIT 30 OFFSET 120", nil}},
		{"page ten size thirty", args{30, 10}, want{" LIMIT 30 OFFSET 270", nil}},
		{"page twenty size thirty", args{30, 20}, want{" LIMIT 30 OFFSET 570", nil}},
		{"page thirty size thirty", args{30, 30}, want{" LIMIT 30 OFFSET 870", nil}},
		{"page forty size thirty", args{30, 40}, want{" LIMIT 30 OFFSET 1170", nil}},
		{"page fifty size thirty", args{30, 50}, want{" LIMIT 30 OFFSET 1470", nil}},
		{"page one hundred size thirty", args{30, 100}, want{" LIMIT 30 OFFSET 2970", nil}},
		{"page one hundred one size thirty", args{30, 101}, want{" LIMIT 30 OFFSET 3000", nil}},

		{"page one size fifty", args{50, 1}, want{" LIMIT 50 OFFSET 0", nil}},
		{"page two size fifty", args{50, 2}, want{" LIMIT 50 OFFSET 50", nil}},
		{"page three size fifty", args{50, 3}, want{" LIMIT 50 OFFSET 100", nil}},
		{"page four size fifty", args{50, 4}, want{" LIMIT 50 OFFSET 150", nil}},
		{"page five size fifty", args{50, 5}, want{" LIMIT 50 OFFSET 200", nil}},

		{"page one size one hundred", args{100, 1}, want{" LIMIT 100 OFFSET 0", nil}},
		{"page two size one hundred", args{100, 2}, want{" LIMIT 100 OFFSET 100", nil}},
		{"page three size one hundred", args{100, 3}, want{" LIMIT 100 OFFSET 200", nil}},
		{"page four size one hundred", args{100, 4}, want{" LIMIT 100 OFFSET 300", nil}},
		{"page five size one hundred", args{100, 5}, want{" LIMIT 100 OFFSET 400", nil}},

		{"page one size one thousand", args{1000, 1}, want{" LIMIT 1000 OFFSET 0", nil}},
		{"page two size one thousand", args{1000, 2}, want{" LIMIT 1000 OFFSET 1000", nil}},
		{"page three size one thousand", args{1000, 3}, want{" LIMIT 1000 OFFSET 2000", nil}},
		{"page four size one thousand", args{1000, 4}, want{" LIMIT 1000 OFFSET 3000", nil}},
		{"page five size one thousand", args{1000, 5}, want{" LIMIT 1000 OFFSET 4000", nil}},

		{"page one size ten thousand", args{10000, 1}, want{" LIMIT 10000 OFFSET 0", nil}},
		{"page two size ten thousand", args{10000, 2}, want{" LIMIT 10000 OFFSET 10000", nil}},
		{"page three size ten thousand", args{10000, 3}, want{" LIMIT 10000 OFFSET 20000", nil}},
		{"page four size ten thousand", args{10000, 4}, want{" LIMIT 10000 OFFSET 30000", nil}},
		{"page five size ten thousand", args{10000, 5}, want{" LIMIT 10000 OFFSET 40000", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.LimitToSQL(tt.args.PageSize, tt.args.PageNum)
			sql := p.GetLimitSQL()

			if p.Error() != nil && errors.Is(p.Error(), tt.want.err) {
				t.Errorf("TestLimit SQL Occur Error -> error:%+v", p.Error())
			}

			if sql != tt.want.sql {
				t.Errorf("TestLimit SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestLimit SQL Gen Error -> want:%v", tt.want.sql)
			}
		})
	}
}

func TestOrderBy(t *testing.T) {
	type args struct {
		order any
	}
	type want struct {
		sql string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"one_asc ", args{[]string{"test"}}, want{" ORDER BY `test` ASC"}},
		{"two_asc ", args{[]string{"test", "test2"}}, want{" ORDER BY `test` ASC, `test2` ASC"}},
		{"three_asc ", args{[]string{"test", "test2", "test3"}}, want{" ORDER BY `test` ASC, `test2` ASC, `test3` ASC"}},
		{"one_desc ", args{[]string{"-test"}}, want{" ORDER BY `test` DESC"}},
		{"two_desc ", args{[]string{"-test", "-test2"}}, want{" ORDER BY `test` DESC, `test2` DESC"}},
		{"three_desc ", args{[]string{"-test", "-test2", "-test3"}}, want{" ORDER BY `test` DESC, `test2` DESC, `test3` DESC"}},
		{"two_mix ", args{[]string{"test", "-test2"}}, want{" ORDER BY `test` ASC, `test2` DESC"}},
		{"three_mix ", args{[]string{"test", "-test2", "test3"}}, want{" ORDER BY `test` ASC, `test2` DESC, `test3` ASC"}},
		{"three_mix ", args{[]string{"-test", "test2", "-test3"}}, want{" ORDER BY `test` DESC, `test2` ASC, `test3` DESC"}},
		{"str_one", args{"test"}, want{" ORDER BY test"}},
		{"str_two", args{"test, test2"}, want{" ORDER BY test, test2"}},
		{"str_three", args{"test, test2, test3"}, want{" ORDER BY test, test2, test3"}},
		{"str_one_desc", args{"test desc"}, want{" ORDER BY test desc"}},
		{"str_three_mix", args{"test, test2 desc, test3 asc"}, want{" ORDER BY test, test2 desc, test3 asc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.OrderByToSQL(tt.args.order)
			sql := p.GetOrderBySQL()

			if p.Error() != nil {
				t.Errorf("TestOrderBy SQL Occur Error -> error:%+v", p.Error())
			}

			if sql != tt.want.sql {
				t.Errorf("TestOrderBy SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestOrderBy SQL Gen Error -> want:%v", tt.want.sql)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	type args struct {
		groupby any
	}
	type want struct {
		sql string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"blank string", args{""}, want{""}},
		{"string", args{"test, test2"}, want{" GROUP BY test, test2"}},
		{"zero slice", args{[]string{}}, want{""}},
		{"one slice", args{[]string{"test"}}, want{" GROUP BY `test`"}},
		{"two slice", args{[]string{"test", "test2"}}, want{" GROUP BY `test`, `test2`"}},
		{"three slice", args{[]string{"test", "test2", "test3"}}, want{" GROUP BY `test`, `test2`, `test3`"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.GroupByToSQL(tt.args.groupby)

			sql := p.GetGroupBySQL()

			if p.Error() != nil {
				t.Errorf("TestGroupBy SQL Occur Error -> error:%+v", p.Error())
			}

			if sql != tt.want.sql {
				t.Errorf("TestGroupBy SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestGroupBy SQL Gen Error -> want:%v", tt.want.sql)
			}
		})
	}
}

func TestHaving(t *testing.T) {
	type args struct {
		havingSQL  string
		havingArgs []any
	}
	type want struct {
		havingSQL  string
		havingArgs []any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"blank string", args{"", []any{}}, want{"", []any{}}},
		{"string", args{"SUM(test) > ?", []any{1}}, want{" HAVING SUM(test) > ?", []any{1}}},
		{"string", args{"SUM(test) > ? AND SUM(test2) < ?", []any{1, 2}}, want{" HAVING SUM(test) > ? AND SUM(test2) < ?", []any{1, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.HavingToSQL(tt.args.havingSQL, tt.args.havingArgs...)

			sql, sqlArgs := p.GetHavingSQL()

			if p.Error() != nil {
				t.Errorf("TestHaving SQL Occur Error -> error:%+v", p.Error())
			}

			if sql != tt.want.havingSQL {
				t.Errorf("TestHaving SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestHaving SQL Gen Error -> want:%v", tt.want.havingSQL)
			}

			if !reflect.DeepEqual(sqlArgs, tt.want.havingArgs) {
				t.Errorf("TestHaving SQL Gen Error -> args :%v", sqlArgs)
				t.Errorf("TestHaving SQL Gen Error -> want:%+v", tt.want.havingArgs)
			}
		})
	}
}

func TestWhereError(t *testing.T) {
	tests := []struct {
		name string
		cond string
		args []any
		want error
	}{
		{"args_mismatch", "test = ? AND test2 = ? AND test3 = ?", []any{1, 2}, errors.New(argsLenError)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.WhereToSQL(tt.cond, tt.args...)

			if p.Error() == nil || p.Error().Error() != tt.want.Error() {
				t.Errorf("TestWhereError not works as expected.")
				t.Errorf("want: %s", tt.want)
				t.Errorf("get : %s", p.Error())
			}
		})
	}
}

func TestFilterAndExcludeConflict(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p.WhereToSQL("test = ?", 1)
	p.FilterToSQL(0, Cond{"test": 1})

	if p.Error() == nil || p.Error().Error() != fmt.Errorf(filterOrWhereError, "Filter").Error() {
		t.Errorf("TestFilterAndExcludeConflict not working as expected")
	}

	p = NewQuerySet(mysqlOp.NewOperator())
	p.WhereToSQL("test = ?", 1)
	p.FilterToSQL(1, Cond{"test": 1})

	if p.Error() == nil || p.Error().Error() != fmt.Errorf(filterOrWhereError, "Exclude").Error() {
		t.Errorf("TestFilterAndExcludeConflict not working as expected")
	}
}

func TestSelectError(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p.SelectToSQL(123)

	if p.Error() == nil || p.Error().Error() != paramTypeError {
		t.Errorf("TestSelectError not working as expected")
	}
}

func TestOrderByError(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p.OrderByToSQL(123)

	if p.Error() == nil || p.Error().Error() != paramTypeError {
		t.Errorf("TestOrderByError not working as expected")
	}
}

func TestGroupByError(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p.GroupByToSQL(123)

	if p.Error() == nil || p.Error().Error() != paramTypeError {
		t.Errorf("TestGroupByError not working as expected")
	}
}

func TestFilterWithConjunction(t *testing.T) {
	type args struct {
		isNot  int
		filter []any
	}
	type want struct {
		sql  string
		args []any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"with_and", args{0, []any{"AND", Cond{"test": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"with_or", args{0, []any{"OR", Cond{"test": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"with_invalid", args{0, []any{"INVALID", Cond{"test": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"with_and_not", args{1, []any{"AND", Cond{"test": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
		{"with_or_not", args{1, []any{"OR", Cond{"test": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p = p.FilterToSQL(tt.args.isNot, tt.args.filter...)
			sql, sqlArgs := p.GetQuerySet()

			if p.Error() != nil {
				t.Errorf("TestFilterWithConjunction SQL Occur Error -> error:%+v", p.Error())
			}

			wantSQL := tt.want.sql
			if sql != wantSQL {
				t.Errorf("TestFilterWithConjunction SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestFilterWithConjunction SQL Gen Error -> want:%v", wantSQL)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestFilterWithConjunction Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestFilterWithConjunction Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestFilterWithConjunction Args Length Error -> want:%+v", tt.want.args)
			}
		})
	}
}

func TestInvalidFilterType(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p = p.FilterToSQL(0, 123) // Invalid filter type
	p.GetQuerySet()

	expected := fmt.Errorf(unsupportedFilterTypeError, "int")
	if p.Error() == nil || p.Error().Error() != expected.Error() {
		t.Errorf("TestInvalidFilterType not working as expected")
		t.Errorf("want: %s", expected)
		t.Errorf("get : %s", p.Error())
	}
}

func TestEmptyFilters(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p = p.FilterToSQL(0)
	sql, args := p.GetQuerySet()

	if p.Error() != nil {
		t.Errorf("TestEmptyFilters SQL Occur Error -> error:%+v", p.Error())
	}

	if sql != "" {
		t.Errorf("TestEmptyFilters SQL should be empty, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("TestEmptyFilters Args should be empty, got: %v", args)
	}
}

func TestInvalidIsNot(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p = p.FilterToSQL(2, Cond{"test": 1}) // 2 is invalid for isNot
	p.GetQuerySet()

	if p.Error() == nil || p.Error().Error() != isNotValueError {
		t.Errorf("TestInvalidIsNot not working as expected")
		t.Errorf("want: %s", isNotValueError)
		t.Errorf("get : %s", p.Error())
	}
}

func TestFilterResetAndError(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())

	// Create an error
	p.SelectToSQL(123)
	if p.Error() == nil {
		t.Errorf("Error should be set")
	}

	// Reset should clear the error
	p.Reset()
	if p.Error() != nil {
		t.Errorf("Error should be nil after Reset, got: %v", p.Error())
	}

	// After reset, functions should work properly
	p.FilterToSQL(0, Cond{"test": 1})
	sql, args := p.GetQuerySet()

	if sql != " WHERE (`test` = ?)" || len(args) != 1 || args[0] != 1 {
		t.Errorf("FilterToSQL not working after Reset")
	}
}

func TestComplexFilterCombinations(t *testing.T) {
	type args struct {
		isNot  int
		filter []any
	}
	type want struct {
		sql  string
		args []any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"complex_nested_conditions",
			args{0, []any{
				Cond{"test1": 1},
				AND{SortKey: []string{"test2", "test3"}, "test2": 2, "test3__contains": "val"},
				OR{"test4__in": []int{4, 5, 6}},
			}},
			want{
				" WHERE ((`test1` = ?) AND (`test2` = ? AND `test3` LIKE BINARY ?) OR (`test4` IN (?,?,?)))",
				[]any{1, 2, "%val%", 4, 5, 6},
			},
		},
		{
			"mixed_operators_with_not",
			args{1, []any{
				Cond{"test1__gt": 10},
				OR{"test2__contains": "search"},
				AND{"test3__between": []int{20, 30}},
			}},
			want{
				" WHERE NOT ((`test1` > ?) OR (`test2` LIKE BINARY ?) AND (`test3` BETWEEN ? AND ?))",
				[]any{10, "%search%", 20, 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p = p.FilterToSQL(tt.args.isNot, tt.args.filter...)
			sql, sqlArgs := p.GetQuerySet()

			if p.Error() != nil {
				t.Errorf("TestComplexFilterCombinations SQL Occur Error -> error:%+v", p.Error())
			}

			wantSQL := tt.want.sql
			if sql != wantSQL {
				t.Errorf("TestComplexFilterCombinations SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestComplexFilterCombinations SQL Gen Error -> want: %v", wantSQL)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestComplexFilterCombinations Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestComplexFilterCombinations Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestComplexFilterCombinations Args Length Error -> want:%+v", tt.want.args)
			}
		})
	}
}

// Test that multiple call flags are properly set
func TestMultipleCallFlags(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator()).(*QuerySetImpl)

	// Test initial state
	if p.hasCalled(callFilter) || p.hasCalled(callExclude) || p.hasCalled(callWhere) {
		t.Errorf("Initial call flags should be unset")
	}

	// Test filter flag
	p.FilterToSQL(0, Cond{"test": 1})
	if !p.hasCalled(callFilter) {
		t.Errorf("callFilter flag should be set")
	}

	// Reset and test exclude flag
	p.Reset()
	p.FilterToSQL(1, Cond{"test": 1})
	if !p.hasCalled(callExclude) {
		t.Errorf("callExclude flag should be set")
	}

	// Reset and test where flag
	p.Reset()
	p.WhereToSQL("test = ?", 1)
	if !p.hasCalled(callWhere) {
		t.Errorf("callWhere flag should be set")
	}

	// Test multiple flags
	p.Reset()
	p.SelectToSQL("test")
	p.OrderByToSQL("test")
	p.LimitToSQL(10, 1)
	p.GroupByToSQL("test")
	p.HavingToSQL("test = ?", 1)

	if !p.hasCalled(callSelect) || !p.hasCalled(callOrderBy) ||
		!p.hasCalled(callLimit) || !p.hasCalled(callGroupBy) ||
		!p.hasCalled(callHaving) {
		t.Errorf("Multiple call flags were not set correctly")
	}
}

func TestEmptyConditionsList(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())

	// Add an empty filter condition
	p.FilterToSQL(0, Cond{})
	sql, args := p.GetQuerySet()

	if sql != "" || len(args) != 0 {
		t.Errorf("Empty condition should result in empty SQL, got: %s", sql)
	}
}

func TestConditionalExpressionHandling(t *testing.T) {
	tests := []struct {
		name string
		fn   func(p QuerySet) QuerySet
		want string
	}{
		{
			"select_column_empty",
			func(p QuerySet) QuerySet {
				return p.SelectToSQL("")
			},
			"",
		},
		{
			"select_column_explicit_asterisk",
			func(p QuerySet) QuerySet {
				return p.SelectToSQL("*")
			},
			"*",
		},
		{
			"group_by_empty_slice",
			func(p QuerySet) QuerySet {
				return p.GroupByToSQL([]string{})
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			result := tt.fn(p)

			if p.Error() != nil {
				t.Errorf("Error should be nil, got: %v", p.Error())
			}

			var got string
			if strings.Contains(tt.name, "select") {
				got = result.GetSelectSQL()
			} else if strings.Contains(tt.name, "group") {
				got = result.GetGroupBySQL()
			}

			if got != tt.want {
				t.Errorf("Expected result %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGetHavingSQL(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []any
		want     string
		wantArgs []any
	}{
		{"empty_having", "", nil, "", []any{}},
		{"simple_having", "SUM(test) > ?", []any{100}, " HAVING SUM(test) > ?", []any{100}},
		{"complex_having", "AVG(test) BETWEEN ? AND ?", []any{10, 20}, " HAVING AVG(test) BETWEEN ? AND ?", []any{10, 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			if tt.query != "" {
				p.HavingToSQL(tt.query, tt.args...)
			}

			sql, args := p.GetHavingSQL()

			if sql != tt.want {
				t.Errorf("GetHavingSQL() SQL = %v, want %v", sql, tt.want)
			}

			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("GetHavingSQL() args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}

// Test case for handling empty filter conditions within a filter call
func TestEmptyFiltersInMiddle(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p = p.FilterToSQL(0, Cond{"test1": 1}, Cond{}, Cond{"test3": 3})
	sql, args := p.GetQuerySet()

	// Should ignore empty filter
	expectedSQL := " WHERE ((`test1` = ?) AND (`test3` = ?))"
	expectedArgs := []any{1, 3}

	if sql != expectedSQL {
		t.Errorf("TestEmptyFiltersInMiddle SQL = %v, want %v", sql, expectedSQL)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("TestEmptyFiltersInMiddle args = %v, want %v", args, expectedArgs)
	}
}

// Test behavior when SQL contains multiple filters but all are empty
func TestAllEmptyFilters(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p = p.FilterToSQL(0, Cond{}, AND{}, OR{})
	sql, args := p.GetQuerySet()

	// Should result in empty SQL
	if sql != "" {
		t.Errorf("TestAllEmptyFilters SQL should be empty, got: %v", sql)
	}

	if len(args) != 0 {
		t.Errorf("TestAllEmptyFilters args should be empty, got: %v", args)
	}
}
