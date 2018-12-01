package goUtils

import (
	"fmt"
	"runtime"
	"testing"
)

func TestDBase_Conn(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------%s--------\n", f.Name())
	myconf := MySQLConf{
		Host:    "10.96.114.84",
		User:    "phpmyadmin",
		Passwd:  "123456",
		DbName:  "db_wendao",
		Charset: "utf8",
		Timeout: 5,
		Port:    3306,
	}
	db := NewDBase(myconf)
	db.SetDebug(true)
	_, err := db.Conn()
	defer db.Close()
	fmt.Println(err)
	sql := "select * from videoinfo limit 10"
	ret, err := db.FetchRows(sql)
	fmt.Println(ret, err)
}

func TestDBase_InsertBatchData(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------%s--------\n", f.Name())
	myconf := MySQLConf{
		Host:       "10.96.114.84",
		User:       "phpmyadmin",
		Passwd:     "123456",
		DbName:     "db_wendao",
		Charset:    "utf8",
		Timeout:    5,
		Port:       3306,
		AutoCommit: true,
	}
	db := NewDBase(myconf)
	db.SetDebug(true)
	_, err := db.Conn()
	defer db.Close()
	fmt.Println(err)
	var data [][]interface{}
	fields := []string{"id1", "id2"}
	for i := 0; i < 100; i++ {
		var tmp []interface{}
		tmp = append(tmp, i)
		tmp = append(tmp, i+1)
		data = append(data, tmp)
	}
	ret, b, e := db.InsertBatchData("test", fields, data, true)
	fmt.Println(ret, b, e)
}

func TestGetMySQLTableStruct(t *testing.T) {
	myconf := MySQLConf{
		Host:       "10.96.114.84",
		User:       "phpmyadmin",
		Passwd:     "123456",
		DbName:     "db_runtofu",
		Charset:    "utf8",
		Timeout:    5,
		Port:       3306,
		AutoCommit: true,
	}
	db := NewDBase(myconf)
	_, err := db.Conn()
	if err != nil {
		fmt.Println(err)
		return
	}
	ret, err := GetMySQLTableStruct(db, "admin_menu")
	fmt.Println(err, ret)
}

func TestGetAllMySQLTables(t *testing.T) {
	myconf := MySQLConf{
		Host:       "10.96.114.84",
		User:       "phpmyadmin",
		Passwd:     "123456",
		DbName:     "db_runtofu",
		Charset:    "utf8",
		Timeout:    5,
		Port:       3306,
		AutoCommit: true,
	}
	db := NewDBase(myconf)
	_, err := db.Conn()
	if err != nil {
		fmt.Println(err)
		return
	}
	ret, err := GetAllMySQLTables(db)
	fmt.Println(err, ret)
}

func TestGetMySQLAllTablesStruct(t *testing.T) {
	myconf := MySQLConf{
		Host:       "10.96.114.84",
		User:       "phpmyadmin",
		Passwd:     "123456",
		DbName:     "db_runtofu",
		Charset:    "utf8",
		Timeout:    5,
		Port:       3306,
		AutoCommit: true,
	}
	db := NewDBase(myconf)
	_, err := db.Conn()
	fmt.Println(err)
	str, _ := GetMySQLAllTablesStruct(db)
	fmt.Println(str)
}
