package data

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

// TableSet 数据库表操作
type TableSet[Table any] struct {
	dbContext *internalContext // 上下文（用指针的方式，共享同一个上下文）
	tableName string           // 表名
	ormClient *gorm.DB         // 最外层的ormClient一定是nil的
	layer     int              // 链式第几层
	// 字段筛选（官方再第二次设置时，会覆盖第一次的设置，因此需要暂存）
	selectList  collections.ListAny
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
	// 表名
	name, exists := param["name"]
	if exists {
		receiver.SetTableName(name)
	}

	// 自动创建表
	migrate, exists := param["migrate"]
	if exists {
		receiver.CreateTable(migrate)
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
		return &TableSet[Table]{
			dbContext:   receiver.dbContext,
			tableName:   receiver.tableName,
			ormClient:   gormDB,
			err:         receiver.err,
			layer:       1,
			selectList:  collections.NewListAny(),
			whereList:   collections.NewList[whereQuery](),
			orderList:   collections.NewListAny(),
			primaryName: receiver.primaryName,
		}
	}
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

// CreateTable 创建表（如果不存在）
// 相关链接：https://gorm.cn/zh_CN/docs/migration.html
// 相关链接：https://gorm.cn/zh_CN/docs/indexes.html
func (receiver *TableSet[Table]) CreateTable(engine string) {
	var entity Table
	db := receiver.getOrCreateSession().ormClient
	if engine != "" {
		db = db.Set("gorm:table_options", "ENGINE="+engine)
	}
	err := db.AutoMigrate(&entity)
	if err != nil {
		_ = flog.Errorf("创建表：%s 时，出错：%s", receiver.tableName, err.Error())
	}
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

// Where 条件
func (receiver *TableSet[Table]) Where(query any, args ...any) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.whereList.Add(whereQuery{
		query: query,
		args:  args,
	})
	return session
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

// Desc 倒序
func (receiver *TableSet[Table]) Desc(fieldName string) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.orderList.Add(fieldName + " desc")
	return session
}

// Asc 正序
func (receiver *TableSet[Table]) Asc(fieldName string) *TableSet[Table] {
	session := receiver.getOrCreateSession()
	session.orderList.Add(fieldName + " asc")
	return session
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
	return receiver.getOrCreateSession().getClient().Create(po).Error
}

// InsertList 批量新增记录
func (receiver *TableSet[Table]) InsertList(lst collections.List[Table], batchSize int) error {
	return receiver.getOrCreateSession().getClient().CreateInBatches(lst.ToArray(), batchSize).Error
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
//	exp: AddUp("price", "price * ? + ?", 2, 100)
//	sql: UPDATE "xxx" SET "price" = price * 2 + 100
func (receiver *TableSet[Table]) Expr(field string, expr string, args ...any) (int64, error) {
	result := receiver.getOrCreateSession().getClient().UpdateColumn(field, gorm.Expr(expr, args...))
	return result.RowsAffected, result.Error
}

// UpdateOrInsert 记录存在时更新，不存在时插入
func (receiver *TableSet[Table]) UpdateOrInsert(po Table, fields ...string) error {
	// []string转[]clause.Column
	var clos []clause.Column
	for _, field := range fields {
		clos = append(clos, clause.Column{Name: field})
	}
	return receiver.getOrCreateSession().getClient().Clauses(clause.OnConflict{
		Columns:   clos,
		UpdateAll: true,
	}).Create(&po).Error
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
	rows, _ := receiver.getOrCreateSession().getClient().Select(fieldName).Limit(1).Rows()
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
func (receiver *TableSet[Table]) ExecuteSql(sql string, values ...any) (int64, error) {
	result := receiver.getOrCreateSession().getClient().Exec(sql, values...)
	return result.RowsAffected, result.Error
}

// ExecuteSqlToEntity 返回单个对象(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToEntity(sql string, values ...any) Table {
	var entity Table
	receiver.getOrCreateSession().getClient().Raw(sql, values...).Find(&entity)
	return entity
}

// ExecuteSqlToArray 返回结果集(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToArray(sql string, values ...any) []Table {
	var lst []Table
	receiver.getOrCreateSession().getClient().Raw(sql, values...).Find(&lst)
	return lst
}

// ExecuteSqlToList 返回结果集(执行自定义SQL)
func (receiver *TableSet[Table]) ExecuteSqlToList(sql string, values ...any) collections.List[Table] {
	var lst []Table
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
