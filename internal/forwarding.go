package internal

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"ssh-forwarding/internal/logging"
	"strconv"
	"time"
)

var logger = logging.NewLogger("Forwarding", logging.DEBUG)

type Mapping struct {
	Local  net.Conn
	Remote net.Conn
}
type Forwarding struct {
	Label      string `yaml:"label"`
	LocalHost  string `yaml:"local_host"`
	LocalPort  int    `yaml:"local_port"`
	RemoteHost string `yaml:"remote_host"`
	RemotePort int    `yaml:"remote_port"`

	LocalListener *net.Listener
	Mappings      []*Mapping
}

func (forwarding *Forwarding) GetLocalAddr() string {
	return forwarding.LocalHost + ":" + strconv.Itoa(forwarding.LocalPort)
}
func (forwarding *Forwarding) GetRemoteAddr() string {
	return forwarding.RemoteHost + ":" + strconv.Itoa(forwarding.RemotePort)
}
func (forwarding *Forwarding) ListenLocal(server SshServer) error {
	listener, err := net.Listen("tcp", forwarding.GetLocalAddr())
	if err != nil {
		return err
	}
	logger.Info("[%-25s] Listen local address [%-13s] success", forwarding.Label, forwarding.GetLocalAddr())
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)
	forwarding.LocalListener = &listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.New(fmt.Sprintf("[%-25s] Accept connection failed, err = %v", forwarding.Label, err))
		}
		logger.Debug("[%-25s] Connected [%-21s] <=> [%-13s]", forwarding.Label, conn.RemoteAddr(), conn.LocalAddr())
		if forwarding.Mappings == nil {
			forwarding.Mappings = make([]*Mapping, 0)
		}
		mapping := &Mapping{Local: conn}
		forwarding.Mappings = append(forwarding.Mappings, mapping)
		go func(local net.Conn) {
			defer func() {
				newMappings := make([]*Mapping, 0)
				for _, m := range forwarding.Mappings {
					if mapping != m {
						newMappings = append(newMappings, mapping)
					} else {
						_ = m.Local.Close()
						if m.Remote != nil {
							_ = m.Remote.Close()
						}
					}
				}
				forwarding.Mappings = newMappings
			}()
			remote, err := server.SshClient.Dial("tcp", forwarding.GetRemoteAddr())
			if err != nil {
				err = server.Connect()
				if err != nil {
					logger.Error("[%-25s] Connect to remote address [%-21s] via ssh server [%-21s] failed, err = %v", forwarding.Label, forwarding.GetRemoteAddr(), server.SshClient.RemoteAddr(), err)
					return
				}
				remote, err = server.SshClient.Dial("tcp", forwarding.GetRemoteAddr())
				if err != nil {
					logger.Error("[%-25s] Connect to remote address [%-21s] via ssh server [%-21s] failed, err = %v", forwarding.Label, forwarding.GetRemoteAddr(), server.SshClient.RemoteAddr(), err)
					return
				}
			}
			mapping.Remote = remote
			go func() {
				_, _ = io.Copy(remote, local)
			}()
			_, _ = io.Copy(local, remote)
			logger.Warn("[%-25s] Mapping closed [%-21s] <=> [%-21s]", forwarding.Label, local.RemoteAddr(), forwarding.GetRemoteAddr())
		}(conn)
	}
}

type SshServer struct {
	Host        string       `yaml:"host"`
	Port        int          `yaml:"port"`
	Username    string       `yaml:"username"`
	Password    string       `yaml:"password"`
	Forwardings []Forwarding `yaml:"forwardings"`

	SshClient *ssh.Client
}

func (server *SshServer) GetAddr() string {
	return server.Host + ":" + strconv.Itoa(server.Port)
}

func (server *SshServer) Connect() error {
	c, e := ssh.Dial("tcp", server.GetAddr(), &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(server.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if e != nil {
		return e
	}
	server.SshClient = c
	return nil
}
