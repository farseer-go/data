package data

import (
	"sync"
	"time"

	"github.com/farseer-go/fs/flog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// schemaMigrateTableName 系统表名称：记录每个上下文中各PO表的当前schema版本号
const schemaMigrateTableName = "data_schema_migrate"

// schemaMigratePO 系统表 data_schema_migrate 的映射，记录每张表最近一次迁移的版本号
type schemaMigratePO struct {
	KeyName   string    `gorm:"column:key_name;type:varchar(64);primaryKey"`    // 上下文配置名（config.yaml中的Database节点名），区分同库多上下文
	TblName   string    `gorm:"column:table_name;type:varchar(128);primaryKey"` // 表名
	Version   string    `gorm:"column:version;type:varchar(128)"`               // 标签声明的版本号（来自 data:"version=xxx"）
	PoType    string    `gorm:"column:po_type;type:varchar(128)"`               // PO结构体类型名，便于排查
	DbName    string    `gorm:"column:db_name;type:varchar(128)"`               // 库名
	DataType  string    `gorm:"column:data_type;type:varchar(32)"`              // 驱动类型（mysql/clickhouse等）
	Remark    string    `gorm:"column:remark;type:varchar(256)"`                // 备注，预留扩展，避免以后频繁改表结构
	MigrateAt time.Time `gorm:"column:migrate_at"`                              // 最近一次迁移时间（ClickHouse下同时作为ReplacingMergeTree的版本列）
}

// TableName 指定系统表名
func (schemaMigratePO) TableName() string {
	return schemaMigrateTableName
}

// 进程级版本缓存与加载状态，避免每张表都查一次系统表
// 以 keyName（上下文配置名）隔离，多租户独立库时天然区分
var (
	schemaMigrateCache  = make(map[string]map[string]string) // keyName -> (表名 -> 版本号)
	schemaMigrateLoaded = make(map[string]bool)              // keyName -> 是否已加载系统表
	schemaMigrateLock   sync.Mutex                           // 保护上述两个map的并发安全
)

// ensureSchemaMigrateLoaded 确保系统表存在，并把当前上下文的版本记录加载进内存
// 注意：调用方必须已持有 schemaMigrateLock；每个keyName只真正执行一次
func (receiver *internalContext) ensureSchemaMigrateLoaded() {
	key := receiver.dbConfig.keyName
	// 已加载过，直接返回
	if schemaMigrateLoaded[key] {
		return
	}

	db, err := receiver.Original()
	if err != nil {
		flog.Warningf("加载系统表%s失败，连接数据库出错：%s", schemaMigrateTableName, err.Error())
		return
	}

	// 1. 系统表不存在时先创建（解决首次上线自举问题）
	if !db.Migrator().HasTable(schemaMigrateTableName) {
		if receiver.dbConfig.DataType == "clickhouse" {
			// ClickHouse无真正的UPDATE，用ReplacingMergeTree按(key_name,table_name)去重，migrate_at作为版本列（取最新）
			ddl := "CREATE TABLE IF NOT EXISTS " + schemaMigrateTableName + " (" +
				"key_name String, table_name String, version String, po_type String, " +
				"db_name String, data_type String, remark String, migrate_at DateTime" +
				") ENGINE = ReplacingMergeTree(migrate_at) ORDER BY (key_name, table_name)"
			if err = db.Exec(ddl).Error; err != nil {
				flog.Warningf("创建系统表%s失败：%s", schemaMigrateTableName, err.Error())
				return
			}
		} else {
			// 其余驱动交给GORM按类型自动建表（复合主键 key_name+table_name）
			if err = db.AutoMigrate(&schemaMigratePO{}); err != nil {
				flog.Warningf("创建系统表%s失败：%s", schemaMigrateTableName, err.Error())
				return
			}
		}
	}

	// 2. 一次性把本上下文的版本记录读进内存，后续比对全部走内存
	var rows []schemaMigratePO
	query := db.Table(schemaMigrateTableName)
	if receiver.dbConfig.DataType == "clickhouse" {
		// ClickHouse需FINAL去重，确保读到最新版本
		query = query.Clauses(FinalHint{})
	}
	query.Where("key_name = ?", key).Find(&rows)

	// 整理成 表名->版本号 的map
	versions := make(map[string]string)
	for _, row := range rows {
		versions[row.TblName] = row.Version
	}
	schemaMigrateCache[key] = versions
	schemaMigrateLoaded[key] = true
}

// NeedSchemaMigrate 判断指定表是否需要执行自动建表/建索引
// version为空时强制迁移（保持旧行为）；否则与系统表记录的版本一致则跳过
func (receiver *internalContext) NeedSchemaMigrate(tableName string, version string) bool {
	// 空版本：强制迁移，完全兼容历史标签 data:"migrate"
	if version == "" {
		return true
	}

	schemaMigrateLock.Lock()
	defer schemaMigrateLock.Unlock()

	// 懒加载：第一次遇到带版本的表时，自举系统表并加载版本缓存
	receiver.ensureSchemaMigrateLoaded()

	// 已记录的版本与当前声明版本一致 -> 跳过迁移
	if versions, exists := schemaMigrateCache[receiver.dbConfig.keyName]; exists {
		if recorded, ok := versions[tableName]; ok && recorded == version {
			return false
		}
	}
	// 无记录或版本不一致 -> 需要迁移
	return true
}

// RecordSchemaMigrate 记录某张表本次迁移后的版本号，供下次启动比对
func (receiver *internalContext) RecordSchemaMigrate(tableName, version, poType string) {
	// 空版本无需记录（每次都强制迁移）
	if version == "" {
		return
	}

	db, err := receiver.Original()
	if err != nil {
		flog.Warningf("记录系统表%s失败，连接数据库出错：%s", schemaMigrateTableName, err.Error())
		return
	}

	// 组装待写入的版本记录
	po := schemaMigratePO{
		KeyName:   receiver.dbConfig.keyName,
		TblName:   tableName,
		Version:   version,
		PoType:    poType,
		DbName:    receiver.dbConfig.databaseName,
		DataType:  receiver.dbConfig.DataType,
		MigrateAt: time.Now(),
	}

	if receiver.dbConfig.DataType == "clickhouse" {
		// ClickHouse：直接追加一行，由ReplacingMergeTree+FINAL保证下次读到最新
		err = db.Table(schemaMigrateTableName).Transaction(func(tx *gorm.DB) error { // Transaction必须这么使用,否则数据库查不到数据
			result := tx.Create(&po) // 不能使用batchSize,会出现code: 101, message: Unexpected packet Query received from client
			if result.Error != nil {
				return result.Error
			}
			return nil
		})
		//err = db.Table(schemaMigrateTableName).Create(&po).Error
	} else {
		// 其余驱动：按主键(key_name,table_name)做upsert
		err = db.Table(schemaMigrateTableName).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key_name"}, {Name: "table_name"}},
			UpdateAll: true,
		}).Create(&po).Error
	}
	if err != nil {
		flog.Warningf("记录系统表%s失败，表：%s，错误：%s", schemaMigrateTableName, tableName, err.Error())
		return
	}

	// 同步更新内存缓存，使本次启动后续比对生效
	schemaMigrateLock.Lock()
	if _, exists := schemaMigrateCache[receiver.dbConfig.keyName]; !exists {
		schemaMigrateCache[receiver.dbConfig.keyName] = make(map[string]string)
	}
	schemaMigrateCache[receiver.dbConfig.keyName][tableName] = version
	schemaMigrateLock.Unlock()
}
