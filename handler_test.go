package norm

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	mysqlOp "github.com/leisurelicht/norm/operator/mysql"
	"github.com/leisurelicht/norm/test"
)

const (
	mysqlAddress = "root:123456@tcp(127.0.0.1:6033)/test?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
)

func TestCharacterEncoding(t *testing.T) {
	db := sqlx.NewMysql(mysqlAddress)
	var variables []struct {
		Variable string `db:"Variable_name"`
		Value    string `db:"Value"`
	}

	err := db.QueryRows(&variables, "SHOW VARIABLES WHERE Variable_name IN ('character_set_client', 'character_set_connection', 'character_set_results')")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range variables {
		if v.Value != "utf8mb4" {
			t.Errorf("Expected utf8mb4 for %s, got %s", v.Variable, v.Value)
		}
	}
}

func TestQuery(t *testing.T) {
	ctl := NewController(sqlx.NewMysql(mysqlAddress), mysqlOp.NewOperator(), test.Source{})

	if num, err := ctl(nil).Count(); err != nil {
		t.Error(err)
	} else if num != 8 {
		t.Errorf("expect 8 but got %d", num)
	}

	if res, err := ctl(nil).Filter(Cond{"id": 11}).FindOne(); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(res, map[string]any{}) {
		t.Error("expect not nil")
	} else if res["id"].(int64) != int64(11) {
		t.Errorf("expect 11 but got %d", res["id"])
	}

	if res, err := ctl(nil).Filter(Cond{"is_deleted": false}).FindOne(); err != nil {
		t.Error(err)
	} else if res["name"].(string) != "Acfun" {
		t.Errorf("expect Acfun but got %s", res["name"])
	}

	created, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
	updated, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
	source1 := test.Source{Id: 11, Name: "Acfun", Type: 1, Description: "A 站", IsDeleted: false, CreateTime: created, UpdateTime: updated}
	source2 := test.Source{}

	if err := ctl(nil).Filter(Cond{"id": 11}).FindOneModel(&source2); err != nil {
		t.Error(err)
	} else if source1.Id != source2.Id {
		t.Errorf("expect 11 but got %d", source2.Id)
	} else if source1.Name != source2.Name {
		t.Errorf("expect Acfun but got %s", source2.Name)
	} else if source1.Type != source2.Type {
		t.Errorf("expect 1 but got %d", source2.Type)
	} else if source1.Description != source2.Description {
		t.Errorf("expect A 站 but got %s", source2.Description)
	} else if source1.IsDeleted != source2.IsDeleted {
		t.Errorf("expect false but got %t", source2.IsDeleted)
	} else if source1.CreateTime.Format("2006-01-02 15:04:05") != source2.CreateTime.Format("2006-01-02 15:04:05") {
		t.Errorf("expect %s but got %s", source1.CreateTime, source2.CreateTime)
	} else if source1.UpdateTime.Format("2006-01-02 15:04:05") != source2.UpdateTime.Format("2006-01-02 15:04:05") {
		t.Errorf("expect %s but got %s", source1.UpdateTime, source2.UpdateTime)
	}

	if res, err := ctl(nil).Filter(Cond{"id": 11}, OR{"id": 12}).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 2 {
		t.Errorf("expect 2 but got %d\ngot res: %+v", len(res), res)
	} else if res[0]["id"].(int64) != int64(11) {
		t.Errorf("expect 11 but got %d", res[0]["id"])
	} else if res[1]["id"].(int64) != int64(12) {
		t.Errorf("expect 12 but got %d", res[1]["id"])
	}

	if res, err := ctl(nil).Filter(Cond{"id": 11}, AND{"id": 12}).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 0 {
		t.Errorf("expect 0 but got %d\ngot res: %+v", len(res), res)
	}

	if res, err := ctl(nil).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 6 {
		t.Errorf("expect 6 but got %d\ngot res: %+v", len(res), res)
	} else if res[0]["id"].(int64) != 11 {
		t.Errorf("expect 11 but got %d", res[0]["id"])
	} else if res[len(res)-1]["id"].(int64) != 23 {
		t.Errorf("expect 4 but got %d", res[2]["id"])
	}

}

func TestHandlerError(t *testing.T) {
	ctl := NewController(sqlx.NewMysql(mysqlAddress), mysqlOp.NewOperator(), test.Source{})

	if _, err := ctl(nil).Filter(Cond{}).Where("").FindOne(); err != nil && err.Error() != fmt.Sprintf(filterOrWhereError, "Filter") {
		t.Errorf("expect nil but got %v", err)
	}

	if _, err := ctl(nil).Exclude(Cond{}).Where("").FindOne(); err != nil && err.Error() != fmt.Sprintf(filterOrWhereError, "Exclude") {
		t.Errorf("expect nil but got %v", err)
	}

	// FindOne unsupported operations
	if res, err := ctl(nil).GroupBy("").FindOne(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[GroupBy] not supported for FindOne" {
		t.Error(err)
	}

	if res, err := ctl(nil).GroupBy("").Select("").FindOne(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[Select] not supported for FindOne" {
		t.Error(err)
	}

	// FindAll unsupported operations
	if res, err := ctl(nil).GroupBy("").FindAll(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[GroupBy] not supported for FindAll" {
		t.Error(err)
	}

	if res, err := ctl(nil).GroupBy("").Select("").FindAll(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[Select] not supported for FindAll" {
		t.Error(err)
	}

	// Insert unsupported operations
	if id, err := ctl(nil).Filter(Cond{}).Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(nil).Filter(Cond{}).Where("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(nil).Filter(Cond{}).Where("").OrderBy("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(nil).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy GroupBy] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(nil).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Select("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where Select OrderBy GroupBy] not supported for Insert" {
		t.Error(err)
	}

	// Update unsupported operations
	if num, err := ctl(nil).GroupBy("").Update(map[string]any{}); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Update" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").Update(map[string]any{}); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy] not supported for Update" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").OrderBy("").Update(map[string]any{}); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy OrderBy] not supported for Update" {
		t.Error(err)
	}

	// Remove unsupported operations
	if num, err := ctl(nil).GroupBy("").Remove(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Remove" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").Remove(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy] not supported for Remove" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").OrderBy("").Remove(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy OrderBy] not supported for Remove" {
		t.Error(err)
	}

	// Delete unsupported operations
	if num, err := ctl(nil).GroupBy("").Delete(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Delete" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").Delete(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy Select] not supported for Delete" {
		t.Error(err)
	}

	if num, err := ctl(nil).GroupBy("").Select("").OrderBy("").Delete(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy Select OrderBy] not supported for Delete" {
		t.Error(err)
	}

	// send not exist columns to Select
	if err := ctl(nil).Select([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).Select([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).Select([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns to OrderBy
	if err := ctl(nil).OrderBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).OrderBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).OrderBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns to GroupBy
	if err := ctl(nil).GroupBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).GroupBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).GroupBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns and display last error
	if err := ctl(nil).GroupBy([]string{"test"}).OrderBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).OrderBy([]string{"age"}).GroupBy([]string{"test"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).OrderBy([]string{"age", "happy"}).GroupBy([]string{"test", "test2"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test; test2] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(nil).OrderBy([]string{"age", "happy", "damnit"}).GroupBy([]string{"test", "test2", "test3"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test; test2; test3] not exist" {
			t.Error(err)
		}
	}

	// test reset
	cli := ctl(nil)
	_ = cli.Filter().FindOneModel(&test.Source{})

	if _, err := cli.Insert(map[string]any{}); err != nil && err.Error() != "[Filter] not supported for Insert" {
		t.Error(err)
	}

	cli = cli.Reset()

	if _, err := cli.Insert(map[string]any{}); err != nil && err.Error() != "insert data is empty" {
		t.Error(err)
	}

	if _, err := cli.Insert(map[string]any{"id": 1000, "name": "rest", "description": "test rest"}); err != nil {
		t.Error(err)
	}

	if _, err := cli.Filter(Cond{"id": 1000}).Remove(); err != nil {
		t.Error(err)
	}

	if res, err := cli.Filter(Cond{"id": 1000}).FindOne(); err != nil {
		t.Error(err)
	} else if len(res) != 0 {
		t.Errorf("expect 0 but got %d", len(res))
		t.Errorf("expect empty but got %+v", res)
	}

}
