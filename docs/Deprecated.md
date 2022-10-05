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