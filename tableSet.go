package data

import (
	"github.com/farseer-go/collections"
	"gorm.io/gorm"
	"time"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	// 上下文（用指针的方式，共享同一个上下文）
	dbContext *DbContext
	// 表名
	tableName string
	db        *gorm.DB
	err       error
}

// Init 在反射的时候会调用此方法
func (table *TableSet[Table]) Init(dbContext *DbContext, tableName string) {
	table.dbContext = dbContext
	table.reInit()
	table.SetTableName(tableName)
}

// 重新初始化
func (table *TableSet[Table]) reInit() *gorm.DB {
	// Data Source ClientName，参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
	table.db, table.err = gorm.Open(table.dbContext.getDriver(), &gorm.Config{})
	if table.err != nil {
		panic(table.err.Error())
	}
	table.db = table.db.Table(table.tableName)
	table.setPool()
	return table.db
}

// SetTableName 设置表名
func (table *TableSet[Table]) SetTableName(tableName string) *TableSet[Table] {
	table.tableName = tableName
	if table.db != nil {
		table.db.Table(table.tableName)
	}
	return table
}

// GetTableName 获取表名称
func (table *TableSet[Table]) GetTableName() string {
	return table.tableName
}

// 设置池大小
func (table *TableSet[Table]) setPool() {
	sqlDB, _ := table.db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	if table.dbContext.dbConfig.PoolMinSize > 0 {
		sqlDB.SetMaxIdleConns(table.dbContext.dbConfig.PoolMinSize)
	}
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	if table.dbContext.dbConfig.PoolMaxSize > 0 {
		sqlDB.SetMaxOpenConns(table.dbContext.dbConfig.PoolMaxSize)
	}
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}

// Select 筛选字段
func (table *TableSet[Table]) Select(query interface{}, args ...interface{}) *TableSet[Table] {
	table.db.Select(query, args...)
	return table
}

// Where 条件
func (table *TableSet[Table]) Where(query interface{}, args ...interface{}) *TableSet[Table] {
	table.db.Where(query, args...)
	return table
}

// Order 排序
func (table *TableSet[Table]) Order(value interface{}) *TableSet[Table] {
	table.db.Order(value)
	return table
}

// Desc 倒序
func (table *TableSet[Table]) Desc(fieldName string) *TableSet[Table] {
	table.db.Order(fieldName + " desc")
	return table
}

// Asc 正序
func (table *TableSet[Table]) Asc(fieldName string) *TableSet[Table] {
	table.Order(fieldName + " asc")
	return table
}

// Limit 限制记录数
func (table *TableSet[Table]) Limit(limit int) *TableSet[Table] {
	table.db.Limit(limit)
	return table
}

// ToList 返回结果集
func (table *TableSet[Table]) ToList() collections.List[Table] {
	defer table.reInit()
	var lst []Table
	table.db.Find(&lst)
	return collections.NewList(lst...)
}

// ToArray 返回结果集
func (table *TableSet[Table]) ToArray() []Table {
	defer table.reInit()
	var lst []Table
	table.db.Find(&lst)
	return lst
}

// ToPageList 返回分页结果集
func (table *TableSet[Table]) ToPageList(pageSize int, pageIndex int) collections.PageList[Table] {
	defer table.reInit()
	offset := (pageIndex - 1) * pageSize
	var lst []Table
	table.db.Offset(offset).Limit(pageSize).Find(&lst)
	return collections.NewPageList[Table](collections.NewList(lst...), table.Count())
}

// ToEntity 返回单个对象
func (table *TableSet[Table]) ToEntity() Table {
	defer table.reInit()
	var entity Table
	table.db.Take(&entity)
	return entity
}

// Count 返回表中的数量
func (table *TableSet[Table]) Count() int64 {
	defer table.reInit()
	var count int64
	table.db.Count(&count)
	return count
}

// IsExists 是否存在记录
func (table *TableSet[Table]) IsExists() bool {
	defer table.reInit()
	var count int64
	table.db.Count(&count)
	return count > 0
}

// Insert 新增记录
func (table *TableSet[Table]) Insert(po *Table) {
	defer table.reInit()
	table.db.Create(po)
}

// Update 修改记录
func (table *TableSet[Table]) Update(po Table) int64 {
	defer table.reInit()
	result := table.db.Updates(po)
	return result.RowsAffected
}

// UpdateValue 修改单个字段
func (table *TableSet[Table]) UpdateValue(column string, value interface{}) {
	defer table.reInit()
	table.db.Update(column, value)
}

// Delete 删除记录
func (table *TableSet[Table]) Delete() int64 {
	defer table.reInit()
	result := table.db.Delete(nil)
	return result.RowsAffected
}

// GetString 获取单条记录中的单个string类型字段值
func (table *TableSet[Table]) GetString(fieldName string) string {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val string
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}

// GetInt 获取单条记录中的单个int类型字段值
func (table *TableSet[Table]) GetInt(fieldName string) int {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val int
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}

// GetLong 获取单条记录中的单个int64类型字段值
func (table *TableSet[Table]) GetLong(fieldName string) int64 {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val int64
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}

// GetBool 获取单条记录中的单个bool类型字段值
func (table *TableSet[Table]) GetBool(fieldName string) bool {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val bool
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}

// GetFloat32 获取单条记录中的单个float32类型字段值
func (table *TableSet[Table]) GetFloat32(fieldName string) float32 {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float32
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}

// GetFloat64 获取单条记录中的单个float64类型字段值
func (table *TableSet[Table]) GetFloat64(fieldName string) float64 {
	defer table.reInit()
	rows, _ := table.db.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float64
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//table.ScanRows(rows, &user)
	}
	return val
}
