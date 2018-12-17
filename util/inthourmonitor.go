/// @file inthourmonitor.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-03-01

package util
import (
	"time"
	"gitee.com/jntse/gotoolkit/log"
)

// --------------------------------------------------------------------------
/// @brief 整点回调管理池，非线程安全
// --------------------------------------------------------------------------
type IntHourCallBack func(now int64)		// 整点回调函数
type IntHourMonitorPool struct {
	pool map[int64]*IntHourMonitor
}

func NewIntHourMonitorPool() *IntHourMonitorPool {
	return &IntHourMonitorPool{}
}

func (p *IntHourMonitorPool) Init() {
	p.pool = make(map[int64]*IntHourMonitor)
}

func (p *IntHourMonitorPool) Regist(clock int64, handler IntHourCallBack) bool {
	if clock < 0 || clock >= 24 {
		log.Error("Regist IntHourMonitor clock[%d] Must be 'clock < 0 || clock >= 24' ", clock)
		return false
	}
	_, ok := p.pool[clock]
	if ok == true {
		log.Error("Regist IntHourMonitor clock[%d] is repeated", clock)
		return false 
	}

	monitor := NewIntHourMonitor()
	monitor.Init(handler, clock)
	p.pool[clock] = monitor
	log.Info("Regist IntHourMonitor[%d] Sucess", clock)
	return true
}

func (p *IntHourMonitorPool) Run(now int64) {
	for _, v := range p.pool {
		v.Tick(now)
	}
}


// --------------------------------------------------------------------------
/// @brief 整点回调监听，非导出
// --------------------------------------------------------------------------
type IntHourMonitor struct {
	clock		int64
	tm_zero  	int64      	// 今日零点
	tm_fix   	int64       // 修正时间
	tm_delay 	int64       // 延迟时间
	processed 	bool        // 是否已经处理
	reached 	bool    	// 到达处理时间
	handler		IntHourCallBack
}

func NewIntHourMonitor() *IntHourMonitor {
	return &IntHourMonitor{}
}

func (p *IntHourMonitor) Init(handler IntHourCallBack, clock int64)	{
	// 检查注册时间是否已经过了
	if clock < int64(time.Now().Hour()) {
		p.tm_zero = GetDayStart() + DaySec  	// 明日零点
	}else {
		p.tm_zero = GetDayStart()     		// 今日零点
	}
	p.tm_delay = 20		// 给20秒冗余区间
	p.processed = false
	p.reached = false
	p.handler = handler
	p.clock = clock
	p.tm_fix = HourSec * clock
}

func (p *IntHourMonitor) Tick(now int64)	{
	if ( now >= (p.tm_zero + p.tm_fix) && now <= (p.tm_zero + p.tm_fix + p.tm_delay) )	{
		p.reached = true;
	} else {
		p.reached , p.processed = false, false
		if ( p.tm_zero + p.tm_fix < now )	{
			p.tm_zero  = GetDayStart() + DaySec; // 明日0点
		}
	}
	p.process(now);
}

func (p *IntHourMonitor) process(now int64)	{
	if ( !p.reached  || p.processed )	{ return }
	p.processed = true;
	p.handler(now);
}


