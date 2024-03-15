package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"ssh-forwarding/internal"
	"ssh-forwarding/internal/util"
)

type Config struct {
	SshServers []internal.SshServer `yaml:"ssh_servers"`
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
