package main

import (
	"context"
	"github.com/miekg/dns"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jpillora/opts"
	log "github.com/sirupsen/logrus"
)

type (
	AudchMap  map[string]string
	AudchOpts struct {
		HostsFile string `opts:"help=hosts file to use for country lookups, short=c, default=/etc/hosts"`
	}
	AudchApp struct {
		lastStatus                                                      map[string]string // 创建容器有两个事件，再创建容器的时候容器并不存在与容器列表，所以还需监听start事件
		HostBytes                                                       []byte
		HostAll                                                         []string
		HostNow, HostLast                                               AudchMap
		Cli                                                             *client.Client
		HostNowData                                                     types.Container
		HostStr, Name, DnsmasqID, BridgeNetworkID, BridgeNetworkGateway string
		AudchOpts
	}
)

var (
	versionData string
	err         error
	Audch       AudchApp
)

// 初始化日志方法
func init() {
	opts.New(&Audch.AudchOpts).Name("Audch").PkgRepo().Version(versionData).Complete().Parse()
	if Audch.HostsFile == "" {
		Audch.HostsFile = "/hosts"
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// ClientDocker 连接客户端
func (AudchApp) ClientDocker() {
	Audch.Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Errorf("Unable to connect to Docker: %v", err)
		os.Exit(1)
	}
	log.Infoln("Successfully connected to Docker")
}

// EventListen 监听事件
func (AudchApp) EventListen() {
	log.Infoln("EventListen")
	msgs, errs := Audch.Cli.Events(context.Background(), types.EventsOptions{})

	for {
		select {
		case err := <-errs:
			log.Infoln("err", err)
		case msg := <-msgs:
			// 非容器事件跳过
			if msg.Type != "container" {
				break
			}
			// 非创建/销毁事件跳过
			tmpName := msg.Actor.Attributes["name"]
			if msg.Action == "create" {
				Audch.lastStatus[tmpName] = msg.Action
				break
			}
			if (msg.Action == "start" && Audch.lastStatus[tmpName] == "create") || msg.Action == "destroy" {
			} else {
				break
			}

			log.Infoln(msg.Action, msg.Actor.Attributes["name"])
			Audch.GetHostNow()
			Audch.GetHostLast()
			Audch.GetHostDiff()
			delete(Audch.lastStatus, msg.Actor.Attributes["name"])
			break

		}
	}
}

// GetBridge 获取桥接网络的网关
func (AudchApp) GetBridge() {
	networkList, err := Audch.Cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Errorf("Unable to get bridge network, error: %v", err)
		return
	}

	Audch.lastStatus = make(map[string]string)

	for _, v := range networkList {
		if v.Name == "bridge" {
			Audch.BridgeNetworkID = v.ID
			for _, v1 := range v.IPAM.Config {
				if v1.Gateway != "" {
					Audch.BridgeNetworkGateway = v1.Gateway
					break
				}
			}
		}
	}
	if Audch.BridgeNetworkID == "" || Audch.BridgeNetworkGateway == "" {
		log.Errorln("bridgeID/bridgeIP is empty")
		os.Exit(1)
	}
}

// ReturnName 获取容器名称
func (AudchApp) ReturnName() string {
	return strings.Replace(Audch.HostNowData.Names[len(Audch.HostNowData.Names)-1], "/", "", -1) + ".docker.shared" // TODO name: ["/adminer/db", "mysql"] *docker run --link
}

// GetIPAddress 获取容器IP
func (AudchApp) GetIPAddress() {
	if Audch.HostNowData.HostConfig.NetworkMode == "host" {
		Audch.HostNow[Audch.ReturnName()] = Audch.BridgeNetworkGateway // 172.17.0.1 like 127.0.0.1
		goto check
	}
	if k, ok := Audch.HostNowData.NetworkSettings.Networks["bridge"]; ok {
		Audch.HostNow[Audch.ReturnName()] = k.IPAddress
	} else {
		Audch.ConnectBridgeNetWork()
		Audch.GetIPAddress()
		// 等待下次执行
	}
check:
	if Audch.HostNow[Audch.ReturnName()] == "" {
		log.Infof("GetIPAddress Error: [%v]", Audch.ReturnName())
	}
}

// ConnectBridgeNetWork 连接桥接网络
func (AudchApp) ConnectBridgeNetWork() {
	err := Audch.Cli.NetworkConnect(context.Background(), Audch.BridgeNetworkID, Audch.HostNowData.ID, nil)
	if err != nil {
		log.Errorf("Unable connect to bridge network, name: %v, error: %v", Audch.ReturnName(), err)
		return
	}
	log.Infof("Connect bridge NetWork: %v", Audch.ReturnName())
	inspect, err := Audch.Cli.ContainerInspect(context.Background(), Audch.HostNowData.ID)
	if err != nil {
		return
	}
	Audch.HostNowData.NetworkSettings.Networks = inspect.NetworkSettings.Networks
}

// GetHostNow 获取当前容器
func (AudchApp) GetHostNow() {
	// 容器列表
	list, err := Audch.Cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Errorf("Unable to list containers: %v", err)
		Audch.ClientDocker() // 重新连接
		Audch.GetHostNow()
		return
	}
	Audch.HostNow = make(AudchMap)
	for _, Audch.HostNowData = range list {
		Audch.GetIPAddress()
	}
}

func (AudchApp) GetHostBytes() {
	Audch.HostBytes, err = os.ReadFile(Audch.HostsFile)
	if err != nil {
		log.Errorf("Failed to read %v: %v", Audch.HostsFile, err)
		return
	}
}

func (AudchApp) GetHostAll() {
	Audch.GetHostBytes()
	Audch.HostAll = strings.Split(string(Audch.HostBytes), "\n")
}

func (AudchApp) GetHostLast() {
	Audch.HostLast = make(AudchMap)
	Audch.GetHostAll()
	for i := 0; i < len(Audch.HostAll); i++ {
		if strings.Contains(Audch.HostAll[i], "# Audch") {
			row := strings.Split(Audch.HostAll[i], "\t")
			if len(row) != 3 {
				return
			}
			name := strings.Replace(row[1], " ", "", -1)
			ipaddress := strings.Replace(row[0], " ", "", -1)
			Audch.HostLast[name] = ipaddress
			if !strings.Contains(strings.Replace(row[1], " ", "", -1), "docker.shared") {
				log.Errorf("alias not match, please check: %v", name)
			}
			Audch.HostAll = append(Audch.HostAll[:i], Audch.HostAll[i+1:]...)
			i--
		}
		if i > 0 && Audch.HostAll[i] == Audch.HostAll[i-1] && Audch.HostAll[i] == "" {
			Audch.HostAll = append(Audch.HostAll[:i], Audch.HostAll[i+1:]...)
			i--
		}
	}
}

func (AudchApp) GetHostDiff() {
	del := 0
	update := 0
	if reflect.DeepEqual(Audch.HostLast, Audch.HostNow) {
		log.Infoln("Nothing to update")
		return
	}
	for k, v := range Audch.HostLast {
		if h, ok := Audch.HostNow[k]; !ok {
			log.Infof("Del: %v %v", k, v)
			del++
		} else {
			if h != v {
				log.Infof("Update: [%v %v] -> %v", k, v, h)
				update++
			}
		}
	}

	for k, v := range Audch.HostNow {
		if _, ok := Audch.HostLast[k]; !ok {
			log.Infof("Add: %v %v", k, v)
			update++
		}
	}

	if len(Audch.HostLast) != 0 && del == 0 && update == 0 && len(Audch.HostNow) == len(Audch.HostLast) {
		log.Infoln("Nothing to update")
		return
	}
	log.Infof("Last %v、Del: %v、Update: %v records", len(Audch.HostLast), del, update)

	for k, v := range Audch.HostNow {
		Audch.HostAll = append(Audch.HostAll, v+"\t"+k+"\t# Audch")
	}
	Audch.GetHostStr()
	Audch.HostWrite()
}

func (AudchApp) GetHostStr() {
	Audch.HostStr = ""
	for _, v := range Audch.HostAll {
		Audch.HostStr += v + "\n"
	}
}

func (AudchApp) HostWrite() {
	err = os.WriteFile(Audch.HostsFile, []byte(Audch.HostStr), 0644)
	if err != nil {
		log.Errorf("Failed to write %v, %v", Audch.HostsFile, err)
		return
	}
	log.Debugf("Write %v success", Audch.HostsFile)
}

func (AudchApp) dnsServer() {
	dns.HandleFunc("docker.shared.", func(w dns.ResponseWriter, m *dns.Msg) {
		r := new(dns.Msg)
		r.SetReply(m)
		r.Compress = true
		r.Authoritative = true
		r.RecursionAvailable = true
		IPAddress := "NULL"
		Domain := strings.TrimRight(r.Question[0].Name, ".")
		dnsType := r.Question[0].Qtype

		defer func(w dns.ResponseWriter, msg *dns.Msg) {
			log.Infof("Query: [%v ? %s --> %v]\n", w.RemoteAddr(), Domain, IPAddress)
			err := w.WriteMsg(msg)
			if err != nil {
				log.Errorf("Error writing response: %v", err)
				return
			}
		}(w, r)

		// not resolve
		if !(dnsType == dns.TypeA || dnsType == dns.TypeAAAA) {
			r.Rcode = dns.RcodeNotImplemented
			return
		}

		if addr := Audch.HostNow[Domain]; addr != "" {
			IPAddress = addr
			r.Answer = append(r.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   m.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(86400),
				},
				A: net.ParseIP(IPAddress).To4(),
			})
		}

		if len(r.Answer) == 0 {
			r.Rcode = dns.RcodeNameError
		}
	})

	// start server
	udpServer := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("Starting at %v\n", udpServer.Addr)
	if err := udpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to setup the udp server: %s\n", err.Error())
	}

	defer func(udpServer *dns.Server) {
		err := udpServer.Shutdown()
		if err != nil {
			log.Errorf("Failed to shutdown the udp server: %s\n", err.Error())
			return
		}
	}(udpServer)
}

func main() {
	Audch.ClientDocker()
	Audch.GetBridge()
	Audch.GetHostNow()
	Audch.GetHostLast()
	Audch.GetHostDiff()

	// Start DNS Server
	go Audch.EventListen()
	Audch.dnsServer()
}
