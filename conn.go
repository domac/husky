package husky

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

//conn的功能封装层
type HSession struct {
	id           int64
	conn         *net.TCPConn //物理连接
	remoteAddr   string
	buffreader   *bufio.Reader //读缓冲层
	buffwriter   *bufio.Writer //写缓冲层
	ReadChannel  chan *Packet  //消息源通道
	WriteChannel chan *Packet  //目标消息通道
	closeFlag    int32
	closeMutex   sync.Mutex
	lasttime     time.Time //上次会话工作时间
	hc           *HConfig
	codec        *HCodec
	limiter      *RateLimiter //限流阀
}

//创建会话连接
func NewHSession(conn *net.TCPConn, hc *HConfig) *HSession {

	//物理连接调优
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(hc.IdleTime * 2)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(hc.ReadBufferSize)
	conn.SetWriteBuffer(hc.WriteBufferSize)

	sessionId, _ := NewGUIDFactory(0).NewGUID() //生成唯一的sessionID

	//session是连接的处理层,husky传输的交互通过session处理
	session := &HSession{
		id:           int64(sessionId),
		conn:         conn,
		buffreader:   bufio.NewReaderSize(conn, hc.ReadBufferSize),
		buffwriter:   bufio.NewWriterSize(conn, hc.WriteBufferSize),
		ReadChannel:  make(chan *Packet, hc.ReadChannelSize),
		WriteChannel: make(chan *Packet, hc.WriteChannelSize),
		remoteAddr:   conn.RemoteAddr().String(),
		codec:        &HCodec{MAX_BYTES},
		hc:           hc,
	}
	return session
}

func (session *HSession) ID() int64 {
	return session.id
}

func (session *HSession) RemotingAddr() string {
	return session.remoteAddr
}

func (session *HSession) Idle() bool {
	return time.Now().After(session.lasttime.Add(session.hc.IdleTime))
}

func (session *HSession) Closed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

//数据包读取
func (session *HSession) ReadPacket() {

	for session != nil && !session.Closed() {

		//由于有for, 所以有defer的情况,最后用匿名函数包起来
		func() {

			defer func() {
				if err := recover(); nil != err {
					log.Fatalf("session read packet : %s recover == fail :%s", session.remoteAddr, err)
				}
			}()

			buffer, err := session.codec.Read(session.buffreader)

			if err != nil {
				session.Close()
				return
			}

			//限流功能
			if session.limiter != nil {
				ok := session.limiter.GetQuota()
				if !ok {
					log.Printf("reject packet, current quota :%d", session.limiter.QuotaPerSecond())
					return
				}
			}

			//通过译码器解码
			p, err := session.codec.UnmarshalPacket(buffer)
			if nil != err {
				session.Close()
				log.Fatalf("session read packet marshal packet : %s == fail close session: %s", session.remoteAddr, err)
				return
			}
			session.ReadChannel <- p
		}()
	}

}

//写出数据
func (session *HSession) Write(p *Packet) error {

	if session != nil && !session.Closed() {
		select {
		case session.WriteChannel <- p:
			return nil
		default:
			return errors.New(fmt.Sprintf("WRITE CHANNLE [%s] FULL", session.remoteAddr))
		}
	}
	return errors.New(fmt.Sprintf("session [%s] closed", session.remoteAddr))
}

//数据包发送
func (session *HSession) WritePackets() {

	//批量bulk
	packets := make([]*Packet, 0, 100)

	for session != nil && !session.Closed() {
		//从写出通道那里获取发送的消息包任务
		p := <-session.WriteChannel
		if nil != p {
			packets = append(packets, p)
		}
		l := int(math.Min(float64(len(session.WriteChannel)), 100))
		//如果channel的长度还有数据批量最多读取100合并写出
		//减少系统调用
		for i := 0; i < l; i++ {
			p := <-session.WriteChannel
			if nil != p {
				packets = append(packets, p)
			}
		}

		if len(packets) > 0 {
			session.writeBulk(packets)
			session.lasttime = time.Now()
			//清空包
			packets = packets[:0]
		}
	}

	//榨干剩下的数据包
	for {
		_, ok := <-session.WriteChannel
		if !ok {
			break
		}
	}
}

//批量写入到网络流
func (session *HSession) writeBulk(tlv []*Packet) {
	batch := make([]byte, 0, len(tlv)*128)
	for _, t := range tlv {
		p := session.codec.MarshalPacket(t)
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
		length, err := session.buffwriter.Write(tmp)
		if err != nil {
			log.Printf("session write conn %s fail %s=%d=%d", session.remoteAddr, err, length, len(tmp))
			if err != io.ErrShortWrite {
				session.Close()
				return
			}

			if err == io.ErrShortWrite {
				session.buffwriter.Reset(session.conn)
			}
		}

		l += length
		//写完
		if l == len(batch) {
			break
		}
		tmp = tmp[l:]
	}
	session.buffwriter.Flush()
}

//会话关闭
func (session *HSession) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.buffwriter.Flush()
		session.conn.Close()
		close(session.WriteChannel)
		close(session.ReadChannel)
		log.Printf("client (%s) was closed | sessionid-> %d \n", session.remoteAddr, session.ID())
	}
	return nil
}
