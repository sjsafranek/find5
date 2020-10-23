package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paulmach/orb/geojson"
)

type Filter struct {
	Logical    string            `json:"logical,omitempty"`
	Test       string            `json:"test,omitempty"`
	ColumnId   string            `json:"column_id,omitempty"`
	Value      string            `json:"value,omitempty"`
	Values     []string          `json:"values,omitempty"`
	Min        float64           `json:"min,omitempty"`
	Max        float64           `json:"max,omitempty"`
	Start      time.Time         `json:"start,omitempty"`
	End        time.Time         `json:"end,omitempty"`
	Check      bool              `json:"check,omitempty"`
	Geometry   *geojson.Geometry `json:"geometry,omitempty"`
	Conditions []*Filter         `json:"conditions,omitempty"`
}

func (self *Filter) isTest() bool {
	return "" != self.Test && "" == self.Logical
}

func (self *Filter) isLogicalBlock() bool {
	return "" == self.Test && "" != self.Logical
}

func (self *Filter) ToSQL(table string) (string, error) {
	_sql := ""

	if "" == self.Test && "" == self.Logical {
		return "", errors.New("Must specify 'test' or 'logical'")
	}

	if "" != self.Test && "" != self.Logical {
		return "", errors.New("Filter has both 'test' and 'logical'")
	}

	if self.isLogicalBlock() {
		conditions := []string{}
		for _, filter := range self.Conditions {
			fSql, err := filter.ToSQL(table)
			if nil != err {
				return _sql, err
			}
			conditions = append(conditions, fSql)
		}
		_sql += fmt.Sprintf(` (
			%v
		)`, strings.Join(conditions, strings.ToUpper(self.Logical)))
	}

	if self.isTest() {
		switch self.Test {

		case "within":

			geom, err := self.Geometry.MarshalJSON()
			if nil != err {
				return _sql, err
			}

			_sql += fmt.Sprintf(`
				ST_Contains(
					ST_SetSRID(ST_GeomFromGeoJSON('%v'), 4326),
					places_view.geom
				)
			`, string(geom))
			break

		case "boolean":
			_sql += fmt.Sprintf(`
				places_view.%v = %v
			`, self.ColumnId, self.Check)
			break

		case "range":
			_sql += fmt.Sprintf(`
				(
					%v.%v >= %v
					AND
					%v.%v <= %v
				)
			`, table, self.ColumnId, self.Min, table, self.ColumnId, self.Max)
			break

		case "date_range":
			// TODO: BETWEEN operator
			// https://stackoverflow.com/questions/15817871/postgresql-filter-a-date-range
			_sql += fmt.Sprintf(`
				(
					%v.%v >= to_timestamp('%v', 'YYYY-MM-DD HH24:MI:SS')
					AND
					%v.%v <= to_timestamp('%v', 'YYYY-MM-DD HH24:MI:SS')
				)
			`, table, self.ColumnId, self.Start, table, self.ColumnId, self.End)
			break

		case "in":
			_sql += fmt.Sprintf(`
				%v.%v IN ( '%v' )
			`, table, self.ColumnId, strings.Join(self.Values, "','"))
			break

		case "not_in":
			_sql += fmt.Sprintf(`
				%v.%v NOT IN ( '%v' )
			`, table, self.ColumnId, strings.Join(self.Values, "','"))
			break

		case "is_null":
			_sql += fmt.Sprintf(`
				%v.%v IS NULL
			`, table, self.ColumnId)
			break

		case "not_null":
			_sql += fmt.Sprintf(`
				%v.%v IS NOT NULL
			`, table, self.ColumnId)
			break

		case "equals":
			_sql += fmt.Sprintf(`
				%v.%v = '%v'
			`, table, self.ColumnId, self.Value)
			break

		case "not_equals":
			_sql += fmt.Sprintf(`
				%v.%v != '%v'
			`, table, self.ColumnId, self.Value)
			break

		}

	}

	return _sql, nil
}

// {"method":"get_places","limit":1,"filter":{"test":"within","geometry":{"type":"Polygon","coordinates":[[[-122.6946,43.9592],[-122.6946,44.1565],[-123.478,44.1565],[-123.478,43.9592],[-122.6946,43.9592]]]}}}

/*
{
"test": "boolean",
"column_id": "is_deleted",
"check": false
}
{"method":"get_places","limit":1,"filter":{"logical":"and","conditions":[{"test":"boolean","column_id":"is_deleted","check":false},{"test":"range","column_id":"price","min":100,"max":200}]}}
{"method":"get_places","limit":1,"filter":{"logical":"and","conditions":[{"test":"boolean","column_id":"is_deleted","check":false},{"test":"match","column_id":"name","values":["appartamento piano terra"]}]}}
{"method":"get_places","limit":2,"filter":{"logical":"and","conditions":[{"test":"boolean","column_id":"is_deleted","check":false},{"logical":"or","conditions":[{"test":"range","column_id":"price","min":100,"max":200},{"test":"match","column_id":"name","values":["appartamento piano terra"]}]}]}}
*/
