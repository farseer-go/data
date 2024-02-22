package data

import (
	"fmt"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/container"
	"github.com/farseer-go/fs/parse"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	dbContext    *internalContext  // 上下文（用指针的方式，共享同一个上下文）
	dbName       string            // 库名
	tableName    string            // 表名
	nameReplacer *strings.Replacer // 替换dbName、tableName
	ormClient    *gorm.DB          // 最外层的ormClient一定是nil的
	layer        int               // 链式第几层
	// 字段筛选（官方再第二次设置时，会覆盖第一次的设置，因此需要暂存）
	selectList  collections.ListAny
	omitList    collections.List[string]
	whereList   collections.List[whereQuery]
	orderList   collections.ListAny
	limit       int
	err         error
	primaryName string
}

// where条件
type whereQuery struct {
	query any
	args  []any
}

// Init 在反射的时候会调用此方法
func (receiver *TableSet[Table]) Init(dbContext *internalContext, param map[string]string) {
	receiver.dbContext = dbContext
	receiver.GetPrimaryName()
	db := receiver.getOrCreateSession().ormClient
	// 表名
	if name, exists := param["name"]; exists {
		receiver.SetTableName(name)
	} else {
		receiver.SetTableName(db.Statement.Table)
	}
	receiver.dbName = db.Migrator().CurrentDatabase()
	receiver.nameReplacer = strings.NewReplacer("{database}", receiver.dbName, "{table}", receiver.tableName)
	// 自动创建表
	migrate, exists := param["migrate"]
	if exists {

		// 创建表
		receiver.CreateTable(db, migrate)
		// 创建索引
		receiver.CreateIndex(db)
	}
}

// CreateTable 创建表（如果不存在）
// 相关链接：https://gorm.io/zh_CN/docs/migration.html
// 相关链接：https://gorm.io/zh_CN/docs/indexes.html
func (receiver *TableSet[Table]) CreateTable(db *gorm.DB, engine string) {
	var entity Table
	if engine != "" {
		db = db.Set("gorm:table_options", "ENGINE="+engine)
	}
	// 如果继承了IMigrator，则使用自定义的SQL来创建表
	if mig, exists := any(&entity).(IMigratorCreate); exists {
		if !db.Migrator().HasTable(receiver.tableName) {
			SqlScript := receiver.nameReplacer.Replace(mig.CreateTable())
			receiver.err = db.Exec(SqlScript).Error
		}
	} else {
		receiver.err = db.AutoMigrate(&entity)
	}
	if receiver.err != nil {
		panic(fmt.Sprintf("创建或修改表：%s 时，出错：%s", receiver.tableName, receiver.err.Error()))
	}
}

func (receiver *TableSet[Table]) CreateIndex(db *gorm.DB) {
	var entity Table

	// 创建索引
	if mig, exists := any(&entity).(IMigratorIndex); exists {
		if !container.IsRegister[IDataDriver](receiver.dbContext.dbConfig.DataType) {
			panic(fmt.Sprintf("要使用%s，请加载模块：对应的驱动，通常位置在：github.com/farseer-go/data/driver/%s", receiver.dbContext.dbConfig.DataType, receiver.dbContext.dbConfig.DataType))
		}

		// 得到要创建的索引字段
		idx := mig.CreateIndex()
		for idxName, idxFields := range idx {
			// 索引已存在时，不创建
			if db.Migrator().HasIndex(db.Statement.Table, idxName) {
				continue
			}

			// 得到创建索引的SQL脚本
			sqlScript := container.Resolve[IDataDriver](receiver.dbContext.dbConfig.DataType).CreateIndex(db.Statement.Table, idxName, idxFields)
			// 执行
			if receiver.err = db.Exec(sqlScript).Error; receiver.err != nil {
				panic(fmt.Sprintf("创建索引，表：%s，索引名称：%s 时，出错：%s", receiver.tableName, idxName, receiver.err.Error()))
			}
		}
	}
}

// 初始化一个Session
func (receiver *TableSet[Table]) getOrCreateSession() *TableSet[Table] {
	if receiver.layer == 0 {
		// 先从上下文中读取事务
		gormDB := routineOrmClient[receiver.dbContext.dbConfig.dbName].Get()

		// 上下文没有开启事务
		if gormDB == nil {
			gormDB, receiver.err = open(receiver.dbContext.dbConfig)
			if len(receiver.tableName) > 0 {
				gormDB = gormDB.Table(receiver.tableName)
			} else {
				gormDB = gormDB.Session(&gorm.Session{
					SkipDefaultTransaction: gormDB.SkipDefaultTransaction,
					Logger:                 gormDB.Logger,
				})
			}
		} else {
			if len(receiver.tableName) > 0 {
				gormDB = gormDB.Table(receiver.tableName)
			}
		}

		gormDB.InstanceSet("ConnectionString", receiver.dbContext.dbConfig.ConnectionString)
		gormDB.InstanceSet("DbName", receiver.dbContext.dbConfig.databaseName)
		return &TableSet[Table]{
			dbContext:   receiver.dbContext,
			tableName:   receiver.tableName,
			ormClient:   gormDB,
			err:         receiver.err,
			layer:       1,
			selectList:  collections.NewListAny(),
			omitList:    collections.NewList[string](),
			whereList:   collections.NewList[whereQuery](),
			orderList:   collections.NewListAny(),
			primaryName: receiver.primaryName,
		}
	}

	receiver.ormClient.InstanceSet("ConnectionString", receiver.dbContext.dbConfig.ConnectionString)
	receiver.ormClient.InstanceSet("DbName", receiver.dbContext.dbConfig.databaseName)
	return receiver
}

func (receiver *TableSet[Table]) getClient() *gorm.DB {
	// 设置Select
	if receiver.selectList.Any() {
		lst := receiver.selectList.Distinct()
		if lst.Count() > 1 {
			args := lst.RangeStart(1).ToArray()
			receiver.ormClient.Select(lst.First(), args...)
		} else {
			receiver.ormClient.Select(lst.First())
		}
	}
	// 设置Omit
	if receiver.omitList.Any() {
		lst := receiver.omitList.Distinct()
		receiver.ormClient.Omit(lst.ToArray()...)
	}

	// 设置Where
	if receiver.whereList.Any() {
		for _, query := range receiver.whereList.ToArray() {
			receiver.ormClient.Where(query.query, query.args...)
		}
	}

	// 设置Order
	if receiver.orderList.Any() {
		for _, order := range receiver.orderList.ToArray() {
			receiver.ormClient.Order(order)
		}
	}

	// 设置limit
	if receiver.limit > 0 {
		receiver.ormClient.Limit(receiver.limit)
	}

	return receiver.ormClient
}

// SetTableName 设置表名
func (receiver *TableSet[Table]) SetTableName(tableName string) *TableSet[Table] {
	receiver.tableName = tableName
	return receiver
}

// GetTableName 获取表名称
func (receiver *TableSet[Table]) GetTableName() string {
	return receiver.tableName
}

// Select 筛选字段
func (receiver *TableSet[Table]) Select(query any, args ...any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
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

// Omit 忽略字段
func (receiver *TableSet[Table]) Omit(columns ...string) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	for _, s := range columns {
		session.omitList.Add(s)
	}
	return session
}

// Where 条件
func (receiver *TableSet[Table]) Where(query any, args ...any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: query,
		args:  args,
	})
	return session
}

// WhereEq 条件
func (receiver *TableSet[Table]) WhereEq(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v = ?", columnName),
		args:  []any{args},
	})
	return session
}

// WhereGt 大于条件
func (receiver *TableSet[Table]) WhereGt(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v > ?", columnName),
		args:  []any{args},
	})
	return session
}

// WhereGte 大于等于条件
func (receiver *TableSet[Table]) WhereGte(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v >= ?", columnName),
		args:  []any{args},
	})
	return session
}

// WhereLt 小于条件
func (receiver *TableSet[Table]) WhereLt(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v < ?", columnName),
		args:  []any{args},
	})
	return session
} // WhereLte 小于等于条件
func (receiver *TableSet[Table]) WhereLte(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v <= ?", columnName),
		args:  []any{args},
	})
	return session
} // WhereIn in条件
func (receiver *TableSet[Table]) WhereIn(columnName any, args ...any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v in ?", columnName),
		args:  args,
	})
	return session
} // WhereLike like条件("%?%")
func (receiver *TableSet[Table]) WhereLike(columnName any, args any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v like ?", columnName),
		args:  []any{fmt.Sprintf("%%%v%%", args)},
	})
	return session
} // WhereEq between条件(>= and <=)
func (receiver *TableSet[Table]) WhereBetween(columnName any, min, max any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: fmt.Sprintf("%v >= ? and %v <= ?", columnName, columnName),
		args:  []any{min, max},
	})
	return session
}

// WhereIf 当conditional==true时，使用条件
func (receiver *TableSet[Table]) WhereIf(conditional bool, query any, args ...any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.Where(query, args...)
}

// WhereEqIf 当conditional==true时，使用等于条件
func (receiver *TableSet[Table]) WhereEqIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereEq(columnName, args)
}

// WhereGtIf 当conditional==true时，使用大于条件
func (receiver *TableSet[Table]) WhereGtIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereGt(columnName, args)
}

// WhereGteIf 当conditional==true时，使用大于等于条件
func (receiver *TableSet[Table]) WhereGteIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereGte(columnName, args)
}

// WhereLtIf 当conditional==true时，使用小于条件
func (receiver *TableSet[Table]) WhereLtIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereLt(columnName, args)
}

// WhereLteIf 当conditional==true时，使用小于等于条件
func (receiver *TableSet[Table]) WhereLteIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereLte(columnName, args)
}

// WhereInIf 当conditional==true时，使用in条件
func (receiver *TableSet[Table]) WhereInIf(conditional bool, columnName any, args ...any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereIn(columnName, args)
}

// WhereLikeIf 当conditional==true时，使用like条件("%?%"匹配)
func (receiver *TableSet[Table]) WhereLikeIf(conditional bool, columnName any, args any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereLike(columnName, args)
}

// WhereBetweenIf 当conditional==true时，使用between条件(>=and<=)
func (receiver *TableSet[Table]) WhereBetweenIf(conditional bool, columnName any, min, max any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.WhereBetween(columnName, min, max)
}

// WhereIgnoreLessZero 条件，自动忽略小于等于0的
func (receiver *TableSet[Table]) WhereIgnoreLessZero(query any, val int) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	if val > 0 {
		session.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
	return session
}

// WhereIgnoreNil 条件，自动忽略nil条件
func (receiver *TableSet[Table]) WhereIgnoreNil(query any, val any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	if val != nil {
		session.whereList.Add(whereQuery{
			query: query,
			args:  []any{val},
		})
	}
	return session
}

// Order 排序
func (receiver *TableSet[Table]) Order(value any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.orderList.Add(value)
	return session
}

// OrderIf 排序，当conditional==true时，使用排序
func (receiver *TableSet[Table]) OrderIf(conditional bool, value any) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.Order(value)
}

// Desc 倒序
func (receiver *TableSet[Table]) Desc(fieldName string) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.orderList.Add(fieldName + " desc")
	return session
}

// DescIf 倒序，当conditional==true时，使用倒序
func (receiver *TableSet[Table]) DescIf(conditional bool, fieldName string) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.Desc(fieldName)
}

// DescIfElse 倒序，当conditional==true时，使用trueFieldName倒序，否则使用falseFieldName倒序
func (receiver *TableSet[Table]) DescIfElse(conditional bool, trueFieldName string, falseFieldName string) *TableSet[Table] {
	if conditional {
		return receiver.Desc(trueFieldName)
	}
	return receiver.Desc(falseFieldName)
}

// Asc 正序
func (receiver *TableSet[Table]) Asc(fieldName string) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.orderList.Add(fieldName + " asc")
	return session
}

// AscIf 正序，当conditional==true时，使用正序
func (receiver *TableSet[Table]) AscIf(conditional bool, fieldName string) *TableSet[Table] {
	if !conditional {
		return receiver
	}
	return receiver.Asc(fieldName)
}

// AscIfElse 正序，当conditional==true时，使用trueFieldName正序，否则使用falseFieldName正序
func (receiver *TableSet[Table]) AscIfElse(conditional bool, trueFieldName string, falseFieldName string) *TableSet[Table] {
	if conditional {
		return receiver.Asc(trueFieldName)
	}
	return receiver.Asc(falseFieldName)
}

// Limit 限制记录数
func (receiver *TableSet[Table]) Limit(limit int) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.limit = limit
	return session
}

// ToList 返回结果集
func (receiver *TableSet[Table]) ToList() collections.List[Table] {
	var lst []Table
	receiver.getOrCreateSession().getClient().Find(&lst)
	return collections.NewList(lst...)
}

// ToArray 返回结果集
func (receiver *TableSet[Table]) ToArray() []Table {
	var lst []Table
	receiver.getOrCreateSession().getClient().Find(&lst)
	return lst
}

// ToPageList 返回分页结果集
func (receiver *TableSet[Table]) ToPageList(pageSize int, pageIndex int) collections.PageList[Table] {
	var count int64
	client := receiver.getOrCreateSession().getClient()
	client.Count(&count)

	offset := (pageIndex - 1) * pageSize
	var lst []Table
	client.Offset(offset).Limit(pageSize).Find(&lst)
	return collections.NewPageList[Table](collections.NewList(lst...), count)
}

// ToEntity 返回单个对象
func (receiver *TableSet[Table]) ToEntity() Table {
	var entity Table
	receiver.getOrCreateSession().getClient().Limit(1).Find(&entity)
	return entity
}

// Count 返回表中的数量
func (receiver *TableSet[Table]) Count() int64 {
	var count int64
	receiver.getOrCreateSession().getClient().Count(&count)
	return count
}

// IsExists 是否存在记录
func (receiver *TableSet[Table]) IsExists() bool {
	var count int64
	receiver.getOrCreateSession().getClient().Count(&count)
	return count > 0
}

// Insert 新增记录
func (receiver *TableSet[Table]) Insert(po *Table) error {
	result := receiver.getOrCreateSession().getClient().Create(po)
	return result.Error
}

// InsertList 批量新增记录
func (receiver *TableSet[Table]) InsertList(lst collections.List[Table], batchSize int) (int64, error) {
	// 在clickhouse数据库中，gorm官方包会出现异常：当batchSize小于lst.Count时。会收到：code: 101, message: Unexpected packet Query received from client的错误
	if receiver.dbContext.dbConfig.DataType == "clickhouse" {
		batchSize = lst.Count()
	}
	result := receiver.getOrCreateSession().getClient().CreateInBatches(lst.ToArray(), batchSize)
	return result.RowsAffected, result.Error
}

// Update 修改记录
// 如果只更新部份字段，需使用Select进行筛选
func (receiver *TableSet[Table]) Update(po Table) (int64, error) {
	mapPO := ToMap(po)
	//result := receiver.getOrCreateSession().getClient().Save(po)
	result := receiver.getOrCreateSession().getClient().Updates(mapPO)
	return result.RowsAffected, result.Error
}

// Expr 对字段做表达式操作
//
//	exp: Expr("price", "price * ? + ?", 2, 100)
//	sql: UPDATE "xxx" SET price = price * 2 + 100
func (receiver *TableSet[Table]) Expr(field string, expr string, args ...any) (int64, error) {
	result := receiver.getOrCreateSession().getClient().UpdateColumn(field, gorm.Expr(expr, args...))
	return result.RowsAffected, result.Error
}

// Exprs 对多个字段做表达式操作
//
//	exp: Exprs(map[string][]any{"price": {"price - ?", 10}, "count": {"count + ?", 5}})
//	sql: UPDATE "xxx" SET price = price - 10, count = count + 5
func (receiver *TableSet[Table]) Exprs(fields map[string][]any) (int64, error) {
	var args []any
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("UPDATE %s SET ", receiver.tableName))

	// SET
	var setSql []string
	for k, v := range fields {
		setSql = append(setSql, fmt.Sprintf("%s = ?", k))
		args = append(args, gorm.Expr(parse.ToString(v[0]), v[1:]...))
	}
	builder.WriteString(strings.Join(setSql, ", "))

	// WHERE
	if receiver.whereList.Any() {
		var whereSql []string
		builder.WriteString(" WHERE ")
		for _, query := range receiver.whereList.ToArray() {
			whereSql = append(whereSql, query.query.(string))
			args = append(args, query.args...)
		}
		builder.WriteString(strings.Join(whereSql, " AND "))
	}
	rowsAffected, err := receiver.ExecuteSql(builder.String(), args...)
	return rowsAffected, err
}

// UpdateOrInsert 记录存在时更新，不存在时插入
func (receiver *TableSet[Table]) UpdateOrInsert(po Table, fields ...string) error {
	// []string转[]clause.Column
	var clos []clause.Column
	for _, field := range fields {
		clos = append(clos, clause.Column{Name: field})
	}
	result := receiver.getOrCreateSession().getClient().Clauses(clause.OnConflict{
		Columns:   clos,
		UpdateAll: true,
	}).Create(&po)
	return result.Error
}

// UpdateValue 修改单个字段
func (receiver *TableSet[Table]) UpdateValue(column string, value any) (int64, error) {
	result := receiver.getOrCreateSession().getClient().UpdateColumn(column, value)
	return result.RowsAffected, result.Error
}

// Delete 删除记录
func (receiver *TableSet[Table]) Delete() (int64, error) {
	result := receiver.getOrCreateSession().getClient().Delete(nil)
	return result.RowsAffected, result.Error
}

// GetString 获取单条记录中的单个string类型字段值
func (receiver *TableSet[Table]) GetString(fieldName string) string {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return ""
	}
	defer rows.Close()
	var val string
	for rows.Next() {
		_ = rows.Scan(&val)
		// ScanRows 方法用于将一行记录扫描至结构体
		//receiver.ScanRows(rows, &user)
	}
	return val
}

// GetInt 获取单条记录中的单个int类型字段值
func (receiver *TableSet[Table]) GetInt(fieldName string) int {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return 0
	}
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
func (receiver *TableSet[Table]) GetLong(fieldName string) int64 {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return int64(0)
	}
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
func (receiver *TableSet[Table]) GetBool(fieldName string) bool {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return false
	}
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
func (receiver *TableSet[Table]) GetFloat32(fieldName string) float32 {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return float32(0)
	}
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
func (receiver *TableSet[Table]) GetFloat64(fieldName string) float64 {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return float64(0)
	}
	defer func() {
		_ = rows.Close()
	}()
	var val float64
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// GetDecimal 获取单条记录中的单个decimal.Decimal类型字段值
func (receiver *TableSet[Table]) GetDecimal(fieldName string) decimal.Decimal {
	result := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1)
	rows, _ := result.Rows()
	if rows == nil {
		return decimal.NewFromInt(0)
	}
	defer func() {
		_ = rows.Close()
	}()
	var val decimal.Decimal
	for rows.Next() {
		_ = rows.Scan(&val)
	}
	return val
}

// ExecuteSql 执行自定义SQL
func (receiver *TableSet[Table]) ExecuteSql(sql string, values ...any) (int64, error) {
	sql = receiver.nameReplacer.Replace(sql)
	result := receiver.getOrCreateSession().getClient().Exec(sql, values...)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToEntity 返回单个对象(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToEntity(sql string, values ...any) Table {
	var entity Table
	sql = receiver.nameReplacer.Replace(sql)
	receiver.getOrCreateSession().getClient().Raw(sql, values...).Find(&entity)
	return entity
}

// ExecuteSqlToArray 返回结果集(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToArray(sql string, values ...any) []Table {
	var lst []Table
	sql = receiver.nameReplacer.Replace(sql)
	receiver.getOrCreateSession().getClient().Raw(sql, values...).Find(&lst)
	return lst
}

// ExecuteSqlToList 返回结果集(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToList(sql string, values ...any) collections.List[Table] {
	var lst []Table
	sql = receiver.nameReplacer.Replace(sql)
	receiver.getOrCreateSession().getClient().Raw(sql, values...).Find(&lst)
	return collections.NewList(lst...)
}

// Original 返回原生的对象
func (receiver *TableSet[Table]) Original() *gorm.DB {
	return receiver.getOrCreateSession().getClient()
}

// GetPrimaryName 获取主键
func (receiver *TableSet[Table]) GetPrimaryName() {
	var tableIns Table
	tableType := reflect.TypeOf(tableIns)

	for i := 0; i < tableType.NumField(); i++ {
		field := tableType.Field(i)
		fieldTags := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		if _, existsPrimaryKey := fieldTags["PRIMARYKEY"]; existsPrimaryKey {
			if c, existsColumn := fieldTags["COLUMN"]; existsColumn {
				receiver.primaryName = c
				return
			}
			receiver.primaryName = schema.NamingStrategy{IdentifierMaxLength: 64}.ColumnName("", field.Name)
			return
		}
	}
	return
}

// 大写字母，转蛇形
func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90(A-Z)，97-122(a-z)
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	// 统一转小写
	return strings.ToLower(string(data[:]))
}

// Clickhouse 返回Clickhouse的对象
func (receiver *TableSet[Table]) Clickhouse() *mergeTreeSet {
	return newClickhouse(receiver.getOrCreateSession())
}
