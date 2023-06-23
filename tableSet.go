package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"strings"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	// 上下文（用指针的方式，共享同一个上下文）
	dbContext *InternalContext
	// 表名
	tableName string
	// 最外层的ormClient一定是nil的
	ormClient *gorm.DB
	err       error
	// 字段筛选（官方再第二次设置时，会覆盖第一次的设置，因此需要暂存）
	selectList collections.ListAny
	whereList  collections.List[whereQuery]
	orderList  collections.ListAny
	limit      int
}

// where条件
type whereQuery struct {
	query any
	args  []any
}

// Init 在反射的时候会调用此方法
func (table *TableSet[Table]) Init(dbContext *InternalContext, tableName string, autoCreateTable bool) {
	table.dbContext = dbContext
	table.SetTableName(tableName)

	// 自动创建表
	if autoCreateTable {
		table.CreateTable()
	}

}

// 初始化一个Session
func (table *TableSet[Table]) getOrCreateSession() *TableSet[Table] {
	if table.ormClient == nil {
		var gormDB *gorm.DB
		var err error

		// 上下文没有开启事务
		if routineOrmClient.Get() == nil {
			gormDB, err = open(table.dbContext.dbConfig)
			if len(table.tableName) > 0 {
				gormDB = gormDB.Table(table.tableName)
			} else {
				gormDB = gormDB.Session(&gorm.Session{})
			}
		} else {
			gormDB = routineOrmClient.Get()
			if len(table.tableName) > 0 {
				gormDB = gormDB.Table(table.tableName)
			}
		}

		return &TableSet[Table]{
			dbContext:  table.dbContext,
			tableName:  table.tableName,
			ormClient:  gormDB,
			err:        err,
			selectList: collections.NewListAny(),
			whereList:  collections.NewList[whereQuery](),
			orderList:  collections.NewListAny(),
		}
	}
	return table
}

func (table *TableSet[Table]) getClient() *gorm.DB {
	// 设置Select
	if table.selectList.Any() {
		lst := table.selectList.Distinct()
		if lst.Count() > 1 {
			args := lst.RangeStart(1).ToArray()
			table.ormClient.Select(lst.First(), args...)
		} else {
			table.ormClient.Select(lst.First())
		}
	}

	// 设置Where
	if table.whereList.Any() {
		for _, query := range table.whereList.ToArray() {
			table.ormClient.Where(query.query, query.args...)
		}
	}

	// 设置Order
	if table.orderList.Any() {
		for _, order := range table.orderList.ToArray() {
			table.ormClient.Order(order)
		}
	}

	// 设置limit
	if table.limit > 0 {
		table.ormClient.Limit(table.limit)
	}

	return table.ormClient.Debug()
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
	var entity Table
	err := table.getOrCreateSession().ormClient.AutoMigrate(&entity)
	if err != nil {
		_ = flog.Errorf("创建表：%s 时，出错：%s", table.tableName, err.Error())
	}
}

// Select 筛选字段
func (table *TableSet[Table]) Select(query any, args ...any) *TableSet[Table] {
	session := table.getOrCreateSession()
	switch query.(type) {
	case []string:
		selects := query.([]string)
		for _, s := range selects {
			session.selectList.Add(s)
		}
	default:
		session.selectList.Add(query)
	}
	if len(args) > 0 {
		session.selectList.Add(args...)
	}
	return session
}

// Where 条件
func (table *TableSet[Table]) Where(query any, args ...any) *TableSet[Table] {
	session := table.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: query,
		args:  args,
	})
	return session
}

// WhereIgnoreLessZero 条件，自动忽略小于等于0的
func (table *TableSet[Table]) WhereIgnoreLessZero(query any, val int) *TableSet[Table] {
	session := table.getOrCreateSession()
	if val > 0 {
		session.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
	return session
}

// WhereIgnoreNil 条件，自动忽略nil条件
func (table *TableSet[Table]) WhereIgnoreNil(query any, val any) *TableSet[Table] {
	session := table.getOrCreateSession()
	if val != nil {
		session.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
	return session
}

// Order 排序
func (table *TableSet[Table]) Order(value any) *TableSet[Table] {
	session := table.getOrCreateSession()
	session.orderList.Add(value)
	return session
}

// Desc 倒序
func (table *TableSet[Table]) Desc(fieldName string) *TableSet[Table] {
	session := table.getOrCreateSession()
	session.orderList.Add(fieldName + " desc")
	return session
}

// Asc 正序
func (table *TableSet[Table]) Asc(fieldName string) *TableSet[Table] {
	session := table.getOrCreateSession()
	session.orderList.Add(fieldName + " asc")
	return session
}

// Limit 限制记录数
func (table *TableSet[Table]) Limit(limit int) *TableSet[Table] {
	session := table.getOrCreateSession()
	session.limit = limit
	return session
}

// ToList 返回结果集
func (table *TableSet[Table]) ToList() collections.List[Table] {
	var lst []Table
	table.getOrCreateSession().getClient().Find(&lst)
	return collections.NewList(lst...)
}

// ToArray 返回结果集
func (table *TableSet[Table]) ToArray() []Table {
	var lst []Table
	table.getOrCreateSession().getClient().Find(&lst)
	return lst
}

// ToPageList 返回分页结果集
func (table *TableSet[Table]) ToPageList(pageSize int, pageIndex int) collections.PageList[Table] {
	var count int64
	client := table.getOrCreateSession().getClient()
	client.Count(&count)

	offset := (pageIndex - 1) * pageSize
	var lst []Table
	client.Offset(offset).Limit(pageSize).Find(&lst)

	return collections.NewPageList[Table](collections.NewList(lst...), count)
}

// ToEntity 返回单个对象
func (table *TableSet[Table]) ToEntity() Table {
	var entity Table
	table.getOrCreateSession().getClient().Limit(1).Find(&entity)
	return entity
}

// Count 返回表中的数量
func (table *TableSet[Table]) Count() int64 {
	var count int64
	table.getOrCreateSession().getClient().Count(&count)
	return count
}

// IsExists 是否存在记录
func (table *TableSet[Table]) IsExists() bool {
	var count int64
	table.getOrCreateSession().getClient().Count(&count)
	return count > 0
}

// Insert 新增记录
func (table *TableSet[Table]) Insert(po *Table) error {
	return table.getOrCreateSession().getClient().Create(po).Error
}

// InsertList 批量新增记录
func (table *TableSet[Table]) InsertList(lst collections.List[Table], batchSize int) error {
	return table.getOrCreateSession().getClient().CreateInBatches(lst.ToArray(), batchSize).Error
}

// Update 修改记录
// 如果只更新部份字段，需使用Select进行筛选
func (table *TableSet[Table]) Update(po Table) int64 {
	result := table.getOrCreateSession().getClient().Save(po)
	return result.RowsAffected
}

// Expr 对字段做表达式操作
//
//	exp: AddUp("price", "price * ? + ?", 2, 100)
//	sql: UPDATE "xxx" SET "price" = price * 2 + 100
func (table *TableSet[Table]) Expr(field string, expr string, args ...any) int64 {
	result := table.getOrCreateSession().getClient().Update(field, gorm.Expr(expr, args...))
	return result.RowsAffected
}

// UpdateOrInsert 记录存在时更新，不存在时插入
func (table *TableSet[Table]) UpdateOrInsert(po Table, fields ...string) error {
	// []string转[]clause.Column
	var clos []clause.Column
	for _, field := range fields {
		clos = append(clos, clause.Column{Name: field})
	}
	return table.getOrCreateSession().getClient().Clauses(clause.OnConflict{
		Columns:   clos,
		UpdateAll: true,
	}).Create(po).Error
}

// UpdateValue 修改单个字段
func (table *TableSet[Table]) UpdateValue(column string, value any) {
	table.getOrCreateSession().getClient().Update(column, value)
}

// Delete 删除记录
func (table *TableSet[Table]) Delete() int64 {
	result := table.getOrCreateSession().getClient().Delete(nil)
	return result.RowsAffected
}

// GetString 获取单条记录中的单个string类型字段值
func (table *TableSet[Table]) GetString(fieldName string) string {
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := table.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
	defer func() {
		_ = rows.Close()
	}()
	var val float64
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// ExecuteSql 执行自定义SQL
func (table *TableSet[Table]) ExecuteSql(sql string, values ...any) {
	table.getOrCreateSession().getClient().Exec(sql, values...)
}

// ExecuteSqlToEntity 返回单个对象(执行自定义SQL)
func (table *TableSet[Table]) ExecuteSqlToEntity(sql string, values ...any) Table {
	var entity Table
	table.getOrCreateSession().getClient().Raw(sql, values...).Find(&entity)
	return entity
}

// ExecuteSqlToArray 返回结果集(执行自定义SQL)
func (table *TableSet[Table]) ExecuteSqlToArray(sql string, values ...any) []Table {
	var lst []Table
	table.getOrCreateSession().getClient().Raw(sql, values...).Find(&lst)
	return lst
}

// ExecuteSqlToList 返回结果集(执行自定义SQL)
func (table *TableSet[Table]) ExecuteSqlToList(sql string, values ...any) collections.List[Table] {
	var lst []Table
	table.getOrCreateSession().getClient().Raw(sql, values...).Find(&lst)
	return collections.NewList(lst...)
}

// Original 返回原生的对象
func (table *TableSet[Table]) Original() *gorm.DB {
	return table.getOrCreateSession().getClient()
}

// GetPrimaryName 获取主键
func (table *TableSet[Table]) GetPrimaryName() string {
	var tableIns Table
	tableType := reflect.TypeOf(tableIns)

	for i := 0; i < tableType.NumField(); i++ {
		field := tableType.Field(i)
		tag := field.Tag.Get("gorm")
		// 找到主键ID（目前只支持单个主键）
		if strings.Contains(tag, "primaryKey") {
			return field.Name
		}
	}
	return ""
}
