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
	"is_null": "`%s` IS NULL",

	"in":          "`%s`%s in",
	"between":     "`%s`%s BETWEEN ? AND ?",
	"contains":    "`%s`%s like ?",
	"icontains":   "`%s`%s ilike ?",
	"startswith":  "`%s`%s like ?",
	"istartswith": "`%s`%s ilike ?",
	"endswith":    "`%s`%s like ?",
	"iendswith":   "`%s`%s ilike ?",
}

var Methods = map[string]string{
	"toDate":     "toDate(?)",
	"toDateTime": "toDateTime(?)",
}
