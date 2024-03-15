package main

import (
	"fmt"
	"path/filepath"
	"ssh-forwarding/internal"
	"ssh-forwarding/internal/config"
	"ssh-forwarding/internal/logging"
	"sync"
)

func main() {
	logger := logging.NewLogger("Forwarding", logging.DEBUG)
	confFile, err := config.SearchConfigYaml()
	if err != nil {
		logger.Error("Start failed, err = %v", err)
		return
	}
	conf, err := config.FromYaml(confFile)
	if err != nil {
		logger.Error("Load config failed, err = %v", err)
	}
	p, _ := filepath.Abs(confFile)
	logger.Info(fmt.Sprintf("Load config file [%s] success", p))
	connected := make([]internal.SshServer, 0)
	for _, server := range conf.SshServers {
		err := server.Connect()
		if err != nil {
			logger.Error("Connect to ssh server [%-21s] failed, err = %v", server.GetAddr(), err)
			continue
		}
		logger.Info("Connect to ssh server [%-21s] success", server.GetAddr())
		connected = append(connected, server)
	}
	if len(connected) > 0 {
		var wg sync.WaitGroup
		wg.Add(1)
		for _, c := range connected {
			for _, forwarding := range c.Forwardings {
				if len(forwarding.LocalHost) == 0 {
					forwarding.LocalHost = "0.0.0.0"
				}
				server := c
				f := forwarding
				go func(forwarding internal.Forwarding) {
					err := forwarding.ListenLocal(server)
					if err != nil {
						logger.Info("[%-25s] Listen local address [%-13s] failed, err = %v", forwarding.Label, forwarding.GetLocalAddr(), err)
					}
				}(f)
			}
		}
		wg.Wait()
	} else {
		logger.Info("No effective forwarding config found")
	}
}
