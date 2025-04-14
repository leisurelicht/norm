package norm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	mysqlOp "github.com/leisurelicht/norm/operator/mysql"
)

func TestFilter(t *testing.T) {
	type args struct {
		state  int
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
		{"empty", args{isFilter, []any{}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{Cond{}}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{AND{}}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{OR{}}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{Cond{}, AND{}}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{Cond{}, OR{}}}, want{"", []any{}}},
		{"default_cond", args{isFilter, []any{Cond{}, AND{}, OR{}}}, want{"", []any{}}},

		{"default_cond", args{isFilter, []any{Cond{"test": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"default_list_cond", args{isFilter, []any{Cond{"test": []any{1, 2}}}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"exact_cond", args{isFilter, []any{Cond{"test__exact": 1}}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"exact_list_cond", args{isFilter, []any{Cond{"test__exact": []any{1, 2}}}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"exclude_cond", args{isFilter, []any{Cond{"test__exclude": 1}}}, want{" WHERE (`test` != ?)", []any{1}}},
		{"exclude_list_cond", args{isFilter, []any{Cond{"test__exclude": []any{1, 2}}}}, want{" WHERE (`test` != ? AND `test` != ?)", []any{1, 2}}},
		{"iexact_cond", args{isFilter, []any{Cond{"test__iexact": 1}}}, want{" WHERE (`test` LIKE ?)", []any{1}}},
		{"iexact_list_cond", args{isFilter, []any{Cond{"test__iexact": []any{1, 2}}}}, want{" WHERE (`test` LIKE ? AND `test` LIKE ?)", []any{1, 2}}},
		{"gt_cond", args{isFilter, []any{Cond{"test__gt": 1}}}, want{" WHERE (`test` > ?)", []any{1}}},
		{"gte_cond", args{isFilter, []any{Cond{"test__gte": 1}}}, want{" WHERE (`test` >= ?)", []any{1}}},
		{"lt_cond", args{isFilter, []any{Cond{"test__lt": 1}}}, want{" WHERE (`test` < ?)", []any{1}}},
		{"lte_cond", args{isFilter, []any{Cond{"test__lte": 1}}}, want{" WHERE (`test` <= ?)", []any{1}}},
		{"len_cond", args{isFilter, []any{Cond{"test__len": 1}}}, want{" WHERE (LENGTH(`test`) = ?)", []any{1}}},
		{"in_string_cond", args{isFilter, []any{Cond{"test__in": "1,2,3"}}}, want{" WHERE (`test` IN (1,2,3))", []any{}}},
		{"in_list_cond", args{isFilter, []any{Cond{"test__in": []int{1, 2}}}}, want{" WHERE (`test` IN (?,?))", []any{1, 2}}},
		{"not_in_string_cond", args{isFilter, []any{Cond{"test__not_in": "1,2,3"}}}, want{" WHERE (`test` NOT IN (1,2,3))", []any{}}},
		{"not_list_in_cond", args{isFilter, []any{Cond{"test__not_in": []int{1, 2}}}}, want{" WHERE (`test` NOT IN (?,?))", []any{1, 2}}},
		{"between_cond", args{isFilter, []any{Cond{"test__between": []int{1, 2}}}}, want{" WHERE (`test` BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_between_cond", args{isFilter, []any{Cond{"test__not_between": []int{1, 2}}}}, want{" WHERE (`test` NOT BETWEEN ? AND ?)", []any{1, 2}}},
		{"contains_cond", args{isFilter, []any{Cond{"test__contains": "e"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"%e%"}}},
		{"list_contains cond", args{isFilter, []any{Cond{"test__contains": []string{"e", "s"}}}}, want{" WHERE (`test` LIKE BINARY ? AND `test` LIKE BINARY ?)", []any{"%e%", "%s%"}}},
		{"not_contains cond", args{isFilter, []any{Cond{"test__not_contains": "e"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"%e%"}}},
		{"icontains_cond", args{isFilter, []any{Cond{"test__icontains": "E"}}}, want{" WHERE (`test` LIKE ?)", []any{"%E%"}}},
		{"not_icontains cond", args{isFilter, []any{Cond{"test__not_icontains": "E"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"%E%"}}},
		{"startswith_cond", args{isFilter, []any{Cond{"test__startswith": "te"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"te%"}}},
		{"not_startswith cond", args{isFilter, []any{Cond{"test__not_startswith": "te"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"te%"}}},
		{"istartswith_cond", args{isFilter, []any{Cond{"test__istartswith": "tE"}}}, want{" WHERE (`test` LIKE ?)", []any{"tE%"}}},
		{"not_istartswith cond", args{isFilter, []any{Cond{"test__not_istartswith": "tE"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"tE%"}}},
		{"endswith_cond", args{isFilter, []any{Cond{"test__endswith": "st"}}}, want{" WHERE (`test` LIKE BINARY ?)", []any{"%st"}}},
		{"not_endswith_cond", args{isFilter, []any{Cond{"test__not_endswith": "st"}}}, want{" WHERE (`test` NOT LIKE BINARY ?)", []any{"%st"}}},
		{"iendswith_cond", args{isFilter, []any{Cond{"test__iendswith": "sT"}}}, want{" WHERE (`test` LIKE ?)", []any{"%sT"}}},
		{"not_iendswith_cond", args{isFilter, []any{Cond{"test__not_iendswith": "sT"}}}, want{" WHERE (`test` NOT LIKE ?)", []any{"%sT"}}},

		{"not_default_cond", args{isExclude, []any{Cond{"test": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
		{"not_exact_cond", args{isExclude, []any{Cond{"test__exact": 1}}}, want{" WHERE NOT (`test` = ?)", []any{1}}},
		{"not_exclude_cond", args{isExclude, []any{Cond{"test__exclude": 1}}}, want{" WHERE NOT (`test` != ?)", []any{1}}},
		{"not_iexact_cond", args{isExclude, []any{Cond{"test__iexact": 1}}}, want{" WHERE NOT (`test` LIKE ?)", []any{1}}},
		{"not_gt_cond", args{isExclude, []any{Cond{"test__gt": 1}}}, want{" WHERE NOT (`test` > ?)", []any{1}}},
		{"not_gte_cond", args{isExclude, []any{Cond{"test__gte": 1}}}, want{" WHERE NOT (`test` >= ?)", []any{1}}},
		{"not_lt_cond", args{isExclude, []any{Cond{"test__lt": 1}}}, want{" WHERE NOT (`test` < ?)", []any{1}}},
		{"not_lte_cond", args{isExclude, []any{Cond{"test__lte": 1}}}, want{" WHERE NOT (`test` <= ?)", []any{1}}},
		{"not_len_cond", args{isExclude, []any{Cond{"test__len": 1}}}, want{" WHERE NOT (LENGTH(`test`) = ?)", []any{1}}},
		{"not_in_cond", args{isExclude, []any{Cond{"test__in": []int{1, 2}}}}, want{" WHERE NOT (`test` IN (?,?))", []any{1, 2}}},
		{"not_not_in_cond", args{isExclude, []any{Cond{"test__not_in": []int{1, 2}}}}, want{" WHERE NOT (`test` NOT IN (?,?))", []any{1, 2}}},
		{"not_between_cond", args{isExclude, []any{Cond{"test__between": []int{1, 2}}}}, want{" WHERE NOT (`test` BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_not_between_cond", args{isExclude, []any{Cond{"test__not_between": []int{1, 2}}}}, want{" WHERE NOT (`test` NOT BETWEEN ? AND ?)", []any{1, 2}}},
		{"not_contains_cond", args{isExclude, []any{Cond{"test__contains": "e"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"%e%"}}},
		{"not_not_contains_cond", args{isExclude, []any{Cond{"test__not_contains": "e"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"%e%"}}},
		{"not_icontains_cond", args{isExclude, []any{Cond{"test__icontains": "E"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"%E%"}}},
		{"not_not_icontains_cond", args{isExclude, []any{Cond{"test__not_icontains": "E"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"%E%"}}},
		{"not_startswith_cond", args{isExclude, []any{Cond{"test__startswith": "te"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"te%"}}},
		{"not_not_startswith_cond", args{isExclude, []any{Cond{"test__not_startswith": "te"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"te%"}}},
		{"not_istartswith_cond", args{isExclude, []any{Cond{"test__istartswith": "tE"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"tE%"}}},
		{"not_not_istartswith_cond", args{isExclude, []any{Cond{"test__not_istartswith": "tE"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"tE%"}}},
		{"not_endswith_cond", args{isExclude, []any{Cond{"test__endswith": "st"}}}, want{" WHERE NOT (`test` LIKE BINARY ?)", []any{"%st"}}},
		{"not_not_endswith_cond", args{isExclude, []any{Cond{"test__not_endswith": "st"}}}, want{" WHERE NOT (`test` NOT LIKE BINARY ?)", []any{"%st"}}},
		{"not_not_iendswith_cond", args{isExclude, []any{Cond{"test__iendswith": "sT"}}}, want{" WHERE NOT (`test` LIKE ?)", []any{"%sT"}}},
		{"not_not_iendswith_cond", args{isExclude, []any{Cond{"test__not_iendswith": "sT"}}}, want{" WHERE NOT (`test` NOT LIKE ?)", []any{"%sT"}}},

		{"two_default_column", args{isFilter, []any{Cond{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2}}}, want{" WHERE (`test1` = ? AND `test2` = ?)", []any{1, 2}}},
		{"reverse_default_column", args{isFilter, []any{Cond{SortKey: []string{"test2", "test1"}, "test1": 1, "test2": 2}}}, want{" WHERE (`test2` = ? AND `test1` = ?)", []any{2, 1}}},
		{"three_default_column", args{isFilter, []any{Cond{SortKey: []string{"test1", "test2", "test3"}, "test1__gt": 1, "test2": 2, "test3": 3}}}, want{" WHERE (`test1` > ? AND `test2` = ? AND `test3` = ?)", []any{1, 2, 3}}},
		{"reverse_three_default_column", args{isFilter, []any{Cond{SortKey: []string{"test3", "test2", "test1"}, "test1__lt": 1, "test2": 2, "test3": 3}}}, want{" WHERE (`test3` = ? AND `test2` = ? AND `test1` < ?)", []any{3, 2, 1}}},
		{"out_order_three_default_column", args{isFilter, []any{Cond{SortKey: []string{"test1", "test3", "test2"}, "test3__in": []int{1, 4, 6}, "test2": 2, "test1": 3}}}, want{" WHERE (`test1` = ? AND `test3` IN (?,?,?) AND `test2` = ?)", []any{3, 1, 4, 6, 2}}},

		{"default_conj", args{isFilter, []any{Cond{"test": 1}, Cond{"test2": 2}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?))", []any{1, 2}}},
		{"and_conj", args{isFilter, []any{Cond{"test": 1}, AND{"test2": 2}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?))", []any{1, 2}}},
		{"or_conj", args{isFilter, []any{Cond{"test": 1}, OR{"test2": 2}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?))", []any{1, 2}}},
		{"3_default_conj", args{isFilter, []any{Cond{"test": 1}, Cond{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3_and_conj", args{isFilter, []any{AND{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3_or_conj", args{isFilter, []any{OR{"test": 1}, OR{"test2": 2}, OR{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) OR (`test3` = ?))", []any{1, 2, "3"}}},
		{"d_and_d_conj", args{isFilter, []any{Cond{"test": 1}, AND{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&2and_conj", args{isFilter, []any{Cond{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d_or_d_conj", args{isFilter, []any{Cond{"test": 1}, OR{"test2": 2}, Cond{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&2or_conj", args{isFilter, []any{Cond{"test": 1}, OR{"test2": 2}, OR{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) OR (`test3` = ?))", []any{1, 2, "3"}}},
		{"d&or&and_conj", args{isFilter, []any{Cond{"test": 1}, OR{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) OR (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},
		{"3and_conj", args{isFilter, []any{AND{"test": 1}, AND{"test2": 2}, AND{"test3": "3"}}}, want{" WHERE ((`test` = ?) AND (`test2` = ?) AND (`test3` = ?))", []any{1, 2, "3"}}},

		{"default_cond", args{isFilter, []any{Cond{"test": []any{1, 2}}, AND{"test2": 3}, OR{"test3": []any{4, 5}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ?) OR (`test3` = ? AND `test3` = ?))", []any{1, 2, 3, 4, 5}}},
		{"exact_cond", args{isFilter, []any{Cond{"test__exact": []any{1, 2}}, AND{"test2__exact": 3}, OR{"test3__exact": []any{4, 5}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ?) OR (`test3` = ? AND `test3` = ?))", []any{1, 2, 3, 4, 5}}},
		{"exclude_cond", args{isFilter, []any{Cond{"test__exclude": []any{1, 2}}, AND{"test2__exclude": 3}, OR{"test3__exclude": []any{4, 5}}}}, want{" WHERE ((`test` != ? AND `test` != ?) AND (`test2` != ?) OR (`test3` != ? AND `test3` != ?))", []any{1, 2, 3, 4, 5}}},
		{"iexact_cond", args{isFilter, []any{Cond{"test__iexact": []any{1, 2}}, AND{"test2__iexact": 3}, OR{"test3__iexact": []any{4, 5}}}}, want{" WHERE ((`test` LIKE ? AND `test` LIKE ?) AND (`test2` LIKE ?) OR (`test3` LIKE ? AND `test3` LIKE ?))", []any{1, 2, 3, 4, 5}}},
		{"gt_cond", args{isFilter, []any{Cond{"test__gt": 1}, AND{"test2__gt": 2}, OR{"test3__gt": 3}}}, want{" WHERE ((`test` > ?) AND (`test2` > ?) OR (`test3` > ?))", []any{1, 2, 3}}},
		{"in_cond", args{isFilter, []any{Cond{"test__in": "1,2,3"}, AND{"test2__in": []any{4, 5}}, OR{"test3__in": []any{6, 7}}}}, want{" WHERE ((`test` IN (1,2,3)) AND (`test2` IN (?,?)) OR (`test3` IN (?,?)))", []any{4, 5, 6, 7}}},

		{"default_mix_contains_conj", args{isFilter, []any{Cond{"test": 1}, Cond{"test2__contains": []string{"e", "s"}}}}, want{" WHERE ((`test` = ?) AND (`test2` LIKE BINARY ? AND `test2` LIKE BINARY ?))", []any{1, "%e%", "%s%"}}},

		{"exact_one_and_one_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, "test2": 2}}}, want{" WHERE (`test` = ? AND `test2` = ?)", []any{1, 2}}},
		{"exact_one_and_list_and_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, "test2": []any{3, 4}}}}, want{" WHERE (`test` = ? AND (`test2` = ? AND `test2` = ?))", []any{1, 3, 4}}},
		{"exact_list_and_list_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": []any{1, 2}, "test2": []any{3, 4}}}}, want{" WHERE ((`test` = ? AND `test` = ?) AND (`test2` = ? AND `test2` = ?))", []any{1, 2, 3, 4}}},

		{"test_value_is_null", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, "test2": nil}}}, want{" WHERE (`test` = ? AND `test2` IS NULL)", []any{1}}},
		{"test_value_is_null", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": nil, "test2": nil}}}, want{" WHERE (`test` IS NULL AND `test2` IS NULL)", []any{}}},

		{"test_empty_in_midst", args{isFilter, []any{Cond{"test1": 1}, Cond{}, Cond{"test3": 3}}}, want{" WHERE ((`test1` = ?) AND (`test3` = ?))", []any{1, 3}}},

		// ToOR
		{"exact_one_or_one_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, ToOR("test2"): 2}}}, want{" WHERE (`test` = ? OR `test2` = ?)", []any{1, 2}}},
		{"exact_one_or_list_and_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": 1, ToOR("test2"): []any{3, 4}}}}, want{" WHERE (`test` = ? OR (`test2` = ? AND `test2` = ?))", []any{1, 3, 4}}},
		{"exact_list_or_list_cond", args{isFilter, []any{Cond{SortKey: []string{"test", "test2"}, "test": []any{1, 2}, ToOR("test2"): []any{3, 4}}}}, want{" WHERE ((`test` = ? AND `test` = ?) OR (`test2` = ? AND `test2` = ?))", []any{1, 2, 3, 4}}},

		// EachOR
		{"each_or", args{isFilter, []any{EachOR(Cond{"test": 1})}}, want{" WHERE (`test` = ?)", []any{1}}},
		{"each_or_list", args{isFilter, []any{EachOR(Cond{"test": []any{1, 2}})}}, want{" WHERE (`test` = ? AND `test` = ?)", []any{1, 2}}},
		{"each_or_and", args{isFilter, []any{Cond{"test": 1}, EachOR(AND{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2})}}, want{" WHERE ((`test` = ?) AND (`test1` = ? OR `test2` = ?))", []any{1, 1, 2}}},
		{"each_or_or", args{isFilter, []any{Cond{"test": 1}, EachOR(OR{SortKey: []string{"test1", "test2"}, "test1": 1, "test2": 2})}}, want{" WHERE ((`test` = ?) OR (`test1` = ? OR `test2` = ?))", []any{1, 1, 2}}},

		// complex
		{
			"complex_nested_conditions",
			args{isFilter, []any{
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
			args{isExclude, []any{
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
			p = p.FilterToSQL(tt.args.state, tt.args.filter...)
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
		state  int
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
		{"double call", []args{{isFilter, []any{Cond{"test1": 1}}}, {0, []any{Cond{"test2": 1}}}}, want{" WHERE (`test1` = ?) AND (`test2` = ?)", []any{1, 1}}},

		// meet by accident
		{"meet1", []args{{isFilter, []any{Cond{SortKey: []string{"delete_flag", "devise_sn"}, "delete_flag": 0, "devise_sn__len": 22}}}, {0, []any{Cond{SortKey: []string{"device_name", "devise_sn", "belong_to_company"}, "device_name__icontains": "test", "devise_sn__icontains": "test", "belong_to_company__icontains": "test"}}}}, want{" WHERE (`delete_flag` = ? AND LENGTH(`devise_sn`) = ?) AND (`device_name` LIKE ? AND `devise_sn` LIKE ? AND `belong_to_company` LIKE ?)", []any{0, 22, "%test%", "%test%", "%test%"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			for _, f := range tt.args {
				p = p.FilterToSQL(f.state, f.filter...)
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
		{"empty", args{isFilter, []any{}}, want{nil}},
		{"InvalidStat", args{2, []any{Cond{"test": 1}}}, want{errors.New(isNotValueError)}},
		{"order key type", args{isFilter, []any{Cond{SortKey: []int{1, 2}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyTypeError)}},
		{"order key len", args{isFilter, []any{Cond{SortKey: []string{"1"}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyLenError)}},
		{"field lookup", args{isFilter, []any{Cond{"1__not__contains": "b"}}}, want{fmt.Errorf(fieldLookupError, "1__not__contains")}},
		{"unknown operator", args{isFilter, []any{Cond{"1__contain": "b"}}}, want{fmt.Errorf(unknownOperatorError, "contain")}},
		{"operator value len", args{isFilter, []any{Cond{"test__between": []int{}}}}, want{fmt.Errorf(operatorValueLenError, "between", 2)}},
		{"operator value len less", args{isFilter, []any{Cond{"test": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exact", 0)}},
		{"operator value len less1", args{isFilter, []any{Cond{"test__exclude": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exclude", 0)}},
		{"operator value len less2", args{isFilter, []any{Cond{"test__iexact": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "iexact", 0)}},
		{"operator value len less3", args{isFilter, []any{Cond{"test__in": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "in", 0)}},
		{"operator value len less4", args{isFilter, []any{Cond{"test__contains": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value len less5", args{isFilter, []any{Cond{"test__contains": [0]int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value type", args{isFilter, []any{Cond{"test__contains": []int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"operator value type1", args{isFilter, []any{Cond{"test__contains": [2]int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"not implementd operator", args{isFilter, []any{Cond{"test__unimplemented": 1}}}, want{fmt.Errorf(notImplementedOperatorError, "UNIMPLEMENTED")}},
		{"unsupported value", args{isFilter, []any{Cond{"test__exclude": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exclude", "map")}},
		{"unsupported value1", args{isFilter, []any{Cond{"test": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exact", "map")}},
		{"unsupported value2", args{isFilter, []any{Cond{"test__iexact": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "iexact", "map")}},
		{"unsupported value3", args{isFilter, []any{Cond{"test__gt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gt", "string")}},
		{"unsupported value4", args{isFilter, []any{Cond{"test__gte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gte", "string")}},
		{"unsupported value5", args{isFilter, []any{Cond{"test__lt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lt", "string")}},
		{"unsupported value6", args{isFilter, []any{Cond{"test__lte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lte", "string")}},
		{"unsupported value7", args{isFilter, []any{Cond{"test__len": "10"}}}, want{fmt.Errorf(unsupportedValueError, "len", "string")}},
		{"unsupported value8", args{isFilter, []any{Cond{"test__gt": true}}}, want{fmt.Errorf(unsupportedValueError, "gt", "bool")}},
		{"unsupported value9", args{isFilter, []any{Cond{"test__gte": true}}}, want{fmt.Errorf(unsupportedValueError, "gte", "bool")}},
		{"unsupported value10", args{isFilter, []any{Cond{"test__lt": true}}}, want{fmt.Errorf(unsupportedValueError, "lt", "bool")}},
		{"unsupported value11", args{isFilter, []any{Cond{"test__lte": true}}}, want{fmt.Errorf(unsupportedValueError, "lte", "bool")}},
		{"unsupported value12", args{isFilter, []any{Cond{"test__len": true}}}, want{fmt.Errorf(unsupportedValueError, "len", "bool")}},
		{"unsupported value13", args{isFilter, []any{Cond{"test__gt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "slice")}},
		{"unsupported value14", args{isFilter, []any{Cond{"test__gte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "slice")}},
		{"unsupported value15", args{isFilter, []any{Cond{"test__lt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "slice")}},
		{"unsupported value16", args{isFilter, []any{Cond{"test__lte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "slice")}},
		{"unsupported value17", args{isFilter, []any{Cond{"test__len": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "slice")}},
		{"unsupported value18", args{isFilter, []any{Cond{"test__gt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "array")}},
		{"unsupported value19", args{isFilter, []any{Cond{"test__gte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "array")}},
		{"unsupported value20", args{isFilter, []any{Cond{"test__lt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "array")}},
		{"unsupported value21", args{isFilter, []any{Cond{"test__lte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "array")}},
		{"unsupported value22", args{isFilter, []any{Cond{"test__len": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "array")}},
		{"unsupported value23", args{isFilter, []any{Cond{"test__in": 1}}}, want{err: fmt.Errorf(unsupportedValueError, "in", "int")}},
		{"unsupported value25", args{isFilter, []any{Cond{"test__between": "test"}}}, want{err: fmt.Errorf(unsupportedValueError, "between", "string")}},
		{"unsupported value26", args{isFilter, []any{Cond{"test__contains": ""}}}, want{err: fmt.Errorf(unsupportedValueError, "contains", "blank string")}},
		{"unsupported value27", args{isFilter, []any{Cond{"test__contains": true}}}, want{err: fmt.Errorf(unsupportedValueError, "contains", "bool")}},
		{"unsupported value28", args{isFilter, []any{Cond{"test__contains": 0}}}, want{err: fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value29", args{isFilter, []any{Cond{"test__contains": 1}}}, want{err: fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value30", args{isFilter, []any{Cond{"test__contains": 1.0}}}, want{err: fmt.Errorf(unsupportedValueError, "contains", "float64")}},
		{"order key type", args{isFilter, []any{Cond{SortKey: []int{1, 2}, "1": "b", "2": "b"}}}, want{err: errors.New(orderKeyTypeError)}},
		{"order key len", args{isFilter, []any{Cond{SortKey: []string{"1"}, "1": "b", "2": "b"}}}, want{err: errors.New(orderKeyLenError)}},
		{"field lookup", args{isFilter, []any{Cond{"1__not__contains": "b"}}}, want{err: fmt.Errorf(fieldLookupError, "1__not__contains")}},
		{"unknown operator", args{isFilter, []any{Cond{"1__contain": "b"}}}, want{fmt.Errorf(unknownOperatorError, "contain")}},
		{"operator value len", args{isFilter, []any{Cond{"test__between": []int{}}}}, want{fmt.Errorf(operatorValueLenError, "between", 2)}},
		{"operator value len less", args{isFilter, []any{Cond{"test": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exact", 0)}},
		{"operator value len less1", args{isFilter, []any{Cond{"test__exclude": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exclude", 0)}},
		{"operator value len less2", args{isFilter, []any{Cond{"test__iexact": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "iexact", 0)}},
		{"operator value len less3", args{isFilter, []any{Cond{"test__in": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "in", 0)}},
		{"operator value len less4", args{isFilter, []any{Cond{"test__contains": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value len less5", args{isFilter, []any{Cond{"test__contains": [0]int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value type", args{isFilter, []any{Cond{"test__contains": []int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"operator value type1", args{isFilter, []any{Cond{"test__contains": [2]int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"not implementd operator", args{isFilter, []any{Cond{"test__unimplemented": 1}}}, want{fmt.Errorf(notImplementedOperatorError, "UNIMPLEMENTED")}},
		{"unsupported value", args{isFilter, []any{Cond{"test__exclude": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exclude", "map")}},
		{"unsupported value1", args{isFilter, []any{Cond{"test": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exact", "map")}},
		{"unsupported value2", args{isFilter, []any{Cond{"test__iexact": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "iexact", "map")}},
		{"unsupported value3", args{isFilter, []any{Cond{"test__gt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gt", "string")}},
		{"unsupported value4", args{isFilter, []any{Cond{"test__gte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gte", "string")}},
		{"unsupported value5", args{isFilter, []any{Cond{"test__lt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lt", "string")}},
		{"unsupported value6", args{isFilter, []any{Cond{"test__lte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lte", "string")}},
		{"unsupported value7", args{isFilter, []any{Cond{"test__len": "10"}}}, want{fmt.Errorf(unsupportedValueError, "len", "string")}},
		{"unsupported value8", args{isFilter, []any{Cond{"test__gt": true}}}, want{fmt.Errorf(unsupportedValueError, "gt", "bool")}},
		{"unsupported value9", args{isFilter, []any{Cond{"test__gte": true}}}, want{fmt.Errorf(unsupportedValueError, "gte", "bool")}},
		{"unsupported value10", args{isFilter, []any{Cond{"test__lt": true}}}, want{fmt.Errorf(unsupportedValueError, "lt", "bool")}},
		{"unsupported value11", args{isFilter, []any{Cond{"test__lte": true}}}, want{fmt.Errorf(unsupportedValueError, "lte", "bool")}},
		{"unsupported value12", args{isFilter, []any{Cond{"test__len": true}}}, want{fmt.Errorf(unsupportedValueError, "len", "bool")}},
		{"unsupported value13", args{isFilter, []any{Cond{"test__gt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "slice")}},
		{"unsupported value14", args{isFilter, []any{Cond{"test__gte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "slice")}},
		{"unsupported value15", args{isFilter, []any{Cond{"test__lt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "slice")}},
		{"unsupported value16", args{isFilter, []any{Cond{"test__lte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "slice")}},
		{"unsupported value17", args{isFilter, []any{Cond{"test__len": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "slice")}},
		{"unsupported value18", args{isFilter, []any{Cond{"test__gt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "array")}},
		{"unsupported value19", args{isFilter, []any{Cond{"test__gte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "array")}},
		{"unsupported value20", args{isFilter, []any{Cond{"test__lt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "array")}},
		{"unsupported value21", args{isFilter, []any{Cond{"test__lte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "array")}},
		{"unsupported value22", args{isFilter, []any{Cond{"test__len": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "array")}},
		{"unsupported value23", args{isFilter, []any{Cond{"test__in": 1}}}, want{fmt.Errorf(unsupportedValueError, "in", "int")}},
		{"unsupported value25", args{isFilter, []any{Cond{"test__between": "test"}}}, want{fmt.Errorf(unsupportedValueError, "between", "string")}},
		{"unsupported value26", args{isFilter, []any{Cond{"test__contains": ""}}}, want{fmt.Errorf(unsupportedValueError, "contains", "blank string")}},
		{"unsupported value27", args{isFilter, []any{Cond{"test__contains": true}}}, want{fmt.Errorf(unsupportedValueError, "contains", "bool")}},
		{"unsupported value28", args{isFilter, []any{Cond{"test__contains": 0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value29", args{isFilter, []any{Cond{"test__contains": 1}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value30", args{isFilter, []any{Cond{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "float64")}},
		{"order key type", args{isFilter, []any{AND{SortKey: []int{1, 2}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyTypeError)}},
		{"order key len", args{isFilter, []any{AND{SortKey: []string{"1"}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyLenError)}},
		{"field lookup", args{isFilter, []any{AND{"1__not__contains": "b"}}}, want{fmt.Errorf(fieldLookupError, "1__not__contains")}},
		{"unknown operator", args{isFilter, []any{AND{"1__contain": "b"}}}, want{fmt.Errorf(unknownOperatorError, "contain")}},
		{"operator value len", args{isFilter, []any{AND{"test__between": []int{}}}}, want{fmt.Errorf(operatorValueLenError, "between", 2)}},
		{"operator value len less", args{isFilter, []any{AND{"test": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exact", 0)}},
		{"operator value len less1", args{isFilter, []any{AND{"test__exclude": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exclude", 0)}},
		{"operator value len less2", args{isFilter, []any{AND{"test__iexact": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "iexact", 0)}},
		{"operator value len less3", args{isFilter, []any{AND{"test__in": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "in", 0)}},
		{"operator value len less4", args{isFilter, []any{AND{"test__contains": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value len less5", args{isFilter, []any{AND{"test__contains": [0]int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value type", args{isFilter, []any{AND{"test__contains": []int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"operator value type1", args{isFilter, []any{AND{"test__contains": [2]int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"not implementd operator", args{isFilter, []any{AND{"test__unimplemented": 1}}}, want{fmt.Errorf(notImplementedOperatorError, "UNIMPLEMENTED")}},
		{"unsupported value", args{isFilter, []any{AND{"test__exclude": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exclude", "map")}},
		{"unsupported value1", args{isFilter, []any{AND{"test": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exact", "map")}},
		{"unsupported value2", args{isFilter, []any{AND{"test__iexact": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "iexact", "map")}},
		{"unsupported value3", args{isFilter, []any{AND{"test__gt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gt", "string")}},
		{"unsupported value4", args{isFilter, []any{AND{"test__gte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gte", "string")}},
		{"unsupported value5", args{isFilter, []any{AND{"test__lt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lt", "string")}},
		{"unsupported value6", args{isFilter, []any{AND{"test__lte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lte", "string")}},
		{"unsupported value7", args{isFilter, []any{AND{"test__len": "10"}}}, want{fmt.Errorf(unsupportedValueError, "len", "string")}},
		{"unsupported value8", args{isFilter, []any{AND{"test__gt": true}}}, want{fmt.Errorf(unsupportedValueError, "gt", "bool")}},
		{"unsupported value9", args{isFilter, []any{AND{"test__gte": true}}}, want{fmt.Errorf(unsupportedValueError, "gte", "bool")}},
		{"unsupported value10", args{isFilter, []any{AND{"test__lt": true}}}, want{fmt.Errorf(unsupportedValueError, "lt", "bool")}},
		{"unsupported value11", args{isFilter, []any{AND{"test__lte": true}}}, want{fmt.Errorf(unsupportedValueError, "lte", "bool")}},
		{"unsupported value12", args{isFilter, []any{AND{"test__len": true}}}, want{fmt.Errorf(unsupportedValueError, "len", "bool")}},
		{"unsupported value13", args{isFilter, []any{AND{"test__gt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "slice")}},
		{"unsupported value14", args{isFilter, []any{AND{"test__gte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "slice")}},
		{"unsupported value15", args{isFilter, []any{AND{"test__lt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "slice")}},
		{"unsupported value16", args{isFilter, []any{AND{"test__lte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "slice")}},
		{"unsupported value17", args{isFilter, []any{AND{"test__len": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "slice")}},
		{"unsupported value18", args{isFilter, []any{AND{"test__gt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "array")}},
		{"unsupported value19", args{isFilter, []any{AND{"test__gte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "array")}},
		{"unsupported value20", args{isFilter, []any{AND{"test__lt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "array")}},
		{"unsupported value21", args{isFilter, []any{AND{"test__lte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "array")}},
		{"unsupported value22", args{isFilter, []any{AND{"test__len": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "array")}},
		{"unsupported value23", args{isFilter, []any{AND{"test__in": 1}}}, want{fmt.Errorf(unsupportedValueError, "in", "int")}},
		{"unsupported value25", args{isFilter, []any{AND{"test__between": "test"}}}, want{fmt.Errorf(unsupportedValueError, "between", "string")}},
		{"unsupported value26", args{isFilter, []any{AND{"test__contains": ""}}}, want{fmt.Errorf(unsupportedValueError, "contains", "blank string")}},
		{"unsupported value27", args{isFilter, []any{AND{"test__contains": true}}}, want{fmt.Errorf(unsupportedValueError, "contains", "bool")}},
		{"unsupported value28", args{isFilter, []any{AND{"test__contains": 0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value29", args{isFilter, []any{AND{"test__contains": 1}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value30", args{isFilter, []any{AND{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "float64")}},
		{"order key type", args{isFilter, []any{OR{SortKey: []int{1, 2}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyTypeError)}},
		{"order key len", args{isFilter, []any{OR{SortKey: []string{"1"}, "1": "b", "2": "b"}}}, want{errors.New(orderKeyLenError)}},
		{"field lookup", args{isFilter, []any{OR{"1__not__contains": "b"}}}, want{fmt.Errorf(fieldLookupError, "1__not__contains")}},
		{"unknown operator", args{isFilter, []any{OR{"1__contain": "b"}}}, want{fmt.Errorf(unknownOperatorError, "contain")}},
		{"operator value len", args{isFilter, []any{OR{"test__between": []int{}}}}, want{fmt.Errorf(operatorValueLenError, "between", 2)}},
		{"operator value len less", args{isFilter, []any{OR{"test": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exact", 0)}},
		{"operator value len less1", args{isFilter, []any{OR{"test__exclude": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "exclude", 0)}},
		{"operator value len less2", args{isFilter, []any{OR{"test__iexact": []string{}}}}, want{fmt.Errorf(operatorValueLenLessError, "iexact", 0)}},
		{"operator value len less3", args{isFilter, []any{OR{"test__in": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "in", 0)}},
		{"operator value len less4", args{isFilter, []any{OR{"test__contains": []int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value len less5", args{isFilter, []any{OR{"test__contains": [0]int{}}}}, want{fmt.Errorf(operatorValueLenLessError, "contains", 0)}},
		{"operator value type", args{isFilter, []any{OR{"test__contains": []int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"operator value type1", args{isFilter, []any{OR{"test__contains": [2]int{1, 2}}}}, want{fmt.Errorf(operatorValueTypeError, "contains")}},
		{"not implementd operator", args{isFilter, []any{OR{"test__unimplemented": 1}}}, want{fmt.Errorf(notImplementedOperatorError, "UNIMPLEMENTED")}},
		{"unsupported value", args{isFilter, []any{OR{"test__exclude": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exclude", "map")}},
		{"unsupported value1", args{isFilter, []any{OR{"test": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "exact", "map")}},
		{"unsupported value2", args{isFilter, []any{OR{"test__iexact": map[string]int{"1": 1}}}}, want{fmt.Errorf(unsupportedValueError, "iexact", "map")}},
		{"unsupported value3", args{isFilter, []any{OR{"test__gt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gt", "string")}},
		{"unsupported value4", args{isFilter, []any{OR{"test__gte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "gte", "string")}},
		{"unsupported value5", args{isFilter, []any{OR{"test__lt": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lt", "string")}},
		{"unsupported value6", args{isFilter, []any{OR{"test__lte": "10"}}}, want{fmt.Errorf(unsupportedValueError, "lte", "string")}},
		{"unsupported value7", args{isFilter, []any{OR{"test__len": "10"}}}, want{fmt.Errorf(unsupportedValueError, "len", "string")}},
		{"unsupported value8", args{isFilter, []any{OR{"test__gt": true}}}, want{fmt.Errorf(unsupportedValueError, "gt", "bool")}},
		{"unsupported value9", args{isFilter, []any{OR{"test__gte": true}}}, want{fmt.Errorf(unsupportedValueError, "gte", "bool")}},
		{"unsupported value10", args{isFilter, []any{OR{"test__lt": true}}}, want{fmt.Errorf(unsupportedValueError, "lt", "bool")}},
		{"unsupported value11", args{isFilter, []any{OR{"test__lte": true}}}, want{fmt.Errorf(unsupportedValueError, "lte", "bool")}},
		{"unsupported value12", args{isFilter, []any{OR{"test__len": true}}}, want{fmt.Errorf(unsupportedValueError, "len", "bool")}},
		{"unsupported value13", args{isFilter, []any{OR{"test__gt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "slice")}},
		{"unsupported value14", args{isFilter, []any{OR{"test__gte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "slice")}},
		{"unsupported value15", args{isFilter, []any{OR{"test__lt": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "slice")}},
		{"unsupported value16", args{isFilter, []any{OR{"test__lte": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "slice")}},
		{"unsupported value17", args{isFilter, []any{OR{"test__len": []int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "slice")}},
		{"unsupported value18", args{isFilter, []any{OR{"test__gt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gt", "array")}},
		{"unsupported value19", args{isFilter, []any{OR{"test__gte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "gte", "array")}},
		{"unsupported value20", args{isFilter, []any{OR{"test__lt": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lt", "array")}},
		{"unsupported value21", args{isFilter, []any{OR{"test__lte": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "lte", "array")}},
		{"unsupported value22", args{isFilter, []any{OR{"test__len": [1]int{1}}}}, want{fmt.Errorf(unsupportedValueError, "len", "array")}},
		{"unsupported value23", args{isFilter, []any{OR{"test__in": 1}}}, want{fmt.Errorf(unsupportedValueError, "in", "int")}},
		{"unsupported value25", args{isFilter, []any{OR{"test__between": "test"}}}, want{fmt.Errorf(unsupportedValueError, "between", "string")}},
		{"unsupported value26", args{isFilter, []any{OR{"test__contains": ""}}}, want{fmt.Errorf(unsupportedValueError, "contains", "blank string")}},
		{"unsupported value27", args{isFilter, []any{OR{"test__contains": true}}}, want{fmt.Errorf(unsupportedValueError, "contains", "bool")}},
		{"unsupported value28", args{isFilter, []any{OR{"test__contains": 0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value29", args{isFilter, []any{OR{"test__contains": 1}}}, want{fmt.Errorf(unsupportedValueError, "contains", "int")}},
		{"unsupported value30", args{isFilter, []any{OR{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedValueError, "contains", "float64")}},
		{"unsupported supported filter type0", args{isFilter, []any{map[string]any{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]interface {}")}},
		{"unsupported supported filter type1", args{isFilter, []any{map[string]any{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]interface {}")}},
		{"unsupported supported filter type2", args{isFilter, []any{map[string]int{"test__contains": 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]int")}},
		{"unsupported supported filter type3", args{isFilter, []any{map[string]string{"test__contains": "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]string")}},
		{"unsupported supported filter type4", args{isFilter, []any{map[string]bool{"test__contains": true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]bool")}},
		{"unsupported supported filter type5", args{isFilter, []any{map[string]float64{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]float64")}},
		{"unsupported supported filter type6", args{isFilter, []any{map[int]any{1: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]interface {}")}},
		{"unsupported supported filter type7", args{isFilter, []any{map[int]string{1: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]string")}},
		{"unsupported supported filter type8", args{isFilter, []any{map[int]bool{1: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]bool")}},
		{"unsupported supported filter type9", args{isFilter, []any{map[int]float64{1: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]float64")}},
		{"unsupported supported filter type10", args{isFilter, []any{map[bool]any{true: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]interface {}")}},
		{"unsupported supported filter type11", args{isFilter, []any{map[bool]string{true: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]string")}},
		{"unsupported supported filter type12", args{isFilter, []any{map[bool]bool{true: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]bool")}},
		{"unsupported supported filter type13", args{isFilter, []any{map[bool]float64{true: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]float64")}},
		{"unsupported supported filter type14", args{isFilter, []any{map[float64]any{1.0: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]interface {}")}},
		{"unsupported supported filter type15", args{isFilter, []any{map[float64]string{1.0: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]string")}},
		{"unsupported supported filter type16", args{isFilter, []any{map[float64]bool{1.0: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]bool")}},
		{"unsupported supported filter type17", args{isFilter, []any{map[float64]float64{1.0: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]float64")}},
		{"unsupported supported filter type18", args{isFilter, []any{map[any]any{"test__contains": 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]interface {}")}},
		{"unsupported supported filter type19", args{isFilter, []any{map[any]string{"test__contains": "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]string")}},
		{"unsupported supported filter type20", args{isFilter, []any{map[any]bool{"test__contains": true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]bool")}},
		{"unsupported supported filter type21", args{isFilter, []any{map[any]float64{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]float64")}},
		{"unsupported supported filter type22", args{isFilter, []any{Cond{}, []any{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]interface {}")}},
		{"unsupported supported filter type23", args{isFilter, []any{Cond{}, []any{"test__contains"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]interface {}")}},
		{"unsupported supported filter type24", args{isFilter, []any{Cond{}, 1}}, want{fmt.Errorf(unsupportedFilterTypeError, "int")}},
		{"unsupported supported filter type25", args{isFilter, []any{Cond{}, "test__contains"}}, want{fmt.Errorf(unsupportedFilterTypeError, "string")}},
		{"unsupported supported filter type26", args{isFilter, []any{Cond{}, 1.0}}, want{fmt.Errorf(unsupportedFilterTypeError, "float64")}},
		{"unsupported supported filter type27", args{isFilter, []any{Cond{}, true}}, want{fmt.Errorf(unsupportedFilterTypeError, "bool")}},
		{"unsupported supported filter type28", args{isFilter, []any{Cond{}, false}}, want{fmt.Errorf(unsupportedFilterTypeError, "bool")}},
		{"unsupported supported filter type29", args{isFilter, []any{Cond{}, nil}}, want{fmt.Errorf(unsupportedFilterTypeError, "nil")}},
		{"unsupported supported filter type30", args{isFilter, []any{[]int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]int")}},
		{"unsupported supported filter type31", args{isFilter, []any{[]string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]string")}},
		{"unsupported supported filter type32", args{isFilter, []any{[]bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]bool")}},
		{"unsupported supported filter type33", args{isFilter, []any{[]float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]float64")}},
		{"unsupported supported filter type34", args{isFilter, []any{[1]int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]int")}},
		{"unsupported supported filter type35", args{isFilter, []any{[1]string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]string")}},
		{"unsupported supported filter type36", args{isFilter, []any{[1]bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]bool")}},
		{"unsupported supported filter type37", args{isFilter, []any{[1]float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]float64")}},
		{"unsupported supported filter type38", args{isFilter, []any{[2]int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]int")}},
		{"unsupported supported filter type39", args{isFilter, []any{[2]string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]string")}},
		{"unsupported supported filter type40", args{isFilter, []any{[2]bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]bool")}},
		{"unsupported supported filter type41", args{isFilter, []any{[2]float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]float64")}},
		{"unsupported supported filter type0", args{isFilter, []any{Cond{}, map[string]any{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]interface {}")}},
		{"unsupported supported filter type1", args{isFilter, []any{Cond{}, map[string]any{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]interface {}")}},
		{"unsupported supported filter type2", args{isFilter, []any{Cond{}, map[string]int{"test__contains": 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]int")}},
		{"unsupported supported filter type3", args{isFilter, []any{Cond{}, map[string]string{"test__contains": "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]string")}},
		{"unsupported supported filter type4", args{isFilter, []any{Cond{}, map[string]bool{"test__contains": true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]bool")}},
		{"unsupported supported filter type5", args{isFilter, []any{Cond{}, map[string]float64{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[string]float64")}},
		{"unsupported supported filter type6", args{isFilter, []any{Cond{}, map[int]any{1: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]interface {}")}},
		{"unsupported supported filter type7", args{isFilter, []any{Cond{}, map[int]string{1: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]string")}},
		{"unsupported supported filter type8", args{isFilter, []any{Cond{}, map[int]bool{1: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]bool")}},
		{"unsupported supported filter type9", args{isFilter, []any{Cond{}, map[int]float64{1: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[int]float64")}},
		{"unsupported supported filter type10", args{isFilter, []any{Cond{}, map[bool]any{true: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]interface {}")}},
		{"unsupported supported filter type11", args{isFilter, []any{Cond{}, map[bool]string{true: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]string")}},
		{"unsupported supported filter type12", args{isFilter, []any{Cond{}, map[bool]bool{true: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]bool")}},
		{"unsupported supported filter type13", args{isFilter, []any{Cond{}, map[bool]float64{true: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[bool]float64")}},
		{"unsupported supported filter type14", args{isFilter, []any{Cond{}, map[float64]any{1.0: 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]interface {}")}},
		{"unsupported supported filter type15", args{isFilter, []any{Cond{}, map[float64]string{1.0: "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]string")}},
		{"unsupported supported filter type16", args{isFilter, []any{Cond{}, map[float64]bool{1.0: true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]bool")}},
		{"unsupported supported filter type17", args{isFilter, []any{Cond{}, map[float64]float64{1.0: 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[float64]float64")}},
		{"unsupported supported filter type18", args{isFilter, []any{Cond{}, map[any]any{"test__contains": 1}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]interface {}")}},
		{"unsupported supported filter type19", args{isFilter, []any{Cond{}, map[any]string{"test__contains": "1"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]string")}},
		{"unsupported supported filter type20", args{isFilter, []any{Cond{}, map[any]bool{"test__contains": true}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]bool")}},
		{"unsupported supported filter type21", args{isFilter, []any{Cond{}, map[any]float64{"test__contains": 1.0}}}, want{fmt.Errorf(unsupportedFilterTypeError, "map[interface {}]float64")}},
		{"unsupported supported filter type22", args{isFilter, []any{Cond{}, []any{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]interface {}")}},
		{"unsupported supported filter type23", args{isFilter, []any{Cond{}, []any{"test__contains"}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]interface {}")}},
		{"unsupported supported filter type24", args{isFilter, []any{Cond{}, 1}}, want{fmt.Errorf(unsupportedFilterTypeError, "int")}},
		{"unsupported supported filter type25", args{isFilter, []any{Cond{}, "test__contains"}}, want{fmt.Errorf(unsupportedFilterTypeError, "string")}},
		{"unsupported supported filter type26", args{isFilter, []any{Cond{}, 1.0}}, want{fmt.Errorf(unsupportedFilterTypeError, "float64")}},
		{"unsupported supported filter type27", args{isFilter, []any{Cond{}, true}}, want{fmt.Errorf(unsupportedFilterTypeError, "bool")}},
		{"unsupported supported filter type28", args{isFilter, []any{Cond{}, false}}, want{fmt.Errorf(unsupportedFilterTypeError, "bool")}},
		{"unsupported supported filter type29", args{isFilter, []any{Cond{}, nil}}, want{fmt.Errorf(unsupportedFilterTypeError, "nil")}},
		{"unsupported supported filter type30", args{isFilter, []any{Cond{}, []int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]int")}},
		{"unsupported supported filter type31", args{isFilter, []any{Cond{}, []string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]string")}},
		{"unsupported supported filter type32", args{isFilter, []any{Cond{}, []bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]bool")}},
		{"unsupported supported filter type33", args{isFilter, []any{Cond{}, []float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[]float64")}},
		{"unsupported supported filter type34", args{isFilter, []any{Cond{}, [1]int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]int")}},
		{"unsupported supported filter type35", args{isFilter, []any{Cond{}, [1]string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]string")}},
		{"unsupported supported filter type36", args{isFilter, []any{Cond{}, [1]bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]bool")}},
		{"unsupported supported filter type37", args{isFilter, []any{Cond{}, [1]float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[1]float64")}},
		{"unsupported supported filter type38", args{isFilter, []any{Cond{}, [2]int{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]int")}},
		{"unsupported supported filter type39", args{isFilter, []any{Cond{}, [2]string{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]string")}},
		{"unsupported supported filter type40", args{isFilter, []any{Cond{}, [2]bool{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]bool")}},
		{"unsupported supported filter type41", args{isFilter, []any{Cond{}, [2]float64{}}}, want{fmt.Errorf(unsupportedFilterTypeError, "[2]float64")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p = p.FilterToSQL(tt.args.isNot, tt.args.filter...)
			p.GetQuerySet()

			if p.Error() == nil && !errors.Is(p.Error(), tt.want.err) {
				t.Errorf("TestFilterError SQL Occur Error -> error: %s", p.Error())
				t.Errorf("TestFilterError SQL Occur Error -> want: %s", tt.want.err)
			}

			if p.Error() != nil && p.Error().Error() != tt.want.err.Error() {
				t.Errorf("TestFilterError SQL Occur Error -> error: %s", p.Error())
				t.Errorf("TestFilterError SQL Occur Error -> want: %s", tt.want.err)
			}
		})
	}

}

func TestWhere(t *testing.T) {
	type args struct {
		cond string
		args []any
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
		{"one", args{"`test` = ?", []any{1}}, want{" WHERE `test` = ?", []any{1}}},
		{"two", args{"`test` = ? AND test2 = ?", []any{1, 2}}, want{" WHERE `test` = ? AND test2 = ?", []any{1, 2}}},
		{"three", args{"test = ? AND `test2` = ? AND test3 = ?", []any{1, 2, 3}}, want{" WHERE test = ? AND `test2` = ? AND test3 = ?", []any{1, 2, 3}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.WhereToSQL(tt.args.cond, tt.args.args...)

			sql, sqlArgs := p.GetQuerySet()

			if sql != tt.want.sql {
				t.Errorf("TestWhere SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestWhere SQL Gen Error -> want:%v", tt.want.sql)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestWhere Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestWhere Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestWhere Args Length Error -> want:%+v", tt.want.args)
			}

			for i, a := range sqlArgs {
				if !reflect.DeepEqual(a, tt.want.args[i]) {
					t.Errorf("TestWhere Arg Value Error -> args:%+v", sqlArgs)
					t.Errorf("TestWhere Arg Value Error -> want:%+v", tt.want.args)
				}
			}
		})
	}
}

func TestWhereError(t *testing.T) {
	type args struct {
		cond string
		args []any
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
		{"args num mismatch", args{"test = ? AND test2 = ? AND test3 = ?", []any{1, 2}}, want{"", []any{}, errors.New(argsLenError)}},
		{"args num mismatch1", args{"test = ? AND test2 = ? AND test3 = ?", []any{1, 2, 3, 4}}, want{"", []any{}, errors.New(argsLenError)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())
			p.WhereToSQL(tt.args.cond, tt.args.args...)

			sql, sqlArgs := p.GetQuerySet()

			if !errors.Is(p.Error(), tt.want.err) && p.Error().Error() != tt.want.err.Error() {
				t.Errorf("TestWhereError SQL Occur Error -> error:%+v", p.Error())
				t.Errorf("TestWhereError SQL Occur Error -> want:%+v", tt.want.err)
			}

			if sql != tt.want.sql {
				t.Errorf("TestWhereError SQL Gen Error -> sql :%v", sql)
				t.Errorf("TestWhereError SQL Gen Error -> want:%v", tt.want.sql)
			}

			if len(sqlArgs) != len(tt.want.args) {
				t.Errorf("TestWhereError Args Length Error -> len:%+v, want:%+v", len(sqlArgs), len(tt.want.args))
				t.Errorf("TestWhereError Args Length Error -> args:%+v", sqlArgs)
				t.Errorf("TestWhereError Args Length Error -> want:%+v", tt.want.args)
			}

			for i, a := range sqlArgs {
				if !reflect.DeepEqual(a, tt.want.args[i]) {
					t.Errorf("TestWhereError Arg Value Error -> args:%+v", sqlArgs)
					t.Errorf("TestWhereError Arg Value Error -> want:%+v", tt.want.args)
				}
			}
		})
	}
}

func TestFilterAndWhereConflict(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())
	p.WhereToSQL("test = ?", 1)
	p.FilterToSQL(notNot, Cond{"test": 1})
	if p.Error() == nil {
		t.Error("Test Where Filter Conflict Not Occur Error")
	} else if p.Error().Error() != fmt.Errorf(filterOrWhereError, "Filter").Error() {
		t.Errorf("TestFilterAndExcludeConflict not working as expected")
	}

	p = NewQuerySet(mysqlOp.NewOperator())
	p.WhereToSQL("test = ?", 1)
	p.FilterToSQL(isNot, Cond{"test": 1})
	if p.Error() == nil {
		t.Error("Test Where Exclude Conflict Not Occur Error")
	} else if p.Error().Error() != fmt.Errorf(filterOrWhereError, "Exclude").Error() {
		t.Errorf("TestFilterAndExcludeConflict not working as expected")
	}

	p = NewQuerySet(mysqlOp.NewOperator())
	p.FilterToSQL(notNot, Cond{"test": 1})
	p.WhereToSQL("test = ?", 1)
	if p.Error() == nil {
		t.Error("Test Filter Where Conflict Not Occur Error")
		return
	} else if p.Error().Error() != fmt.Errorf(filterOrWhereError, "Filter").Error() {
		t.Errorf("TestFilterAndWhereConflict SQL Occur Error -> error:%+v", p.Error().Error())
		t.Errorf("TestFilterAndWhereConflict SQL Occur Error -> want:%+v", fmt.Errorf(filterOrWhereError, "Filter").Error())
	}
}

func TestSelect(t *testing.T) {
	type args struct {
		selects any
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
		{"string", args{"test, test2 as test3"}, want{"test, test2 as test3"}},
		{"zero slice", args{[]string{}}, want{"*"}},
		{"one slice", args{[]string{"test"}}, want{"`test`"}},
		{"two slice", args{[]string{"test", "test2"}}, want{"`test`, `test2`"}},
		{"three slice", args{[]string{"test", "test2", "test3"}}, want{"`test`, `test2`, `test3`"}},
		{"four slice", args{[]string{"test", "test2", "test3", "test4"}}, want{"`test`, `test2`, `test3`, `test4`"}},
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

			if sql != tt.want.sql {
				t.Errorf("TestSelect SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestSelect SQL Gen Error -> want: %v", tt.want.sql)
			}
		})
	}
}

func TestSelectError(t *testing.T) {
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
		{"array", args{[1]string{"test"}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{123}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{[]int{1, 2}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[string]int{"test": 1}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[int]string{1: "test"}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[bool]string{true: "test"}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[string]bool{"test": true}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": "test"}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1.0}}, want{sql: "`test`", err: errors.New(paramTypeError)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())

			sql := p.GetSelectSQL()
			if sql != "*" {
				t.Errorf("TestSelectError SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestSelectError SQL Gen Error -> want: %v", "*")
			}

			p.SelectToSQL(tt.args.selects)
			sql = p.GetSelectSQL()

			if p.Error() != nil {
				if p.Error().Error() != tt.want.err.Error() {
					t.Errorf("TestSelectError SQL Occur Error -> error: %+v", p.Error())
				}

				return
			}

			if sql != tt.want.sql {
				t.Errorf("TestSelectError SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestSelectError SQL Gen Error -> want: %v", tt.want.sql)
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
		{"one_asc ", args{[]string{}}, want{""}},
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

func TestOrderByError(t *testing.T) {
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
		{"array", args{[1]string{"test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{123}, want{"", errors.New(paramTypeError)}},
		{"array", args{[]int{1, 2}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]int{"test": 1}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[int]string{1: "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[bool]string{true: "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]bool{"test": true}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1.0}}, want{"", errors.New(paramTypeError)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())

			p.OrderByToSQL(tt.args.selects)
			sql := p.GetOrderBySQL()

			if p.Error() != nil {
				if p.Error().Error() != tt.want.err.Error() {
					t.Errorf("TestOrderByError SQL Occur Error -> error: %+v", p.Error())
				}
				return
			}

			if sql != tt.want.sql {
				t.Errorf("TestOrderByError SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestOrderByError SQL Gen Error -> want: %v", tt.want.sql)
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

func TestGroupByError(t *testing.T) {
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
		{"array", args{[1]string{"test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{123}, want{"", errors.New(paramTypeError)}},
		{"array", args{[]int{1, 2}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]int{"test": 1}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[int]string{1: "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[bool]string{true: "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]bool{"test": true}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": "test"}}, want{"", errors.New(paramTypeError)}},
		{"array", args{map[string]any{"test": 1.0}}, want{"", errors.New(paramTypeError)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewQuerySet(mysqlOp.NewOperator())

			p.GroupByToSQL(tt.args.selects)
			sql := p.GetGroupBySQL()

			if p.Error() != nil {
				if p.Error().Error() != tt.want.err.Error() {
					t.Errorf("TestGroupByError SQL Occur Error -> error: %+v", p.Error())
				}
				return
			}

			if sql != tt.want.sql {
				t.Errorf("TestGroupByError SQL Gen Error -> sql : %v", sql)
				t.Errorf("TestGroupByError SQL Gen Error -> want: %v", tt.want.sql)
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

func TestFilterResetAndError(t *testing.T) {
	p := NewQuerySet(mysqlOp.NewOperator())

	// Create an error
	p.SelectToSQL("test")
	p.FilterToSQL(notNot, Cond{"test": 1})
	p.WhereToSQL("test1 = ?", 1)
	p.OrderByToSQL("test")
	p.GroupByToSQL("test")
	p.HavingToSQL("test = ?", 1)
	p.LimitToSQL(10, 1)

	// Reset should clear the error
	p.Reset()
	if p.Error() != nil {
		t.Errorf("Error should be nil after Reset, got: %v", p.Error())
	}
	if p.GetSelectSQL() != "*" {
		t.Errorf("SelectToSQL should be reset to default")
	}
	if p.GetOrderBySQL() != "" {
		t.Errorf("OrderByToSQL should be reset to default")
	}
	if p.GetGroupBySQL() != "" {
		t.Errorf("GroupByToSQL should be reset to default")
	}
	if sql, args := p.GetHavingSQL(); sql != "" || len(args) != 0 {
		t.Errorf("HavingToSQL should be reset to default")
	}
	if p.GetLimitSQL() != "" {
		t.Errorf("LimitToSQL should be reset to default")
	}

	// After reset, functions should work properly
	p.FilterToSQL(notNot, Cond{"test": 1})
	sql, args := p.GetQuerySet()

	if sql != " WHERE (`test` = ?)" || len(args) != 1 || args[0] != 1 {
		t.Errorf("FilterToSQL not working after Reset")
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
	p.FilterToSQL(notNot, Cond{"test": 1})
	if !p.hasCalled(callFilter) {
		t.Errorf("callFilter flag should be set")
	}

	// Reset and test exclude flag
	p.Reset()
	p.FilterToSQL(isNot, Cond{"test": 1})
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
