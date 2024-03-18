package main

import (
	"fmt"
	"path/filepath"
	"ssh-forwarding/internal/config"
	"ssh-forwarding/internal/forwarding"
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
	var wg sync.WaitGroup
	for _, server := range conf.SshServers {
		for _, f := range server.Forwardings {
			wg.Add(1)
			if len(f.LocalHost) == 0 {
				f.LocalHost = "0.0.0.0"
			}
			holder := config.ForwardingHolder{}
			go func(server config.SshServer, f config.Forwarding) {
				if err := forwarding.ListenLocal(server, f, &holder); err != nil {
					logger.Info("[%-25s] Listen local address [%-13s] failed, err = %v", f.Label, f.GetLocalAddr(), err)
					wg.Done()
				}
			}(server, f)
		}
	}
	wg.Wait()
}
