package husky

import (
	"errors"
	"sync/atomic"
	"time"
)

const (
	CONCURRENT_LEVEL = 8
)

var TIMEOUT_ERROR = errors.New("WAIT RESPONSE TIMEOUT ")
var ERROR_OVER_FLOW = errors.New("Group Over Flow")
var ERROR_NO_HOSTS = errors.New("NO VALID RemoteClient")

//响应式钩子
type Future struct {
	seqId      int32
	response   chan interface{}
	TargetHost string
	Err        error
}

func NewFuture(seqId int32, TargetHost string) *Future {
	return &Future{
		seqId,
		make(chan interface{}, 1),
		TargetHost,
		nil}
}

//创建有错误的future
func NewErrFuture(seqId int32, TargetHost string, err error) *Future {
	return &Future{
		seqId,
		make(chan interface{}, 1),
		TargetHost,
		err}
}

func (f Future) Error(err error) {
	f.Err = err
	f.response <- err
}

func (f Future) SetResponse(resp interface{}) {
	f.response <- resp

}

//异步获取
func (f Future) Get(timeout <-chan time.Time) (interface{}, error) {
	if nil != f.Err {
		return nil, f.Err
	}

	select {
	case <-timeout:
		select {
		case resp := <-f.response:
			return resp, nil
		default:
			//如果是已经超时了但是当前还是没有响应也认为超时
			return nil, TIMEOUT_ERROR
		}
	case resp := <-f.response:
		e, ok := resp.(error)
		if ok {
			return nil, e
		} else {
			//如果没有错误直接等待结果
			return resp, nil
		}
	}
}

//配置信息
type HuskyConfig struct {
	MaxSchedulerNum  chan int
	ReadBufferSize   int
	WriteBufferSize  int
	WriteChannelSize int
	ReadChannelSize  int
	IdleTime         time.Duration //连接空闲时间
	RequestHolder    *ReqHolder
}

func NewConfig(maxSchedulerNum, readbuffersize, writebuffersize,
	writechannelsize, readchannelsize int, idletime time.Duration, maxseqId int) *HuskyConfig {

	//定义holder
	holders := make([]map[int32]*Future, 0, CONCURRENT_LEVEL)
	locks := make([]chan *interface{}, 0, CONCURRENT_LEVEL)
	for i := 0; i < CONCURRENT_LEVEL; i++ {
		splitMap := make(map[int32]*Future, maxseqId/CONCURRENT_LEVEL)
		holders = append(holders, splitMap)
		locks = append(locks, make(chan *interface{}, 1))
	}
	rh := &ReqHolder{
		seqId:    0,
		locks:    locks,
		holders:  holders,
		maxseqId: maxseqId}

	hc := &HuskyConfig{
		MaxSchedulerNum:  make(chan int, maxSchedulerNum),
		ReadBufferSize:   readbuffersize,
		WriteBufferSize:  writebuffersize,
		WriteChannelSize: writebuffersize,
		ReadChannelSize:  readchannelsize,
		IdleTime:         idletime,
		RequestHolder:    rh,
	}
	return hc
}

func NewDefaultConfig() *HuskyConfig {
	return NewConfig(1000, 4*1024, 4*1024, 10000, 10000, 10*time.Second, 160000)
}

//请求信息寄存器
type ReqHolder struct {
	maxseqId int
	seqId    uint32
	locks    []chan *interface{}
	holders  []map[int32]*Future
}

func (rh *ReqHolder) CurrentSeqId() int32 {
	return int32((atomic.AddUint32(&rh.seqId, 1) % uint32(rh.maxseqId)))
}

func (rh *ReqHolder) locker(id int32) (chan *interface{}, map[int32]*Future) {
	return rh.locks[id%CONCURRENT_LEVEL], rh.holders[id%CONCURRENT_LEVEL]
}

//从requesthold中移除
func (rh *ReqHolder) ReleaseFuture(seqId int32, obj interface{}) {
	l, m := rh.locker(seqId)
	l <- nil
	defer func() { <-l }()

	future, ok := m[seqId]
	if ok {
		delete(m, seqId)
		future.SetResponse(obj)
	}
}

func (rh *ReqHolder) AddFuture(seqId int32, future *Future) {
	l, m := rh.locker(seqId)
	l <- nil
	defer func() { <-l }()
	m[seqId] = future
}
