package mysql

var operators = map[string]string{
	"exact":   "`%s` = ?",
	"exclude": "`%s` != ?",
	"iexact":  "`%s` LIKE ?",
	"gt":      "`%s` > ?",
	"gte":     "`%s` >= ?",
	"lt":      "`%s` < ?",
	"lte":     "`%s` <= ?",
	"len":     "LENGTH(`%s`) = ?",

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

type Operator struct{}

func NewOperator() *Operator {
	return &Operator{}
}

func (d *Operator) OperatorSQL(operator string) string {
	return operators[operator]
}
