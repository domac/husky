package husky

import (
	"errors"
	"golang.org/x/time/rate"
	"log"
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
	listener              *HuskyServerListener
	limiter               *RateLimiter
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

//服务限速器
func (hs *HuskyServer) SetLimiter(limiter *RateLimiter) {
	hs.limiter = limiter
}

func (hs *HuskyServer) ListenAndServer() error {

	addr, err := net.ResolveTCPAddr("tcp4", hs.hostport)
	if nil != err {
		log.Fatalf("server resolve tcp addr fail %s\n", err)
		return err
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if nil != err {
		log.Fatalf("Server ListenTCP fail >> %s\n", err)
		return err
	}

	sl := &HuskyServerListener{listener, hs.StopChan, hs.keepalive}
	hs.listener = sl
	log.Printf("开始监听连接\n")
	go hs.serve()
	return nil
}

func (hs *HuskyServer) serve() error {
	sl := hs.listener

	for !hs.isShutdown {
		if hs.limiter != nil {
			ok := hs.limiter.GetQuota()
			if !ok {
				log.Printf("reject connection, current quota :%d", hs.limiter.QuotaPerSecond())
				continue
			}
		}
		conn, err := sl.Accept()
		if nil != err {
			log.Fatalf("Server serve accept fail %s\n", err)
			continue
		} else {
			remoteClient := NewClient(conn, hs.packetReceiveCallBack, hs.rc)
			remoteClient.Start()

		}
	}
	return nil
}

//服务端关闭
func (hs *HuskyServer) Shutdown() {
	hs.isShutdown = true
	close(hs.StopChan)
	hs.listener.Close()
	log.Printf("Server Shutdown...\n")
}

//服务端监听器
type HuskyServerListener struct {
	*net.TCPListener
	stop      chan bool
	keepalive time.Duration
}

func (hsl *HuskyServerListener) Accept() (*net.TCPConn, error) {
	for {
		conn, err := hsl.AcceptTCP()
		select {
		case <-hsl.stop:
			return nil, errors.New("Stop listen now ...")
		default:
		}

		if nil == err {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(hsl.keepalive)
		} else {
			return nil, err
		}

		return conn, err
	}

}

type RateLimiter struct {
	limiter     *rate.Limiter
	rejectCount int64
}

func NewRateLimiter(initQuota int, quotaPerSecond int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(quotaPerSecond), initQuota),
	}
}

func (rl *RateLimiter) QuotaPerSecond() int {
	return int(rl.limiter.Limit())
}

func (rl *RateLimiter) GetQuota() bool {
	return rl.limiter.Allow()
}

func (rl *RateLimiter) GetQuotas(quotaCount int) bool {
	return rl.limiter.AllowN(time.Now(), quotaCount)
}
