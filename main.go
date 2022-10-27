package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
)

type (
	AudchMap map[string]string
	AudchApp struct {
		defaultHost                                                     string
		Recreate                                                        bool
		HostBytes                                                       []byte // hosts Byte 内容
		HostAll                                                         []string
		HostNow, HostLast                                               AudchMap
		Cli                                                             *client.Client
		HostNowData                                                     types.Container
		HostStr, Name, DnsmasqID, BridgeNetworkID, BridgeNetworkGateway string
	}
)

var (
	err              error
	Audch            AudchApp
	version          = flag.Bool("v", false, "show version")
	defaultHostsFile = flag.String("f", "/etc/hosts", "hosts filepath")
	Debug            = flag.Bool("d", false, "debug")
	buildTime        = "2022-10-14/09:52:49"
	author           = "XRSec"
	commitId         = "2de23d5054449f77ae88c8c3371586b3e0a941c4"
	versionData      = "preview"
)

func init() {
	flag.Parse()
	// Version
	if *version {
		fmt.Printf("\033[1;34m %-12v\033[1;36m %v\n", "Version:", versionData)
		fmt.Printf("\033[1;34m %-12v\033[1;36m %v\n", "BuildTime:", buildTime)
		fmt.Printf("\033[1;34m %-12v\033[1;36m %v\n", "Author:", author)
		fmt.Printf("\033[1;34m %-12v\033[1;36m %v\n", "CommitId:", commitId)
		os.Exit(0)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *Debug {
		DebugModel()
	}
}

func (AudchApp) ClientDocker() {
	Audch.Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Errorf("Unable to connect to Docker: %v", err)
		os.Exit(1)
	}
	log.Infoln("Successfully connected to Docker")
}

func (AudchApp) GetBridge() {
	networkList, err := Audch.Cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Errorf("Unable to get bridge network, error: %v", err)
		return
	}
	for _, v := range networkList {
		if v.Name == "bridge" {
			Audch.BridgeNetworkID = v.ID
			Audch.BridgeNetworkGateway = v.IPAM.Config[0].Gateway
		}
	}
	if Audch.BridgeNetworkID == "" || Audch.BridgeNetworkGateway == "" {
		log.Errorln("bridgeID/bridgeIP is empty")
		return
	}
}

func (AudchApp) ReturnName() string {
	return strings.Replace(Audch.HostNowData.Names[len(Audch.HostNowData.Names)-1], "/", "", -1) + ".docker.shared" // TODO name: ["/adminer/db", "mysql"] *docker run --link
}

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

func (AudchApp) GetHostNow() {
	// hostNow
	list, err := Audch.Cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Errorf("Unable to list containers: %v", err)
		return
	}
	Audch.HostNow = make(AudchMap)
	for _, Audch.HostNowData = range list {
		Audch.GetIPAddress()
	}
}

func (AudchApp) GetHostBytes() {
	Audch.HostBytes, err = ioutil.ReadFile(*defaultHostsFile)
	if err != nil {
		log.Errorf("Failed to read %v: %v", *defaultHostsFile, err)
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

	if len(Audch.HostLast) != 0 && del == 0 && update == 0 && len(Audch.HostNow) == len(Audch.HostLast) {
		log.Infoln("Nothing to update")
		return
	}
	log.Infof("Last %v、Del: %v、Update: %v records", len(Audch.HostLast), del, len(Audch.HostNow)-update)

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
	err = ioutil.WriteFile(*defaultHostsFile, []byte(Audch.HostStr), 0644)
	if err != nil {
		log.Errorf("Failed to write %v, %v", *defaultHostsFile, err)
		return
	}
	log.Infof("Write %v success", *defaultHostsFile)
}

func DebugModel() {
	log.Infof("DebugModel Start.")
	defer func() {
		os.Exit(1)
	}()

	udpServer, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 53})
	if err != nil {
		log.Errorf("Failed to setup the udp server: %s\n", err.Error())
		return
	}
	defer func() {
		if err := udpServer.Close(); err != nil {
			log.Errorf("Failed to close the udp server: %s\n", err.Error())
			return
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, remoteAddr, err := udpServer.ReadFromUDP(buf)
		if err != nil {
			log.Errorf("Failed to read from udp server: %s\n", err.Error())
			return
		}
		if n <= 0 {
			continue
		}
		log.Infof("get: \n  {\n\tn: %v,\n\tremoteAddr: %v,\n\tbytes: %v\n  }\n", n, remoteAddr, string(buf))
		if _, err := udpServer.WriteToUDP([]byte("Hello Word"), remoteAddr); err != nil {
			log.Errorf("Failed to write to udp server: %s\n", err.Error())
			return
		}
	}
}

func audchServer() {
	Audch.ClientDocker()
	Audch.GetBridge()
	Audch.GetHostNow()
	Audch.GetHostLast()
	Audch.GetHostDiff()
}

func dnsServer() {
	// 本着极限的原则，我准备自己写一个 精简版的dns_server，但是目前已有的库都存在一个问题，无法解析，也许是我姿势不对，如果你有办法，请 提交 issues
	// 关于一些 库 可以参考 docs/Deprecated.md 中的 DNS 下面的 dnsServer
}

func main() {
	audchServer() // Run once when starting up
	c := cron.New()
	if _, err = c.AddFunc("*/5 * * * *", func() {
		audchServer()
	}); err != nil {
		log.Errorf("Cron Add Error: [%v]", err)
	}
	//启动定时器
	c.Start()
	log.Infof("Cron Start")

	dnsServer()
}
