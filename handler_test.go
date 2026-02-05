package norm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/leisurelicht/norm/internal/queryset"
	go_zero "github.com/leisurelicht/norm/operator/mysql/go-zero"
	"github.com/leisurelicht/norm/test"
)

func TestCharacterEncoding(t *testing.T) {
	db := sqlx.NewMysql(getMysqlAddress())
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

// ============================================================================
// 以下是优化后的测试用例
// ============================================================================

func getMysqlAddress() string {
	if addr := os.Getenv("MYSQL_ADDRESS"); addr != "" {
		return addr
	}
	return "root:123456@tcp(127.0.0.1:6033)/test?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
}

// TestGoZeroMysqlTransaction_Refactored 优化后的事务测试
func TestGoZeroMysqlTransaction_Refactored(t *testing.T) {
	ctx := context.Background()
	conn := sqlx.NewMysql(getMysqlAddress())
	sourceCli := NewController(go_zero.NewOperator(conn), test.Source{})
	propertyCli := NewController(go_zero.NewOperator(conn), test.Property{})

	t.Run("commit", func(t *testing.T) {
		err := conn.Transact(func(tx sqlx.Session) error {
			_, err := sourceCli(ctx).WithSession(tx).Create(map[string]any{"id": 1000, "name": "transaction2", "description": "test transaction2"})
			return err
		})
		if err != nil {
			t.Fatalf("Transaction error: %v", err)
		}

		tmpConn := sourceCli(ctx).Filter(Cond{"id": 1000, "name": "transaction2"})
		exist, err := tmpConn.Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if !exist {
			t.Error("got not exist, want exist")
		}

		num, err := tmpConn.Remove()
		if err != nil {
			t.Fatalf("Remove error: %v", err)
		}
		if num != 1 {
			t.Errorf("got num %d, want 1", num)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		err := conn.Transact(func(tx sqlx.Session) error {
			if _, err := sourceCli(ctx).WithSession(tx).Create(map[string]any{"id": 1000, "name": "transaction2", "description": "test transaction2"}); err != nil {
				return err
			}
			if _, _, err := propertyCli(ctx).WithSession(tx).CreateOrUpdate(map[string]any{"column_name": "test65432", "description": "Test65432"}); err != nil {
				return err
			}
			return errors.New("rollback transaction")
		})
		if err == nil {
			t.Error("expected error, got nil")
		}

		exist, err := sourceCli(ctx).Filter(Cond{"id": 1000, "name": "transaction2"}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})

	t.Run("rollback with ctx", func(t *testing.T) {
		err := conn.TransactCtx(ctx, func(innerCtx context.Context, tx sqlx.Session) error {
			if _, err := sourceCli(innerCtx).WithSession(tx).Create(map[string]any{"id": 1000, "name": "transaction2", "description": "test transaction2"}); err != nil {
				return err
			}
			if _, _, err := propertyCli(innerCtx).WithSession(tx).CreateOrUpdate(map[string]any{"column_name": "test65432", "description": "Test65432"}); err != nil {
				return err
			}
			return errors.New("rollback transaction")
		})
		if err == nil {
			t.Error("expected error, got nil")
		}

		exist, err := sourceCli(ctx).Filter(Cond{"id": 1000, "name": "transaction2"}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})

	t.Run("with session rollback visibility", func(t *testing.T) {
		err := conn.Transact(func(tx sqlx.Session) error {
			if _, err := sourceCli(ctx).WithSession(tx).Create(map[string]any{"id": 2000, "name": "session", "description": "session tx"}); err != nil {
				return err
			}
			exist, err := sourceCli(ctx).WithSession(tx).Filter(Cond{"id": 2000, "name": "session"}).Exist()
			if err != nil {
				return err
			}
			if !exist {
				return errors.New("expected exist within tx")
			}
			return errors.New("force rollback")
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		exist, err := sourceCli(ctx).Filter(Cond{"id": 2000, "name": "session"}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})

	t.Run("with session update rollback", func(t *testing.T) {
		err := conn.Transact(func(tx sqlx.Session) error {
			cli := sourceCli(ctx).WithSession(tx)
			if _, err := cli.Create(map[string]any{"id": 2001, "name": "session_u", "description": "session tx"}); err != nil {
				return err
			}
			if _, err := cli.Filter(Cond{"id": 2001}).Update(map[string]any{"name": "session_u2"}); err != nil {
				return err
			}
			res, err := cli.Filter(Cond{"id": 2001}).FindOne()
			if err != nil {
				return err
			}
			if res["name"].(string) != "session_u2" {
				return errors.New("expected updated name within tx")
			}
			return errors.New("force rollback")
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		exist, err := sourceCli(ctx).Filter(Cond{"id": 2001}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})

	t.Run("with session reset keeps session", func(t *testing.T) {
		err := conn.Transact(func(tx sqlx.Session) error {
			cli := sourceCli(ctx).WithSession(tx)
			if _, err := cli.Create(map[string]any{"id": 2002, "name": "session_r", "description": "session tx"}); err != nil {
				return err
			}
			cli = cli.Reset()
			exist, err := cli.Filter(Cond{"id": 2002}).Exist()
			if err != nil {
				return err
			}
			if !exist {
				return errors.New("expected exist after reset within tx")
			}
			return errors.New("force rollback")
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		exist, err := sourceCli(ctx).Filter(Cond{"id": 2002}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})
}

// TestGoZeroMysqlMethods_Refactored 优化后的方法测试
func TestGoZeroMysqlMethods_Refactored(t *testing.T) {
	SetLevel(Debug)
	mysqlAddr := getMysqlAddress()
	sourceCli := NewController(go_zero.NewOperator(go_zero.NewMysql(mysqlAddr), go_zero.WithTableName("source")), test.Source{})
	propertyCli := NewController(go_zero.NewOperator(go_zero.NewMysql(mysqlAddr)), test.Property{})
	ctx := context.Background()

	t.Run("Count", func(t *testing.T) {
		num, err := sourceCli(nil).Count()
		if err != nil {
			t.Fatalf("Count error: %v", err)
		}
		if num != 15 {
			t.Errorf("got %d, want 15", num)
		}
	})

	t.Run("Exist", func(t *testing.T) {
		exist, err := sourceCli(ctx).Filter(Cond{"id": 11}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if !exist {
			t.Error("got not exist, want exist")
		}

		exist, err = sourceCli(ctx).Filter(Cond{"id": 12345}).Exist()
		if err != nil {
			t.Fatalf("Exist error: %v", err)
		}
		if exist {
			t.Error("got exist, want not exist")
		}
	})

	t.Run("FindOne", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"id": 11}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 11 {
			t.Errorf("got id %d, want 11", res["id"])
		}

		res, err = sourceCli(ctx).Filter(Cond{"is_deleted": false}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["name"].(string) != "Acfun" {
			t.Errorf("got name %s, want Acfun", res["name"])
		}
	})

	t.Run("Where", func(t *testing.T) {
		res, err := sourceCli(ctx).Where("id = ?", 11).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 11 {
			t.Errorf("got id %d, want 11", res["id"])
		}
	})

	t.Run("FindOneModel", func(t *testing.T) {
		created, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
		updated, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-19 15:16:23", time.Local)
		want := test.Source{Id: 11, Name: "Acfun", Type: 1, Description: "A 站", IsDeleted: false, CreateTime: created, UpdateTime: updated}

		var got test.Source
		if err := sourceCli(ctx).Filter(Cond{"id": 11}).FindOneModel(&got); err != nil {
			t.Fatalf("FindOneModel error: %v", err)
		}
		if got.Id != want.Id {
			t.Errorf("got id %d, want %d", got.Id, want.Id)
		}
		if got.Name != want.Name {
			t.Errorf("got name %s, want %s", got.Name, want.Name)
		}
		if got.Type != want.Type {
			t.Errorf("got type %d, want %d", got.Type, want.Type)
		}
		if got.Description != want.Description {
			t.Errorf("got description %s, want %s", got.Description, want.Description)
		}
		if got.IsDeleted != want.IsDeleted {
			t.Errorf("got is_deleted %t, want %t", got.IsDeleted, want.IsDeleted)
		}
		if got.CreateTime.Format("2006-01-02 15:04:05") != want.CreateTime.Format("2006-01-02 15:04:05") {
			t.Errorf("got create_time %s, want %s", got.CreateTime, want.CreateTime)
		}
		if got.UpdateTime.Format("2006-01-02 15:04:05") != want.UpdateTime.Format("2006-01-02 15:04:05") {
			t.Errorf("got update_time %s, want %s", got.UpdateTime, want.UpdateTime)
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"id": 11}, OR{"id": 12}).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 2 {
			t.Errorf("got len %d, want 2", len(res))
		}
		if res[0]["id"].(int64) != 11 {
			t.Errorf("got id %d, want 11", res[0]["id"])
		}
		if res[1]["id"].(int64) != 12 {
			t.Errorf("got id %d, want 12", res[1]["id"])
		}
	})

	t.Run("FindAllModel with AND", func(t *testing.T) {
		sources := []test.Source{}
		if err := sourceCli(ctx).Filter(Cond{"id": 11}, AND{"id": 12}).FindAllModel(&sources); err != nil {
			t.Fatalf("FindAllModel error: %v", err)
		}
		if len(sources) != 0 {
			t.Errorf("got len %d, want 0", len(sources))
		}
	})

	t.Run("FindAll with Limit", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 9 {
			t.Errorf("got len %d, want 9", len(res))
		}
		if res[0]["id"].(int64) != 11 {
			t.Errorf("got first id %d, want 11", res[0]["id"])
		}
		if res[len(res)-1]["id"].(int64) != 53 {
			t.Errorf("got last id %d, want 53", res[len(res)-1]["id"])
		}
	})

	t.Run("FindAll empty result", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"id": 12345}).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}
	})

	t.Run("FindOne empty result", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"id": 54321}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}
	})

	t.Run("FindAllModel empty result", func(t *testing.T) {
		var sources []test.Source
		if err := sourceCli(ctx).Filter(Cond{"id": 54321}).FindAllModel(&sources); err != nil {
			t.Fatalf("FindAllModel error: %v", err)
		}
		if len(sources) != 0 {
			t.Errorf("got len %d, want 0", len(sources))
		}
	})

	t.Run("Contains filter", func(t *testing.T) {
		res, err := sourceCli(ctx).Filter(Cond{"name__contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 6 {
			t.Errorf("got len %d, want 6", len(res))
		}
	})

	t.Run("Not contains and Exclude", func(t *testing.T) {
		resNotContains, err := sourceCli(ctx).Filter(Cond{"name__not_contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(resNotContains) != 9 {
			t.Errorf("got len %d, want 9", len(resNotContains))
		}

		resExclude, err := sourceCli(ctx).Exclude(Cond{"name__contains": []string{"Ac", "Ap"}}).OrderBy("id").Limit(10, 1).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(resExclude) != 9 {
			t.Errorf("got len %d, want 9", len(resExclude))
		}

		for i, v := range resExclude {
			if v["id"] != resNotContains[i]["id"] {
				t.Errorf("Exclude and NotContains mismatch at %d", i)
			}
		}
	})

	t.Run("Select slice", func(t *testing.T) {
		selectSources := []test.Source{}
		if err := sourceCli(ctx).Select([]string{"id", "name"}).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAllModel(&selectSources); err != nil {
			t.Fatalf("FindAllModel error: %v", err)
		}
		if len(selectSources) != 9 {
			t.Errorf("got len %d, want 9", len(selectSources))
		}
	})

	t.Run("Select string", func(t *testing.T) {
		selectSourcesSlice := []test.Source{}
		if err := sourceCli(ctx).Select([]string{"id", "name"}).Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAllModel(&selectSourcesSlice); err != nil {
			t.Fatalf("FindAllModel error: %v", err)
		}

		selectSources := []test.Source{}
		if err := sourceCli(ctx).Select("id, name").Filter(Cond{"is_deleted": false}).OrderBy("id").Limit(10, 1).FindAllModel(&selectSources); err != nil {
			t.Fatalf("FindAllModel error: %v", err)
		}
		if len(selectSources) != 9 {
			t.Errorf("got len %d, want 9", len(selectSources))
		}

		if !reflect.DeepEqual(selectSourcesSlice, selectSources) {
			t.Errorf("Select slice and string results differ:\nslice: %+v\nstring: %+v", selectSourcesSlice, selectSources)
		}

		selectSource := test.Source{}
		if err := sourceCli(ctx).Select("id, name").Filter(Cond{"is_deleted": false}).OrderBy("id").FindOneModel(&selectSource); err != nil {
			t.Fatalf("FindOneModel error: %v", err)
		}
		if selectSource.Id != 11 {
			t.Errorf("got id %d, want 11", selectSource.Id)
		}
		if selectSource.Name != "Acfun" {
			t.Errorf("got name %s, want Acfun", selectSource.Name)
		}
	})

	t.Run("OrderBy", func(t *testing.T) {
		if _, err := sourceCli(ctx).OrderBy([]string{}).FindAll(); err != nil {
			t.Fatalf("OrderBy empty error: %v", err)
		}

		res, err := sourceCli(ctx).OrderBy([]string{"-id"}).FindAll()
		if err != nil {
			t.Fatalf("OrderBy error: %v", err)
		}
		if res[0]["id"].(int64) != 53 {
			t.Errorf("got first id %d, want 53", res[0]["id"])
		}
	})

	t.Run("GroupBy", func(t *testing.T) {
		groupbyNames := []struct {
			Name string `db:"name"`
		}{}
		if err := sourceCli(ctx).Select([]string{"name"}).GroupBy("name").FindAllModel(&groupbyNames); err != nil {
			t.Fatalf("GroupBy error: %v", err)
		}
		if len(groupbyNames) != 5 {
			t.Errorf("got len %d, want 5", len(groupbyNames))
		}
		for _, v := range groupbyNames {
			if v.Name == "" {
				t.Error("got empty name, want non-empty")
			}
		}

		groupbyNames1 := []struct {
			Name string `db:"name"`
		}{}
		if err := sourceCli(ctx).Select([]string{"name"}).GroupBy([]string{"name"}).FindAllModel(&groupbyNames1); err != nil {
			t.Fatalf("GroupBy error: %v", err)
		}
		if len(groupbyNames1) != 5 {
			t.Errorf("got len %d, want 5", len(groupbyNames1))
		}
		for _, v := range groupbyNames1 {
			if v.Name == "" {
				t.Error("got empty name, want non-empty")
			}
		}

		groupbyNames2 := []struct {
			Name string `db:"name"`
		}{}
		if err := sourceCli(ctx).Select("name").GroupBy([]string{}).FindAllModel(&groupbyNames2); err != nil {
			t.Fatalf("GroupBy error: %v", err)
		}
		if len(groupbyNames2) != 15 {
			t.Errorf("got len %d, want 15", len(groupbyNames2))
		}
		for _, v := range groupbyNames2 {
			if v.Name == "" {
				t.Error("got empty name, want non-empty")
			}
		}
	})

	t.Run("CRUD with map", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 666}).Remove() })

		if _, err := sourceCli(ctx).Create(map[string]any{"id": 666, "name": "666", "description": "2333"}); err != nil {
			t.Fatalf("Create error: %v", err)
		}

		res, err := sourceCli(ctx).Filter(Cond{"id": 666}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 666 || res["name"].(string) != "666" || res["description"].(string) != "2333" {
			t.Errorf("got id=%d name=%s desc=%s, want id=666 name=666 desc=2333", res["id"], res["name"], res["description"])
		}

		num, err := sourceCli(ctx).Filter(Cond{"id": 666}).Update(map[string]any{"name": "test"})
		if err != nil {
			t.Fatalf("Update error: %v", err)
		}
		if num != 1 {
			t.Errorf("got updated %d, want 1", num)
		}

		res, err = sourceCli(ctx).Filter(Cond{"id": 666}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["name"].(string) != "test" || res["description"].(string) != "2333" {
			t.Errorf("got name=%s desc=%s, want name=test desc=2333", res["name"], res["description"])
		}
	})

	t.Run("Delete soft delete", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 666}).Remove() })

		if _, err := sourceCli(ctx).Create(map[string]any{"id": 666, "name": "666", "description": "2333"}); err != nil {
			t.Fatalf("Create error: %v", err)
		}

		if num, err := sourceCli(ctx).Filter(Cond{"id": 666}).Update(map[string]any{"name": "test"}); err != nil {
			t.Fatalf("Update error: %v", err)
		} else if num != 1 {
			t.Errorf("got updated %d, want 1", num)
		}

		num, err := sourceCli(ctx).Filter(Cond{"id": 666}).Delete()
		if err != nil {
			t.Fatalf("Delete error: %v", err)
		}
		if num != 1 {
			t.Errorf("got deleted %d, want 1", num)
		}

		res, err := sourceCli(ctx).Filter(Cond{"id": 666, "is_deleted": true}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 666 || res["is_deleted"].(bool) != true {
			t.Errorf("got id=%d is_deleted=%t, want id=666 is_deleted=true", res["id"], res["is_deleted"])
		}
		if res["name"].(string) != "test" {
			t.Errorf("got name=%s, want test", res["name"])
		}

		if _, err := sourceCli(ctx).Filter(Cond{"id": 666, "is_deleted": true}).Remove(); err != nil {
			t.Fatalf("Remove error: %v", err)
		}

		res, err = sourceCli(ctx).Filter(Cond{"id": 666}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}
	})

	t.Run("CRUD with model", func(t *testing.T) {
		if _, err := sourceCli(ctx).Create(&test.Source{Id: 777, Name: "777", Description: "2333", IsDeleted: false, CreateTime: time.Now(), UpdateTime: time.Now()}); err != nil {
			t.Fatalf("Create error: %v", err)
		}

		res, err := sourceCli(ctx).Filter(Cond{"id": 777, "is_deleted": false}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 777 || res["name"].(string) != "777" || res["description"].(string) != "2333" {
			t.Errorf("got id=%d name=%s desc=%s, want id=777 name=777 desc=2333", res["id"], res["name"], res["description"])
		}

		if _, err := sourceCli(ctx).Filter(Cond{"id": 777}).Remove(); err != nil {
			t.Fatalf("Remove error: %v", err)
		}

		res, err = sourceCli(ctx).Filter(Cond{"id": 777}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}
	})

	t.Run("List", func(t *testing.T) {
		num, res, err := sourceCli(ctx).Filter(Cond{"id__in": []int64{11, 12, 13}}).OrderBy("id DESC").List()
		if err != nil {
			t.Fatalf("List error: %v", err)
		}
		if num != 3 {
			t.Errorf("got num %d, want 3", num)
		}
		if len(res) != 3 {
			t.Errorf("got len %d, want 3", len(res))
		}
		if res[0]["id"].(int64) != 13 || res[1]["id"].(int64) != 12 || res[2]["id"].(int64) != 11 {
			t.Errorf("got ids %d,%d,%d, want 13,12,11", res[0]["id"], res[1]["id"], res[2]["id"])
		}
	})

	t.Run("GetOrCreate existing", func(t *testing.T) {
		res, err := sourceCli(ctx).GetOrCreate(map[string]any{"id": 11, "description": "A 站"})
		if err != nil {
			t.Fatalf("GetOrCreate error: %v", err)
		}
		if res["id"].(int64) != 11 {
			t.Errorf("got id %d, want 11", res["id"])
		}
		if res["name"].(string) != "Acfun" {
			t.Errorf("got name %s, want Acfun", res["name"])
		}
	})

	t.Run("GetOrCreate new", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 12345}).Remove() })

		res, err := sourceCli(ctx).GetOrCreate(map[string]any{"id": 12345, "name": "12345", "description": "12345"})
		if err != nil {
			t.Fatalf("GetOrCreate error: %v", err)
		}
		if res["name"].(string) != "12345" {
			t.Errorf("got name %s, want 12345", res["name"])
		}

		res, err = sourceCli(ctx).Filter(Cond{"id": 12345}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 12345 || res["name"].(string) != "12345" {
			t.Errorf("got id=%d name=%s, want id=12345 name=12345", res["id"], res["name"])
		}
	})

	t.Run("CreateOrUpdate create", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 23456}).Remove() })

		created, num, err := sourceCli(ctx).Filter(Cond{"id": 23456}).CreateOrUpdate(map[string]any{"id": 23456, "description": "Test23456"})
		if err != nil {
			t.Fatalf("CreateOrUpdate error: %v", err)
		}
		if !created {
			t.Error("got not created, want created")
		}
		if num != 0 {
			t.Errorf("got num %d, want 0", num)
		}

		res, err := sourceCli(ctx).Filter(Cond{"id": 23456}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["id"].(int64) != 23456 || res["description"].(string) != "Test23456" || res["name"].(string) != "" {
			t.Errorf("got id=%d desc=%s name=%s, want id=23456 desc=Test23456 name=", res["id"], res["description"], res["name"])
		}
	})

	t.Run("CreateOrUpdate update", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 23456}).Remove() })

		sourceCli(ctx).Filter(Cond{"id": 23456}).CreateOrUpdate(map[string]any{"id": 23456, "description": "Test23456"})

		created, num, err := sourceCli(ctx).Filter(Cond{"id": 23456}).CreateOrUpdate(map[string]any{"id": 23456, "name": "test65432", "description": "Test65432"})
		if err != nil {
			t.Fatalf("CreateOrUpdate error: %v", err)
		}
		if created {
			t.Error("got created, want not created")
		}
		if num != 1 {
			t.Errorf("got num %d, want 1", num)
		}

		res, err := sourceCli(ctx).Filter(Cond{"id": 23456}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if res["name"].(string) != "test65432" || res["description"].(string) != "Test65432" {
			t.Errorf("got name=%s desc=%s, want name=test65432 desc=Test65432", res["name"], res["description"])
		}
	})

	t.Run("CreateOrUpdate multiple", func(t *testing.T) {
		filter := Cond{"source_id": 11}
		bakTmp, err := propertyCli(ctx).Filter(filter).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		t.Cleanup(func() {
			for _, v := range bakTmp {
				propertyCli(ctx).Filter(Cond{"id": v["id"]}).Update(v)
			}
		})

		created, num, err := propertyCli(ctx).Filter(filter).CreateOrUpdate(map[string]any{"column_name": "test65432", "description": "Test65432"})
		if err != nil {
			t.Fatalf("CreateOrUpdate error: %v", err)
		}
		if created {
			t.Error("got created, want not created")
		}
		if num != 3 {
			t.Errorf("got num %d, want 3", num)
		}

		res, err := propertyCli(ctx).Filter(Cond{"source_id": 11}).FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 3 {
			t.Errorf("got len %d, want 3", len(res))
		}
		for i, r := range res {
			if r["column_name"].(string) != "test65432" || r["description"].(string) != "Test65432" {
				t.Errorf("row %d: got column_name=%s desc=%s, want column_name=test65432 desc=Test65432", i, r["column_name"], r["description"])
			}
		}
	})

	t.Run("CreateIfNotExist", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 11111}).Remove() })

		res, err := sourceCli(ctx).Filter(Cond{"id": 11111, "name": "test11111"}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}

		id, created, err := sourceCli(ctx).CreateIfNotExist(map[string]any{"id": 11111, "name": "test11111", "description": "Test11111"})
		if err != nil {
			t.Fatalf("CreateIfNotExist error: %v", err)
		}
		if !created {
			t.Error("got not created, want created")
		}
		if id != 0 {
			t.Errorf("got id %d, want 0", id)
		}

		res, err = sourceCli(ctx).Filter(Cond{"id": 11111, "name": "test11111"}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 7 {
			t.Errorf("got len %d, want 7", len(res))
		}
		if res["id"].(int64) != 11111 || res["description"].(string) != "Test11111" || res["name"].(string) != "test11111" {
			t.Errorf("got id=%d desc=%s name=%s, want id=11111 desc=Test11111 name=test11111", res["id"], res["description"], res["name"])
		}
	})

	t.Run("CreateIfNotExist existing", func(t *testing.T) {
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id": 11112}).Remove() })

		if _, err := sourceCli(ctx).Create(map[string]any{"id": 11112, "name": "test11112", "description": "Test11112"}); err != nil {
			t.Fatalf("Create error: %v", err)
		}

		id, created, err := sourceCli(ctx).CreateIfNotExist(map[string]any{"id": 11112, "name": "test11112", "description": "Test11112"})
		if err != nil {
			t.Fatalf("CreateIfNotExist error: %v", err)
		}
		if created {
			t.Error("got created, want not created")
		}
		if id != 0 {
			t.Errorf("got id %d, want 0", id)
		}
	})

	t.Run("Batch Create", func(t *testing.T) {
		t.Cleanup(func() { propertyCli(ctx).Filter(Cond{"source_id__between": []int64{11111, 11116}}).Remove() })

		num, err := propertyCli(ctx).Create([]map[string]any{
			{"source_id": 11111, "column_name": "test11111", "description": "Test11111"},
			{"source_id": 11112, "column_name": "test11112", "description": "Test11112"},
			{"source_id": 11113, "column_name": "test11113", "description": "Test11113"},
			{"source_id": 11114, "column_name": "test11114", "description": "Test11114"},
			{"source_id": 11115, "column_name": "test11115", "description": "Test11115"},
			{"source_id": 11116, "column_name": "test11116", "description": "Test11116"},
		})
		if err != nil {
			t.Fatalf("Create error: %v", err)
		}
		if num != 6 {
			t.Errorf("got num %d, want 6", num)
		}

		res, err := propertyCli(ctx).Filter(Cond{"source_id__between": []int64{11111, 11116}}).OrderBy("source_id").FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != 6 {
			t.Errorf("got len %d, want 6", len(res))
		}
		if res[0]["source_id"].(int64) != 11111 || res[1]["source_id"].(int64) != 11112 || res[5]["source_id"].(int64) != 11116 {
			t.Errorf("got source_ids %d,%d,%d, want 11111,11112,11116", res[0]["source_id"], res[1]["source_id"], res[5]["source_id"])
		}
	})

	t.Run("Batch Create with model slice", func(t *testing.T) {
		ids := []int64{8881, 8882}
		_, _ = sourceCli(ctx).Filter(Cond{"id__in": ids}).Remove()
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id__in": ids}).Remove() })

		type sourceLite struct {
			Id          int64  `db:"id"`
			Name        string `db:"name"`
			Description string `db:"description"`
		}
		data := []sourceLite{
			{Id: ids[0], Name: "batch1", Description: "batch1"},
			{Id: ids[1], Name: "batch2", Description: "batch2"},
		}
		num, err := sourceCli(ctx).Create(data)
		if err != nil {
			t.Fatalf("Create error: %v", err)
		}
		if num != int64(len(data)) {
			t.Errorf("got num %d, want %d", num, len(data))
		}

		res, err := sourceCli(ctx).Filter(Cond{"id__in": ids}).OrderBy("id").FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != len(data) {
			t.Errorf("got len %d, want %d", len(res), len(data))
		} else if res[0]["id"].(int64) != ids[0] || res[1]["id"].(int64) != ids[1] {
			t.Errorf("got ids %d,%d, want %d,%d", res[0]["id"], res[1]["id"], ids[0], ids[1])
		}
	})

	t.Run("Batch Create with model slice pointer", func(t *testing.T) {
		ids := []int64{8891, 8892}
		_, _ = sourceCli(ctx).Filter(Cond{"id__in": ids}).Remove()
		t.Cleanup(func() { sourceCli(ctx).Filter(Cond{"id__in": ids}).Remove() })

		type sourceLite struct {
			Id          int64  `db:"id"`
			Name        string `db:"name"`
			Description string `db:"description"`
		}
		data := []sourceLite{
			{Id: ids[0], Name: "batchp1", Description: "batchp1"},
			{Id: ids[1], Name: "batchp2", Description: "batchp2"},
		}
		num, err := sourceCli(ctx).Create(&data)
		if err != nil {
			t.Fatalf("Create error: %v", err)
		}
		if num != int64(len(data)) {
			t.Errorf("got num %d, want %d", num, len(data))
		}

		res, err := sourceCli(ctx).Filter(Cond{"id__in": ids}).OrderBy("id").FindAll()
		if err != nil {
			t.Fatalf("FindAll error: %v", err)
		}
		if len(res) != len(data) {
			t.Errorf("got len %d, want %d", len(res), len(data))
		}
	})
}

// TestGoZeroMysqlHandlerError_Refactored 优化后的错误处理测试
// 改进点:
// 1. 使用 table-driven 测试
// 2. 使用 t.Run 分离子测试
// 3. 修复错误断言逻辑 (先检查 err == nil)
// 4. 使用非空字符串触发实际错误检查
func TestGoZeroMysqlHandlerError_Refactored(t *testing.T) {
	ctl := NewController(go_zero.NewOperator(sqlx.NewMysql(getMysqlAddress())), test.Source{})
	ctx := context.Background()

	t.Run("FilterOrWhereError", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Filter with Where", func() error { _, err := ctl(nil).Filter(Cond{}).Where("").FindOne(); return err }, fmt.Sprintf(queryset.FilterOrWhereError, "Filter")},
			{"Exclude with Where", func() error { _, err := ctl(ctx).Exclude(Cond{}).Where("").FindOne(); return err }, fmt.Sprintf(queryset.FilterOrWhereError, "Exclude")},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("FindOne unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Select", func() error { _, err := ctl(ctx).Select("").FindOne(); return err }, "[Select] not supported for FindOne"},
			{"Having", func() error { _, err := ctl(ctx).Having("").FindOne(); return err }, "[Having] not supported for FindOne"},
			{"GroupBy+Select", func() error { _, err := ctl(ctx).GroupBy("").Select("").FindOne(); return err }, "[Select] not supported for FindOne"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("FindAll unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Select", func() error { _, err := ctl(ctx).Select("").FindAll(); return err }, "[Select] not supported for FindAll"},
			{"Having", func() error { _, err := ctl(ctx).Having("").FindAll(); return err }, "[Having] not supported for FindAll"},
			{"GroupBy+Having", func() error { _, err := ctl(ctx).GroupBy("").Having("").FindAll(); return err }, "[Having] not supported for FindAll"},
			{"GroupBy+Select", func() error { _, err := ctl(ctx).GroupBy("").Select("").FindAll(); return err }, "[Select] not supported for FindAll"},
			{"Limit without OrderBy", func() error { _, err := ctl(ctx).Limit(0, 1).FindAll(); return err }, fmt.Errorf(MustBeCalledError, "Limit", "OrderBy").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("List unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Select", func() error { _, _, err := ctl(ctx).Select("").List(); return err }, fmt.Errorf(UnsupportedControllerError, ctlSelect.Name, "List").Error()},
			{"Having", func() error { _, _, err := ctl(ctx).Having("").List(); return err }, fmt.Errorf(UnsupportedControllerError, ctlHaving.Name, "List").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Create unsupported map", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"Filter", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Create(map[string]any{}) }, "[Filter] not supported for Create"},
			{"Filter+Where", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Where("").Create(map[string]any{}) }, "[Filter, Where] not supported for Create"},
			{"Filter+Where+OrderBy", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").Create(map[string]any{}) }, "[Filter, Where, OrderBy] not supported for Create"},
			{"Filter+Where+OrderBy+GroupBy", func() (int64, error) {
				return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Create(map[string]any{})
			}, "[Filter, Where, OrderBy, GroupBy] not supported for Create"},
			{"Filter+Where+OrderBy+GroupBy+Select", func() (int64, error) {
				return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Select("").Create(map[string]any{})
			}, "[Filter, Where, Select, OrderBy, GroupBy] not supported for Create"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := tt.fn()
				if id != 0 {
					t.Errorf("got id %d, want 0", id)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Create unsupported model", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"Filter", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Create(&test.Source{}) }, "[Filter] not supported for Create"},
			{"Filter+Where", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Where("").Create(&test.Source{}) }, "[Filter, Where] not supported for Create"},
			{"Filter+Where+OrderBy", func() (int64, error) { return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").Create(&test.Source{}) }, "[Filter, Where, OrderBy] not supported for Create"},
			{"Filter+Where+OrderBy+GroupBy", func() (int64, error) {
				return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Create(&test.Source{})
			}, "[Filter, Where, OrderBy, GroupBy] not supported for Create"},
			{"Filter+Where+OrderBy+GroupBy+Select", func() (int64, error) {
				return ctl(ctx).Filter(Cond{}).Where("").OrderBy("").GroupBy("").Select("").Create(&test.Source{})
			}, "[Filter, Where, Select, OrderBy, GroupBy] not supported for Create"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := tt.fn()
				if id != 0 {
					t.Errorf("got id %d, want 0", id)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Create unsupported other methods", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"Exclude", func() (int64, error) { return ctl(ctx).Exclude(Cond{}).Create(map[string]any{"id": 1}) }, "[Exclude] not supported for Create"},
			{"Where", func() (int64, error) { return ctl(ctx).Where("id = ?", 1).Create(map[string]any{"id": 1}) }, "[Where] not supported for Create"},
			{"OrderBy", func() (int64, error) { return ctl(ctx).OrderBy("id").Create(map[string]any{"id": 1}) }, "[OrderBy] not supported for Create"},
			{"GroupBy", func() (int64, error) { return ctl(ctx).GroupBy("id").Create(map[string]any{"id": 1}) }, "[GroupBy] not supported for Create"},
			{"Having", func() (int64, error) { return ctl(ctx).Having("id = ?", 1).Create(map[string]any{"id": 1}) }, "[Having] not supported for Create"},
			{"Select", func() (int64, error) { return ctl(ctx).Select("id").Create(map[string]any{"id": 1}) }, "[Select] not supported for Create"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, err := tt.fn()
				if id != 0 {
					t.Errorf("got id %d, want 0", id)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Update unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"GroupBy", func() (int64, error) { return ctl(ctx).GroupBy("").Update(map[string]any{}) }, "[GroupBy] not supported for Update"},
			{"GroupBy+Select", func() (int64, error) { return ctl(ctx).GroupBy("").Select("").Update(map[string]any{}) }, "[Select, GroupBy] not supported for Update"},
			{"Select", func() (int64, error) { return ctl(ctx).Select("").Update(map[string]any{}) }, "[Select] not supported for Update"},
			{"Having", func() (int64, error) { return ctl(ctx).Having("").Update(map[string]any{}) }, "[Having] not supported for Update"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				num, err := tt.fn()
				if num != 0 {
					t.Errorf("got num %d, want 0", num)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Remove unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"GroupBy", func() (int64, error) { return ctl(ctx).GroupBy("").Remove() }, "[GroupBy] not supported for Remove"},
			{"GroupBy+Select", func() (int64, error) { return ctl(ctx).GroupBy("").Select("").Remove() }, "[Select, GroupBy] not supported for Remove"},
			{"Select", func() (int64, error) { return ctl(ctx).Select("").Remove() }, "[Select] not supported for Remove"},
			{"Having", func() (int64, error) { return ctl(ctx).Having("").Remove() }, "[Having] not supported for Remove"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				num, err := tt.fn()
				if num != 0 {
					t.Errorf("got num %d, want 0", num)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Delete unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() (int64, error)
			wantErr string
		}{
			{"GroupBy", func() (int64, error) { return ctl(ctx).GroupBy("").Delete() }, "[GroupBy] not supported for Delete"},
			{"GroupBy+Select", func() (int64, error) { return ctl(ctx).GroupBy("").Select("").Delete() }, "[Select, GroupBy] not supported for Delete"},
			{"Select", func() (int64, error) { return ctl(ctx).Select("").Delete() }, "[Select] not supported for Delete"},
			{"Having", func() (int64, error) { return ctl(ctx).Having("").Delete() }, "[Having] not supported for Delete"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				num, err := tt.fn()
				if num != 0 {
					t.Errorf("got num %d, want 0", num)
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Exist unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"GroupBy", func() error { _, err := ctl(ctx).GroupBy("").Exist(); return err }, fmt.Errorf(UnsupportedControllerError, ctlGroupBy.Name, "Exist").Error()},
			{"Select", func() error { _, err := ctl(ctx).Select("").Exist(); return err }, fmt.Errorf(UnsupportedControllerError, ctlSelect.Name, "Exist").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("GetOrCreate unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"GroupBy+Having", func() error { _, err := ctl(ctx).GroupBy("").Having("").GetOrCreate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, strings.Join([]string{ctlGroupBy.Name, ctlHaving.Name}, ", "), "GetOrCreate").Error()},
			{"GroupBy", func() error { _, err := ctl(ctx).GroupBy("").GetOrCreate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlGroupBy.Name, "GetOrCreate").Error()},
			{"Having", func() error { _, err := ctl(ctx).Having("").GetOrCreate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlHaving.Name, "GetOrCreate").Error()},
			{"Select", func() error { _, err := ctl(ctx).Select("").GetOrCreate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlSelect.Name, "GetOrCreate").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("CreateOrUpdate unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"GroupBy+Having", func() error {
				_, _, err := ctl(ctx).GroupBy("").Having("").CreateOrUpdate(map[string]any{})
				return err
			}, fmt.Errorf(UnsupportedControllerError, strings.Join([]string{ctlGroupBy.Name, ctlHaving.Name}, ", "), "CreateOrUpdate").Error()},
			{"GroupBy", func() error { _, _, err := ctl(ctx).GroupBy("").CreateOrUpdate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlGroupBy.Name, "CreateOrUpdate").Error()},
			{"Having", func() error { _, _, err := ctl(ctx).Having("").CreateOrUpdate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlHaving.Name, "CreateOrUpdate").Error()},
			{"Select", func() error { _, _, err := ctl(ctx).Select("").CreateOrUpdate(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlSelect.Name, "CreateOrUpdate").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("CreateIfNotExist unsupported", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"GroupBy+Having", func() error {
				_, _, err := ctl(ctx).GroupBy("").Having("").CreateIfNotExist(map[string]any{})
				return err
			}, fmt.Errorf(UnsupportedControllerError, strings.Join([]string{ctlGroupBy.Name, ctlHaving.Name}, ", "), "CreateIfNotExist").Error()},
			{"GroupBy", func() error { _, _, err := ctl(ctx).GroupBy("").CreateIfNotExist(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlGroupBy.Name, "CreateIfNotExist").Error()},
			{"Having", func() error { _, _, err := ctl(ctx).Having("").CreateIfNotExist(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlHaving.Name, "CreateIfNotExist").Error()},
			{"Select", func() error { _, _, err := ctl(ctx).Select("").CreateIfNotExist(map[string]any{}); return err }, fmt.Errorf(UnsupportedControllerError, ctlSelect.Name, "CreateIfNotExist").Error()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Select column validation", func(t *testing.T) {
		tests := []struct {
			name    string
			cols    []string
			wantErr string
		}{
			{"single invalid", []string{"age"}, "select columns validate error: [age] not exist"},
			{"two invalid", []string{"age", "happy"}, "select columns validate error: [age; happy] not exist"},
			{"three invalid", []string{"age", "happy", "damnit"}, "select columns validate error: [age; happy; damnit] not exist"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ctl(ctx).Select(tt.cols).FindOneModel(&test.Source{})
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("OrderBy column validation", func(t *testing.T) {
		tests := []struct {
			name    string
			cols    []string
			wantErr string
		}{
			{"single invalid", []string{"age"}, "orderBy columns validate error: [age] not exist"},
			{"two invalid", []string{"age", "happy"}, "orderBy columns validate error: [age; happy] not exist"},
			{"three invalid", []string{"age", "happy", "damnit"}, "orderBy columns validate error: [age; happy; damnit] not exist"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ctl(ctx).OrderBy(tt.cols).FindOneModel(&test.Source{})
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("GroupBy column validation", func(t *testing.T) {
		tests := []struct {
			name    string
			cols    []string
			wantErr string
		}{
			{"single invalid", []string{"age"}, "groupBy columns validate error: [age] not exist"},
			{"two invalid", []string{"age", "happy"}, "groupBy columns validate error: [age; happy] not exist"},
			{"three invalid", []string{"age", "happy", "damnit"}, "groupBy columns validate error: [age; happy; damnit] not exist"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ctl(ctx).GroupBy(tt.cols).FindOneModel(&test.Source{})
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Combined column validation", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"GroupBy+OrderBy 1", func() error {
				return ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age"}).FindOneModel(&test.Source{})
			}, "orderBy columns validate error: [age] not exist"},
			{"GroupBy+OrderBy 2", func() error {
				return ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy"}).FindOneModel(&test.Source{})
			}, "orderBy columns validate error: [age; happy] not exist"},
			{"GroupBy+OrderBy 3", func() error {
				return ctl(ctx).GroupBy([]string{"test"}).OrderBy([]string{"age", "happy", "damnit"}).FindOneModel(&test.Source{})
			}, "orderBy columns validate error: [age; happy; damnit] not exist"},
			{"OrderBy+GroupBy 1", func() error {
				return ctl(ctx).OrderBy([]string{"age"}).GroupBy([]string{"test"}).FindOneModel(&test.Source{})
			}, "groupBy columns validate error: [test] not exist"},
			{"OrderBy+GroupBy 2", func() error {
				return ctl(ctx).OrderBy([]string{"age", "happy"}).GroupBy([]string{"test", "test2"}).FindOneModel(&test.Source{})
			}, "groupBy columns validate error: [test; test2] not exist"},
			{"OrderBy+GroupBy 3", func() error {
				return ctl(ctx).OrderBy([]string{"age", "happy", "damnit"}).GroupBy([]string{"test", "test2", "test3"}).FindOneModel(&test.Source{})
			}, "groupBy columns validate error: [test; test2; test3] not exist"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("type errors", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Select empty slice", func() error { return ctl(ctx).Select([]string{}).FindOneModel(&test.Source{}) }, ""},
			{"Select with int slice", func() error { return ctl(ctx).Select([]int{1}).FindOneModel(&test.Source{}) }, SelectColumnsTypeError},
			{"OrderBy with array", func() error { _, err := ctl(ctx).OrderBy([1]string{"id"}).FindAll(); return err }, OrderByColumnsTypeError},
			{"OrderBy contains empty", func() error { _, err := ctl(ctx).OrderBy([]string{""}).FindAll(); return err }, ""},
			{"GroupBy with array", func() error {
				return ctl(ctx).Select([]string{"name"}).GroupBy([1]string{"name"}).FindAllModel(&[]struct {
					Name string `db:"name"`
				}{})
			}, GroupByColumnsTypeError},
			{"FindOneModel non-pointer", func() error { return ctl(ctx).FindOneModel(test.Source{}) }, ModelTypeNotStructError},
			{"FindOneModel slice pointer", func() error { return ctl(ctx).FindOneModel(&[]test.Source{}) }, ModelTypeNotStructError},
			{"FindAllModel non-pointer", func() error { return ctl(ctx).FindAllModel([]test.Source{}) }, ModelTypeNotSliceError},
			{"FindAllModel struct pointer", func() error { return ctl(ctx).FindAllModel(&test.Source{}) }, ModelTypeNotSliceError},
			{"GetOrCreate empty data", func() error { _, err := ctl(ctx).GetOrCreate(map[string]any{}); return err }, strings.ToLower("GetOrCreate") + " " + DataEmptyError},
			{"CreateOrUpdate empty data", func() error { _, _, err := ctl(ctx).CreateOrUpdate(map[string]any{}); return err }, strings.ToLower("CreateOrUpdate") + " " + DataEmptyError},
			{"CreateIfNotExist empty data", func() error { _, _, err := ctl(ctx).CreateIfNotExist(map[string]any{}); return err }, strings.ToLower("CreateIfNotExist") + " " + DataEmptyError},
			{"Count with nil filter", func() error { _, err := ctl(ctx).Filter(nil).Count(); return err }, fmt.Errorf(queryset.UnsupportedFilterTypeError, "nil").Error()},
			{"Exist with nil filter", func() error { _, err := ctl(ctx).Filter(nil).Exist(); return err }, fmt.Errorf(queryset.UnsupportedFilterTypeError, "nil").Error()},
			{"List with nil filter", func() error { _, _, err := ctl(ctx).Filter(nil).List(); return err }, fmt.Errorf(queryset.UnsupportedFilterTypeError, "nil").Error()},
			{"Update with nonexistent column", func() error {
				_, err := ctl(ctx).Filter(Cond{"id": 11}).Update(map[string]any{"name": "test", "age": 18})
				return err
			}, fmt.Errorf(UpdateColumnNotExistError, "age").Error()},
			{"Update with empty map", func() error { _, err := ctl(ctx).Filter(Cond{"id": 11}).Update(map[string]any{}); return err }, "update " + DataEmptyError},
			{"Create with empty map", func() error { _, err := ctl(ctx).Create(map[string]any{}); return err }, "create " + DataEmptyError},
			{"Create with empty slice map", func() error { _, err := ctl(ctx).Create([]map[string]any{}); return err }, "bulk create " + DataEmptyError},
			{"Create with invalid type", func() error { _, err := ctl(ctx).Create(1); return err }, fmt.Sprintf(CreateDataTypeError, "int")},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if tt.wantErr == "" {
					if err != nil {
						t.Errorf("expected no error, got %q", err.Error())
					}
					return
				}
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("prior error propagation", func(t *testing.T) {
		tests := []struct {
			name    string
			fn      func() error
			wantErr string
		}{
			{"Remove with OrderBy error", func() error { _, err := ctl(ctx).Filter(Cond{"id": 1}).OrderBy([1]string{"-id"}).Remove(); return err }, OrderByColumnsTypeError},
			{"Update with OrderBy error", func() error {
				_, err := ctl(ctx).Filter(Cond{"id": 1}).OrderBy([1]string{"-id"}).Update(map[string]any{})
				return err
			}, OrderByColumnsTypeError},
			{"Count with OrderBy error", func() error { _, err := ctl(ctx).OrderBy([1]string{"-id"}).Count(); return err }, OrderByColumnsTypeError},
			{"List with OrderBy error", func() error { _, _, err := ctl(ctx).OrderBy([1]string{"-id"}).List(); return err }, OrderByColumnsTypeError},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("got %q, want %q", err.Error(), tt.wantErr)
				}
			})
		}
	})

	t.Run("Contains empty value", func(t *testing.T) {
		res, err := ctl(ctx).Filter(Cond{"name__contains": []string{"Ac", ""}}).OrderBy("id").Limit(10, 1).FindAll()
		if err != nil {
			if err.Error() != "operator [contains] unsupported value empty" {
				t.Errorf("got %q, want %q", err.Error(), "operator [contains] unsupported value empty")
			}
		} else if len(res) != 5 {
			t.Errorf("got len %d, want 5", len(res))
		}
	})

	t.Run("Reset", func(t *testing.T) {
		cli := ctl(ctx)
		_ = cli.Filter().FindOneModel(&test.Source{})

		_, err := cli.Create(map[string]any{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "[Filter] not supported for Create" {
			t.Errorf("got %q, want %q", err.Error(), "[Filter] not supported for Create")
		}

		cli = cli.Reset()

		_, err = cli.Create(map[string]any{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "create data is empty" {
			t.Errorf("got %q, want %q", err.Error(), "create data is empty")
		}

		if _, err := cli.Create(map[string]any{"id": 1000, "name": "rest", "description": "test rest"}); err != nil {
			t.Fatalf("Create error: %v", err)
		}

		if _, err := cli.Filter(Cond{"id": 1000}).Remove(); err != nil {
			t.Fatalf("Remove error: %v", err)
		}

		res, err := cli.Filter(Cond{"id": 1000}).FindOne()
		if err != nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got len %d, want 0", len(res))
		}
	})

	t.Run("DB error branches", func(t *testing.T) {
		mysqlAddr := getMysqlAddress()

		type badSource struct {
			ID  int64  `db:"id"`
			Bad string `db:"not_exist"`
		}

		badTableCtl := NewController(
			go_zero.NewOperator(go_zero.NewMysql(mysqlAddr), go_zero.WithTableName("not_exist_table")),
			test.Source{},
		)
		badModelCtl := NewController(
			go_zero.NewOperator(go_zero.NewMysql(mysqlAddr), go_zero.WithTableName("source")),
			badSource{},
		)

		if _, err := badTableCtl(ctx).FindOne(); err == nil {
			t.Fatal("expected FindOne error, got nil")
		}

		if _, err := badTableCtl(ctx).FindAll(); err == nil {
			t.Fatal("expected FindAll error, got nil")
		}
		var badRows []badSource
		if err := badTableCtl(ctx).FindAllModel(&badRows); err == nil {
			t.Fatal("expected FindAllModel error, got nil")
		}

		if _, err := badModelCtl(ctx).Filter(Cond{"id": 11}).Update(map[string]any{"description": nil}); err == nil {
			t.Fatal("expected Update error, got nil")
		}
		if impl := ctl(ctx).(*Impl); impl != nil {
			if _, err := impl.update(map[string]any{}); err == nil {
				t.Fatal("expected update empty map error, got nil")
			}
		}

		if _, _, err := badTableCtl(ctx).List(); err == nil {
			t.Fatal("expected List count error, got nil")
		}
		if _, _, err := badModelCtl(ctx).Filter(Cond{"id": 11}).List(); err == nil {
			t.Fatal("expected List FindAll error, got nil")
		}

		if _, err := badModelCtl(ctx).GetOrCreate(map[string]any{"id": 98765}); err == nil {
			t.Fatal("expected GetOrCreate error, got nil")
		}

		if _, _, err := badTableCtl(ctx).Filter(Cond{"id": 1}).CreateOrUpdate(map[string]any{"id": 1}); err == nil {
			t.Fatal("expected CreateOrUpdate exist error, got nil")
		}
		if _, _, err := ctl(ctx).Filter(Cond{"id": 11}).CreateOrUpdate(map[string]any{"description": nil}); err == nil {
			t.Fatal("expected CreateOrUpdate update error, got nil")
		}
		if _, _, err := ctl(ctx).Filter(Cond{"id": 98766}).CreateOrUpdate(map[string]any{"id": 98766}); err == nil {
			t.Fatal("expected CreateOrUpdate create error, got nil")
		}

		if _, _, err := badTableCtl(ctx).CreateIfNotExist(map[string]any{"id": 1}); err == nil {
			t.Fatal("expected CreateIfNotExist exist error, got nil")
		}
		if _, _, err := ctl(ctx).CreateIfNotExist(map[string]any{"id": 98767}); err == nil {
			t.Fatal("expected CreateIfNotExist create error, got nil")
		}
	})
}
