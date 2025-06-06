// This file was generated by go-dao-code-gen

package {{ .Pkg }}

import (
	"context"
	"errors"
	"reflect"
	"strings"
    "database/sql"

	"github.com/yangsai7/go-mysql"
	"github.com/yangsai7/go-sqlbuilder"
)

const (
    // ShadowCtxKey specifies shadow identity in context.
	ShadowCtxKey string = "X-Shadow"
    // ForceMasterIdenti specifies force master tag in sql comment.
    ForceMasterIdenti string = "/*force_master*/ "
)

var (
	globalDB     *sql.DB
)

// InitDao initialize database connection and adding shadow mapping.
func InitDao(ctx context.Context, cfg *mysql.Config) (err error) {
    if cfg == nil {
        err = errors.New("mysql config is nil")
        return
    }
    connector, err := mysql.NewConnector(cfg)
    if err != nil {
        return
    }
	globalDB = sql.OpenDB(connector)
	return
}

// Close the connection pool, use in the main function via defer.
func Close() {
    if globalDB != nil {
        globalDB.Close()
    }
}

// InitTableFields initialize feild names list from table entity.  
func InitTableFields(v interface{}, fields *[]string) {
	t := reflect.TypeOf(v)
	*fields = make([]string, 0, t.NumField())
	var fieldName string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		fieldName = strings.Split(tag, ",")[0]
		if fieldName == "" {
			continue
		}
		*fields = append(*fields, strings.TrimSpace(fieldName))
	}
}

// InitTableUpdateFields Initialize modifiable field names from table entity. 
func InitTableUpdateFields(v interface{}, fields map[string]struct{}) {
	if fields == nil {
		panic("update fields not initial property")
	}
    cols := sqlbuilder.NewStruct(v).ColumnsForTag("update")
    for _, col := range cols {
        fields[col] = struct{}{}
    }
}

// InitTableOmitEmptyFields Initialize omitempty field names from table entity.
func InitTableOmitEmptyFields(v interface{}, fields map[string]struct{}) {
	if fields == nil {
		panic("update fields not initial property")
	}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		DBTag := t.Field(i).Tag.Get("db")
		opts := strings.Split(DBTag, ",")
		for _, opt := range opts {
			if strings.TrimSpace(opt) == "omitempty" {
				fields[opts[0]] = struct{}{}
			}
		}
	}
}

// InitTableAliasFields initialize alias of feilds from table entity.
func InitTableAlias(v interface{}, alias interface{}) {
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(alias)
	for i := 0; i < rt.NumField(); i++ {
		DBTag := rt.Field(i).Tag.Get("db")
		tags := strings.Split(DBTag, ",")
		if len(tags) < 1 {
			continue
		}
		rv.Elem().FieldByName(rt.Field(i).Name).SetString(tags[0])
	}
}

// GetOffset get pagination offset.
func GetOffset(pageIndex int, pageLimit int) (offset int) {
    if pageLimit < 1 {
        pageLimit = 10
    }
    if pageIndex < 1 {
        pageIndex = 1
    }
    return (pageIndex - 1) * pageLimit
}
