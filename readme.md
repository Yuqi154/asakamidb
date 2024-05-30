
# ASAKAMIDB

ASAKAMIDB 是一个 Go 包，提供了管理 SQLite 数据库的简单接口。它允许用户定义自定义表结构，并执行创建表和插入数据等基本数据库操作。

## 特性

- 初始化和打开 SQLite 数据库。
- 使用 Go 接口定义自定义表结构。
- 从结构或模式创建表。
- 向表中插入数据。

## 安装

使用 `go get` 安装 ASAKAMIDB：

```sh
go get github.com/Yuqi154/asakamidb
```

## 使用方法

### 导入包

```go
import "github.com/Yuqi154/asakamidb"
```

### 初始化数据库

创建 `ASAKAMIDB` 实例并打开数据库连接：

```go
db := asakamidb.NewDB("path/to/your/database/")
err := db.OpenDB("yourdbname")
if err != nil {
    panic(err)
}
defer db.Closedb()
```

### 定义表

实现 `Table` 接口来定义你的表结构：

```go
type Table struct {}

func (t Table) Name() string {
    return "users"
}

func (t Table) Schema() string {
    return "id INTEGER PRIMARY KEY, name TEXT, age INTEGER"
}

func (t Table) Columns() []string {
    return []string{"name", "age"}
}

func (t Table) Values() []Value {
    return []Value{"John Doe", 30}    
}

type Value struct {
    Name string
    Age int
}
```

### 创建表

使用 `CreateTable` 方法创建表：

```go
Table := Table{}
err = db.CreateTable(Table)
if (err != nil) {
    panic(err)
}
```

### 插入数据

向表中插入数据：

```go
Table := Table{}
err = db.Insert(Table)
if err != nil {
    panic(err)
}
```

### 查询数据

使用 `SelectData` 方法查询数据：  
假设你要查找一个名为 `John Doe` 的用户，你可以这样做：

```go
Table := Table{}
Table.Values = []interface{}{"John Doe", nil}
rows, err := db.SelectData(Table)
if err != nil {
    panic(err)
}
defer rows.Close()
```

**注意：** 当Values中的项为`nil`时，会忽略这些项。

### 更新数据

使用 `Update` 方法更新数据：
    
```go
Table := Table{}
NewValues = []Value{"Jane Doe", 31}
err = db.Update(Table,NewValues)
if err != nil {
    panic(err)
}
```

## 贡献

欢迎贡献！请在 GitHub 上提交问题或拉取请求。

## 许可证

此项目使用 MIT 许可证。有关详细信息，请参阅 [LICENSE](LICENSE) 文件。

请根据您的具体项目设置和需求调整路径、包导入路径和其他详细信息。如果您需要进一步的定制或在 README 中添加其他部分，请告诉我！