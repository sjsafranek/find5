package database

import (
    "fmt"
    "strings"
    // 
    // "log"
)

type Table struct {
    Name string `json:"string"`
    Columns []Column `json:"columns"`
}

func (self *Table) IsTimeSeries() bool {
    for _, column := range self.Columns {
        if "timestamp" == column.Type || "timestamptz" == column.Type {
            return true
        }
    }
    return false
}

func (self *Table) IsGeoSpatial() bool {
    for _, column := range self.Columns {
        if "point" == column.Type || "linestring" == column.Type || "polygon" == column.Type {
            return true
        }
    }
    return false
}

func (self *Table) ToSQL() string {
    _sqlColumns := []string{}
    for _, column := range self.Columns {
        _sqlColumns = append(_sqlColumns, column.ToSQL())
    }

    return fmt.Sprintf(`
CREATE TABLE %v (
        %v
)`, self.Name, strings.Join(_sqlColumns, ",\n\t"))
}

type Column struct {
    Name string `json:"name"`
    Type string `json:"type"`
    IsKey bool `json:"is_key"`
}

func (self *Column) ToSQL() string {
    dtype := self.Type

    switch(self.Type) {
    case "float32":
        dtype = "REAL"
        break
    case "float64":
        dtype = "DOUBLE PRECISION"
        break
    case "int32":
        dtype = "INTEGER"
        break
    case "int64":
        dtype = "BIG INT"
        break
    case "varchar":
        dtype = "VARCHAR"
        break
    }

    if self.IsKey {
        dtype += " PRIMARY KEY"
    }

    return fmt.Sprintf("%v      %v", self.Name, dtype)
}

//
// func init() {
//     table := Table{
//         Name: "mytable",
//         Columns: []Column{
//             Column{
//                 Name: "column1",
//                 Type: "float64",
//                 IsKey: true,
//             },
//             Column{
//                 Name: "column2",
//                 Type: "float32",
//             },
//             Column{
//                 Name: "column3",
//                 Type: "int64",
//             },
//             Column{
//                 Name: "column4",
//                 Type: "int32",
//             },
//             Column{
//                 Name: "column3",
//                 Type: "varchar",
//             },
//         },
//     }
//
//     log.Println(table.ToSQL())
// }
