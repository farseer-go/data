package data

import (
	"context"
	"fmt"
	"time"

	"github.com/farseer-go/fs/core"
)

// 确保实现了IConnectionChecker接口
var _ core.IConnectionChecker = (*connectionChecker)(nil)

type connectionChecker struct{}

// Check 检查连接字符串是否能成功连接到数据库
// 实现IConnectionChecker接口
func (c *connectionChecker) Check(configString string) (bool, error) {
	if configString == "" {
		return false, fmt.Errorf("连接字符串不能为空")
	}

	// 使用NewInternalContext解析配置字符串并验证基础配置
	internalCtx := NewInternalContext(configString)

	if internalCtx.dbConfig.ConnectionString == "" {
		return false, fmt.Errorf("连接字符串配置不正确：%s", configString)
	}

	if internalCtx.dbConfig.DataType == "" {
		return false, fmt.Errorf("数据库类型配置不正确：%s", configString)
	}

	// 使用open函数建立数据库连接
	gormDB, err := open(internalCtx.dbConfig)
	if err != nil {
		return false, err
	}

	// 获取数据库连接并进行ping测试
	sqlDB, err := gormDB.DB()
	if err != nil {
		return false, fmt.Errorf("获取数据库连接失败[%s]：%s", internalCtx.dbConfig.DataType, err.Error())
	}
	defer func() { sqlDB.Close() }()

	if err := sqlDB.Ping(); err != nil {
		return false, fmt.Errorf("Ping数据库失败[%s]：%s", internalCtx.dbConfig.DataType, err.Error())
	}

	return true, nil
}

// CheckConnection 检查连接字符串是否能成功连接到数据库
// configString 格式参考：
// DataType=mysql,PoolMaxSize=5,PoolMinSize=1,ConnectionString=user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
// DataType=sqlserver,PoolMaxSize=5,PoolMinSize=1,ConnectionString=sqlserver://user:123456@127.0.0.1:9930?database=dbname
// DataType=clickhouse,PoolMaxSize=5,PoolMinSize=1,ConnectionString=clickhouse://user:123456@127.0.0.1:9000/dbname?dial_timeout=10s&read_timeout=60s
// DataType=postgresql,PoolMaxSize=5,PoolMinSize=1,ConnectionString=host=127.0.0.1 user=user password=123456 dbname=dbname port=9920 sslmode=disable TimeZone=Asia/Shanghai
// DataType=sqlite,PoolMaxSize=5,PoolMinSize=1,ConnectionString=gorm.db
// 返回值：(成功, 错误信息)
// CheckWithTimeout 带超时时间的连接检查
// timeout 为0时使用默认的10秒超时，参数类型为 time.Duration
func (c *connectionChecker) CheckWithTimeout(configString string, timeout time.Duration) (bool, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		success bool
		err     error
	}
	resultChan := make(chan result, 1)

	go func() {
		success, err := c.Check(configString)
		resultChan <- result{success: success, err: err}
	}()

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("连接检查超时，超时时间：%v", timeout)
	case res := <-resultChan:
		return res.success, res.err
	}
}
