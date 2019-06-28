package mysql
import (
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net"
)


type MysqlDriverPool struct {
	pools map[string]*MysqlDriver
}

func (p *MysqlDriverPool) Init(confs []network.MysqlConf) bool {
	p.pools = make(map[string]*MysqlDriver)

	for _, conf := range confs {
		if conf.Enable == false {
			log.Warn("mysql[%s] not enable", conf.Name)
			continue
		}

		if p.DB(conf.Name) != nil {
			log.Error("mysql conf error，duplicate lable[%s]", conf.Name)
			return false
		}

		db := &MysqlDriver{}
		db.Init(conf)
		if err := db.Open(); err != nil {
			log.Error("connect mysql database fail[%#v] error[%s]", conf, err)
			return false
		}
		p.pools[db.Name()] = db
		log.Trace("connect mysql database success[%#v]", conf)
	}
	return true
}

func (p *MysqlDriverPool) DB(name string) *MysqlDriver {
	db, find := p.pools[name]
	if find == false {
		return nil
	}
	return db
}

func (p *MysqlDriverPool) Stop() {
	for _, v := range p.pools {
		v.Close()
	}
}


