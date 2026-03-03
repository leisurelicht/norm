package norm

import (
	"context"
	"testing"

	ioperator "github.com/leisurelicht/norm/internal/operator"
)

// benchModel is a small representative struct to benchmark handler operations.
type benchModel struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	IsDeleted   int64  `db:"is_deleted"`
}

// benchOperator is a lightweight in‑memory implementation of Operator used only for benchmarks.
// It focuses on handler/queryset overhead rather than real DB cost.
type benchOperator struct {
	tableName   string
	placeholder string
	dbTag       string
}

func newBenchOperator() Operator {
	return &benchOperator{
		tableName:   "`bench_model`",
		placeholder: "?",
		dbTag:       "db",
	}
}

// Operator interface implementation

func (op *benchOperator) OperatorSQL(operator, method string) string {
	// Minimal subset that keeps queryset happy for common operators.
	switch operator {
	case "exact":
		return "`%s` = ?"
	case "gt":
		return "`%s` > ?"
	case "gte":
		return "`%s` >= ?"
	case "lt":
		return "`%s` < ?"
	case "lte":
		return "`%s` <= ?"
	default:
		// Fallback simple equality to avoid errors in benchmarks that do not care about exact SQL.
		return "`%s` = ?"
	}
}

func (op *benchOperator) GetPlaceholder() string {
	return op.placeholder
}

func (op *benchOperator) GetDBTag() string {
	return op.dbTag
}

func (op *benchOperator) GetTableName() string {
	return op.tableName
}

func (op *benchOperator) SetTableName(tableName string) ioperator.Operator {
	op.tableName = tableName
	return op
}

func (op *benchOperator) WithSession(session any) ioperator.Operator {
	// Session is ignored in benchmarks.
	return op
}

func (op *benchOperator) Insert(ctx context.Context, query string, args ...any) (int64, error) {
	return 1, nil
}

func (op *benchOperator) BulkInsert(ctx context.Context, query string, args []string, data []map[string]any) (int64, error) {
	return int64(len(data)), nil
}

func (op *benchOperator) Remove(ctx context.Context, query string, args ...any) (int64, error) {
	return 1, nil
}

func (op *benchOperator) Update(ctx context.Context, query string, args ...any) (int64, error) {
	return 1, nil
}

func (op *benchOperator) Count(ctx context.Context, condition string, args ...any) (int64, error) {
	return 100, nil
}

func (op *benchOperator) Exist(ctx context.Context, condition string, args ...any) (bool, error) {
	return true, nil
}

func (op *benchOperator) FindOne(ctx context.Context, model any, query string, args ...any) error {
	// Leave model as zero value; handler will still exercise mapping logic.
	return nil
}

func (op *benchOperator) FindAll(ctx context.Context, model any, query string, args ...any) error {
	// Zero-value slice is fine for exercising handler/queryset paths.
	return nil
}

// Benchmarks

// BenchmarkHandler_Create_Struct measures Create with a struct input (using reflection + Struct2Map).
func BenchmarkHandler_Create_Struct(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().Create(benchModel{
			ID:          int64(i),
			Name:        "bench",
			Description: "create_struct",
			IsDeleted:   0,
		})
	}
}

// BenchmarkHandler_Create_Map measures Create with a map input.
func BenchmarkHandler_Create_Map(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	data := map[string]any{
		"id":          int64(1),
		"name":        "bench",
		"description": "create_map",
		"is_deleted":  int64(0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data["id"] = int64(i)
		_, _ = ctrl.Reset().Create(data)
	}
}

// BenchmarkHandler_BulkCreate_Maps measures bulkCreate using []map[string]any.
func BenchmarkHandler_BulkCreate_Maps(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	const batchSize = 20
	batch := make([]map[string]any, batchSize)
	for i := 0; i < batchSize; i++ {
		batch[i] = map[string]any{
			"id":          int64(i),
			"name":        "bench",
			"description": "bulk_create",
			"is_deleted":  int64(0),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().Create(batch)
	}
}

// BenchmarkHandler_FindOne_Map measures FindOne returning a map.
func BenchmarkHandler_FindOne_Map(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Where("`id` = ?", 1).
			FindOne()
	}
}

// BenchmarkHandler_FindOne_Model measures FindOneModel into a struct pointer.
func BenchmarkHandler_FindOne_Model(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	var m benchModel

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctrl.Reset().
			Where("`id` = ?", 1).
			FindOneModel(&m)
	}
}

// BenchmarkHandler_FindAll_Map measures FindAll returning []map.
func BenchmarkHandler_FindAll_Map(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Where("`is_deleted` = ?", 0).
			OrderBy([]string{"id"}).
			Limit(10, 1).
			FindAll()
	}
}

// BenchmarkHandler_FindOne_Map_SelectNoAlias measures FindOne map path with Select(string) without alias.
func BenchmarkHandler_FindOne_Map_SelectNoAlias(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Select("`id`, `name`").
			Where("`id` = ?", 1).
			FindOne()
	}
}

// BenchmarkHandler_FindOne_Map_SelectAliasError measures FindOne alias validation path (expected error).
func BenchmarkHandler_FindOne_Map_SelectAliasError(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Select("`id` AS `uid`").
			Where("`id` = ?", 1).
			FindOne()
	}
}

// BenchmarkHandler_FindAll_Map_SelectNoAlias measures FindAll map path with Select(string) without alias.
func BenchmarkHandler_FindAll_Map_SelectNoAlias(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Select("`id`, `name`").
			Where("`is_deleted` = ?", 0).
			OrderBy([]string{"id"}).
			Limit(10, 1).
			FindAll()
	}
}

// BenchmarkHandler_FindAll_Map_SelectAliasError measures FindAll alias validation path (expected error).
func BenchmarkHandler_FindAll_Map_SelectAliasError(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctrl.Reset().
			Select("`id` uid").
			Where("`is_deleted` = ?", 0).
			OrderBy([]string{"id"}).
			Limit(10, 1).
			FindAll()
	}
}

// BenchmarkHandler_FindAll_Model measures FindAllModel into a slice pointer.
func BenchmarkHandler_FindAll_Model(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	var list []benchModel

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctrl.Reset().
			Where("`is_deleted` = ?", 0).
			OrderBy([]string{"id"}).
			Limit(10, 1).
			FindAllModel(&list)
	}
}

// BenchmarkHandler_List measures the combined List (Count + FindAll) path.
func BenchmarkHandler_List(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ctrl.Reset().
			Where("`is_deleted` = ?", 0).
			OrderBy([]string{"id"}).
			Limit(10, 1).
			List()
	}
}

// BenchmarkHandler_CreateOrUpdate measures CreateOrUpdate path.
func BenchmarkHandler_CreateOrUpdate(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	data := map[string]any{
		"id":          int64(1),
		"name":        "bench",
		"description": "create_or_update",
		"is_deleted":  int64(0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data["id"] = int64(i)
		_, _, _ = ctrl.Reset().
			Where("`id` = ?", data["id"]).
			CreateOrUpdate(data)
	}
}

// BenchmarkHandler_GetOrCreate measures GetOrCreate path.
func BenchmarkHandler_GetOrCreate(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	data := map[string]any{
		"id":          int64(1),
		"name":        "bench",
		"description": "get_or_create",
		"is_deleted":  int64(0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data["id"] = int64(i)
		_, _ = ctrl.Reset().GetOrCreate(data)
	}
}

// BenchmarkHandler_CreateIfNotExist measures CreateIfNotExist path.
func BenchmarkHandler_CreateIfNotExist(b *testing.B) {
	b.ReportAllocs()

	ctrlFactory := NewController(newBenchOperator(), benchModel{})
	ctrl := ctrlFactory(context.Background())

	data := map[string]any{
		"id":          int64(1),
		"name":        "bench",
		"description": "create_if_not_exist",
		"is_deleted":  int64(0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data["id"] = int64(i)
		_, _, _ = ctrl.Reset().CreateIfNotExist(data)
	}
}
