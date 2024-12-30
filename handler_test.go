package norm

import (
	mysqlOp "github.com/leisurelicht/norm/operator/mysql"
	"github.com/leisurelicht/norm/test"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"reflect"
	"testing"
	"time"
)

const (
	mysqlAddress = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&loc=Asia%2FShanghai&parseTime=true"
)

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
		t.Errorf("expect 2 but got %d", len(res))
	} else if res[0]["id"].(int64) != int64(11) {
		t.Errorf("expect 11 but got %d", res[0]["id"])
	} else if res[1]["id"].(int64) != int64(12) {
		t.Errorf("expect 12 but got %d", res[1]["id"])
	}

	if res, err := ctl(nil).Filter(Cond{"id": 11}, AND{"id": 12}).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 0 {
		t.Errorf("expect 0 but got %d", len(res))
	}

	if res, err := ctl(nil).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 6 {
		t.Errorf("expect 6 but got %d", len(res))
	} else if res[0]["id"].(int64) != 11 {
		t.Errorf("expect 11 but got %d", res[0]["id"])
	} else if res[len(res)-1]["id"].(int64) != 23 {
		t.Errorf("expect 4 but got %d", res[2]["id"])
	}

}
