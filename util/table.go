package util

import (
	"errors"
	"log"
	"reflect"
	"strings"
)

/**


Define
	name 	字段名称
	pk		主键(bool)
	unique 	独立键(bool)
	index   索引(bool)
	type	数据类型+长度
	null  	是否为NULL(bool)
	default 默认值/自增(auto)
	comment 注释
	exclude 不否包含到sql
	id
**/

//Desc 数据表字段描述
type Desc struct {
	Name     string
	Typ      string
	PK       bool
	Defaulty string
	Unique   bool
	Index    bool
	Null     bool
	Comment  string
	Exclude  bool
}

var (
	//ErrNotStruct .
	ErrNotStruct = errors.New("params is not a struct")
	//ErrNoType .
	ErrNoType = errors.New("Column don't have type")
	//ErrTableName .
	ErrTableName = errors.New("table error name")
	//ErrKeyNotFound .
	ErrKeyNotFound = errors.New("key not found")
)

//GetSQL 获取数据库创建sql
func GetSQL(st interface{}) (string, error) {
	cls, name, err := GetInfo(st)
	if err != nil {
		return "", err
	}
	return GenerateSQL(cls, name), nil
}

//GenerateSQL 生成创建SQL
func GenerateSQL(cls []Desc, tableName string) string {
	if len(cls) == 0 || tableName == "" {
		return ""
	}

	var (
		sql = "create table if not exists " + tableName + "("
		err error
		sub string
	)

	for _, cl := range cls {

		if cl.Exclude {
			continue
		}

		if sub, err = getCL(cl); err != nil {
			log.Println(err)
			err = nil
			continue
		}
		sql += sub
	}

	sql = sql[:len(sql)-1] + `)engine=innodb charset=utf8;`

	return sql
}

//生成配置
func getCL(cl Desc) (string, error) {

	var (
		NotPk = true
	)

	name := strings.ToLower(cl.Name)
	sql := "`" + name + "`"
	if cl.Typ == "" {
		return "", ErrNoType
	}
	sql += ` ` + cl.Typ

	if cl.PK {
		sql += ` primary key`
		NotPk = false
	}

	if NotPk && !cl.Null {
		sql += ` not null`
	}

	if cl.Defaulty != "" {
		if cl.Defaulty == "auto" {
			sql += ` auto_increment`
		} else {
			sql += ` default "` + cl.Defaulty + `"`
		}
	}

	if cl.Unique {
		sql += ` unique`
	}

	if cl.Comment != "" {
		sql += ` comment "` + cl.Comment + `"`
	}

	sql += `,`

	if cl.Index {
		sql += `index ` + name + "(`" + name + "`),"
	}

	return sql, nil
}

//GetInfo 分析结构体
func GetInfo(st interface{}) ([]Desc, string, error) {

	ref := reflect.TypeOf(st)
	val := reflect.ValueOf(st)

	typ := ref.Kind()
	if typ != reflect.Struct {
		return nil, "", ErrNotStruct
	}

	tableName := strings.ToLower(ref.Name())
	num := ref.NumField()
	cls := []Desc{}
	for i := 0; i < num; i++ {

		field := ref.Field(i)

		//如果sub结构体没有被exclude
		if !getBoolColumn(field.Tag.Get("exclude")) && val.Field(i).Kind() == reflect.Struct {
			desc, _, err := GetInfo(val.Field(i).Interface())
			if err != nil {
				return nil, "", err
			}
			cls = append(cls, desc...)
			continue
		}

		temp := Desc{
			Name:     field.Name,
			Typ:      getOriginType(field.Type.Name()),
			PK:       getBoolColumn(field.Tag.Get("pk")),
			Unique:   getBoolColumn(field.Tag.Get("unique")),
			Index:    getBoolColumn(field.Tag.Get("index")),
			Null:     getBoolColumn(field.Tag.Get("null")),
			Comment:  field.Tag.Get("comment"),
			Defaulty: field.Tag.Get("default"),
			Exclude:  getBoolColumn(field.Tag.Get("exclude")),
		}
		customType := field.Tag.Get("type")
		if customType != "" {
			temp.Typ = customType
		}

		cls = append(cls, temp)
	}
	return cls, tableName, nil
}

//获取字段布尔值
func getBoolColumn(param string) bool {
	if param == "true" {
		return true
	}
	return false
}

//获取Golang定义的字段类型 作为默认的
func getOriginType(orgin string) string {
	switch orgin {
	case "string":
		return "varchar(32)"
	case "int", "uint32", "uint64", "int32", "int64":
		return "int(11)"
	default:
		return "varchar(64)"
	}
}

//GetFields 获取列
func GetFields(in interface{}) (out []string) {
	fs := reflect.TypeOf(in)
	vs := reflect.ValueOf(in)
	for i := 0; i < fs.NumField(); i++ {

		//结构体的话，使用子字段
		if vs.Field(i).Kind() == reflect.Struct {
			fields := GetFields(vs.Field(i).Interface())
			out = append(out, fields...)
			continue
		}

		//如果列的值没有设置
		if getBoolColumn(fs.Field(i).Tag.Get("exclude")) || IsZero(vs.Field(i).Interface()) {
			continue
		}

		name := fs.Field(i).Name
		out = append(out, strings.ToLower(name))
	}
	return
}

//GetValues 获取值
func GetValues(in interface{}) (out []interface{}) {
	vs := reflect.ValueOf(in)
	fs := reflect.TypeOf(in)
	for i := 0; i < vs.NumField(); i++ {

		//结构体的话，使用子字段
		if vs.Field(i).Kind() == reflect.Struct {
			values := GetValues(vs.Field(i).Interface())
			out = append(out, values...)
			continue
		}

		//如果列的值没有设置
		if getBoolColumn(fs.Field(i).Tag.Get("exclude")) || IsZero(vs.Field(i).Interface()) {
			continue
		}

		t := vs.Field(i).Kind()
		switch t {
		case reflect.Int:
			out = append(out, vs.Field(i).Interface().(int))
		case reflect.Int64:
			out = append(out, vs.Field(i).Interface().(int64))
		case reflect.Uint64:
			out = append(out, vs.Field(i).Interface().(uint64))
		case reflect.String:
			out = append(out, vs.Field(i).Interface().(string))
		default:
			continue
		}
	}
	return
}

//GetTableName 获取表名
func GetTableName(in interface{}) (string, error) {
	ref := reflect.TypeOf(in)
	typ := ref.Kind()
	if typ != reflect.Struct {
		return "", ErrTableName
	}
	return strings.ToLower(ref.Name()), nil
}

//GetFieldValue 获取字段值
func GetFieldValue(in interface{}, key string) (interface{}, error) {
	if in == nil || key == "" {
		return "", ErrKeyNotFound
	}
	return reflect.ValueOf(in).FieldByName(key).Interface(), nil
}

//IsZero 判断interface{}是否为0、""、nil
func IsZero(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
