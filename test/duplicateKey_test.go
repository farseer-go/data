package test

import (
	"github.com/farseer-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestDuplicateContext struct {
	Account data.TableSet[AccountPO] `data:"migrate"`
}

type AccountPO struct {
	// 用户名称
	Name string `gorm:"primaryKey"`
	// 用户年龄
	Age int
}

// 创建索引
func (*AccountPO) CreateIndex() map[string]data.IdxField {
	return map[string]data.IdxField{
		"idx_age": {true, "age"},
	}
}

func TestDuplicateKey(t *testing.T) {
	context := data.NewContext[TestDuplicateContext]("test")
	_, _ = context.Account.Delete()

	if err := context.Account.Insert(&AccountPO{Name: "aaa", Age: 8}); err != nil {
		assert.Error(t, err, "insert error")
	}

	if err := context.Account.InsertIgnoreDuplicateKey(&AccountPO{Name: "aaa", Age: 2}); err != nil {
		// number=1062
		// Duplicate entry 'aaa' for key 'account.PRIMARY'
		assert.Error(t, err, "insert error")
	}

	if err := context.Account.InsertIgnoreDuplicateKey(&AccountPO{Name: "bbb", Age: 8}); err != nil {
		// number=1062
		// Duplicate entry '8' for key 'account.idx_age'
		assert.Error(t, err, "insert error")
	}
}
