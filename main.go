package main

import (
	"flag"
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

type App map[string]string

var (
	version                                  = flag.Bool("v", false, "show version")
	buildTime, commitId, versionData, author string
	Cli                                      *client.Client
	err                                      error
	defaultHostsFile                         = flag.String("f", "/etc/hosts", "hosts filepath")
	hostNow, hostLast                        App      // 集合
	hostBytes                                []byte   // hosts Byte 内容
	hostStr                                  string   // hosts 字符串内容
	hostAll                                  []string // 数组
	bridgeID                                 string
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
}

func clientDocker() {
	Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Errorf("Unable to connect to Docker: %v", err)
		os.Exit(1)
	}
	log.Infoln("Successfully connected to Docker")
}

func main() {
	//c := cron.New()
	//if _, err = c.AddFunc("*/5 * * * *", func() {
	clientDocker()
	bridgeID = getBridge()

	if bridgeID == "" {
		log.Errorln("bridgeID is empty")
		return
	}
	// hostNow
	//err := Cli.NetworkConnect(context.Background(), "", "", c)
	list, err := Cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Errorf("Unable to list containers: %v", err)
		return
	}
	hostNow = make(App)
	for i := 0; i < len(list); i++ {
		name := strings.Replace(list[i].Names[len(list[i].Names)-1], "/", "", -1) + ".docker.shared" // TODO name: ["/adminer/db", "mysql"] *docker run --link
		if name == "audch.docker.shared" {
			continue
		}
		for _, v := range list[i].NetworkSettings.Networks {
			if v.IPAddress != "" {
				hostNow[name] = v.IPAddress
			}
			break // 只记录第一个IP
		}
		//if dataType, _ := json.Marshal(list[i].NetworkSettings.Networks); !strings.Contains(string(dataType), bridgeID) {
		//	updateNetwork(list[i].ID)
		//}
		if list[i].HostConfig.NetworkMode != "default" { // TODO 修复容器网络模式
			updateNetwork(list[i].ID)
		}
		updateHostName(list[i])
	}
	// hostBytes
	hostBytes, err = ioutil.ReadFile(*defaultHostsFile)
	if err != nil {
		log.Errorf("Failed to read %v: %v", defaultHostsFile, err)
		return
	}

	// hostAll
	hostAll = strings.Split(string(hostBytes), "\n")

	// hostLast
	hostLast = make(App)
	//for k, v := range hostAll {
	for i := 0; i < len(hostAll); i++ {
		if strings.Contains(hostAll[i], "# AUDCH") {
			row := strings.Split(hostAll[i], "\t")
			if len(row) != 3 {
				return
			}
			name := strings.Replace(row[1], " ", "", -1)
			ipaddress := strings.Replace(row[0], " ", "", -1)
			hostLast[name] = ipaddress
			if !strings.Contains(strings.Replace(row[1], " ", "", -1), "docker.shared") {
				log.Errorf("alias not match, please check: %v", name)
			}
			hostAll = append(hostAll[:i], hostAll[i+1:]...)
			i--
		}
		if i != 0 && hostAll[i] == hostAll[i-1] && hostAll[i] == "" {
			hostAll = append(hostAll[:i], hostAll[i+1:]...)
			i--
		}
	}

	// diff
	del := 0
	update := 0
	for k, v := range hostLast {
		if h, ok := hostNow[k]; !ok {
			log.Infof("Del: %v %v", k, v)
			del++
		} else {
			if h != v {
				log.Infof("Update: [%v %v] -> %v", k, v, h)
				update++
			}
		}
	}

	if len(hostLast) != 0 && del == 0 && update == 0 {
		log.Infoln("Nothing to update")
		return
	}
	log.Infof("Last %v、Del: %v、Update: %v records", len(hostLast), del, len(hostNow)-update)

	for k, v := range hostNow {
		hostAll = append(hostAll, v+"\t"+k+"\t# AUDCH")
	}

	//hostStr
	hostStr = ""
	for _, v := range hostAll {
		hostStr += v + "\n"
	}
	err = ioutil.WriteFile(*defaultHostsFile, []byte(hostStr), 0644)
	if err != nil {
		log.Errorf("Failed to write %v, %v", *defaultHostsFile, err)
		return
	}
	//}); err != nil {
	//	log.Errorf("Cron Add Error: [%v]", err)
	//}
	//log.Infof("Cron Start")
	////启动定时器
	//c.Run()
}

func updateHostName(containers types.Container) {
	name := strings.Replace(containers.Names[len(containers.Names)-1], "/", "", -1)
	inspect, err := Cli.ContainerInspect(context.Background(), containers.ID)
	if err != nil {
		log.Errorf("ContainerInspect Error: [%v]", err)
		return
	}

	config := inspect.Config
	if config.Hostname == name {
		return
	}
	config.Hostname = name + ".docker.shared"
	commit, err := Cli.ContainerCommit(context.Background(), containers.ID, types.ContainerCommitOptions{
		Comment:   "AUDCH Update hostname",
		Pause:     true,
		Changes:   []string{`ENTRYPOINT ["true"]`},
		Reference: inspect.Config.Image,
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

	err = Cli.ContainerRemove(context.Background(), containers.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Errorf("ContainerRemove Error: [%v]", err)
		return
	}

	create, err := Cli.ContainerCreate(context.Background(), config, inspect.HostConfig, simpleNetworkConfig, nil, containers.Names[len(containers.Names)-1])
	if err != nil {
		log.Errorf("ContainerCreate Error: [%v]", err)
		return
	}

	err = Cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Errorf("ContainerStart Error: [%v]", err)
		return
	}
	log.Infof("Update hostname: %v", name)
}

func getBridge() string {
	networkList, err := Cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return ""
	}
	for _, v := range networkList {
		if v.Name == "bridge" {
			return v.ID
		}
	}
	return ""
}

func updateNetwork(containerID string) {
	err := Cli.NetworkConnect(context.Background(), bridgeID, containerID, nil)
	if err != nil {
		log.Errorf("Unable to connect to Docker: %v", err)
		return
	}
}
