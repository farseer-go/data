package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	// 上下文（用指针的方式，共享同一个上下文）
	dbContext *DbContext
	// 表名
	tableName string
	//gormDB    *gorm.DB
	err error
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
func (table *TableSet[Table]) session() *gorm.DB {
	gormDB, err := open(table.dbContext.dbConfig)
	if err != nil {
		return gormDB
	}
	if len(table.tableName) > 0 {
		gormDB = gormDB.Table(table.tableName)
	} else {
		gormDB = gormDB.Session(&gorm.Session{})
	}

	// 设置Select
	if table.selectList.Any() {
		lst := table.selectList.Distinct()
		if lst.Count() > 1 {
			args := lst.RangeStart(1).ToArray()
			gormDB.Select(lst.First(), args...)
		} else {
			gormDB.Select(lst.First())
		}
	}

	// 设置Where
	if table.whereList.Any() {
		for _, query := range table.whereList.ToArray() {
			gormDB.Where(query.query, query.args...)
		}
	}

	// 设置Order
	if table.orderList.Any() {
		for _, order := range table.orderList.ToArray() {
			gormDB.Order(order)
		}
	}

	// 设置limit
	if table.limit > 0 {
		gormDB.Limit(table.limit)
	}
	return gormDB
}

// 关闭数据库
func (table *TableSet[Table]) clear() {
	table.selectList.Clear()
	table.whereList.Clear()
	table.orderList.Clear()
	table.limit = 0
}

// SetTableName 设置表名
func (table *TableSet[Table]) SetTableName(tableName string) *TableSet[Table] {
	table.tableName = tableName
	return table
}

// GetTableName 获取表名称
func (table *TableSet[Table]) GetTableName() string {
	return table.tableName
}

// CreateTable 创建表（如果不存在）
// 相关链接：https://gorm.cn/zh_CN/docs/migration.html
// 相关链接：https://gorm.cn/zh_CN/docs/indexes.html
func (table *TableSet[Table]) CreateTable() {
	defer table.clear()
	var entity Table
	err := table.session().AutoMigrate(&entity)
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

// WhereIgnoreLessZero 条件，自动忽略小于等于0的
func (table *TableSet[Table]) WhereIgnoreLessZero(query any, val int) *TableSet[Table] {
	if val > 0 {
		table.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
	return table
}

// WhereIgnoreNil 条件，自动忽略nil条件
func (table *TableSet[Table]) WhereIgnoreNil(query any, val any) *TableSet[Table] {
	if val != nil {
		table.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
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
	defer table.clear()

	var lst []Table
	table.session().Find(&lst)
	return collections.NewList(lst...)
}

// ToListBySql 返回结果集
func (table *TableSet[Table]) ToListBySql() collections.List[Table] {
	defer table.clear()

	var lst []Table
	table.session().Find(&lst)
	return collections.NewList(lst...)
}

// ToArray 返回结果集
func (table *TableSet[Table]) ToArray() []Table {
	defer table.clear()

	var lst []Table
	table.session().Find(&lst)
	return lst
}

// ToPageList 返回分页结果集
func (table *TableSet[Table]) ToPageList(pageSize int, pageIndex int) collections.PageList[Table] {
	defer table.clear()

	var count int64
	table.session().Count(&count)

	offset := (pageIndex - 1) * pageSize
	var lst []Table
	table.session().Offset(offset).Limit(pageSize).Find(&lst)

	return collections.NewPageList[Table](collections.NewList(lst...), count)
}

// ToEntity 返回单个对象
func (table *TableSet[Table]) ToEntity() Table {
	defer table.clear()

	var entity Table
	table.session().Limit(1).Find(&entity)
	return entity
}

// Count 返回表中的数量
func (table *TableSet[Table]) Count() int64 {
	defer table.clear()

	var count int64
	table.session().Count(&count)
	return count
}

// IsExists 是否存在记录
func (table *TableSet[Table]) IsExists() bool {
	defer table.clear()

	var count int64
	table.session().Count(&count)
	return count > 0
}

// Insert 新增记录
func (table *TableSet[Table]) Insert(po *Table) error {
	defer table.clear()
	return table.session().Create(po).Error
}

// InsertList 批量新增记录
func (table *TableSet[Table]) InsertList(lst collections.List[Table], batchSize int) error {
	defer table.clear()
	return table.session().CreateInBatches(lst.ToArray(), batchSize).Error
}

// Update 修改记录
// 如果只更新部份字段，需使用Select进行筛选
func (table *TableSet[Table]) Update(po Table) int64 {
	defer table.clear()

	result := table.session().Save(po)
	return result.RowsAffected
}

// UpdateOrInsert 记录存在时更新，不存在时插入
func (table *TableSet[Table]) UpdateOrInsert(po Table, fields ...string) error {
	defer table.clear()

	// []string转[]clause.Column
	var clos []clause.Column
	for _, field := range fields {
		clos = append(clos, clause.Column{Name: field})
	}
	return table.session().Clauses(clause.OnConflict{
		Columns:   clos,
		UpdateAll: true,
	}).Create(po).Error
}

// UpdateValue 修改单个字段
func (table *TableSet[Table]) UpdateValue(column string, value any) {
	defer table.clear()

	table.session().Update(column, value)
}

// Delete 删除记录
func (table *TableSet[Table]) Delete() int64 {
	defer table.clear()

	result := table.session().Delete(nil)
	return result.RowsAffected
}

// GetString 获取单条记录中的单个string类型字段值
func (table *TableSet[Table]) GetString(fieldName string) string {
	defer table.clear()

	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
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
	defer table.clear()

	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
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
	defer table.clear()

	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
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
	defer table.clear()

	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
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
	defer table.clear()

	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
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
	defer table.clear()
	rows, _ := table.session().Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float64
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}
