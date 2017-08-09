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
	huskyConfig           *HuskyConfig
	AttachChannel         chan interface{} //用于处理统一个连接上返回信息
}

func NewClient(conn *net.TCPConn, packetReceiveCallBack func(remoteClient *HuskyClient, p *Packet),
	hc *HuskyConfig) *HuskyClient {

	//省缺配置
	if hc == nil {
		hc = NewDefaultConfig()
	}

	if packetReceiveCallBack == nil {
		packetReceiveCallBack = func(remoteClient *HuskyClient, resp *Packet) {
			remoteClient.FinishReq(resp.Header.PacketId, resp.Data)
		}
	}

	//会话状态
	remoteSession := NewHuskySession(conn, hc)

	Client := &HuskyClient{
		heartbeat:             0,
		conn:                  conn,
		session:               remoteSession,
		huskyConfig:           hc,
		packetReceiveCallBack: packetReceiveCallBack,
		AttachChannel:         make(chan interface{}, 100),
	}
	return Client
}

func (hclient *HuskyClient) RemoteAddr() string {
	return hclient.remoteAddr
}

func (hclient *HuskyClient) LocalAddr() string {
	return hclient.localAddr
}

func (hclient *HuskyClient) Idle() bool {
	return hclient.session.Idle()
}

func (hclient *HuskyClient) GetSession() *HuskySession {
	return hclient.session
}

//带Context的客户端启动
func (hclient *HuskyClient) StartWithContext(ctx context.Context) {
	go func() {
		hclient.Start()
		log.GetLogger().Info("wait context")
		<-ctx.Done()
		hclient.Shutdown()
		log.GetLogger().Errorln(ctx.Err())
	}()
}

//客户端服务启动
func (hclient *HuskyClient) Start() {

	laddr := hclient.conn.LocalAddr().(*net.TCPAddr)
	raddr := hclient.conn.RemoteAddr().(*net.TCPAddr)
	hclient.localAddr = fmt.Sprintf("%s:%d", laddr.IP, laddr.Port)
	hclient.remoteAddr = fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	//开启写操作
	go hclient.session.WritePackets()

	//开启转发
	go hclient.schedulerPackets()

	//启动读取
	go hclient.session.ReadPacket()

	log.GetLogger().Infof("client start success ! local: %s , remote: %s \n", hclient.LocalAddr(), hclient.RemoteAddr())
}

//重连
func (hclient *HuskyClient) Reconnect() (bool, error) {

	//重新创建物理连接
	conn, err := net.DialTCP("tcp4", nil, hclient.conn.RemoteAddr().(*net.TCPAddr))
	if nil != err {
		log.GetLogger().Fatalf("client reconnect (%s) fail : %s\n", hclient.RemoteAddr(), err)
		return false, err
	}

	//重新设置conn
	hclient.conn = conn
	//创建session
	hclient.session = NewHuskySession(hclient.conn, hclient.huskyConfig)
	//create an new channel
	hclient.AttachChannel = make(chan interface{}, 100)

	//再次启动remoteClient
	hclient.Start()
	return true, nil
}

//包分发
func (hclient *HuskyClient) schedulerPackets() {

	for hclient.session != nil && !hclient.session.Closed() {

		p := <-hclient.session.ReadChannel
		if p == nil {
			continue
		}

		hclient.huskyConfig.MaxSchedulerNum <- 1

		go func() {
			defer func() {
				<-hclient.huskyConfig.MaxSchedulerNum
			}()
			//调用自定义的包消息处理函数

			if hclient.packetReceiveCallBack == nil {
				log.GetLogger().Fatalf("packetDispatcher is null \n")
			}

			hclient.packetReceiveCallBack(hclient, p)
		}()
	}
}

//直接释放请求
func (hclient *HuskyClient) FinishReq(seqId int32, obj interface{}) {
	defer func() {
		if err := recover(); nil != err {
			log.GetLogger().Fatalf("release packet fail : %s - %s\n", err, obj)
		}
	}()
	hclient.huskyConfig.RequestHolder.ReleaseFuture(seqId, obj)
}

var ERROR_PONG = errors.New("ERROR PONG TYPE !")

//同步发起ping的命令
func (hclient *HuskyClient) Ping(heartbeat *Packet, timeout time.Duration) error {
	pong, err := hclient.SyncWrite(*heartbeat, timeout)
	if nil != err {
		return err
	}
	version, ok := pong.(int64)
	if !ok {
		log.GetLogger().Fatalf("client ping error |%s\n", pong)
		return ERROR_PONG
	}
	hclient.updateHeartBeat(version)
	return nil
}

func (hclient *HuskyClient) updateHeartBeat(version int64) {
	if version > hclient.heartbeat {
		hclient.heartbeat = version
	}
}

func (hclient *HuskyClient) Pong(seqId int32, version int64) {
	hclient.updateHeartBeat(version)
}

//填充序列号并返回
func (hclient *HuskyClient) fillseqId(p *Packet) int32 {
	tid := p.Header.PacketId
	if tid < 0 {
		id := hclient.huskyConfig.RequestHolder.CurrentSeqId()
		p.Header.PacketId = id
		tid = id
	}
	return tid
}

//异步写
func (hclient *HuskyClient) Write(p Packet) (*Future, error) {
	pp := &p
	seqId := hclient.fillseqId(pp)
	future := NewFuture(seqId, hclient.localAddr)
	hclient.huskyConfig.RequestHolder.AddFuture(seqId, future)
	return future, hclient.session.Write(pp)

}

//同步写
func (hclient *HuskyClient) SyncWrite(p Packet, timeout time.Duration) (interface{}, error) {

	pp := &p
	seqId := hclient.fillseqId(pp)
	future := NewFuture(seqId, hclient.localAddr)
	hclient.huskyConfig.RequestHolder.AddFuture(seqId, future)
	err := hclient.session.Write(pp)
	// //同步写出
	if nil != err {
		return nil, err
	}
	tchan := time.After(timeout)
	resp, err := future.Get(tchan)
	return resp, err

}

func (hclient *HuskyClient) IsClosed() bool {
	return hclient.session.Closed()
}

//客户端关闭==>断开会话
func (hclient *HuskyClient) Shutdown() {
	hclient.session.Close()
	log.GetLogger().Printf("client shutdown %s...\n", hclient.RemoteAddr())
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
