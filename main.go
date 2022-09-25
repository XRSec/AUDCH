package main

import (
	"flag"
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
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
	c := cron.New()
	if _, err = c.AddFunc("*/5 * * * *", func() {
		clientDocker()
		// hostNow
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
	}); err != nil {
		log.Errorf("Cron Add Error: [%v]", err)
	}
	log.Infof("Cron Start")
	//启动定时器
	c.Run()
}
