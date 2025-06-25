package clickhouse

var Operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",

	"in":              "`%s`%s IN",
	"not_in":          "`%s`%s IN",
	"contains":        "`%s`%s LIKE ?",
	"not_contains":    "`%s`%s LIKE ?",
	"icontains":       "`%s`%s ILIKE ?",
	"not_icontains":   "`%s`%s ILIKE ?",
	"startswith":      "`%s`%s LIKE ?",
	"not_startswith":  "`%s`%s LIKE ?",
	"istartswith":     "`%s`%s ILIKE ?",
	"not_istartswith": "`%s`%s ILIKE ?",
	"endswith":        "`%s`%s LIKE ?",
	"not_endswith":    "`%s`%s LIKE ?",
	"iendswith":       "`%s`%s ILIKE ?",
	"not_iendswith":   "`%s`%s ILIKE ?",
}
