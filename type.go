package main

import (
	"regexp"
	"strconv"
	"strings"
)

// PrecScaleRE is the regexp that matches "(precision[,scale])" definitions in a
// database.
var PrecScaleRE = regexp.MustCompile(`\(([0-9]+)(\s*,[0-9]+)?\)$`)

func GetGOType(dt string, nullable bool) string {
	precision := 0

	// extract unsigned
	if strings.HasSuffix(dt, " unsigned") {
		dt = dt[:len(dt)-len(" unsigned")]
	}

	// extract precision
	dt, precision, _ = ParsePrecision(dt)

	var typ string

switchDT:
	switch dt {
	case "bit":
		if precision == 1 {
			typ = "bool"
			if nullable {
				typ = "sql.NullBool"
			}
			break switchDT
		} else if precision <= 8 {
			typ = "uint8"
		} else if precision <= 16 {
			typ = "uint16"
		} else if precision <= 32 {
			typ = "uint32"
		} else {
			typ = "uint64"
		}
		if nullable {
			typ = "sql.NullInt64"
		}

	case "bool", "boolean":
		typ = "bool"
		if nullable {
			typ = "sql.NullBool"
		}

	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext", "json":
		typ = "string"
		if nullable {
			typ = "sql.NullString"
		}

	case "tinyint":
		//people using tinyint(1) really want a bool
		if precision == 1 {
			typ = "bool"
			if nullable {
				typ = "sql.NullBool"
			}
			break
		}
		typ = "int64"
		if nullable {
			typ = "sql.NullInt64"
		}

	// Integer numbers are uniformly converted to int64 for easier processing.
	case "smallint", "mediumint", "int", "integer", "bigint":
		typ = "int64"
		if nullable {
			typ = "sql.NullInt64"
		}

	case "float":
		typ = "float32"
		if nullable {
			typ = "sql.NullFloat64"
		}

	case "decimal", "double":
		typ = "float64"
		if nullable {
			typ = "sql.NullFloat64"
		}

	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		typ = "[]byte"

	case "timestamp", "datetime", "date":
		typ = "time.Time"
		if nullable {
			typ = "sql.NullTime"
		}

	case "enum", "set", "time":
		// time is not supported by the MySQL driver. Can parse the string to time.Time in the user code.
		typ = "string"
		if nullable {
			typ = "sql.NullString"
		}

	default:
		if strings.HasPrefix(dt, "mysql"+".") {
			// in the same schema, so chop off
			typ = initialisms.SnakeToCamelIdentifier(dt[len("mysql")+1:])
		} else {
			typ = initialisms.SnakeToCamelIdentifier(dt)
		}
	}
	return typ
}

// ParsePrecision extracts (precision[,scale]) strings from a data type and
// returns the data type without the string.
func ParsePrecision(dt string) (string, int, int) {
	var err error

	precision := -1
	scale := -1

	m := PrecScaleRE.FindStringSubmatchIndex(dt)
	if m != nil {
		// extract precision
		precision, err = strconv.Atoi(dt[m[2]:m[3]])
		if err != nil {
			panic("could not convert precision")
		}

		// extract scale
		if m[4] != -1 {
			scale, err = strconv.Atoi(dt[m[4]+1 : m[5]])
			if err != nil {
				panic("could not convert scale")
			}
		}

		// change dt
		dt = dt[:m[0]] + dt[m[1]:]
	}
	// special enum and set
	if strings.HasPrefix(dt, "enum") {
		dt = "enum"
	}
	if strings.HasPrefix(dt, "set") {
		dt = "set"
	}

	return dt, precision, scale
}
