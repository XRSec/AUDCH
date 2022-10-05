package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jpillora/opts"
	"github.com/jpillora/webproc/agent"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type (
	AudchMap  map[string]string
	AudchOpts struct {
		HostsFile     string `opts:"help=hosts file to use for country lookups, short=c, default=/etc/hosts"`
		EnableDnsmasq bool   `opts:"help=enable dnsmasq, default=false"`
		agent.Config  `opts:"mode=cmd, help=enable dnsmasq, name=dnsmasq"`
	}
	AudchApp struct {
		Recreate                                                        bool
		HostBytes                                                       []byte // hosts Byte 内容
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

func init() {
	opts.New(&Audch.AudchOpts).Name("Audch").PkgRepo().Version(versionData).Complete().Parse()
	if Audch.HostsFile == "" {
		Audch.HostsFile = "/etc/hosts"
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
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
		if Audch.ReturnName() == "audch.docker.shared" {
			continue
		}
		if strings.Contains(Audch.HostNowData.Image, "webproc") || strings.Contains(Audch.HostNowData.Image, "dnsmasq") {
			Audch.DnsmasqID = Audch.HostNowData.ID
		}
		Audch.GetIPAddress()
	}
}

func (AudchApp) GetHostBytes() {
	Audch.HostBytes, err = ioutil.ReadFile(Audch.HostsFile)
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
	Audch.DnsmasqRestart()
}

func (AudchApp) GetHostStr() {
	Audch.HostStr = ""
	for _, v := range Audch.HostAll {
		Audch.HostStr += v + "\n"
	}
}

func (AudchApp) HostWrite() {
	err = ioutil.WriteFile(Audch.HostsFile, []byte(Audch.HostStr), 0644)
	if err != nil {
		log.Errorf("Failed to write %v, %v", Audch.HostsFile, err)
		return
	}
	log.Infof("Write %v success", Audch.HostsFile)
}
func (AudchApp) DnsmasqRestart() {
	if !Audch.EnableDnsmasq {
		return
	}
	//curl -u admin:admin 'http://127.0.0.1:80/restart' -X 'PUT'
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://127.0.0.1:%v/restart", Audch.Port), nil)
	if err != nil {
		log.Errorf("Unable To Create Restart Request: %v", err)
		return
	}
	req.SetBasicAuth(Audch.User, Audch.Pass)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Unable To Send Restart Message To Server: %v", err)
		return
	}
	log.Infof("Restart dnsmasq: %v", res.Status)
}

func audchServer() {
	Audch.ClientDocker()
	Audch.GetBridge()
	Audch.GetHostNow()
	Audch.GetHostLast()
	Audch.GetHostDiff()
}

func dnsmasqServer() {
	args := Audch.ProgramArgs
	if len(args) == 1 {
		path := args[0]
		if info, err := os.Stat(path); err == nil && info.Mode()&0111 == 0 {
			Audch.ProgramArgs = nil
			if err := agent.LoadConfig(path, &Audch.Config); err != nil {
				log.Fatalf("[webproc] load config error: %s", err)
			}
		}
	}
	//validate and apply defaults
	if err := agent.ValidateConfig(&Audch.Config); err != nil {
		log.Fatalf("[webproc] load config error: %s", err)
	}
	//server listener
	if err := agent.Run(versionData, Audch.Config); err != nil {
		log.Fatalf("[webproc] agent error: %s", err)
	}
}

func main() {
	audchServer() // Run once when starting up
	c := cron.New()
	if _, err = c.AddFunc("*/5 * * * *", func() {
		audchServer()
	}); err != nil {
		log.Errorf("Cron Add Error: [%v]", err)
	}
	log.Infof("Cron Start")

	//启动定时器
	if Audch.EnableDnsmasq {
		c.Start()
		dnsmasqServer()
	} else {
		c.Run()
	}
}
