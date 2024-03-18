package config

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"path"
	"ssh-forwarding/internal/util"
	"strconv"
)

type Config struct {
	SshServers []SshServer `yaml:"ssh_servers"`
}

type SshServer struct {
	Host        string       `yaml:"host"`
	Port        int          `yaml:"port"`
	Username    string       `yaml:"username"`
	Password    string       `yaml:"password"`
	Forwardings []Forwarding `yaml:"forwardings"`
}

func (server *SshServer) GetAddr() string {
	return server.Host + ":" + strconv.Itoa(server.Port)
}

type Forwarding struct {
	Label      string `yaml:"label"`
	LocalHost  string `yaml:"local_host"`
	LocalPort  int    `yaml:"local_port"`
	RemoteHost string `yaml:"remote_host"`
	RemotePort int    `yaml:"remote_port"`
}

func (forwarding *Forwarding) GetLocalAddr() string {
	return forwarding.LocalHost + ":" + strconv.Itoa(forwarding.LocalPort)
}
func (forwarding *Forwarding) GetRemoteAddr() string {
	return forwarding.RemoteHost + ":" + strconv.Itoa(forwarding.RemotePort)
}

type ForwardingHolder struct {
	LocalListener *net.Listener
	Mappings      []*Mapping
}

type Mapping struct {
	Local     net.Conn
	Remote    net.Conn
	SshClient *ssh.Client
	Listener  net.Listener
}

func FromYaml(file string) (*Config, error) {
	conf := &Config{}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, conf)
	return conf, err
}

func SearchConfigYaml() (string, error) {
	// 当前执行目录
	var dirs = []string{""}
	// 程序所在目录
	execPath, err := os.Executable()
	if err == nil {
		dirs = append(dirs, path.Dir(execPath))
	}
	// 用户家目录
	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, path.Join(home, ".ssh-forwarding"))
	}
	for _, dir := range dirs {
		p, r := searchConfigYamlInDir(dir)
		if r {
			return p, nil
		}
	}
	return "", errors.New("config file not found")
}

func searchConfigYamlInDir(dir string) (string, bool) {
	var names = [...]string{"forwarding.yaml", "forwarding.yml"}
	for _, name := range names {
		if util.FileExists(path.Join(dir, name)) {
			return path.Join(dir, name), true
		}
	}
	return "", false
}
