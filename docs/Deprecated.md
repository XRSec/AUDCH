## Deprecated Function

### CheckHostnameV1

```go
// Deprecated: CheckHostnameV1 commit container to image
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
```

### CheckHostnameV2

```go
// Deprecated: CheckHostnameV2 delete container and create new container
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
```

### CheckNetWork

```go
// Deprecated: Use ConnectBridgeNetWork instead.
func (AUDCHAPP) CheckNetWork() {
	if h1 := AUDCH.HostNowData.HostConfig.NetworkMode; h1 == "host" || strings.Contains(h1, "default") { // TODO 修复容器网络模式  二选一
		return
	}
	if dataType, _ := json.Marshal(AUDCH.HostNowData.NetworkSettings.Networks); !strings.Contains(string(dataType), AUDCH.BridgeNetworkID) {
		err := AUDCH.Cli.NetworkConnect(context.Background(), AUDCH.BridgeNetworkID, AUDCH.HostNowData.ID, nil)
		if err != nil {
			log.Errorf("Unable connect to bridge network, name: %v, error: %v", AUDCH.ReturnName(), err)
			return
		}
	}
}
```


### dnsmasqServer

```go
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
func (AudchApp) DnsmasqRestart() {
    if !Audch.EnableDnsmasq {
        return
    }
    if Audch.Port == 0 {
        Audch.Port = 80
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
```

### DNS

```go
if IPAddress != "" {
    rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
    if err == nil {
        r.Answer = append(r.Answer, rr)
    }
}
```

#### dnsServer

```go
func dnsServer() {
	dns.HandleFunc("docker.shared.", func(w dns.ResponseWriter, m *dns.Msg) {
		r := new(dns.Msg)
		r.SetReply(m)
		r.Compress = true
		r.Authoritative = true
		r.RecursionAvailable = true
		IPAddress := "NULL"
		Domain := strings.TrimRight(r.Question[0].Name, ".")

		defer func(w dns.ResponseWriter, msg *dns.Msg) {
			log.Infof("Query: [%v ? %s --> %v]\n", w.RemoteAddr(), Domain, IPAddress)
			err := w.WriteMsg(msg)
			if err != nil {
				log.Errorf("Error writing response: %v", err)
				return
			}
		}(w, r)

		// not resolve
		if r.Question[0].Qtype != dns.TypeA {
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
```

#### dnsServerV2

```go
func dnsServerV2() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Errorf("Failed to setup the udp server: %s\n", err.Error())
		return
	}
	defer func(udpServer *net.UDPConn) {
		err := udpServer.Close()
		if err != nil {
			log.Errorf("Failed to close the udp server: %s\n", err.Error())
			return
		}
	}(conn)
	for {
		buf := make([]byte, 512)
		_, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Errorf("Failed to read from udp server: %s\n", err.Error())
			continue
		}
		var msg dnsmessage.Message
		if err := msg.Unpack(buf); err != nil {
			fmt.Println(err)
			continue
		}

		if !strings.Contains(msg.Questions[0].Name.String(), "docker.shared") {
			return
		}

		func(addr *net.UDPAddr, conn *net.UDPConn, m dnsmessage.Message) {
			var (
				packed    []byte
				IPAddress = "NULL"
				question  = m.Questions[0]
				queryName = strings.TrimRight(question.Name.String(), ".")
			)

			defer func() {
				log.Infof("Query: [%v ? %s --> %v]\n", conn.RemoteAddr(), queryName, IPAddress)

				if len(m.Answers) == 0 {
					m.RCode = dnsmessage.RCodeNameError
				}

				packed, err = m.Pack()
				if err != nil {
					log.Errorf("Failed to pack the dns message: %s\n", err.Error())
					return
				}

				if _, err = conn.WriteToUDP(packed, addr); err != nil {
					log.Errorf("Failed to write to udp server: %v\n", err)
					return
				}
			}()

			if len(m.Questions) == 0 {
				m.RCode = dnsmessage.RCodeServerFailure
				return
			}

			if !strings.Contains(queryName, "docker.shared") {
				m.RCode = dnsmessage.RCodeNameError
				return
			}

			if question.Type != dnsmessage.TypeA {
				m.RCode = dnsmessage.RCodeNotImplemented
				return
			}
			if k, ok := Audch.HostNow[queryName]; !ok {
				return
			} else {
				k1 := net.ParseIP(k)
				IPAddress = k1.String()
				k2 := k1.To4()
				answer := dnsmessage.Resource{
					Header: dnsmessage.ResourceHeader{
						Name:  question.Name,
						Type:  question.Type,
						Class: dnsmessage.ClassINET,
						TTL:   600,
					},
					Body: &dnsmessage.AResource{
						A: [4]byte{k2[0], k2[1], k2[2], k2[3]}},
				}
				m.Answers = append(m.Answers, answer)
				m.Response = true
			}
		}(addr, conn, msg)
		//go ServerDNS(addr, conn, msg)
	}
}
```

#### dnsServerV3

```go
func (d AudchApp) ServeDNS(respWriter dns.ResponseWriter, messageFromClient *dns.Msg) {
	var (
		domain      string
		respMessage dns.Msg
		address     string
	)

	respMessage = dns.Msg{}
	respMessage.SetReply(messageFromClient)

	domain = strings.ToLower(messageFromClient.Question[0].Name)

	switch messageFromClient.Question[0].Qtype {
	case dns.TypeA:
		respMessage.Authoritative = true
		address = Audch.HostNow[strings.TrimRight(domain, ".")]
		if address != "" {
			respMessage.Answer = append(respMessage.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
			respMessage.SetRcode(messageFromClient, dns.RcodeSuccess)
			_ = respWriter.WriteMsg(&respMessage)
		}
	}
	log.Infof(domain)
}

func dnsServerV3() {
	srv := &dns.Server{Addr: ":53", Net: "udp"}
	srv.Handler = AudchApp{}
	log.Infof("starting %s listener on %v", srv.Net, srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %v\n", err)
	}
}
```

#### dnsServerV4

```go
func dnsServerV4() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Errorf("Failed to setup the udp server: %s\n", err.Error())
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("Failed to close the udp server: %v\n", err)
		}
	}(conn)
	for {
		buf := make([]byte, 512)
		_, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		var m dnsmessage.Message
		err = m.Unpack(buf)

		if err != nil {
			log.Println(err)
			continue
		}

		if len(m.Questions) > 0 {
			question := m.Questions[0]

			log.Println("Name: ", question.Name)
			log.Println("Type: ", question.Type)

			answer := dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:   question.Name,
					Type:   question.Type,
					Class:  question.Class,
					TTL:    0,
					Length: 0,
				},
				Body: &dnsmessage.AResource{A: [4]byte{192, 168, 0, 1}},
			}
			m.RCode = dnsmessage.RCodeSuccess
			m.Answers = []dnsmessage.Resource{answer}
			m.Response = true
			packed, _ := m.Pack()
			go conn.WriteToUDP(packed, addr)
		}
	}
}
```

