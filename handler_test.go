package norm

import (
	"context"
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
	} else if num != 15 {
		t.Errorf("Count error: Expect [count] 15 but got %d", num)
	}

	ctx := context.Background()

	if exist, err := ctl(ctx).Filter(Cond{"id": 11}).Exist(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("Exist error: Expect [exist] but got [not exist]")
	} else if exist, err := ctl(ctx).Filter(Cond{"id": 12345}).Exist(); err != nil {
		t.Error(err)
	} else if exist {
		t.Error("Exist error: Expect [not exist] but got [exist]")
	}

	if res, err := ctl(ctx).Filter(Cond{"id": 11}).FindOne(); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(res, map[string]any{}) {
		t.Error("Expect [not nil]")
	} else if res["id"].(int64) != int64(11) {
		t.Errorf("Expect [ID] 11 but got %d", res["id"])
	}

	if res, err := ctl(ctx).Filter(Cond{"is_deleted": false}).FindOne(); err != nil {
		t.Error(err)
	} else if res["name"].(string) != "Acfun" {
		t.Errorf("Expect [Name] Acfun but got %s", res["name"])
	}

	created, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
	updated, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
	source1 := test.Source{Id: 11, Name: "Acfun", Type: 1, Description: "A 站", IsDeleted: false, CreateTime: created, UpdateTime: updated}
	source2 := test.Source{}

	if err := ctl(ctx).Filter(Cond{"id": 11}).FindOneModel(&source2); err != nil {
		t.Error(err)
	} else if source1.Id != source2.Id {
		t.Errorf("Expect [ID] 11 but got %d", source2.Id)
	} else if source1.Name != source2.Name {
		t.Errorf("Expect [Name] Acfun but got %s", source2.Name)
	} else if source1.Type != source2.Type {
		t.Errorf("Expect [Type] 1 but got %d", source2.Type)
	} else if source1.Description != source2.Description {
		t.Errorf("Expect [Description] A 站 but got %s", source2.Description)
	} else if source1.IsDeleted != source2.IsDeleted {
		t.Errorf("Expect [IsDeleted] false but got %t", source2.IsDeleted)
	} else if source1.CreateTime.Format("2006-01-02 15:04:05") != source2.CreateTime.Format("2006-01-02 15:04:05") {
		t.Errorf("Expect [CreateTime] %s but got %s", source1.CreateTime, source2.CreateTime)
	} else if source1.UpdateTime.Format("2006-01-02 15:04:05") != source2.UpdateTime.Format("2006-01-02 15:04:05") {
		t.Errorf("Expect [Updatetime] %s but got %s", source1.UpdateTime, source2.UpdateTime)
	}

	if res, err := ctl(ctx).Filter(Cond{"id": 11}, OR{"id": 12}).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 2 {
		t.Errorf("Expect [Count] 2 but got %d\ngot res: %+v", len(res), res)
	} else if res[0]["id"].(int64) != int64(11) {
		t.Errorf("Expect [ID] 11 but got %d", res[0]["id"])
	} else if res[1]["id"].(int64) != int64(12) {
		t.Errorf("Expect [ID] 12 but got %d", res[1]["id"])
	}

	sources := []test.Source{}
	if err := ctl(ctx).Filter(Cond{"id": 11}, AND{"id": 12}).FindAllModel(&sources); err != nil {
		t.Error(err)
	} else if len(sources) != 0 {
		t.Errorf("Expect [count] 0 but got %d\ngot res: %+v", len(sources), sources)
	}

	if res, err := ctl(ctx).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 9 {
		t.Errorf("Expect [count] 9 but got %d\ngot res: %+v", len(res), res)
	} else if res[0]["id"].(int64) != 11 {
		t.Errorf("Expect [ID] 11 but got %d", res[0]["id"])
	} else if res[len(res)-1]["id"].(int64) != 53 {
		t.Errorf("Expect [ID] 53 but got %d", res[len(res)-1]["id"])
	}

	if res, err := ctl(ctx).Filter(Cond{"id": 12345}).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 0 {
		t.Errorf("Expect [count] 0 but got %d\ngot res: %+v", len(res), res)
	} else if reflect.DeepEqual(res, map[string]any{}) {
		t.Error("Expect [empty]")
	}

	// test multiple conditions for contains
	if res, err := ctl(ctx).Filter(Cond{"name__contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll(); err != nil {
		t.Error(err)
	} else if len(res) != 6 {
		t.Errorf("Expect [count] 6 but got %d\ngot res: %+v", len(res), res)
	}

	// test not contains and exclude contains, they should be return same result
	resNotContains, err := ctl(ctx).Filter(Cond{"name__not_contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll()
	if err != nil {
		t.Error(err)
	} else if len(resNotContains) != 9 {
		t.Errorf("Expect [count] 9 but got %d\ngot res: %+v", len(resNotContains), resNotContains)
	}

	resExclude, err := ctl(ctx).Exclude(Cond{"name__contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll()
	if err != nil {
		t.Error(err)
	} else if len(resExclude) != 9 {
		t.Errorf("Expect [count] 9 but got %d\ngot res: %+v", len(resExclude), resExclude)
	}

	for i, v := range resExclude {
		if v["id"] != resNotContains[i]["id"] {
			t.Errorf("Expect [not equal] but \ngot: resExclude: %+v\ngot: resNotContains: %+v", v, resNotContains[i])
		}
	}

	// test Select
	selectSources := []test.Source{}
	if err := ctl(ctx).Select([]string{"id", "name"}).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAllModel(&selectSources); err != nil {
		t.Error(err)
	} else if len(selectSources) != 9 {
		t.Errorf("Select error:\nExpect [count] 9 but got %d\ngot res: %+v", len(selectSources), selectSources)
	}

	selectSources1 := []test.Source{}
	if err := ctl(ctx).Select("id, name").Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAllModel(&selectSources1); err != nil {
		t.Error(err)
	} else if len(selectSources1) != 9 {
		t.Errorf("Select error:\nExpect [count] 9 but got %d\ngot res: %+v", len(selectSources1), selectSources1)
	}

	if !reflect.DeepEqual(selectSources, selectSources1) {
		t.Errorf("Select error:\nExpect [not equal] but \ngot: sources: %+v\ngot: sources1: %+v", selectSources, selectSources1)
	}

	selectSource := test.Source{}
	if err := ctl(ctx).Select("id, name").Filter(Cond{"is_deleted": false}).OrderBy("id").FindOneModel(&selectSource); err != nil {
		t.Error(err)
	} else if selectSource.Id != 11 {
		t.Errorf("Select error:\nExpect [ID] 11 but got %d", selectSource.Id)
	} else if selectSource.Name != "Acfun" {
		t.Errorf("Select error:\nExpect [Name] Acfun but got %s", selectSource.Name)
	}

	// test OrderBy
	if _, err := ctl(ctx).OrderBy([]string{}).FindAll(); err != nil {
		t.Error(err)
	}

	if res, err := ctl(ctx).OrderBy([]string{"-id"}).FindAll(); err != nil {
		t.Error(err)
	} else if res[0]["id"].(int64) != 53 {
		t.Errorf("Select error:\nExpect [53] but got %d", res[0]["id"])
	}

	// test GroupBy
	groupbyNames := []struct {
		Name string `db:"name"`
	}{}
	if err := ctl(ctx).Select([]string{"name"}).GroupBy("name").FindAllModel(&groupbyNames); err != nil {
		t.Error(err)
	} else if len(groupbyNames) != 5 {
		t.Errorf("GroupBy error:\nExpect [count] 5 but got %d\ngot res: %+v", len(groupbyNames), groupbyNames)
	} else {
		for _, v := range groupbyNames {
			if v.Name == "" {
				t.Error("GroupBy error:\nExpect [non-empty name] but got empty")
			}
		}
	}

	groupbyNames1 := []struct {
		Name string `db:"name"`
	}{}
	if err := ctl(ctx).Select([]string{"name"}).GroupBy([]string{"name"}).FindAllModel(&groupbyNames1); err != nil {
		t.Error(err)
	} else if len(groupbyNames1) != 5 {
		t.Errorf("GroupBy error:\nExpect [count] 5 but got %d\ngot res: %+v", len(groupbyNames1), groupbyNames1)
	} else {
		for _, v := range groupbyNames1 {
			if v.Name == "" {
				t.Error("GroupBy error:\nExpect [non-empty name] but got empty")
			}
		}
	}

	groupbyNames2 := []struct {
		Name string `db:"name"`
	}{}
	if err := ctl(ctx).Select("name").GroupBy([]string{}).FindAllModel(&groupbyNames2); err != nil {
		t.Error(err)
	} else if len(groupbyNames2) != 15 {
		t.Errorf("GroupBy error:\nExpect [15] but got %d\ngot res: %+v", len(groupbyNames2), groupbyNames2)
	} else {
		for _, v := range groupbyNames2 {
			if v.Name == "" {
				t.Error("GroupBy error:\nExpect [non-empty name] but got empty")
			}
		}
	}

	// test having

	// Insert, update, delete, remove
	if _, err := ctl(ctx).Insert(map[string]any{"id": 666, "name": "666", "description": "2333"}); err != nil {
		t.Errorf("Insert error: %s", err)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 666}).FindOne(); err != nil {
		t.Error(err)
	} else if res["id"].(int64) != 666 {
		t.Errorf("Insert error: \nexpect 666 but got %d", res["id"])
	} else if res["name"].(string) != "666" {
		t.Errorf("Insert error: \nexpect 666 but got %s", res["name"])
	} else if res["description"].(string) != "2333" {
		t.Errorf("Insert error: \nexpect 2333 but got %s", res["description"])
	}
	if res, err := ctl(ctx).Filter(Cond{"id": 666}).Update(map[string]any{"name": "test"}); err != nil {
		t.Errorf("Update error: %s", err)
	} else if res != 1 {
		t.Errorf("Update error: \nexpect 1 but got %d", res)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 666}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if res["name"].(string) != "test" {
		t.Errorf("Update error: \nexpect test but got %s", res["name"])
	} else if res["description"].(string) != "2333" {
		t.Errorf("Update error: \nexpect 2333 but got %s", res["description"])
	}

	if num, err := ctl(ctx).Filter(Cond{"id": 666}).Delete(); err != nil {
		t.Errorf("Delete error: %s", err)
	} else if num != 1 {
		t.Errorf("Delete error: \nexpect 1 but got %d", num)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 666, "is_deleted": true}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if res["id"].(int64) != 666 && res["is_deleted"].(bool) != true {
		t.Errorf("Delete error: \nexpect 666 but got %d\nexpect true but got %t", res["id"], res["is_deleted"])
	} else if res["name"].(string) != "test" {
		t.Errorf("Delete error: \nexpect test but got %s", res["name"])
	}

	if _, err := ctl(ctx).Filter(Cond{"id": 666, "is_deleted": true}).Remove(); err != nil {
		t.Errorf("Error error: %v", err)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 666}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if len(res) != 0 {
		t.Errorf("Remove error: \nexpect 0 but got %d\nexpect empty but got %+v\n", len(res), res)
	}

	if _, err := ctl(ctx).InsertModel(&test.Source{Id: 777, Name: "777", Description: "2333", IsDeleted: false, CreateTime: time.Now(), UpdateTime: time.Now()}); err != nil {
		t.Errorf("InsertModel error: %s", err)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 777, "is_deleted": false}).FindOne(); err != nil {
		t.Error(err)
	} else if res["id"].(int64) != 777 {
		t.Errorf("InsertModel error: \nexpect 777 but got %d", res["id"])
	} else if res["name"].(string) != "777" {
		t.Errorf("InsertModel error: \nexpect 777 but got %s", res["name"])
	} else if res["description"].(string) != "2333" {
		t.Errorf("InsertModel error: \nexpect 2333 but got %s", res["description"])
	}

	if _, err := ctl(ctx).Filter(Cond{"id": 777}).Remove(); err != nil {
		t.Errorf("Remove error: %s", err)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 777}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if len(res) != 0 {
		t.Errorf("Remove error: \nexpect 0 but got %d\nexpect empty but got %+v\n", len(res), res)
	}

	// test List
	if num, res, err := ctl(ctx).Filter(Cond{"id__in": []int64{11, 12, 13}}).OrderBy("-id").List(); err != nil {
		t.Errorf("List error: %s", err)
	} else if num != 3 {
		t.Errorf("List num error: \nexpect 3 but got %d", num)
	} else if len(res) != 3 {
		t.Errorf("List res length error: \nexpect 3 but got %d", len(res))
	} else if res[0]["id"].(int64) != 13 {
		t.Errorf("List res[0] error: \nexpect 13 but got %d", res[0]["id"])
	} else if res[1]["id"].(int64) != 12 {
		t.Errorf("List res[1] error: \nexpect 12 but got %d", res[1]["id"])
	} else if res[2]["id"].(int64) != 11 {
		t.Errorf("List res[2] error: \nexpect 11 but got %d", res[2]["id"])
	}

	// test GetOrCreate
	if res, err := ctl(ctx).GetOrCreate(map[string]any{"id": 11, "description": "A 站"}); err != nil {
		t.Errorf("GetOrCreate error: %s", err)
	} else if res["id"].(int64) != 11 {
		t.Errorf("GetOrCreate error: \nexpect [id] 11 but got %d", res["id"])
	} else if res["name"].(string) != "Acfun" {
		t.Errorf("GetOrCreate error: \nexpect [name] Acfun but got %s", res["name"])
	}

	if res, err := ctl(ctx).GetOrCreate(map[string]any{"id": 12345, "name": "12345", "description": "12345"}); err != nil {
		t.Errorf("GetOrCreate error: %s", err)
	} else if res["name"].(string) != "12345" {
		t.Errorf("GetOrCreate error: \nexpect 12345 but got %s", res["name"])
	} else if res, err := ctl(ctx).Filter(Cond{"id": 12345}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if res["id"].(int64) != 12345 {
		t.Errorf("GetOrCreate error: \nexpect 12345 but got %d", res["id"])
	} else if res["name"].(string) != "12345" {
		t.Errorf("GetOrCreate error: \nexpect 12345 but got %s", res["name"])
	} else if _, err := ctl(ctx).Filter(Cond{"id": 12345}).Remove(); err != nil {
		t.Errorf("Remove error: %s", err)
	}

	// test CreateOrUpdate
	if created, num, err := ctl(ctx).CreateOrUpdate(map[string]any{"id": 23456, "description": "Test23456"}, Cond{"id": 23456}); err != nil {
		t.Errorf("CreateOrUpdate error: %s", err)
	} else if !created {
		t.Errorf("CreateOrUpdate error: \nexpect [created] but got [not created]")
	} else if num != 0 {
		t.Errorf("CreateOrUpdate error: \nexpect 0 but got %d", num)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 23456}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if res["id"].(int64) != 23456 {
		t.Errorf("CreateOrUpdate error: \nexpect 23456 but got %d", res["id"])
	} else if res["description"].(string) != "Test23456" {
		t.Errorf("CreateOrUpdate error: \nexpect Test23456 but got %s", res["description"])
	} else if res["name"].(string) != "" {
		t.Errorf("CreateOrUpdate error: \nexpect empty but got %s", res["name"])
	}

	if created, num, err := ctl(ctx).CreateOrUpdate(map[string]any{"id": 23456, "name": "test65432", "description": "Test65432"}, Cond{"id": 23456}); err != nil {
		t.Errorf("CreateOrUpdate error: %s", err)
	} else if created {
		t.Errorf("CreateOrUpdate error: \nexpect [not created] but got [created]")
	} else if num != 1 {
		t.Errorf("CreateOrUpdate error: \nexpect 1 but got %d", num)
	} else if res, err := ctl(ctx).Filter(Cond{"id": 23456}).FindOne(); err != nil {
		t.Errorf("FindOne error: %s", err)
	} else if res["id"].(int64) != 23456 {
		t.Errorf("CreateOrUpdate error: \nexpect 23456 but got %d", res["id"])
	} else if res["description"].(string) != "Test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect Test65432 but got %s", res["description"])
	} else if res["name"].(string) != "test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect test65432 but got %s", res["name"])
	}

	if _, err := ctl(ctx).Filter(Cond{"id": 23456}).Remove(); err != nil {
		t.Errorf("Remove error: %s", err)
	}

	cliTmp := NewController(sqlx.NewMysql(mysqlAddress), mysqlOp.NewOperator(), test.Property{})
	filter := Cond{"source_id": 11}
	bakTmp, err := cliTmp(ctx).Filter(filter).FindAll()
	if err != nil {
		t.Errorf("FindAll error: %s", err)
	}
	if created, num, err := cliTmp(ctx).CreateOrUpdate(map[string]any{"column_name": "test65432", "description": "Test65432"}, filter); err != nil {
		t.Errorf("CreateOrUpdate error: %s", err)
	} else if created {
		t.Errorf("CreateOrUpdate error: \nexpect [not created] but got [created]")
	} else if num != 3 {
		t.Errorf("CreateOrUpdate error: \nexpect 3 but got %d", num)
	} else if res, err := cliTmp(ctx).Filter(Cond{"source_id": 11}).FindAll(); err != nil {
		t.Errorf("FindAll error: %s", err)
	} else if len(res) != 3 {
		t.Errorf("CreateOrUpdate error: \nexpect 3 but got %d", len(res))
	} else if res[0]["column_name"].(string) != "test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect test65432 but got %s", res[0]["column_name"])
	} else if res[0]["description"].(string) != "Test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect Test65432 but got %s", res[0]["description"])
	} else if res[1]["column_name"].(string) != "test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect test65432 but got %s", res[1]["column_name"])
	} else if res[1]["description"].(string) != "Test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect Test65432 but got %s", res[1]["description"])
	} else if res[2]["column_name"].(string) != "test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect test65432 but got %s", res[2]["column_name"])
	} else if res[2]["description"].(string) != "Test65432" {
		t.Errorf("CreateOrUpdate error: \nexpect Test65432 but got %s", res[2]["description"])
	}

	for _, v := range bakTmp {
		if _, err := cliTmp(ctx).Filter(Cond{"id": v["id"]}).Update(v); err != nil {
			t.Errorf("Update error: %s", err)
		}
	}

}

func TestHandlerError(t *testing.T) {
	ctl := NewController(sqlx.NewMysql(mysqlAddress), mysqlOp.NewOperator(), test.Source{})

	if _, err := ctl(nil).Filter(Cond{}).Where("").FindOne(); err != nil && err.Error() != fmt.Sprintf(filterOrWhereError, "Filter") {
		t.Errorf("expect nil but got %v", err)
	}

	ctx := context.Background()

	if _, err := ctl(ctx).Exclude(Cond{}).Where("").FindOne(); err != nil && err.Error() != fmt.Sprintf(filterOrWhereError, "Exclude") {
		t.Errorf("expect nil but got %v", err)
	}

	// FindOne unsupported operations
	if res, err := ctl(ctx).GroupBy("").FindOne(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[GroupBy] not supported for FindOne" {
		t.Error(err)
	}

	if res, err := ctl(ctx).GroupBy("").Select("").FindOne(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[Select] not supported for FindOne" {
		t.Error(err)
	}

	// FindAll unsupported operations
	if res, err := ctl(ctx).GroupBy("").FindAll(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[GroupBy] not supported for FindAll" {
		t.Error(err)
	}

	if res, err := ctl(ctx).GroupBy("").Select("").FindAll(); reflect.DeepEqual(res, map[string]any{}) {
		t.Errorf("expect map[string]any{} but got %+v", res)
	} else if err != nil && err.Error() != "[Select] not supported for FindAll" {
		t.Error(err)
	}

	if _, err := ctl(ctx).Limit(0, 1).FindAll(); err != nil {
		if err.Error() != fmt.Errorf(MustBeCalledError, "Limit", "OrderBy").Error() {
			t.Errorf("Limit call order check failed. error: %v", err)
		}
	}

	// Insert unsupported operations
	if id, err := ctl(ctx).Filter(Cond{}).Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy GroupBy] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Select("").Insert(map[string]any{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where Select OrderBy GroupBy] not supported for Insert" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).InsertModel(&test.Source{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter] not supported for InsertModel" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").InsertModel(&test.Source{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where] not supported for InsertModel" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").InsertModel(&test.Source{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy] not supported for InsertModel" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").InsertModel(&test.Source{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where OrderBy GroupBy] not supported for InsertModel" {
		t.Error(err)
	}

	if id, err := ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Select("").InsertModel(&test.Source{}); id != 0 {
		t.Errorf("expect 0 but got %d", id)
	} else if err != nil && err.Error() != "[Filter Where Select OrderBy GroupBy] not supported for InsertModel" {
		t.Error(err)
	}

	// Update unsupported operations
	if num, err := ctl(ctx).GroupBy("").Update(map[string]any{}); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Update" {
		t.Error(err)
	}

	if num, err := ctl(ctx).GroupBy("").Select("").Update(map[string]any{}); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy] not supported for Update" {
		t.Error(err)
	}

	// if num, err := ctl(ctx).GroupBy("").Select("").OrderBy("").Update(map[string]any{}); num != 0 {
	// 	t.Errorf("expect 0 but got %d", num)
	// } else if err != nil && err.Error() != "[Select GroupBy OrderBy] not supported for Update" {
	// 	t.Error(err)
	// }

	// Remove unsupported operations
	if num, err := ctl(ctx).GroupBy("").Remove(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Remove" {
		t.Error(err)
	}

	if num, err := ctl(ctx).GroupBy("").Select("").Remove(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[Select GroupBy] not supported for Remove" {
		t.Error(err)
	}

	// Delete unsupported operations
	if num, err := ctl(ctx).GroupBy("").Delete(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy] not supported for Delete" {
		t.Error(err)
	}

	if num, err := ctl(ctx).GroupBy("").Select("").Delete(); num != 0 {
		t.Errorf("expect 0 but got %d", num)
	} else if err != nil && err.Error() != "[GroupBy Select] not supported for Delete" {
		t.Error(err)
	}

	// send not exist columns to Select
	if err := ctl(ctx).Select([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).Select([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).Select([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns to OrderBy
	if err := ctl(ctx).OrderBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).OrderBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).OrderBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns to GroupBy
	if err := ctl(ctx).GroupBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).GroupBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).GroupBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	// send not exist columns and display last error
	if err := ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "OrderBy columns validate error: [age; happy; damnit] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).OrderBy([]string{"age"}).GroupBy([]string{"test"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).OrderBy([]string{"age", "happy"}).GroupBy([]string{"test", "test2"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test; test2] not exist" {
			t.Error(err)
		}
	}

	if err := ctl(ctx).OrderBy([]string{"age", "happy", "damnit"}).GroupBy([]string{"test", "test2", "test3"}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "GroupBy columns validate error: [test; test2; test3] not exist" {
			t.Error(err)
		}
	}

	// test reset
	cli := ctl(ctx)
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

	if res, err := ctl(ctx).Filter(Cond{"name__contains": []string{"Ac", ""}}).OrderBy("id").Limit(10, 1).FindAll(); err != nil {
		if err.Error() != "operator [contains] unsupported value empty" {
			t.Error(err)
		}
	} else if len(res) != 5 {
		t.Errorf("expect 5 but got %d\ngot res: %+v", len(res), res)
	}

	if err := ctl(ctx).Select([]string{}).FindOneModel(&test.Source{}); err != nil {
		t.Error(err)
	}

	if err := ctl(ctx).Select([]int{1}).FindOneModel(&test.Source{}); err != nil {
		if err.Error() != "Select type should be string or string slice" {
			t.Error(err)
		}
	}

	// orderBy type error
	if _, err := ctl(ctx).OrderBy([1]string{"id"}).FindAll(); err != nil {
		if err.Error() != OrderByColumnsTypeError {
			t.Error(err)
		}
	}

	// groupby type error
	if err := ctl(ctx).Select([]string{"name"}).GroupBy([1]string{"name"}).FindAllModel(&[]struct {
		Name string `db:"name"`
	}{}); err != nil {
		if err.Error() != GroupByColumnsTypeError {
			t.Error(err)
		}
	}

	// have error before remove
	if _, err := ctl(ctx).Filter(Cond{"id": 1}).OrderBy([1]string{"-id"}).Remove(); err != nil {
		if err.Error() != OrderByColumnsTypeError {
			t.Error(err)
		}
	}

	// have error before update
	if _, err := ctl(ctx).Filter(Cond{"id": 1}).OrderBy([1]string{"-id"}).Update(map[string]any{}); err != nil {
		if err.Error() != OrderByColumnsTypeError {
			t.Error(err)
		}
	}

	// have error before count
	if _, err := ctl(ctx).Filter(nil).Count(); err != nil {
		if err.Error() != fmt.Errorf(unsupportedFilterTypeError, "nil").Error() {
			t.Error(err)
		}
	}

	// have error before exist
	if _, err := ctl(ctx).Filter(nil).Exist(); err != nil {
		if err.Error() != fmt.Errorf(unsupportedFilterTypeError, "nil").Error() {
			t.Error(err)
		}
	}

	// have not exist column in update
	if _, err := ctl(ctx).Filter(Cond{"id": 11}).Update(map[string]any{"name": "test", "age": 18}); err != nil {
		if err.Error() != fmt.Errorf(UpdateColumnNotExistError, "age").Error() {
			t.Error(err)
		}
	}
}
