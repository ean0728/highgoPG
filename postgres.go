package highgoPG

import (
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Dialector struct {
	*Config
}

type Config struct {
	DriverName           string
	DSN                  string
	PreferSimpleProtocol bool
	WithoutReturning     bool
	Conn                 *sql.DB
}

func Open(dsn string) gorm.Dialector {
	info, err := parseDSN(dsn)
	if err != nil {
		panic(fmt.Errorf("parse dsn error, err: %v", err))
	}
	var newDsn string
	for k, v := range info {
		newDsn += fmt.Sprintf("%s=%s ", k, v)
	}

	return &Dialector{&Config{DSN: dsn}}
}

// ParseDSN 拆解 PostgreSQL DSN
func parseDSN(dsn string) (map[string]string, error) {
	// 解析 DSN
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	// 创建一个映射来存储拆解后的参数
	params := make(map[string]string)

	// 提取用户信息
	if u.User != nil {
		password, _ := u.User.Password()
		params["user"] = u.User.Username()
		params["password"] = password
	}

	// 提取主机和端口
	host := u.Hostname()
	port := u.Port()
	if port != "" {
		params["port"] = port
	} else {
		params["port"] = "5432" // PostgreSQL 默认端口
	}
	params["host"] = host

	// 提取数据库名称
	params["dbname"] = strings.TrimPrefix(u.Path, "/")

	// 提取其他参数
	for key, value := range u.Query() {
		if len(value) > 0 {
			params[key] = value[0]
		}
	}

	return params, nil
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (dialector Dialector) Name() string {
	return "postgres"
}

func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	// register callbacks
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		WithReturning: !dialector.WithoutReturning,
	})

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else if dialector.DriverName != "" {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.Config.DSN)
	} else {
		db.ConnPool, err = sql.Open("postgres", dialector.Config.DSN)
		if err != nil {
			return
		}
	}
	return
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   dialector,
		CreateIndexAfterCreateTable: true,
	}}}
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('$')
	writer.WriteString(strconv.Itoa(len(stmt.Vars)))
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteByte('"')
	if strings.Contains(str, ".") {
		for idx, str := range strings.Split(str, ".") {
			if idx > 0 {
				writer.WriteString(`."`)
			}
			writer.WriteString(str)
			writer.WriteByte('"')
		}
	} else {
		writer.WriteString(str)
		writer.WriteByte('"')
	}
}

var numericPlaceholder = regexp.MustCompile("\\$(\\d+)")

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, numericPlaceholder, `'`, vars...)
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "boolean"
	case schema.Int, schema.Uint:
		size := field.Size
		if field.DataType == schema.Uint {
			size++
		}
		if field.AutoIncrement {
			switch {
			case size <= 16:
				return "smallserial"
			case size <= 32:
				return "serial"
			default:
				return "bigserial"
			}
		} else {
			switch {
			case size <= 16:
				return "smallint"
			case size <= 32:
				return "integer"
			default:
				return "bigint"
			}
		}
	case schema.Float:
		if field.Precision > 0 {
			if field.Scale > 0 {
				return fmt.Sprintf("numeric(%d, %d)", field.Precision, field.Scale)
			}
			return fmt.Sprintf("numeric(%d)", field.Precision)
		}
		return "decimal"
	case schema.String:
		if field.Size > 0 {
			return fmt.Sprintf("varchar(%d)", field.Size)
		}
		return "text"
	case schema.Time:
		if field.Precision > 0 {
			return fmt.Sprintf("timestamptz(%d)", field.Precision)
		}
		return "timestamptz"
	case schema.Bytes:
		return "bytea"
	}

	return string(field.DataType)
}

func (dialectopr Dialector) SavePoint(tx *gorm.DB, name string) error {
	tx.Exec("SAVEPOINT " + name)
	return nil
}

func (dialectopr Dialector) RollbackTo(tx *gorm.DB, name string) error {
	tx.Exec("ROLLBACK TO SAVEPOINT " + name)
	return nil
}
