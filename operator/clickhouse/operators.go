package clickhouse

var Operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",
	"len":     "length(`%s`) = ?",
	"toData":  "`%s` = toDate(?)",

	"in":              "`%s`%s in",
	"not_in":          "`%s`%s in",
	"contains":        "`%s`%s like ?",
	"not_contains":    "`%s`%s like ?",
	"icontains":       "`%s`%s ilike ?",
	"not_icontains":   "`%s`%s ilike ?",
	"startswith":      "`%s`%s like ?",
	"not_startswith":  "`%s`%s like ?",
	"istartswith":     "`%s`%s ilike ?",
	"not_istartswith": "`%s`%s ilike ?",
	"endswith":        "`%s`%s like ?",
	"not_endswith":    "`%s`%s like ?",
	"iendswith":       "`%s`%s ilike ?",
	"not_iendswith":   "`%s`%s ilike ?",
}
