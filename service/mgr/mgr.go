package mgr

import (
	"time"

	"github.com/jj-io/jj/service"

	"github.com/chzyer/reflag"
	"gopkg.in/logex.v1"
)

var Name = "mgr"

type Config struct {
	Listen       string        `flag:"def=:8682;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type MgrService struct {
	*Config
}

func NewMgrService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &MgrService{
		Config: &c,
	}
}

func (a *MgrService) Name() string { return Name }

func (a *MgrService) Run() error {
	logex.Infof("[mgr] listen on %v", a.Listen)
	return nil
}