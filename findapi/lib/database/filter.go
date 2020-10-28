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



/*




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
	Wkt        string            `json:"wkt,omitempty"`
	Conditions []*Filter         `json:"conditions,omitempty"`
}

func (self *Filter) isTest() bool {
	return "" != self.Test && "" == self.Logical
}

func (self *Filter) isLogicalBlock() bool {
	return "" == self.Test && "" != self.Logical
}

func (self *Filter) toGeoSQL(layerSRID int64) (string, error) {
	operation := ""

	switch self.Test {
	case "contains":
		operation = "ST_Contains"
		break
	case "within":
		operation = "ST_Within"
		break
	case "overlaps":
		operation = "ST_Overlaps"
		break
	default:
		return "", errors.New("unknown filter test")
	}

	if "" != self.Wkt {
		if 4269 != layerSRID {
			return fmt.Sprintf("%v(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", operation, layerSRID, self.Wkt), nil
		}
		return fmt.Sprintf("%v(geom, ST_GeomFromText('%v'))", operation, self.Wkt), nil
	} else if nil != self.Geometry {
		geom, err := self.Geometry.MarshalJSON()
		if nil != err {
			return "", err
		}
		if 4269 != layerSRID {
			return fmt.Sprintf("%v(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromGeoJSON('%v'), 4269))", operation, layerSRID, geom), nil
		}
		return fmt.Sprintf("%v(geom, ST_GeomFromGeoJSON('%v'))", operation, geom), nil
	} else {
		return "", errors.New("missing filter parameter: `wkt` or `geometry`")
	}
}

func (self *Filter) ToSQL(layerSRID int64) (string, error) {

	if "" == self.Test && "" == self.Logical {
		return "", errors.New("Must specify 'test' or 'logical'")
	}

	if "" != self.Test && "" != self.Logical {
		return "", errors.New("Filter has both 'test' and 'logical'")
	}

	if self.isLogicalBlock() {
		conditions := []string{}
		for _, filter := range self.Conditions {
			fSql, err := filter.ToSQL(layerSRID)
			if nil != err {
				return "", err
			}
			conditions = append(conditions, fSql)
		}
		return fmt.Sprintf(` (
			%v
		)`, strings.Join(conditions, strings.ToUpper(self.Logical))), nil
	}

	if self.isTest() {
		switch self.Test {

		case "contains":
			return self.toGeoSQL(layerSRID)
			// 	if 4269 != layerSRID {
			// 		return fmt.Sprintf("ST_Contains(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layerSRID, self.Wkt), nil
			// 	}
			// 	return fmt.Sprintf("ST_Contains(geom, ST_GeomFromText('%v'))", self.Wkt), nil

		case "within":
			return self.toGeoSQL(layerSRID)
			// 	if 4269 != layerSRID {
			// 		return fmt.Sprintf("ST_Within(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layerSRID, self.Wkt), nil
			// 	}
			// 	return fmt.Sprintf("ST_Within(geom, ST_GeomFromText('%v'))", self.Wkt), nil

		case "overlaps":
			return self.toGeoSQL(layerSRID)
			// 	if 4269 != layerSRID {
			// 		return fmt.Sprintf("ST_Overlaps(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layerSRID, self.Wkt), nil
			// 	}
			// 	return fmt.Sprintf("ST_Overlaps(geom, ST_GeomFromText('%v'))", self.Wkt), nil

		case "boolean":
			return fmt.Sprintf(`
				%v = %v
			`, self.ColumnId, self.Check), nil

		case "range":
			return fmt.Sprintf(`
				(
					%v >= %v
					AND
					%v <= %v
				)
			`, self.ColumnId, self.Min, self.ColumnId, self.Max), nil

		case "date_range":
			// TODO: BETWEEN operator
			// https://stackoverflow.com/questions/15817871/postgresql-filter-a-date-range
			return fmt.Sprintf(`
				(
					%v >= to_timestamp('%v', 'YYYY-MM-DD HH24:MI:SS')
					AND
					%v <= to_timestamp('%v', 'YYYY-MM-DD HH24:MI:SS')
				)
			`, self.ColumnId, self.Start, self.ColumnId, self.End), nil

		case "in":
			return fmt.Sprintf(`
				%v IN ( '%v' )
			`, self.ColumnId, strings.Join(self.Values, "','")), nil

		case "not_in":
			return fmt.Sprintf(`
				%v NOT IN ( '%v' )
			`, self.ColumnId, strings.Join(self.Values, "','")), nil

		case "is_null":
			return fmt.Sprintf(`
				%v IS NULL
			`, self.ColumnId), nil

		case "not_null":
			return fmt.Sprintf(`
				%v IS NOT NULL
			`, self.ColumnId), nil

		case "equals":
			return fmt.Sprintf(`
				%v = '%v'
			`, self.ColumnId, self.Value), nil

		case "not_equals":
			return fmt.Sprintf(`
				%v != '%v'
			`, self.ColumnId, self.Value), nil

		default:
			return "", errors.New("unknown filter test")

		}
	}

	return "", errors.New("unknown filter test")
}



*/
