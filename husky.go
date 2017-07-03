package husky

import (
	"errors"
	"github.com/domac/husky/log"
	"net"
	"time"
)

type CallBackFunc func(client *HuskyClient, p *Packet)

type HuskyServer struct {
	hostport              string
	keepalive             time.Duration
	isShutdown            bool
	StopChan              chan bool
	packetReceiveCallBack CallBackFunc
	rc                    *HuskyConfig
	listener              *ServerListener
}

func NewServer(hostport string, hc *HuskyConfig, callback CallBackFunc) *HuskyServer {

	if hc == nil {
		hc = NewDefaultConfig()
	}

	server := &HuskyServer{
		hostport:              hostport,
		StopChan:              make(chan bool, 1),
		packetReceiveCallBack: callback,
		isShutdown:            false,
		rc:                    hc,
		keepalive:             5 * time.Minute,
	}
	return server
}

func (self *HuskyServer) ListenAndServer() error {

	addr, err := net.ResolveTCPAddr("tcp4", self.hostport)
	if nil != err {
		log.GetLogger().Fatalf("Server resolveTCPAddr fail %s\n", err)
		return err
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if nil != err {
		log.GetLogger().Fatalf("Server ListenTCP fail >> %s\n", err)
		return err
	}

	sl := &ServerListener{listener, self.StopChan, self.keepalive}
	self.listener = sl
	log.GetLogger().Printf("开始监听连接\n")
	go self.serve()
	return nil
}

func (self *HuskyServer) serve() error {
	sl := self.listener
	for !self.isShutdown {
		conn, err := sl.Accept()
		if nil != err {
			log.GetLogger().Fatalf("Server serve accept fail %s\n", err)
			continue
		} else {
			remoteClient := NewClient(conn, self.packetReceiveCallBack, self.rc)
			remoteClient.Start()

		}
	}
	return nil
}

//服务端关闭
func (self *HuskyServer) Shutdown() {
	self.isShutdown = true
	close(self.StopChan)
	self.listener.Close()
	log.GetLogger().Printf("Server|Shutdown...\n")
}

//服务端监听器
type ServerListener struct {
	*net.TCPListener
	stop      chan bool
	keepalive time.Duration
}

func (self *ServerListener) Accept() (*net.TCPConn, error) {
	for {
		conn, err := self.AcceptTCP()
		select {
		case <-self.stop:
			return nil, errors.New("STOP LISTENING")
		default:
		}

		if nil == err {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(self.keepalive)
		} else {
			return nil, err
		}

		return conn, err
	}

}
