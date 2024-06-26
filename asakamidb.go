package ASAKAMIDB

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// ASAKAMIDB 结构包含数据库连接和数据库路径。
type ASAKAMIDB struct {
	db   *sql.DB
	path string
}

// 接口 Table 由用户定义，用户可以自定义表结构。
type Table interface {
	Name() string
	Schema() string
	Columns() []string
	Values() []interface{}
}

type Value []interface{}

// NewDB 创建一个新的 ASAKAMIDB 结构。
func NewDB(path string) *ASAKAMIDB {
	return &ASAKAMIDB{path: path}
}

func (a *ASAKAMIDB) initdir() {
	_, err := os.Stat(a.path)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(a.path, 0755)
		if err != nil {
			panic(err)
		}
	}
}

// OpenDB 打开一个指定名称的数据库连接。
// 它初始化目录并设置数据库路径。
// 数据库连接存储在 ASAKAMIDB 结构中。
// 如果连接失败，它将返回一个错误。
func (a *ASAKAMIDB) OpenDB(dbname string) error {
	var err error
	a.initdir()
	a.db, err = sql.Open("sqlite3", a.path+dbname+".db")
	return err
}

// Closedb 关闭数据库连接。
func (a *ASAKAMIDB) Closedb() {
	a.db.Close()
}

// CreateTableWithStruct 创建一个表，表名为 Table.Name()，表结构为 model 的结构。
// 如果创建表失败，它将返回一个错误。
// 允许传入任意结构，但是必须是一个结构。
func (a *ASAKAMIDB) CreateTableWithStruct(Table Table, model interface{}) error {
	sqlStmt, err := generateCreateTableSQL(Table.Name(), model)
	if err != nil {
		return err
	}
	_, errdb := a.db.Exec(sqlStmt)
	if errdb != nil {
		return errdb
	}
	return nil
}

// CreateTable 创建一个表，表名为 Table.Name()，表结构为 Table.Schema()。
// 如果创建表失败，它将返回一个错误。
// Table接口由用户定义，用户可以自定义表结构。
func (a *ASAKAMIDB) CreateTable(Table Table) error {
	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", Table.Name(), Table.Schema())
	_, errdb := a.db.Exec(createTableSQL)
	if errdb != nil {
		return errdb
	}
	return nil
}

// Insert 插入一行数据到表 Table。
// 如果插入失败，它将返回一个错误。
// Table接口由用户定义，用户可以自定义表结构。
func (a *ASAKAMIDB) Insert(t Table) error {
	columnsSQL := strings.Join(t.Columns(), ", ")
	placeholders := strings.Repeat("?, ", len(t.Columns()))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", t.Name(), columnsSQL, placeholders)
	_, err := a.db.Exec(insertSQL, t.Values()...)
	return err
}

// InsertWithStruct 插入一行数据到表 Table。
// 如果插入失败，它将返回一个错误。
// 允许传入任意结构，但是必须是一个结构。
func (a *ASAKAMIDB) InsertWithStruct(t Table, model interface{}) error {
	r := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if r.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct")
	}
	var columns []string
	var values []interface{}
	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i).Interface()
		columns = append(columns, fieldName)
		values = append(values, fieldValue)
	}
	columnsSQL := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?, ", len(columns))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", t.Name(), columnsSQL, placeholders)
	_, err := a.db.Exec(insertSQL, values...)
	return err
}

// Deletetable 删除表 Table。
func (a *ASAKAMIDB) Deletetable(t Table) error {
	_, err := a.db.Exec("DROP TABLE " + t.Name())
	return err
}

// Delete 删除表 Table 中的数据。
func (a *ASAKAMIDB) Delete(t Table) error {
	_, err := a.db.Exec("DELETE FROM " + t.Name() + " WHERE " + awhere(t))
	return err
}

// Update 更新表 Table 中的数据。
// 先前的数据将被新数据替换。
func (a *ASAKAMIDB) Update(Table Table) error {
	a.Deletetable(Table)
	return a.Insert(Table)
}

// Selectall 从表 Table 中选择所有数据。
func (a *ASAKAMIDB) Selectall(t Table) ([]interface{}, error) {
	rows, err := a.db.Query("SELECT * FROM " + t.Name())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []interface{}
	for rows.Next() {
		err := rows.Scan(result...)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// SelectData 从表 Table 中选择数据。
func (a *ASAKAMIDB) SelectData(t Table) (*sql.Rows, error) {
	return a.db.Query("SELECT * FROM "+t.Name()+" WHERE "+awhere(t), t.Values()...)
}

func generateCreateTableSQL(Name string, model interface{}) (string, error) {
	t := reflect.TypeOf(model)

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("model must be a struct")
	}

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		sqlType, err := goTypeToSQLType(fieldType)
		if err != nil {
			return "", err
		}

		columns = append(columns, fmt.Sprintf("%s %s", fieldName, sqlType))
	}

	columnsSQL := strings.Join(columns, ", ")
	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", Name, columnsSQL)

	return createTableSQL, nil
}

func goTypeToSQLType(t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER", nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER", nil
	case reflect.Float32, reflect.Float64:
		return "REAL", nil
	case reflect.String:
		return "TEXT", nil
	case reflect.Bool:
		return "INTEGER", nil // SQLite does not have a separate Boolean storage class
	default:
		return "", fmt.Errorf("unsupported Go type: %s", t.Kind())
	}
}

// awhere 生成 WHERE 子句。
// 如果值为 nil，则不包含在 WHERE 子句中。
func awhere(t Table) string {
	columns := t.Columns()
	var where []string
	for i, column := range columns {
		if t.Values()[i] != nil {
			where = append(where, fmt.Sprintf("%s = ?", column))
		}
	}
	return strings.Join(where, " AND ")
}
