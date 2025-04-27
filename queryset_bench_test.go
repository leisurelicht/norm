package norm

import (
	"context"
	"testing"

	"github.com/leisurelicht/norm/operator/mysql" // Import mysql package
)

type MockOperator struct {
	mysql.Operator // Embed mysql.Operator to inherit its methods
}

// IsSelectKey implements the Operator interface with the correct signature
func (m *MockOperator) IsSelectKey(column string) bool {
	return true // Always return true for testing purposes
}

// BulkInsert implements the Operator interface with the correct signature
func (m *MockOperator) BulkInsert(ctx context.Context, conn any, sql string, args []string, data []map[string]any) (int64, error) {
	return 0, nil // Just a stub implementation for testing
}

// setupQuerySet creates a new QuerySet with the MockOperator for benchmarking
func setupQuerySet() QuerySet {
	return NewQuerySet(&MockOperator{})
}

// createSimpleFilterMap creates a map for simple filter conditions
func createSimpleFilterMap() map[string]any {
	return map[string]any{
		"id":   1,
		"name": "test",
	}
}

// createComplexFilterMap creates a map for complex filter conditions
func createComplexFilterMap() map[string]any {
	return map[string]any{
		"id":                  1,
		"name":                "test",
		"age__gte":            18,
		"email__contains":     "example.com",
		"created_at__between": []string{"2020-01-01", "2023-01-01"},
		"status__in":          []int{1, 2, 3, 4, 5},
		"| country":           "USA",
		"| city":              "New York",
	}
}

// BenchmarkQuerySet_SimpleFilter benchmarks filtering with basic conditions
func BenchmarkQuerySet_SimpleFilter(b *testing.B) {
	qs := setupQuerySet()
	filter := Cond(createSimpleFilterMap())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.FilterToSQL(notNot, filter)
		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_ComplexFilter benchmarks filtering with complex conditions
func BenchmarkQuerySet_ComplexFilter(b *testing.B) {
	qs := setupQuerySet()
	filter := Cond(createComplexFilterMap())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.FilterToSQL(notNot, filter)
		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_MultipleFilters benchmarks applying multiple filter conditions
func BenchmarkQuerySet_MultipleFilters(b *testing.B) {
	qs := setupQuerySet()
	filter1 := Cond(createSimpleFilterMap())
	filter2 := Cond(createComplexFilterMap())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.FilterToSQL(notNot, filter1)
		qs.FilterToSQL(notNot, "OR", filter2)
		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_Where benchmarks using direct WHERE conditions
func BenchmarkQuerySet_Where(b *testing.B) {
	qs := setupQuerySet()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.WhereToSQL("`id` = ? AND `name` = ?", 1, "test")
		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_CompleteQuery benchmarks generating a complete query with multiple parts
func BenchmarkQuerySet_CompleteQuery(b *testing.B) {
	qs := setupQuerySet()
	filter := Cond(createComplexFilterMap())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.SelectToSQL("id, name, age, email")
		qs.FilterToSQL(notNot, filter)
		qs.OrderByToSQL([]string{"name", "-age"})
		qs.LimitToSQL(10, 1)
		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_BuildLargeQuery benchmarks building a large query with multiple filter groups
func BenchmarkQuerySet_BuildLargeQuery(b *testing.B) {
	qs := setupQuerySet()

	// Create 5 different filter conditions
	filters := []any{
		Cond(map[string]any{"id": 1, "name": "test1"}),
		Cond(map[string]any{"age__gte": 18, "status": 1}),
		Cond(map[string]any{"email__contains": "example.com"}),
		OR(map[string]any{"country": "USA", "region": "Canada"}), // Fixed duplicate key
		OR(map[string]any{"created_at__between": []string{"2020-01-01", "2023-01-01"}}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()

		// Apply each filter with alternating AND/OR conjunctions
		qs.FilterToSQL(notNot, filters[0])
		qs.FilterToSQL(notNot, filters[1])
		qs.FilterToSQL(notNot, filters[2])
		qs.FilterToSQL(notNot, filters[3])
		qs.FilterToSQL(notNot, filters[4])

		qs.OrderByToSQL([]string{"id", "name", "-age"})
		qs.LimitToSQL(20, 3)
		qs.GroupByToSQL([]string{"status", "country"})
		qs.HavingToSQL("COUNT(*) > ?", 5)

		qs.GetQuerySet()
	}
}

// BenchmarkQuerySet_FilterExclude benchmarks filter and exclude operations
func BenchmarkQuerySet_FilterExclude(b *testing.B) {
	qs := setupQuerySet()
	filter := Cond(createSimpleFilterMap())
	exclude := Cond(map[string]any{"status": 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qs.(*QuerySetImpl).Reset()
		qs.FilterToSQL(notNot, filter)
		qs.FilterToSQL(isNot, exclude) // Using 1 for exclude
		qs.GetQuerySet()
	}
}

// 添加一个简单的测试函数，以确保基准测试可以正确运行
func TestQuerySetFunctionality(t *testing.T) {
	qs := setupQuerySet()
	filter := Cond(createSimpleFilterMap())

	qs.FilterToSQL(notNot, filter)
	sql, args := qs.GetQuerySet()

	if sql == "" || len(args) == 0 {
		t.Error("Expected non-empty SQL and args")
	}
}

// 添加示例运行函数，可以直接在IDE中运行
func ExampleRunBenchmarks() {
	// 此函数仅作为示例，展示如何运行基准测试
	// 在终端中运行: go test -bench=. -benchmem
	// 或者: go test -bench=BenchmarkQuerySet_SimpleFilter -benchmem
}
