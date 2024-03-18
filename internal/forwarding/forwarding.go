package forwarding

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"ssh-forwarding/internal/config"
	"ssh-forwarding/internal/logging"
	"time"
)

var logger = logging.NewLogger("Forwarding", logging.DEBUG)

func ListenLocal(server config.SshServer, forwarding config.Forwarding, holder *config.ForwardingHolder) error {
	listener, err := net.Listen("tcp", forwarding.GetLocalAddr())
	if err != nil {
		return err
	}
	logger.Info("[%-25s] Local server listen [%-13s] success", forwarding.Label, forwarding.GetLocalAddr())
	defer func() {
		_ = listener.Close()
	}()
	holder.LocalListener = &listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.New(fmt.Sprintf("[%-25s] Accept connection failed, err = %v", forwarding.Label, err))
		}
		logger.Debug("[%-25s] Local server connected [%-21s] <=> [%-13s]", forwarding.Label, conn.RemoteAddr(), conn.LocalAddr())
		mapping := &config.Mapping{Local: conn}
		holder.Mappings = append(holder.Mappings, mapping)
		go func() {
			defer func() {
				logger.Debug("[%-25s] Disconnect & Cleanup", forwarding.Label)
				newMappings := make([]*config.Mapping, 0)
				for _, m := range holder.Mappings {
					if mapping != m {
						newMappings = append(newMappings, mapping)
					} else {
						_ = m.Local.Close()
						if m.Remote != nil {
							_ = m.Remote.Close()
						}
						if m.SshClient != nil {
							_ = m.SshClient.Close()
						}
					}
				}
				holder.Mappings = newMappings
			}()
			sshClient, err := ConnectSshServer(server)
			if err != nil {
				logger.Error("[%-25s] Ssh Server connect [%-21s] failed, err = %v", forwarding.Label, forwarding.GetRemoteAddr(), err)
				return
			}
			logger.Debug("[%-25s] Ssh Server connected [%-13s] <=> [%-21s]", forwarding.Label, sshClient.LocalAddr(), sshClient.RemoteAddr())
			mapping.SshClient = sshClient
			remote, err := sshClient.Dial("tcp", forwarding.GetRemoteAddr())
			if err != nil {
				logger.Error("[%-25s] Remote server connect [%-21s] via ssh server [%-21s] failed, err = %v", forwarding.Label, forwarding.GetRemoteAddr(), server.GetAddr(), err)
				return
			}
			logger.Debug("[%-25s] Remote server connected [%-21s] <=> [%-13s] <=> [%-21s]", forwarding.Label, conn.RemoteAddr(), conn.LocalAddr(), forwarding.GetRemoteAddr())
			mapping.Remote = remote
			go func() {
				_, _ = io.Copy(remote, conn)
			}()
			_, _ = io.Copy(conn, remote)
			logger.Warn("[%-25s] Mapping closed [%-21s] <=> [%-21s]", forwarding.Label, conn.RemoteAddr(), forwarding.GetRemoteAddr())
		}()
	}
}

func ConnectSshServer(server config.SshServer) (*ssh.Client, error) {
	return ssh.Dial("tcp", server.GetAddr(), &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(server.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
}
