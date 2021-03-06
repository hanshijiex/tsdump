package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/voidint/tsdump/config"
	"github.com/voidint/tsdump/model"
)

// Repo MySQL的model.IRepo接口实现
type Repo struct {
	engine *xorm.Engine
}

// NewRepo 实例化
func NewRepo(c *config.Config) (model.IRepo, error) {
	engine, err := xorm.NewEngine("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local",
			c.Username,
			c.Password,
			c.Host,
			c.Port,
			"information_schema",
		),
	)
	if err != nil {
		return nil, err
	}
	engine.ShowSQL(c.Debug)
	return &Repo{
		engine: engine,
	}, nil
}

type schema struct {
	Name      string `xorm:"'SCHEMA_NAME'"`
	CharSet   string `xorm:"'DEFAULT_CHARACTER_SET_NAME'"`
	Collation string `xorm:"'DEFAULT_COLLATION_NAME'"`
}

func (schema) TableName() string {
	return "SCHEMATA"
}

func (repo *Repo) getSchemas(cond *schema) (items []schema, err error) {
	if err = repo.engine.Find(&items, cond); err != nil {
		return nil, err
	}

	return items, nil
}

type table struct {
	Schema    string `xorm:"'TABLE_SCHEMA'"`
	Name      string `xorm:"'TABLE_NAME'"`
	Collation string `xorm:"'TABLE_COLLATION'"`
	Comment   string `xorm:"'TABLE_COMMENT'"`
}

func (table) TableName() string {
	return "TABLES"
}

func (repo *Repo) getTables(cond *table) (items []table, err error) {
	if err = repo.engine.Find(&items, cond); err != nil {
		return nil, err
	}
	return items, nil
}

type column struct {
	Schema    string `xorm:"'TABLE_SCHEMA'"`
	Table     string `xorm:"'TABLE_NAME'"`
	Name      string `xorm:"'COLUMN_NAME'"`
	Nullable  string `xorm:"'IS_NULLABLE'"`
	DataType  string `xorm:"'COLUMN_TYPE'"`
	CharSet   string `xorm:"'CHARACTER_SET_NAME'"`
	Collation string `xorm:"'COLLATION_NAME'"`
	Comment   string `xorm:"'COLUMN_COMMENT'"`
}

func (column) TableName() string {
	return "COLUMNS"
}

func (repo *Repo) getColumns(cond *column) (items []column, err error) {
	if err = repo.engine.Find(&items, cond); err != nil {
		return nil, err
	}
	return items, nil
}

// GetDBs 按条件查询数据库信息
func (repo *Repo) GetDBs(cond *model.DB) (items []model.DB, err error) {
	var sCond schema
	if cond != nil {
		sCond.Name = cond.Name
		sCond.CharSet = cond.CharSet
		sCond.Collation = cond.Collation
	}
	schemas, err := repo.getSchemas(&sCond)

	if err != nil {
		return nil, err
	}

	if len(schemas) <= 0 {
		return nil, model.ErrDBNotFound
	}

	for i := range schemas {
		tables, err := repo.GetTables(&model.Table{
			DB: schemas[i].Name,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, model.DB{
			Name:      schemas[i].Name,
			CharSet:   schemas[i].CharSet,
			Collation: schemas[i].Collation,
			Tables:    tables,
		})
	}

	return items, nil
}

// GetTables 按条件查询表信息
func (repo *Repo) GetTables(cond *model.Table) (items []model.Table, err error) {
	var tCond table
	if cond != nil {
		tCond.Schema = cond.DB
		tCond.Name = cond.Name
		tCond.Collation = cond.Collation
		tCond.Comment = cond.Comment
	}

	tables, err := repo.getTables(&tCond)
	if err != nil {
		return nil, err
	}

	for i := range tables {
		cols, err := repo.GetColumns(&model.Column{
			DB:    tables[i].Schema,
			Table: tables[i].Name,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, model.Table{
			DB:        tables[i].Schema,
			Name:      tables[i].Name,
			Collation: tables[i].Collation,
			Comment:   tables[i].Comment,
			Columns:   cols,
		})
	}
	return items, nil
}

// GetColumns 按条件查询列信息
func (repo *Repo) GetColumns(cond *model.Column) (items []model.Column, err error) {
	var cCond column
	if cond != nil {
		cCond.Schema = cond.DB
		cCond.Table = cond.Table
		cCond.Name = cond.Name
		cCond.Nullable = cond.Nullable
		cCond.CharSet = cond.CharSet
		cCond.Collation = cond.Collation
		cCond.DataType = cond.DataType
		cCond.Comment = cond.Comment
	}
	cols, err := repo.getColumns(&cCond)
	if err != nil {
		return nil, err
	}

	for i := range cols {
		items = append(items, model.Column{
			DB:        cols[i].Schema,
			Table:     cols[i].Table,
			Name:      cols[i].Name,
			Nullable:  cols[i].Nullable,
			DataType:  cols[i].DataType,
			CharSet:   cols[i].CharSet,
			Collation: cols[i].Collation,
			Comment:   cols[i].Comment,
		})
	}
	return items, nil
}
