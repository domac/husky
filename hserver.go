package husky

import (
	"errors"
	"golang.org/x/time/rate"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

type CallBackFunc func(client *HClient, p *Packet)

var GuidFactory *guidFactory

type HServer struct {
	hostport              string
	keepalive             time.Duration
	isShutdown            bool
	StopChan              chan bool
	packetReceiveCallBack CallBackFunc
	rc                    *HConfig
	listener              *HServerListener
	limiter               *RateLimiter
}

func NewServer(hostport string, hc *HConfig, callback CallBackFunc) *HServer {

	if hc == nil {
		hc = NewDefaultConfig()
	}

	server := &HServer{
		hostport:              hostport,
		StopChan:              make(chan bool, 1),
		packetReceiveCallBack: callback,
		isShutdown:            false,
		rc:                    hc,
		keepalive:             5 * time.Minute,
	}
	return server
}

func (hs *HServer) ListenAndServer() error {

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

	sl := &HServerListener{listener, hs.StopChan, hs.keepalive}
	hs.listener = sl
	log.Printf("start listening\n")
	go hs.serve()
	return nil
}

func (hs *HServer) serve() error {
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
			remoteClient := NewClient(conn, hs.rc, hs.packetReceiveCallBack)
			remoteClient.Start()

		}
	}
	return nil
}

//服务端关闭
func (hs *HServer) Shutdown() {
	hs.isShutdown = true
	close(hs.StopChan)
	hs.listener.Close()
	log.Printf("Server Shutdown...\n")
}

//服务端监听器
type HServerListener struct {
	*net.TCPListener
	stop      chan bool
	keepalive time.Duration
}

func (hsl *HServerListener) Accept() (*net.TCPConn, error) {
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
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Fatalf("temporary Accept() failure - %s", err)
				runtime.Gosched()
				continue
			}
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Fatalf("listener.Accept() - %s", err)
			}
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
