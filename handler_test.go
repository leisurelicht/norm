package norm

import (
	mysqlOp "github.com/leisurelicht/norm/operator/mysql"
	"github.com/leisurelicht/norm/test"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"testing"
)

const (
	mysqlAddress = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&loc=Asia%2FShanghai&parseTime=true"
)

func TestController(t *testing.T) {
	policyCenterConn := sqlx.NewMysql(mysqlAddress)
	ctl := NewController(policyCenterConn, mysqlOp.NewOperator(), test.Source{})
	num, err := ctl(nil).Count()
	if err != nil {
		t.Error(err)
	}
	t.Log(num)
	res, err := ctl(nil).Filter(Cond{"id": 1}).FindOne()
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
