package husky

import (
	"context"
	"errors"
	"fmt"
	"github.com/domac/husky/log"
	"net"
	"time"
)

//网络层客户端
type HuskyClient struct {
	conn                  *net.TCPConn
	localAddr             string
	remoteAddr            string
	heartbeat             int64
	session               *HuskySession
	packetReceiveCallBack func(client *HuskyClient, p *Packet)
	rc                    *HuskyConfig
	AttachChannel         chan interface{} //用于处理统一个连接上返回信息
}

func NewClient(conn *net.TCPConn, packetReceiveCallBack func(remoteClient *HuskyClient, p *Packet),
	hc *HuskyConfig) *HuskyClient {

	if hc == nil {
		hc = NewDefaultConfig()
	}

	if packetReceiveCallBack == nil {
		packetReceiveCallBack = func(remoteClient *HuskyClient, resp *Packet) {
			remoteClient.ReleaseReq(resp.Header.PacketId, resp.Data)
		}
	}

	//会话状态
	remoteSession := NewHuskySession(conn, hc)

	Client := &HuskyClient{
		heartbeat: 0,
		conn:      conn,
		session:   remoteSession,
		rc:        hc,
		packetReceiveCallBack: packetReceiveCallBack,
		AttachChannel:         make(chan interface{}, 100),
	}
	return Client
}

func (self *HuskyClient) RemoteAddr() string {
	return self.remoteAddr
}

func (self *HuskyClient) LocalAddr() string {
	return self.localAddr
}

func (self *HuskyClient) Idle() bool {
	return self.session.Idle()
}

func (self *HuskyClient) GetSession() *HuskySession {
	return self.session
}

//带Context的客户端启动
func (self *HuskyClient) StartWithContext(ctx context.Context) {
	go func() {
		self.Start()
		log.GetLogger().Info("wait context")
		<-ctx.Done()
		self.Shutdown()
		log.GetLogger().Errorln(ctx.Err())
	}()
}

//客户端服务启动
func (self *HuskyClient) Start() {

	laddr := self.conn.LocalAddr().(*net.TCPAddr)
	raddr := self.conn.RemoteAddr().(*net.TCPAddr)
	self.localAddr = fmt.Sprintf("%s:%d", laddr.IP, laddr.Port)
	self.remoteAddr = fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	//开启写操作
	go self.session.WritePackets()

	//开启转发
	go self.schedulerPackets()

	//启动读取
	go self.session.ReadPacket()

	log.GetLogger().Infof("client start success ! local: %s , remote: %s \n", self.LocalAddr(), self.RemoteAddr())
}

//重连
func (self *HuskyClient) reconnect() (bool, error) {

	//重新创建物理连接
	conn, err := net.DialTCP("tcp4", nil, self.conn.RemoteAddr().(*net.TCPAddr))
	if nil != err {
		log.GetLogger().Fatalf("client reconnect (%s) fail : %s\n", self.RemoteAddr(), err)
		return false, err
	}

	//重新设置conn
	self.conn = conn
	//创建session
	self.session = NewHuskySession(self.conn, self.rc)
	//create an new channel
	self.AttachChannel = make(chan interface{}, 100)

	//再次启动remoteClient
	self.Start()
	return true, nil
}

//包分发
func (self *HuskyClient) schedulerPackets() {

	for self.session != nil && !self.session.Closed() {

		p := <-self.session.ReadChannel
		if p == nil {
			continue
		}

		self.rc.MaxSchedulerNum <- 1

		go func() {
			defer func() {
				<-self.rc.MaxSchedulerNum
			}()
			//调用自定义的包消息处理函数

			if self.packetReceiveCallBack == nil {
				log.GetLogger().Fatalf("packetDispatcher is null \n")
			}

			self.packetReceiveCallBack(self, p)
		}()
	}
}

//直接释放请求
func (self *HuskyClient) ReleaseReq(seqId int32, obj interface{}) {
	defer func() {
		if err := recover(); nil != err {
			log.GetLogger().Fatalf("release packet fail : %s - %s\n", err, obj)
		}
	}()
	self.rc.RequestHolder.RemoveFuture(seqId, obj)
}

var ERROR_PONG = errors.New("ERROR PONG TYPE !")

//同步发起ping的命令
func (self *HuskyClient) Ping(heartbeat *Packet, timeout time.Duration) error {
	pong, err := self.SyncWrite(*heartbeat, timeout)
	if nil != err {
		return err
	}
	version, ok := pong.(int64)
	if !ok {
		log.GetLogger().Fatalf("client ping error |%s\n", pong)
		return ERROR_PONG
	}
	self.updateHeartBeat(version)
	return nil
}

func (self *HuskyClient) updateHeartBeat(version int64) {
	if version > self.heartbeat {
		self.heartbeat = version
	}
}

func (self *HuskyClient) Pong(seqId int32, version int64) {
	self.updateHeartBeat(version)
}

//填充序列号并返回
func (self *HuskyClient) fillseqId(p *Packet) int32 {
	tid := p.Header.PacketId
	if tid < 0 {
		id := self.rc.RequestHolder.CurrentSeqId()
		p.Header.PacketId = id
		tid = id
	}
	return tid
}

//异步写
func (self *HuskyClient) Write(p Packet) (*Future, error) {
	pp := &p
	seqId := self.fillseqId(pp)
	future := NewFuture(seqId, self.localAddr)
	self.rc.RequestHolder.AddFuture(seqId, future)
	return future, self.session.Write(pp)

}

//同步写
func (self *HuskyClient) SyncWrite(p Packet, timeout time.Duration) (interface{}, error) {

	pp := &p
	seqId := self.fillseqId(pp)
	future := NewFuture(seqId, self.localAddr)
	self.rc.RequestHolder.AddFuture(seqId, future)
	err := self.session.Write(pp)
	// //同步写出
	if nil != err {
		return nil, err
	}
	tchan := time.After(timeout)
	resp, err := future.Get(tchan)
	return resp, err

}

func (self *HuskyClient) IsClosed() bool {
	return self.session.Closed()
}

//客户端关闭==>断开会话
func (self *HuskyClient) Shutdown() {
	self.session.Close()
	log.GetLogger().Printf("client shutdown %s...\n", self.RemoteAddr())
}

//创建物理连接
func Dial(hostport string) (*net.TCPConn, error) {

	remoteAddr, err_r := net.ResolveTCPAddr("tcp4", hostport)
	if nil != err_r {
		return nil, err_r
	}
	conn, err := net.DialTCP("tcp4", nil, remoteAddr)
	if nil != err {
		return nil, err
	}
	return conn, nil
}
