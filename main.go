package main

import (
	"log"

	"github.com/biexiang/nt/util"
)

/**

Interface{} 可以自定义之后的关系，不包含到建表内


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

type Abcd struct {
	ID   int    `pk:"true" default:"auto" type:"int(8)" comment:"主键"`
	Name string `type:"varchar(255)" default:"golang" index:"true"`
	Pass string `unique:"true"`
	Desc string `type:"text"`
}

func main() {
	sql, err := util.GetSQL(Abcd{})
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(sql)
}
