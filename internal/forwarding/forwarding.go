package forwarding

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"ssh-forwarding/internal/config"
	"ssh-forwarding/internal/logging"
	"strconv"
	"time"
)

var MaxForwardingLabelLength int = 0
var MaxSshServerAddrLength int = 0
var MaxRemoteAddrLength int = 0

func ListenLocal(server config.SshServer, forwarding config.Forwarding, holder *config.ForwardingHolder) error {
	var logger = logging.NewLogger("Forwarding", config.Configuration.Logging.Level)

	listener, err := net.Listen("tcp", forwarding.GetLocalAddr())
	if err != nil {
		return err
	}
	logger.Info("[%-"+strconv.Itoa(MaxForwardingLabelLength)+"s] 监听成功 => [%-13s]", forwarding.Label, forwarding.GetLocalAddr())
	defer func() {
		_ = listener.Close()
	}()
	holder.LocalListener = &listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.New(fmt.Sprintf("[%-"+strconv.Itoa(MaxForwardingLabelLength)+"s] 连接异常 => %v", forwarding.Label, err))
		}
		mapping := &config.Mapping{Local: conn}
		holder.Mappings = append(holder.Mappings, mapping)
		go func() {
			defer closeMapping(holder, mapping)
			mappingFormat := getMappingFormat()
			sshClient, remote, err := createMapping(server, forwarding)
			if err != nil {
				logger.Error("[%-"+strconv.Itoa(MaxForwardingLabelLength)+"s] 映射异常 => "+mappingFormat+"， %v", forwarding.Label, conn.RemoteAddr(), forwarding.GetLocalAddr(), server.GetAddr(), forwarding.GetRemoteAddr(), err)
				return
			}
			logger.Info("[%-"+strconv.Itoa(MaxForwardingLabelLength)+"s] 映射成功 => "+mappingFormat, forwarding.Label, conn.RemoteAddr(), forwarding.GetLocalAddr(), server.GetAddr(), forwarding.GetRemoteAddr())
			mapping.SshClient = sshClient
			mapping.Remote = remote
			go func() {
				_, _ = io.Copy(remote, conn)
			}()
			_, _ = io.Copy(conn, remote)
			logger.Warn("[%-"+strconv.Itoa(MaxForwardingLabelLength)+"s] 映射关闭 => "+mappingFormat, forwarding.Label, conn.RemoteAddr(), forwarding.GetLocalAddr(), server.GetAddr(), forwarding.GetRemoteAddr())
		}()
	}
}

func createMapping(server config.SshServer, forwarding config.Forwarding) (*ssh.Client, net.Conn, error) {
	sshClient, err := ConnectSshServer(server)
	if err != nil {
		return nil, nil, err
	}
	remote, err := sshClient.Dial("tcp", forwarding.GetRemoteAddr())
	if err != nil {
		return nil, nil, err
	}
	return sshClient, remote, nil
}

func closeMapping(holder *config.ForwardingHolder, mapping *config.Mapping) {
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
}

func ConnectSshServer(server config.SshServer) (*ssh.Client, error) {
	return ssh.Dial("tcp", server.GetAddr(), &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(server.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
}

func getMappingFormat() string {
	return "[%-15s] <=> [%-13s] <=> [%-" + strconv.Itoa(MaxSshServerAddrLength) + "s] <=> [%-" + strconv.Itoa(MaxRemoteAddrLength) + "s]"
}

func ResolveMaxAddrLength() {
	for _, server := range config.Configuration.SshServers {
		if len(server.GetAddr()) > MaxSshServerAddrLength {
			MaxSshServerAddrLength = len(server.GetAddr())
		}
		for _, forwarding := range server.Forwardings {
			if len(forwarding.GetRemoteAddr()) > MaxRemoteAddrLength {
				MaxRemoteAddrLength = len(forwarding.GetRemoteAddr())
			}
			if len(forwarding.Label) > MaxForwardingLabelLength {
				MaxForwardingLabelLength = len(forwarding.Label)
			}
		}
	}
}
