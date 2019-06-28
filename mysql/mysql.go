/// @file mysql.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-10-31

package mysql
import (
	"fmt"
	"reflect"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net"
)

// --------------------------------------------------------------------------
/// @brief mysql 配置解析
/// @brief 定义转到conf文件
// --------------------------------------------------------------------------
//type MysqlConf struct {
//	User 	string			`json:"user"`		// 用户
//	Passwd	string			`json:"passwd"`		// 密码
//	Database string			`json:"database"`	// 数据库
//	Host    network.NetHost	`json:"host"`		// 地址断开
//	MaxIdleConn int32		`json:"maxidleconn"`	// 最大闲置连接数
//}

// --------------------------------------------------------------------------
/// @brief mysql初始化选项
// --------------------------------------------------------------------------
//type MysqlInitOption struct {
//	User    	string
//	Passwd  	string
//	Addr    	string
//	Port    	int32
//	Database    string
//	MaxIdleConn int32
//}

type MysqlField struct {
	Name string
	Value interface{}
}


// --------------------------------------------------------------------------
/// @brief mysql 接口封装
// --------------------------------------------------------------------------
type MysqlDriver struct {
	//opt *MysqlInitOption
	conf network.MysqlConf
	db 	*sql.DB
}

func (m *MysqlDriver) Init(conf network.MysqlConf) error {
	m.db = nil
	m.conf = conf
	return nil
}

func (m *MysqlDriver) DB() *sql.DB {
	return m.db
}

func (m *MysqlDriver) IsOpen() bool {
	return m.db != nil
}

func (m *MysqlDriver) Name() string {
	return m.conf.Name;
}

// --------------------------------------------------------------------------
/// @brief 创建mysql连接
// --------------------------------------------------------------------------
func (m *MysqlDriver) Open() error {
	if m.IsOpen() == true {
		return fmt.Errorf("already open")
	}

	conf := m.conf
	strsql := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8", conf.User, conf.Passwd, conf.Host.String(), conf.Database)
	db, err := sql.Open("mysql", strsql)
	if m.CheckErr(err, true) == true {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	db.SetMaxIdleConns(int(conf.MaxIdleConn))
	m.db = db
	return nil
}

func (m *MysqlDriver) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *MysqlDriver) CheckErr(err error, assert...bool) bool {
	if err == nil {
		return false
	}

	log.Error("[mysql] mysql执行报错 MysqlErr[%s]", err)
	if len(assert) != 0 && assert[0] == true {
		panic(err)
	}
	return true
}

// --------------------------------------------------------------------------
/// @brief Stats returns database statistics.
///
/// @param 
// --------------------------------------------------------------------------
func (m *MysqlDriver) Stats() sql.DBStats {
	if m.IsOpen() == false {
		return sql.DBStats{}
	}
	return m.db.Stats()
}


// --------------------------------------------------------------------------
/// @brief  Exec executes a query without returning any rows
///
/// @param 
/// @param 
// --------------------------------------------------------------------------
func (m *MysqlDriver) Exec(sqlstr string) (sql.Result, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}
	result, err := m.db.Exec(sqlstr)
	return result, err
}


// --------------------------------------------------------------------------
/// @brief Query executes a query that returns rows, typically a `SELECT`
///
/// @param sqlstr string
/// @param error
// --------------------------------------------------------------------------
func (m *MysqlDriver) Query(sqlstr string) (*sql.Rows, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	rows, err := m.db.Query(sqlstr)
	return rows, err
}


// --------------------------------------------------------------------------
/// @brief Select查询
/// @brief 返回原生*sql.Rows
///
/// @param 
/// @param []string
/// @param string
/// @param 
/// @param error
// --------------------------------------------------------------------------
func (m *MysqlDriver) RawSelect(tblname string, colname []string, cond string, limit ...int32) (*sql.Rows, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	//
	cols := ""
	if len(colname) == 0 {
		cols = "*"
	}else {
		for k, col := range colname {
			cols += col
			if k != len(colname) - 1 { cols += "," }
		}
	}

	// select
	sqlstr := fmt.Sprintf("SELECT %s FROM `%s`", cols, tblname)

	// where
	if cond != "" {
		sqlstr = fmt.Sprintf("%s WHERE %s", sqlstr, cond)
	}

	// limit
	if len(limit) == 1 {
		sqlstr = fmt.Sprintf("%s LIMIT %d, %d", sqlstr, 0, limit[0])
	}else if len(limit) == 2 {
		sqlstr = fmt.Sprintf("%s LIMIT %d, %d", sqlstr, limit[0], limit[1])
	}

	rows, err := m.db.Query(sqlstr)
	return rows, err;
}


// --------------------------------------------------------------------------
/// @brief Select查询
/// @brief 返回sql.Rows的二次解析
///
/// @param 
/// @param []string
/// @param string
/// @param 
/// @param error
// --------------------------------------------------------------------------
func (m *MysqlDriver) Select(tblname string, colname []string, cond string, limit ...int32) ([][]*util.VarType, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	rows, err := m.RawSelect(tblname, colname, cond, limit...)
	if err != nil {
		return nil, err
	}

	varRows := make([][]*util.VarType, 0)
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			continue
		}

		vFields := make([]*util.VarType, len(columns))
		scanArgs := make([]interface{},  len(columns))
		for i := range scanArgs { 
			scanArgs[i] = &sql.RawBytes{}
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, param := range scanArgs {
			rawdata := param.(*sql.RawBytes)
			vFields[i] = util.NewVarType(*rawdata)
		}
		varRows = append(varRows, vFields)
	}

	return varRows, nil
}



// --------------------------------------------------------------------------
/// @brief 
///
/// @param 
// --------------------------------------------------------------------------
func (m *MysqlDriver) Insert(tblname string, fields ...*MysqlField) (sql.Result, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	cols, vals := "", ""
	for k, f := range fields {
		cols += "`" + f.Name + "`"

		kind := reflect.TypeOf(f.Value).Kind()
		if kind >= reflect.Int && kind <= reflect.Uint64 {
			vals += fmt.Sprintf("%d", f.Value)
		}else if kind == reflect.Float32 || kind == reflect.Float64 {
			vals += fmt.Sprintf("%f", f.Value)
		}else if kind == reflect.String || kind == reflect.Slice {	// slice '[]byte' for 'blob'
			vals += fmt.Sprintf(`"%s"`, f.Value)
		}else if kind == reflect.Bool {
			vals += fmt.Sprintf("%t", f.Value)
		}else {
			err := fmt.Errorf("not support filed datatype[%d]", kind)
			//m.CheckErr(err)
			return nil, err
		}

		if k != len(fields) - 1 { 
			cols += "," 
			vals += ","
		}
	}

	sqlstr := ""
	if len(fields) != 0 {
		sqlstr = fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)", tblname, cols, vals)
	}else {
		sqlstr = fmt.Sprintf("INSERT INTO `%s`() VALUES()", tblname)
	}

	re, err := m.db.Exec(sqlstr)
	//m.CheckErr(err)
	return re, err
}


// --------------------------------------------------------------------------
/// @brief 
///
/// @param 
// --------------------------------------------------------------------------
func (m *MysqlDriver) Update(tblname string, cond string, fields ...*MysqlField) (sql.Result, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	pairs := ""
	for k, f := range fields {
		pair, kind := "", reflect.TypeOf(f.Value).Kind()
		if kind >= reflect.Int && kind <= reflect.Uint64 {
			pair = fmt.Sprintf("`%s`=%d", f.Name, f.Value)
		}else if kind == reflect.Float32 || kind == reflect.Float64 {
			pair = fmt.Sprintf("`%s`=%f", f.Name, f.Value)
		}else if kind == reflect.String || kind == reflect.Slice {	// slice '[]byte' for 'blob', 'timestamp' type use 'string'
			pair = fmt.Sprintf("`%s`=\"%s\"", f.Name, f.Value)
		}else if kind == reflect.Bool {
			pair = fmt.Sprintf("`%s`=%t", f.Name, f.Value)
		}else {
			err := fmt.Errorf("not support filed datatype[%d]", kind)
			//m.CheckErr(err)
			return nil, err
		}

		if k != len(fields) - 1 {  pair += "," }
		pairs += pair
	}

	// select
	sqlstr := fmt.Sprintf("UPDATE `%s` SET %s", tblname, pairs)

	// where
	if cond != "" {
		sqlstr = fmt.Sprintf("%s WHERE %s", sqlstr, cond)
	}

	re, err := m.db.Exec(sqlstr)
	//m.CheckErr(err)
	return re, err
}


// --------------------------------------------------------------------------
/// @brief 删除数据(必须指定条件否则会清空表)
///
/// @param 
// --------------------------------------------------------------------------
func (m *MysqlDriver) Delete(tblname string, cond string) (sql.Result, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	if cond == "" {
		return nil, fmt.Errorf("Must Specify Filter Condition Or Else Will Clear Whole Table")
	}

	sqlstr := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tblname, cond)
	re, err := m.db.Exec(sqlstr)
	//m.CheckErr(err)
	return re, err
}


// --------------------------------------------------------------------------
/// @brief 清空表
///
/// @param 
/// @param error
// --------------------------------------------------------------------------
func (m *MysqlDriver) Truncate(tblname string) (sql.Result, error) {
	if m.IsOpen() == false {
		return nil, fmt.Errorf("db driver dosn't open")
	}

	sqlstr := fmt.Sprintf("TRUNCATE `%s`", tblname)
	re, err := m.db.Exec(sqlstr)
	//m.CheckErr(err)
	return re, err
}


