package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type (
	App      map[string]string
	AUDCHAPP struct {
		Recreate          bool
		Cli               *client.Client
		HostNow, HostLast App      // 集合
		HostBytes         []byte   // hosts Byte 内容
		HostStr           string   // hosts 字符串内容
		HostAll           []string // 数组
		Name              string
		HostNowData       types.Container
		BridgeNetwork     types.NetworkResource
	}
)

var (
	version                                  = flag.Bool("v", false, "show version")
	defaultHostsFile                         = flag.String("f", "/etc/hosts", "hosts filepath")
	buildTime, commitId, versionData, author string
	err                                      error
	AUDCH                                    AUDCHAPP
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
	if os.Getenv("recreate") == "true" {
		AUDCH.Recreate = true
	}
}

func (AUDCHAPP) ClientDocker() {
	AUDCH.Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Errorf("Unable to connect to Docker: %v", err)
		os.Exit(1)
	}
	log.Infoln("Successfully connected to Docker")
}

func (AUDCHAPP) GetBridge() {
	networkList, err := AUDCH.Cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Errorf("Unable to get bridge network, error: %v", err)
		return
	}
	for _, v := range networkList {
		if v.Name == "bridge" {
			AUDCH.BridgeNetwork = v
		}
	}
	if AUDCH.BridgeNetwork.ID == "" {
		log.Errorln("bridgeID is empty")
		return
	}
}

// CheckHostnameV1 commit container to image
func (AUDCHAPP) CheckHostnameV1() {
	name := AUDCH.ReturnName()
	inspect, err := AUDCH.Cli.ContainerInspect(context.Background(), AUDCH.HostNowData.ID)
	if err != nil {
		log.Errorf("ContainerInspect Error: [%v]", err)
		return
	}

	config := inspect.Config
	if config.Hostname == name {
		return
	}

	if strings.Contains(config.Image, "sha256") {
		if k, ok := config.Labels["AUDCH_IMAGE"]; ok {
			config.Image = k
		} else {
			log.Infof("AUDCH_IMAGE not found: [%v]", config.Image)
			config.Image = "audch/" + inspect.Name
		}
	}

	timeout := 10 * time.Second
	if AUDCH.Cli.ContainerStop(context.Background(), AUDCH.HostNowData.ID, &timeout) != nil {
		log.Errorf("ContainerStop Error: [%v]", err)
		return
	}

	config.Hostname = name
	commit, err := AUDCH.Cli.ContainerCommit(context.Background(), AUDCH.HostNowData.ID, types.ContainerCommitOptions{
		Comment:   "AUDCH Update hostname",
		Author:    "AUDCH",
		Pause:     true,
		Reference: config.Image,
		Config:    config,
		Changes:   []string{"LABEL AUDCH_IMAGE=" + inspect.Config.Image},
	})
	if err != nil {
		log.Errorf("ContainerCommit Error: [%v]", err)
		return
	}

	config.Image = commit.ID
	simpleNetworkConfig := func() *network.NetworkingConfig {
		oneEndpoint := make(map[string]*network.EndpointSettings)
		networkConfig := &network.NetworkingConfig{EndpointsConfig: inspect.NetworkSettings.Networks}
		for k, v := range networkConfig.EndpointsConfig {
			oneEndpoint[k] = v
			//we only need 1
			//break
		}
		return &network.NetworkingConfig{EndpointsConfig: oneEndpoint}
	}()

	err = AUDCH.Cli.ContainerRemove(context.Background(), AUDCH.HostNowData.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Errorf("ContainerRemove Error: [%v]", err)
		return
	}

	create, err := AUDCH.Cli.ContainerCreate(context.Background(), config, inspect.HostConfig, simpleNetworkConfig, nil, AUDCH.HostNowData.Names[len(AUDCH.HostNowData.Names)-1])
	if err != nil {
		log.Errorf("ContainerCreate Error: [%v]", err)
		return
	}

	err = AUDCH.Cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Errorf("ContainerStart Error: [%v]", err)
		return
	}
	log.Infof("Update hostname: %v", name)
}

// CheckHostnameV2 delete container and create new container
func (AUDCHAPP) CheckHostnameV2() {
	if !AUDCH.Recreate {
		return
	}
	name := AUDCH.ReturnName()
	inspect, err := AUDCH.Cli.ContainerInspect(context.Background(), AUDCH.HostNowData.ID)
	if err != nil {
		log.Errorf("ContainerInspect Error: [%v]", err)
		return
	}

	config := inspect.Config
	if config.Hostname == name {
		return
	}

	if strings.Contains(config.Image, "sha256") {
		if k, ok := config.Labels["AUDCH_IMAGE"]; ok {
			config.Image = k
		} else {
			log.Infof("AUDCH_IMAGE not found: [%v]", config.Image)
			return
		}
	}

	timeout := 10 * time.Second
	if AUDCH.Cli.ContainerStop(context.Background(), AUDCH.HostNowData.ID, &timeout) != nil {
		log.Errorf("ContainerStop Error: [%v]", err)
		return
	}

	config.Hostname = name
	simpleNetworkConfig := func() *network.NetworkingConfig {
		oneEndpoint := make(map[string]*network.EndpointSettings)
		networkConfig := &network.NetworkingConfig{EndpointsConfig: inspect.NetworkSettings.Networks}
		for k, v := range networkConfig.EndpointsConfig {
			oneEndpoint[k] = v
			//we only need 1
			//break
		}
		return &network.NetworkingConfig{EndpointsConfig: oneEndpoint}
	}()

	err = AUDCH.Cli.ContainerRemove(context.Background(), AUDCH.HostNowData.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Errorf("ContainerRemove Error: [%v]", err)
		return
	}

	create, err := AUDCH.Cli.ContainerCreate(context.Background(), config, inspect.HostConfig, simpleNetworkConfig, nil, AUDCH.HostNowData.Names[len(AUDCH.HostNowData.Names)-1])
	if err != nil {
		log.Errorf("ContainerCreate Error: [%v]", err)
		return
	}

	err = AUDCH.Cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Errorf("ContainerStart Error: [%v]", err)
		return
	}
	log.Infof("Update hostname: %v", name)
}

func (AUDCHAPP) ReturnName() string {
	return strings.Replace(AUDCH.HostNowData.Names[len(AUDCH.HostNowData.Names)-1], "/", "", -1) + ".docker.shared" // TODO name: ["/adminer/db", "mysql"] *docker run --link
}

func (AUDCHAPP) GetIPAddress() {
	for _, v := range AUDCH.HostNowData.NetworkSettings.Networks {
		if v.IPAddress != "" {
			AUDCH.HostNow[AUDCH.ReturnName()] = v.IPAddress
		}
		break // 只记录第一个IP
	}
}

func (AUDCHAPP) CheckNetWork() {
	if h1 := AUDCH.HostNowData.HostConfig.NetworkMode; h1 == "host" || strings.Contains(h1, "default") { // TODO 修复容器网络模式  二选一
		return
	}
	if dataType, _ := json.Marshal(AUDCH.HostNowData.NetworkSettings.Networks); !strings.Contains(string(dataType), AUDCH.BridgeNetwork.ID) {
		err := AUDCH.Cli.NetworkConnect(context.Background(), AUDCH.BridgeNetwork.ID, AUDCH.HostNowData.ID, nil)
		if err != nil {
			log.Errorf("Unable connect to bridge network, name: %v, error: %v", AUDCH.ReturnName(), err)
			return
		}
	}

}

func (AUDCHAPP) GetHostNow() {
	// hostNow
	list, err := AUDCH.Cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Errorf("Unable to list containers: %v", err)
		return
	}
	AUDCH.HostNow = make(App)
	for _, v := range list {
		AUDCH.HostNowData = v
		if AUDCH.ReturnName() == "audch.docker.shared" {
			continue
		}
		AUDCH.GetIPAddress()
		AUDCH.CheckNetWork()
		AUDCH.CheckHostnameV2() // 马勒戈壁，[Nginx Does Not Resolve HostName](https://forums.docker.com/t/nginx-does-not-resolve-hostname/115859)
	}
	for i := 0; i < len(list); i++ {
		AUDCH.HostNowData = list[i]
	}
}

func (AUDCHAPP) GetHostBytes() {
	// HostBytes
	AUDCH.HostBytes, err = ioutil.ReadFile(*defaultHostsFile)
	if err != nil {
		log.Errorf("Failed to read %v: %v", defaultHostsFile, err)
		return
	}
}

func (AUDCHAPP) GetHostAll() {
	AUDCH.HostAll = strings.Split(string(AUDCH.HostBytes), "\n")
}

func (AUDCHAPP) GetHostLast() {
	AUDCH.HostLast = make(App)
	for i := 0; i < len(AUDCH.HostAll); i++ {
		if strings.Contains(AUDCH.HostAll[i], "# AUDCH") {
			row := strings.Split(AUDCH.HostAll[i], "\t")
			if len(row) != 3 {
				return
			}
			name := strings.Replace(row[1], " ", "", -1)
			ipaddress := strings.Replace(row[0], " ", "", -1)
			AUDCH.HostLast[name] = ipaddress
			if !strings.Contains(strings.Replace(row[1], " ", "", -1), "docker.shared") {
				log.Errorf("alias not match, please check: %v", name)
			}
			AUDCH.HostAll = append(AUDCH.HostAll[:i], AUDCH.HostAll[i+1:]...)
			i--
		}
		if i > 0 && AUDCH.HostAll[i] == AUDCH.HostAll[i-1] && AUDCH.HostAll[i] == "" {
			AUDCH.HostAll = append(AUDCH.HostAll[:i], AUDCH.HostAll[i+1:]...)
			i--
		}
	}
}

func (AUDCHAPP) GetHostDiff() {
	del := 0
	update := 0
	for k, v := range AUDCH.HostLast {
		if h, ok := AUDCH.HostNow[k]; !ok {
			log.Infof("Del: %v %v", k, v)
			del++
		} else {
			if h != v {
				log.Infof("Update: [%v %v] -> %v", k, v, h)
				update++
			}
		}
	}

	if len(AUDCH.HostLast) != 0 && del == 0 && update == 0 {
		log.Infoln("Nothing to update")
		return
	}
	log.Infof("Last %v、Del: %v、Update: %v records", len(AUDCH.HostLast), del, len(AUDCH.HostNow)-update)

	for k, v := range AUDCH.HostNow {
		AUDCH.HostAll = append(AUDCH.HostAll, v+"\t"+k+"\t# AUDCH")
	}
	AUDCH.GetHostStr()
	AUDCH.HostWrite()
}

func (AUDCHAPP) GetHostStr() {
	AUDCH.HostStr = ""
	for _, v := range AUDCH.HostAll {
		AUDCH.HostStr += v + "\n"
	}
}

func (AUDCHAPP) HostWrite() {
	err = ioutil.WriteFile(*defaultHostsFile, []byte(AUDCH.HostStr), 0644)
	if err != nil {
		log.Errorf("Failed to write %v, %v", *defaultHostsFile, err)
		return
	}
	log.Infof("Write %v success", *defaultHostsFile)
}

func server() {
	AUDCH.ClientDocker()
	AUDCH.GetBridge()
	AUDCH.GetHostNow()
	AUDCH.GetHostBytes()
	AUDCH.GetHostAll()
	AUDCH.GetHostLast()
	AUDCH.GetHostDiff()
}

func main() {
	server() // Run once when starting up
	c := cron.New()
	if _, err = c.AddFunc("*/5 * * * *", func() {
		server()
	}); err != nil {
		log.Errorf("Cron Add Error: [%v]", err)
	}
	log.Infof("Cron Start")
	//启动定时器
	c.Run()
}
