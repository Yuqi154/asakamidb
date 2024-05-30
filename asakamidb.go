package ASAKAMIDB

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func init() {
	_, err := os.Stat("data/asakamiplugins/db")
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll("data/asakamiplugins/db", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func NewASAKAMIDB() *ASAKAMIDB {
	return &ASAKAMIDB{}
}

type ASAKAMIDB struct {
	db *sql.DB
}

type Table interface {
	Name() string
	Schema() string
	Columns() []string
	Values() []value
}

type value interface{}

func (a *ASAKAMIDB) OpenDB(dbname string) error {
	var err error
	a.db, err = sql.Open("sqlite3", "data/asakamiplugins/db/"+dbname+".db")
	return err
}

func (a *ASAKAMIDB) Closedb() {
	a.db.Close()
}

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

func (a *ASAKAMIDB) CreateTable(Table Table) error {
	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", Table.Name(), Table.Schema())
	_, errdb := a.db.Exec(createTableSQL)
	if errdb != nil {
		return errdb
	}
	return nil
}

func (a *ASAKAMIDB) Insert(Table Table) error {
	columnsSQL := strings.Join(Table.Columns(), ", ")
	placeholders := strings.Repeat("?, ", len(Table.Columns()))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", Table.Name(), columnsSQL, placeholders)
	values := make([]interface{}, len(Table.Values()))
	for i, v := range Table.Values() {
		values[i] = v
	}
	_, err := a.db.Exec(insertSQL, values...)
	return err
}

func (a *ASAKAMIDB) InsertWithStruct(Table Table, model interface{}) error {
	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct")
	}
	var columns []string
	var values []interface{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i).Interface()
		columns = append(columns, fieldName)
		values = append(values, fieldValue)
	}
	columnsSQL := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?, ", len(columns))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", Table.Name(), columnsSQL, placeholders)
	_, err := a.db.Exec(insertSQL, values...)
	return err
}

func (a *ASAKAMIDB) Deletetable(Table Table) error {
	_, err := a.db.Exec("DROP TABLE " + Table.Name())
	return err
}

func (a *ASAKAMIDB) Delete(Table Table) error {
	_, err := a.db.Exec("DELETE FROM " + Table.Name() + " WHERE " + awhere(Table))
	return err
}

func (a *ASAKAMIDB) Update(OldTable Table, Table Table) error {
	columns := Table.Columns()
	var set []string
	for _, column := range columns {
		set = append(set, column+"=?")
	}
	setSQL := strings.Join(set, ", ")
	updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", Table.Name(), setSQL, awhere(OldTable))
	values := make([]interface{}, len(Table.Values()))
	for i, v := range Table.Values() {
		values[i] = v
	}
	_, err := a.db.Exec(updateSQL, values...)
	return err
}

func (a *ASAKAMIDB) Selectall(Table Table) (*sql.Rows, error) {
	return a.db.Query("SELECT * FROM " + Table.Name())
}

func (a *ASAKAMIDB) SelectData(Table Table) (*sql.Rows, error) {
	values := make([]interface{}, len(Table.Values()))
	for i, v := range Table.Values() {
		values[i] = v
	}
	return a.db.Query("SELECT * FROM "+Table.Name()+" WHERE "+awhere(Table), values...)
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

func awhere(Table Table) string {
	columns := Table.Columns()
	var where []string
	for _, column := range columns {
		if Table.Values()[0] == nil {
			continue
		}
		where = append(where, column+"=?")
	}
	return strings.Join(where, " AND ")
}
