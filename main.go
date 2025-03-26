package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"

	"github.com/kenshaw/snaker"
	"github.com/yangsai7/go-mysql"
	"github.com/yangsai7/go-sqlbuilder"
)

const (
	shadowPrefix = "fct_" // 部分老项目
	shadowSuffix = "_shadow"

	Version = "2023-11-14.v3"
)

var (
	tablesMap      map[string]struct{}
	mysqlFactory   *mysql.Factory
	initialisms, _ = snaker.NewDefaultInitialisms()
)

func init() {
	tablesMap = make(map[string]struct{})
	for i := rune('a'); i <= rune('z'); i++ {
		initialisms.Add(strings.ToUpper(string(i) + "id"))
	}
}

type AttrEntity struct {
	Name           string
	NameCamel      string
	NameCamelIdent string
	Type           string
	Tag            string
	Comment        string
	IsPk           bool
	HasIndex       bool
}

func main() {
	flagParse()
	glog.Info("open database connection")
	ctx := context.Background()
	dbConf := &mysql.Config{Dsn: dsn}
	mysqlFactory = mysql.NewFactory()
	err := mysqlFactory.Open(dbConf)
	if err != nil {
		glog.Errorf("error: open db failed, %s", err.Error())
		os.Exit(1)
	}
	defer mysqlFactory.Close()
	mysqlClient := mysqlFactory.New(ctx)
	glog.Info("get tables")
	tableRows, err := mysqlClient.Query("show tables;")
	if err != nil {
		glog.Errorf("error: exec `show tables` failed, %v", err)
		os.Exit(1)
	}
	defer tableRows.Close()
	tableIndex := make(map[string]struct{})
	shadowTables := make(map[string]string)
	for tableRows.Next() {
		var table string
		tableRows.Scan(&table)
		tableIndex[table] = struct{}{}
	}
	var tables []string
	for index := range tableIndex {
		var table string
		if strings.HasPrefix(index, shadowPrefix) {
			table = strings.TrimPrefix(index, shadowPrefix)
		} else if strings.HasSuffix(index, shadowSuffix) {
			table = strings.TrimSuffix(index, shadowSuffix)
		} else {
			tables = append(tables, index)
			continue
		}
		if _, ok := tableIndex[table]; !ok {
			tables = append(tables, index)
			continue
		}
		shadowTables[index] = table
	}
	if len(tables) == 0 {
		glog.Errorf("error: no tables be found in database")
		os.Exit(1)
	}
	glog.Info("gen dao.go")
	pkg := filepath.Base(opDir)
	err = GenInitDao(ctx, pkg, shadowTables)
	if err != nil {
		println(err.Error())
	}
	glog.Info("gen tables")
	for _, table := range tables {
		if len(tablesMap) != 0 {
			if _, ok := tablesMap[table]; !ok {
				continue
			}
		}
		columns, err := GetTableColumns(ctx, table)
		if err != nil {
			glog.Infof("Get table columns failed, %v", err)
			continue
		}
		indexes, err := GetTableIndexes(ctx, table)
		if err != nil {
			glog.Infof("Get table indexes failed, %v", err)
			continue
		}
		rData, err := GetRenderData(ctx, pkg, table, columns, indexes)
		if err != nil {
			glog.Errorf("Get render data failed, %v", err)
			continue
		}
		if rData == nil {
			continue
		}
		glog.Infof("gen table %s", table)
		err = GenTable(ctx, table, rData)
		if err != nil {
			println(err.Error())
			continue
		}
		glog.Infof("gen table conds %s", table)
		err = GenTableConds(ctx, table, rData)
		if err != nil {
			println(err.Error())
			continue
		}
	}
}

const (
	MySQLV5IndexNum = 13
	MySQLV8IndexNum = 15
)

type IndexEntityV5 struct {
	Table         string
	Non_unique    bool
	Key_name      string
	Seq_in_index  int
	Column_name   string
	Collation     string
	Cardinality   int
	Sub_part      sql.NullString
	Packed        sql.NullString
	Null          string
	Index_type    string
	Comment       string
	Index_comment string
}

type IndexEntityV8 struct {
	IndexEntityV5
	Visible    string
	Expression sql.NullString
}

func GetTableIndexes(ctx context.Context, table string) (indexes []*IndexEntityV5, err error) {
	mysqlClient := mysqlFactory.New(ctx)
	indexRows, err := mysqlClient.Query(fmt.Sprintf("show index from `%s`", table))
	if err != nil {
		return indexes, fmt.Errorf("error:  failed to get indexes from table %s, %v", table, err)
	}
	defer indexRows.Close()
	cols, _ := indexRows.Columns()
	if len(cols) == MySQLV8IndexNum {
		for indexRows.Next() {
			indexEntity := IndexEntityV8{}
			indexStruct := sqlbuilder.NewStruct(indexEntity)
			err = indexRows.Scan(indexStruct.Addr(&indexEntity)...)
			if err != nil {
				return
			}
			indexes = append(indexes, &indexEntity.IndexEntityV5)
		}
	} else {
		for indexRows.Next() {
			indexEntity := IndexEntityV5{}
			indexStruct := sqlbuilder.NewStruct(indexEntity)
			err = indexRows.Scan(indexStruct.Addr(&indexEntity)...)
			if err != nil {
				return
			}
			indexes = append(indexes, &indexEntity)
		}
	}

	return
}

type ColumnEntity struct {
	Field      string
	Type       string
	Collation  sql.NullString
	Null       string
	Key        string
	Default    sql.NullString
	Extra      string
	Privileges string
	Comment    string
}

func GetTableColumns(ctx context.Context, table string) (columns []*ColumnEntity, err error) {
	mysqlClient := mysqlFactory.New(ctx)
	columnRows, err := mysqlClient.Query(fmt.Sprintf("show full columns from `%s`", table))
	if err != nil {
		return columns, fmt.Errorf("error:  failed to get full columns from table %s, %v", table, err)
	}
	defer columnRows.Close()
	for columnRows.Next() {
		columnEntity := &ColumnEntity{}
		columnStruct := sqlbuilder.NewStruct(columnEntity)
		columnRows.Scan(columnStruct.Addr(columnEntity)...)
		columns = append(columns, columnEntity)
	}
	return
}

func GetRenderData(ctx context.Context, pkg, table string,
	columns []*ColumnEntity, indexes []*IndexEntityV5) (rData *RenderData, err error) {
	if len(columns) == 0 {
		return rData, nil
	}
	attrs := make([]*AttrEntity, 0, len(columns))
	var primary string
	for _, column := range columns {
		nullable := false
		if column.Null == "YES" {
			nullable = true
		}
		dt := GetGOType(column.Type, nullable)
		var hasIndex, isPk bool
		if column.Key != "" {
			hasIndex = true
			if column.Key == "PRI" {
				isPk = true
				primary = column.Field
			}
		}
		attr := &AttrEntity{
			Name:           initialisms.SnakeToCamelIdentifier(column.Field),
			NameCamel:      ReplaceReserved(initialisms.ForceLowerCamelIdentifier(column.Field)),
			NameCamelIdent: initialisms.ForceCamelIdentifier(column.Field),
			Type:           dt,
			Tag:            column.Field,
			Comment:        column.Comment,
			IsPk:           isPk,
			HasIndex:       hasIndex,
		}
		attrs = append(attrs, attr)
	}
	idxs := make(Indexes)
	for _, index := range indexes {
		if index.Non_unique {
			continue
		}
		idxs[index.Key_name] = append(idxs[index.Key_name], index.Column_name)
	}
	rData = &RenderData{
		Pkg:                  pkg,
		Table:                table,
		TableLowerCamelIdent: initialisms.ForceLowerCamelIdentifier(table),
		TableUpperCamelIdent: initialisms.ForceCamelIdentifier(table),
		Primary:              primary,
		Attrs:                attrs,
		UniqueIndexes:        idxs,
	}
	return
}

func GenTable(ctx context.Context, table string, rData *RenderData) error {
	content, err := RenderTable(table, rData)
	if err != nil {
		return fmt.Errorf("error: render table %s tpl failed, %v", table, err)
	}
	f, err := GetTableFile(table)
	if err != nil {
		return fmt.Errorf("error: generate table %s file failed, %v", table, err)
	}

	f.Write(content)
	f.Close()
	cmd := exec.Command("goimports", "-w", f.Name())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error: format gen file %s failed, %v", f.Name(), err)
	}
	return nil
}

func GenTableConds(ctx context.Context, table string, rData *RenderData) error {
	content, err := RenderTableConds(table, rData)
	if err != nil {
		return fmt.Errorf("error: render table %s conds tpl failed, %v", table, err)
	}
	f, err := GetTableCondsFile(table)
	if err != nil {
		return fmt.Errorf("error: generate table %s conds file failed, %v", table, err)
	}

	f.Write(content)
	f.Close()
	cmd := exec.Command("goimports", "-w", f.Name())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error: format gen file %s failed, %v", f.Name(), err)
	}
	return nil
}

func GenInitDao(ctx context.Context, pkg string, shadowTables map[string]string) error {
	renderData := &RenderData{
		Pkg:          pkg,
		ShadowTables: shadowTables,
	}
	content, err := RenderInitDao(renderData)
	if err != nil {
		return fmt.Errorf("error: render dao tpl failed, %v", err)
	}
	f, err := GetInitDaoFile()
	if err != nil {
		// dao.go已存在时，不报错，直接忽略
		if err == FileAlreadyExistErr {
			return err
		}
		return fmt.Errorf("error: generate dao file failed, %v", err)
	}

	f.Write(content)
	f.Close()
	cmd := exec.Command("goimports", "-w", f.Name())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error: format gen file %s failed, %v", f.Name(), err)
	}
	return nil
}
