package mysql

var Operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",
	"len":     "LENGTH(`%s`) = ?",
	"is_null": "`%s` IS NULL",

	"in":          "`%s`%s IN",
	"between":     "`%s`%s BETWEEN ? AND ?",
	"contains":    "`%s`%s LIKE BINARY ?",
	"icontains":   "`%s`%s LIKE ?",
	"startswith":  "`%s`%s LIKE BINARY ?",
	"istartswith": "`%s`%s LIKE ?",
	"endswith":    "`%s`%s LIKE BINARY ?",
	"iendswith":   "`%s`%s LIKE ?",

	"unimplemented": "UNIMPLEMENTED", // Placeholder for unimplemented operators

}

var Methods = map[string]string{}
