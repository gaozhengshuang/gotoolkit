/// @file ticker.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

// --------------------------------------------------------------------------
/// @brief GameTicker管理
// --------------------------------------------------------------------------
package util
import (
	"time"
	)

type TickerCallbackHandle func(now int64)

/*/
// --------------------------------------------------------------------------
/// @brief 使用官方库Ticker，channel和Select的使用，在大量的ticker下效率很低
/// @brief 精度纳秒级别
// --------------------------------------------------------------------------
type GameTicker struct {
	Duration    time.Duration
	Handler     TickerCallbackHandle
	Ticker      *time.Ticker
	Running		bool
}

func NewGameTicker(d time.Duration, h TickerCallbackHandle) *GameTicker {
	return &GameTicker{Duration: d, Handler: h, Running: false }
}

func (t *GameTicker) Start() {
	if t.Running == true { return }
	t.Ticker = time.NewTicker(t.Duration)
	t.Running = true
}

func (t *GameTicker) Stop() {
	if t.Running == false { return }
	t.Running = false
	t.Ticker.Stop()
}

func (t *GameTicker) Run(now int64) bool {
	//if t.Running == false { panic("'GameTicker' Is Not Running") }
	if t.Running == false { return false }
	select {
	case <-t.Ticker.C:
		t.Handler(now)
		return true
	default:
		break		// 一定要default
	}
	return false
}

/*/
// --------------------------------------------------------------------------
/// @brief 简单ticker，精度毫秒
// --------------------------------------------------------------------------
type IGameTicker interface {
	Run(now int64) bool
	Start()
	Stop()
	IsTimeout(now int64) bool
}

type GameTicker struct { 
	Duration    int64
	Handler     TickerCallbackHandle
	Timestamp	int64
	Running		bool
}

func NewGameTicker(d time.Duration, h TickerCallbackHandle) *GameTicker {
	if d < time.Millisecond { panic("The Ticker precision is 'time.Millisecond' ") }
	ms := d / time.Millisecond
	return &GameTicker{Duration: int64(ms), Handler: h, Running: false }
}

func (t *GameTicker) Start() {
	if t.Running == true { return }
	t.Timestamp = time.Now().UnixNano() / 1000000 + t.Duration
	t.Running = true
}

func (t *GameTicker) Stop() {
	if t.Running == false { return }
	t.Running = false
}

func (t *GameTicker) Run(now int64) bool {
	if t.Running == false { return false }
	if now > t.Timestamp {
		t.Timestamp = now + t.Duration
		t.Handler(now)
		return true
	}
	return false
}

func (t *GameTicker) IsTimeout(now int64) bool {
	if t.Running == false { return false }
	if now > t.Timestamp {
		t.Timestamp = now + t.Duration
		return true
	}
	return false
}
//*/
