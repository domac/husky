package husky

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/domac/husky/log"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var globalSessionId uint64

//Husky会话状态
type HuskySession struct {
	id           uint64
	conn         *net.TCPConn //物理连接
	remoteAddr   string
	buffreader   *bufio.Reader //到时候给codec的
	buffwriter   *bufio.Writer //到时候给codec的
	ReadChannel  chan *Packet  //到时候给codec的
	WriteChannel chan *Packet  //到时候给codec的
	closeFlag    int32
	closeMutex   sync.Mutex
	lasttime     time.Time //上次会话工作时间
	hc           *HuskyConfig
	codec        *HuskyCodec
}

//创建会话连接
func NewHuskySession(conn *net.TCPConn, hc *HuskyConfig) *HuskySession {

	//物理连接调优
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(hc.IdleTime * 2)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(hc.ReadBufferSize)
	conn.SetWriteBuffer(hc.WriteBufferSize)

	session := &HuskySession{
		id:           atomic.AddUint64(&globalSessionId, 1),
		conn:         conn,
		buffreader:   bufio.NewReaderSize(conn, hc.ReadBufferSize),
		buffwriter:   bufio.NewWriterSize(conn, hc.WriteBufferSize),
		ReadChannel:  make(chan *Packet, hc.ReadChannelSize),
		WriteChannel: make(chan *Packet, hc.WriteChannelSize),
		remoteAddr:   conn.RemoteAddr().String(),
		codec:        &HuskyCodec{MAX_BYTES},
		hc:           hc,
	}
	return session
}

func (session *HuskySession) ID() uint64 {
	return session.id
}

func (self *HuskySession) RemotingAddr() string {
	return self.remoteAddr
}

func (self *HuskySession) Idle() bool {
	return time.Now().After(self.lasttime.Add(self.hc.IdleTime))
}

func (self *HuskySession) Closed() bool {
	return atomic.LoadInt32(&self.closeFlag) == 1
}

//数据包读取
func (self *HuskySession) ReadPacket() {

	for !self.Closed() {
		//由于有for, 所以有defer的情况,最后用匿名函数包起来
		func() {
			defer func() {
				if err := recover(); nil != err {
					log.GetLogger().Fatalf("session read packet : %s recover == fail :%s", self.remoteAddr, err)
				}
			}()

			buffer, err := self.codec.Read(self.buffreader)

			if err != nil {
				self.Close()
				return
			}

			//通过译码器解码
			p, err := self.codec.UnmarshalPacket(buffer)
			if nil != err {
				self.Close()
				log.GetLogger().Fatalf("session read packet marshal packet : %s == fail close session: %s", self.remoteAddr, err)
				return
			}
			self.ReadChannel <- p
		}()
	}

}

//写出数据
func (self *HuskySession) Write(p *Packet) error {

	defer func() {
		if err := recover(); nil != err {
			log.GetLogger().Fatalf("session write %s recover fail :%s", self.remoteAddr, err)
		}
	}()

	if !self.Closed() {
		select {
		case self.WriteChannel <- p:
			return nil
		default:
			return errors.New(fmt.Sprintf("WRITE CHANNLE [%s] FULL", self.remoteAddr))
		}
	}
	return errors.New(fmt.Sprintf("session [%s] closed", self.remoteAddr))
}

//数据包发送
func (self *HuskySession) WritePacket() {

	//批量bulk
	packets := make([]*Packet, 0, 100)

	for !self.Closed() {

		//从写出通道那里获取发送的消息包任务
		p := <-self.WriteChannel
		if nil != p {
			packets = append(packets, p)
		}
		l := int(math.Min(float64(len(self.WriteChannel)), 100))
		//如果channel的长度还有数据批量最多读取100合并写出
		//减少系统调用

		for i := 0; i < l; i++ {
			p := <-self.WriteChannel
			if nil != p {
				packets = append(packets, p)
			}
		}

		if len(packets) > 0 {
			self.writeBulk(packets)
			self.lasttime = time.Now()
			//情况bulk
			packets = packets[:0]
		}
	}

	//deal left packet
	for {
		_, ok := <-self.WriteChannel
		if !ok {
			break
		}
	}
}

//批量写入到网络流
func (self *HuskySession) writeBulk(tlv []*Packet) {
	batch := make([]byte, 0, len(tlv)*128)
	for _, t := range tlv {
		p := self.codec.MarshalPacket(t)
		if nil == p || len(p) == 0 {
			continue
		}
		batch = append(batch, p...)
	}

	if len(batch) <= 0 {
		return
	}

	//考虑batch很大的解决方案
	tmp := batch
	l := 0
	for {
		length, err := self.buffwriter.Write(tmp)
		if err != nil {
			log.GetLogger().Printf("session write conn %s fail %s=%d=%d", self.remoteAddr, err, length, len(tmp))
			if err != io.ErrShortWrite {
				self.Close()
				return
			}

			if err == io.ErrShortWrite {
				self.buffwriter.Reset(self.conn)
			}
		}

		l += length
		//写完
		if l == len(batch) {
			break
		}
		tmp = tmp[l:]
	}
	self.buffwriter.Flush()
}

//会话关闭
func (self *HuskySession) Close() error {
	if atomic.CompareAndSwapInt32(&self.closeFlag, 0, 1) {
		log.GetConsoleLogger().Printf("Session is Closing ...| %s\n", self.remoteAddr)
		log.GetLogger().Printf("Session is Closing ...| %s\n", self.remoteAddr)
		self.buffwriter.Flush()
		self.conn.Close()
		close(self.WriteChannel)
		close(self.ReadChannel)
	}
	return nil
}
