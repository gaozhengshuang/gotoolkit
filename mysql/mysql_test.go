/// @file mysql_test.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-10-31

package mysql_test
import (
	"fmt"
	"testing"
	"gitee.com/jntse/gotoolkit/mysql"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net"
)

var printf = fmt.Printf
var printn = fmt.Println
var db *mysql.MysqlDriver

func TestMysql(t *testing.T) {
	log.Info("============= TestMysql =============")
	defer log.Info("============= TestMysql =============")

	// 初始化
	conf := network.MysqlConf{}
	conf.Enable = true
	conf.User = "treasure"
	conf.Passwd = "pp123"
	conf.Database = "test"
	conf.Host = network.NetHost{Ip:"127.0.0.1", Port:3306}
	conf.MaxIdleConn = 100

	db = &mysql.MysqlDriver{}
	db.Init(conf)
	if db.Open() == nil {
		log.Info("init mysql ok")
		log.Info("%#v",db.Stats())
	}

	NewTable()
	Insert()

	// 选择
	Select("")
	//Select("uid=2")

	//Select1("")
	//Select1("uid=2")

	Delete()
	Update()

}

// 创建表格
func NewTable() {
	sql := `DROP TABLE IF EXISTS userinfo;`
	_, err := db.Exec(sql)
	db.CheckErr(err, true)

	sql = `CREATE TABLE userinfo (
		uid INT(10) NOT NULL AUTO_INCREMENT,
		username VARCHAR(64) NULL DEFAULT NULL,
		phone BIGINT(20) NULL DEFAULT NULL,
		score FLOAT NULL DEFAULT NULL,
		codec BLOB NULL DEFAULT NULL,
		empty VARCHAR(64) NULL DEFAULT NULL,
		developer TINYINT(4) NULL DEFAULT NULL,
		login DATE NULL DEFAULT NULL,
		PRIMARY KEY (uid)
	);`
	re, err := db.Exec(sql)
	db.CheckErr(err)

	lastid, err := re.LastInsertId()
	affect, err := re.RowsAffected()
	log.Info("%d %v", lastid, err)
	log.Info("%d %v", affect, err)
}


// 插入
func Insert() {
	defer func(start int64) {
		log.Info("Insert耗时[%d]us", util.CURTIMEUS() - start)
	}(util.CURTIMEUS())

	for i:=0; i < 10; i++ {
		args := make([]*mysql.MysqlField, 0)
		args = append(args, &mysql.MysqlField{Name:"username", Value:"jacky"})
		args = append(args, &mysql.MysqlField{Name:"phone", Value:8613681626939})
		args = append(args, &mysql.MysqlField{Name:"score", Value:88.123456789})
		args = append(args, &mysql.MysqlField{Name:"codec", Value:[]byte{1,2,3,4,5}})
		args = append(args, &mysql.MysqlField{Name:"empty", Value:"empty"})
		args = append(args, &mysql.MysqlField{Name:"developer", Value:true})
		args = append(args, &mysql.MysqlField{Name:"login", Value:"2018-09-03"})
		re, err := db.Insert("userinfo", args...)
		lastid, _ := re.LastInsertId()
		affect, _ := re.RowsAffected()
		log.Info("%d %d", lastid, affect)
		db.CheckErr(err, true)
	}
}

//
func Select(cond string, limit ...int32) {
	defer func(start int64) {
		log.Info("Select耗时[%d]us", util.CURTIMEUS() - start)
	}(util.CURTIMEUS())


	// select * from tbl
	rows, err := db.Select("userinfo", nil, cond, limit...)
	db.CheckErr(err)
	for _, cols := range rows {
		log.Info("%v %v %v %v %v %v %v %v %v",cols[0].Int(),
		cols[1].String(),
		cols[2].Int64(),
		cols[3].Float32(),
		cols[4].Bytes(),
		cols[5].String(),
		cols[5].IsNil(),
		cols[6].Bool(),
		cols[7].String())
	}

	// select [filed...]  from tbl
	rows, err = db.Select("userinfo", []string{"uid", "username","login"}, cond, limit...)
	db.CheckErr(err)
	for _, cols := range rows {
		log.Info("%d %s %s ",
		cols[0].Int(),
		cols[1].String(),
		cols[2].String())
	}

}

func Select1(cond string, limit ...int32) {
	defer func(start int64) {
		log.Info("Select1耗时[%d]us", util.CURTIMEUS() - start)
	}(util.CURTIMEUS())

	rows, err := db.RawSelect("userinfo", nil, cond, limit...)
	db.CheckErr(err)

	for rows.Next() {
		type TableDefine struct {
			uid int
			username string
			phone int64
			socre float32
			codec []byte
			empty string
			developer bool
			login string
		}
		t := &TableDefine{}
		err = rows.Scan(&t.uid, &t.username, &t.phone, &t.socre, &t.codec, &t.empty, &t.developer, &t.login)
		db.CheckErr(err)
		log.Info("%v", t)

		//
		coltypes, _ := rows.ColumnTypes()
		for _, col := range coltypes {
			log.Info("name:%s dt:%s kind:%v", col.Name(), col.DatabaseTypeName(), col.ScanType().Kind())
		}
	}
}

func Delete() {
	defer func(start int64) {
		log.Info("Delete耗时[%d]us", util.CURTIMEUS() - start)
	}(util.CURTIMEUS())
	re, err := db.Delete("userinfo", "uid like '%1%'")
	db.CheckErr(err)

	lastid, err := re.LastInsertId()
	affect, err := re.RowsAffected()
	log.Info("%d %v", lastid, err)
	log.Info("%d %v", affect, err)
}

func Update() {
	defer func(start int64) {
		log.Info("Update耗时[%d]us", util.CURTIMEUS() - start)
	}(util.CURTIMEUS())
	args := make([]*mysql.MysqlField, 0)
	args = append(args, &mysql.MysqlField{Name:"username", Value:"superjakcy"})
	args = append(args, &mysql.MysqlField{Name:"codec", Value:[]byte{5,4,3,2,1}})
	re, err := db.Update("userinfo", "uid=2", args...)
	db.CheckErr(err)

	lastid, err := re.LastInsertId()
	affect, err := re.RowsAffected()
	log.Info("%d %v", lastid, err)
	log.Info("%d %v", affect, err)

}
