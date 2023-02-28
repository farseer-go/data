package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	// 上下文（用指针的方式，共享同一个上下文）
	dbContext *DbContext
	// 表名
	tableName string
	gormDB    *gorm.DB
	err       error
	// 字段筛选（官方再第二次设置时，会覆盖第一次的设置，因此需要暂存）
	selectList collections.ListAny
	whereList  collections.List[whereQuery]
	orderList  collections.ListAny
	limit      int
}

type whereQuery struct {
	query any
	args  []any
}

// Init 在反射的时候会调用此方法
func (table *TableSet[Table]) Init(dbContext *DbContext, tableName string, autoCreateTable bool) {
	table.dbContext = dbContext
	table.selectList = collections.NewListAny()
	table.whereList = collections.NewList[whereQuery]()
	table.orderList = collections.NewListAny()
	table.SetTableName(tableName)

	// 自动创建表
	if autoCreateTable {
		table.CreateTable()
	}
}

// 连接数据库
func (table *TableSet[Table]) open() *gorm.DB {
	if table.gormDB == nil {
		// Data Source ClientName，参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
		table.gormDB, table.err = gorm.Open(table.dbContext.getDriver(), &gorm.Config{
			SkipDefaultTransaction:                   true,
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止自动创建数据库外键约束
		})
		if table.err != nil {
			return table.gormDB
		}
		table.gormDB = table.gormDB.Table(table.tableName)
		table.setPool()

		// 设置Select
		if table.selectList.Any() {
			lst := table.selectList.Distinct()
			if lst.Count() > 1 {
				args := lst.RangeStart(1).ToArray()
				table.gormDB.Select(lst.First(), args...)
			} else {
				table.gormDB.Select(lst.First())
			}
		}

		// 设置Where
		if table.whereList.Any() {
			for _, query := range table.whereList.ToArray() {
				table.gormDB = table.gormDB.Where(query.query, query.args...)
			}
		}

		// 设置Order
		if table.orderList.Any() {
			for _, order := range table.orderList.ToArray() {
				table.gormDB.Order(order)
			}
		}

		// 设置limit
		if table.limit > 0 {
			table.gormDB.Limit(table.limit)
		}
	}
	return table.gormDB
}

// 关闭数据库
func (table *TableSet[Table]) close() {
	table.selectList.Clear()
	table.whereList.Clear()
	table.orderList.Clear()
	table.limit = 0

	if table.gormDB != nil {
		db, _ := table.gormDB.DB()
		_ = db.Close()
	}
	table.gormDB = nil
}

// SetTableName 设置表名
func (table *TableSet[Table]) SetTableName(tableName string) *TableSet[Table] {
	table.tableName = tableName
	if table.gormDB != nil {
		table.gormDB.Table(table.tableName)
	}
	return table
}

// GetTableName 获取表名称
func (table *TableSet[Table]) GetTableName() string {
	return table.tableName
}

// 设置池大小
func (table *TableSet[Table]) setPool() {
	sqlDB, _ := table.gormDB.DB()

	if table.dbContext.dbConfig.PoolMaxSize > 0 {
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(table.dbContext.dbConfig.PoolMaxSize / 3)
		// SetMaxOpenConns 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(table.dbContext.dbConfig.PoolMaxSize)
	}
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}

// CreateTable 创建表（如果不存在）
// 相关链接：https://gorm.cn/zh_CN/docs/migration.html
// 相关链接：https://gorm.cn/zh_CN/docs/indexes.html
func (table *TableSet[Table]) CreateTable() {
	table.open()
	defer table.close()
	var entity Table
	err := table.gormDB.AutoMigrate(&entity)
	if err != nil {
		_ = flog.Errorf("创建表：%s 时，出错：%s", table.tableName, err.Error())
	}
}

// Select 筛选字段
func (table *TableSet[Table]) Select(query any, args ...any) *TableSet[Table] {
	switch query.(type) {
	case []string:
		selects := query.([]string)
		for _, s := range selects {
			table.selectList.Add(s)
		}
	default:
		table.selectList.Add(query)
	}
	if len(args) > 0 {
		table.selectList.Add(args...)
	}
	return table
}

// Where 条件
func (table *TableSet[Table]) Where(query any, args ...any) *TableSet[Table] {
	table.whereList.Add(whereQuery{
		query: query,
		args:  args,
	})
	return table
}

// Order 排序
func (table *TableSet[Table]) Order(value any) *TableSet[Table] {
	table.orderList.Add(value)
	return table
}

// Desc 倒序
func (table *TableSet[Table]) Desc(fieldName string) *TableSet[Table] {
	table.orderList.Add(fieldName + " desc")
	return table
}

// Asc 正序
func (table *TableSet[Table]) Asc(fieldName string) *TableSet[Table] {
	table.orderList.Add(fieldName + " asc")
	return table
}

// Limit 限制记录数
func (table *TableSet[Table]) Limit(limit int) *TableSet[Table] {
	table.limit = limit
	return table
}

// ToList 返回结果集
func (table *TableSet[Table]) ToList() collections.List[Table] {
	table.open()
	defer table.close()

	var lst []Table
	table.gormDB.Find(&lst)
	return collections.NewList(lst...)
}

// ToArray 返回结果集
func (table *TableSet[Table]) ToArray() []Table {
	table.open()
	defer table.close()

	var lst []Table
	table.gormDB.Find(&lst)
	return lst
}

// ToPageList 返回分页结果集
func (table *TableSet[Table]) ToPageList(pageSize int, pageIndex int) collections.PageList[Table] {
	table.open()
	defer table.close()

	var count int64
	table.gormDB.Count(&count)

	offset := (pageIndex - 1) * pageSize
	var lst []Table
	table.gormDB.Offset(offset).Limit(pageSize).Find(&lst)

	return collections.NewPageList[Table](collections.NewList(lst...), count)
}

// ToEntity 返回单个对象
func (table *TableSet[Table]) ToEntity() Table {
	table.open()
	defer table.close()

	var entity Table
	table.gormDB.Limit(1).Find(&entity)
	return entity
}

// Count 返回表中的数量
func (table *TableSet[Table]) Count() int64 {
	table.open()
	defer table.close()

	var count int64
	table.gormDB.Count(&count)
	return count
}

// IsExists 是否存在记录
func (table *TableSet[Table]) IsExists() bool {
	table.open()
	defer table.close()

	var count int64
	table.gormDB.Count(&count)
	return count > 0
}

// Insert 新增记录
func (table *TableSet[Table]) Insert(po *Table) error {
	table.open()
	defer table.close()
	return table.gormDB.Create(po).Error
}

// Insert 新增记录
func (table *TableSet[Table]) InsertList(lst collections.List[Table], batchSize int) error {
	table.open()
	defer table.close()
	return table.gormDB.CreateInBatches(lst.ToArray(), batchSize).Error
}

// Update 修改记录
// 如果只更新部份字段，需使用Select进行筛选
func (table *TableSet[Table]) Update(po Table) int64 {
	table.open()
	defer table.close()

	result := table.gormDB.Save(po)
	return result.RowsAffected
}

// UpdateOrInsert 记录存在时更新，不存在时插入
func (table *TableSet[Table]) UpdateOrInsert(po Table, fields ...string) error {
	table.open()
	defer table.close()

	// []string转[]clause.Column
	var clos []clause.Column
	for _, field := range fields {
		clos = append(clos, clause.Column{Name: field})
	}
	return table.gormDB.Clauses(clause.OnConflict{
		Columns:   clos,
		UpdateAll: true,
	}).Create(po).Error
}

// UpdateValue 修改单个字段
func (table *TableSet[Table]) UpdateValue(column string, value any) {
	table.open()
	defer table.close()

	table.gormDB.Update(column, value)
}

// Delete 删除记录
func (table *TableSet[Table]) Delete() int64 {
	table.open()
	defer table.close()

	result := table.gormDB.Delete(nil)
	return result.RowsAffected
}

// GetString 获取单条记录中的单个string类型字段值
func (table *TableSet[Table]) GetString(fieldName string) string {
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer rows.Close()
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
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val int
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// GetLong 获取单条记录中的单个int64类型字段值
func (table *TableSet[Table]) GetLong(fieldName string) int64 {
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val int64
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// GetBool 获取单条记录中的单个bool类型字段值
func (table *TableSet[Table]) GetBool(fieldName string) bool {
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val bool
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// GetFloat32 获取单条记录中的单个float32类型字段值
func (table *TableSet[Table]) GetFloat32(fieldName string) float32 {
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float32
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// GetFloat64 获取单条记录中的单个float64类型字段值
func (table *TableSet[Table]) GetFloat64(fieldName string) float64 {
	table.open()
	defer table.close()

	rows, _ := table.gormDB.Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float64
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}
