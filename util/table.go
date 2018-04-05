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
}

var (
	//ErrNotStruct .
	ErrNotStruct = errors.New("params is not a struct")
	//ErrNoType .
	ErrNoType = errors.New("Column don't have type")
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
		sql = "create table " + tableName + "("
		err error
		sub string
	)

	for _, cl := range cls {
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

	typ := ref.Kind()
	if typ != reflect.Struct {
		return nil, "", ErrNotStruct
	}

	tableName := ref.Name()
	num := ref.NumField()
	cls := []Desc{}
	for i := 0; i < num; i++ {
		field := ref.Field(i)
		temp := Desc{
			Name:     field.Name,
			Typ:      getOriginType(field.Type.Name()),
			PK:       getBoolColumn(field.Tag.Get("pk")),
			Unique:   getBoolColumn(field.Tag.Get("unique")),
			Index:    getBoolColumn(field.Tag.Get("index")),
			Null:     getBoolColumn(field.Tag.Get("null")),
			Comment:  field.Tag.Get("comment"),
			Defaulty: field.Tag.Get("default"),
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
	} else {
		return false
	}
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
