package main

import (
	"fmt"
	"ssh-forwarding/internal/config"
	"ssh-forwarding/internal/forwarding"
	"ssh-forwarding/internal/logging"
	"sync"
)

func main() {
	file, err := config.Load()
	if err != nil {
		fmt.Printf("配置文件加载异常 => %v\n", err)
		return
	}
	logger := logging.NewLogger("Forwarding", config.Configuration.Logging.Level)
	logger.Info(fmt.Sprintf("配置文件加载成功 => [%s]", file))
	forwarding.ResolveMaxAddrLength()
	var wg sync.WaitGroup
	for _, server := range config.Configuration.SshServers {
		for _, f := range server.Forwardings {
			wg.Add(1)
			if len(f.LocalHost) == 0 {
				f.LocalHost = "0.0.0.0"
			}
			holder := config.ForwardingHolder{}
			go func(server config.SshServer, f config.Forwarding) {
				if err := forwarding.ListenLocal(server, f, &holder); err != nil {
					logger.Info("[%-25s] 监听失败 => [%-13s]， %v", f.Label, f.GetLocalAddr(), err)
					wg.Done()
				}
			}(server, f)
		}
	}
	wg.Wait()
}
