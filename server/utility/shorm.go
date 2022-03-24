package utility

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const (
	ormName = "shorm"
)

func CreateTable(db *sql.DB, object interface{}) error {
	tableName := ""
	var elmts reflect.Value
	if t := reflect.TypeOf(object); t.Kind() == reflect.Ptr {
		tableName = t.Elem().Name()
		elmts = reflect.ValueOf(object).Elem()
	} else {
		tableName = t.Name()
		elmts = reflect.ValueOf(object)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(", strings.ToLower(tableName))
	for i := 0; i < elmts.NumField(); i++ {
		field := elmts.Type().Field(i)
		tag := field.Tag.Get(ormName)

		if len(tag) == 0 {
			return fmt.Errorf("invalid tag[%s]! use \"%s\"", tag, ormName)
		}

		tags := strings.Split(tag, ";")

		query += field.Name
		switch field.Type.Kind() {
		case reflect.Int:
			query += " int"
		case reflect.String:
			check := false
			reg := regexp.MustCompile(`varchar\([0-9]+\)`)
			for i, t := range tags {
				r := reg.FindStringSubmatch(t)
				if len(r) == 1 {
					check = true
					tags = append(tags[:i], tags[i+1:]...)
					query += " " + strings.ToUpper(r[0])
					break
				}
			}
			if !check {
				return fmt.Errorf("invalid contents! use \"varchar(size)\"")
			}
		}
		for _, t := range tags {
			query += " " + t
		}

		if i < elmts.NumField()-1 {
			query += ", "
		}
	}
	query += ")"
	fmt.Println("query : ", query)
	_, err := db.Exec(query)
	return err
}
