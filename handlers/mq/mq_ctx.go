package mq

import (
	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcmux"
	"gopkg.in/logex.v1"
)

type MsgParams struct {
	Topic   string `json:"topic"`
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

type Context struct {
	mux *rpcmux.ClientMux
	*mq.MqClient
}

func NewContext(mux *rpcmux.ClientMux) rpc.Context {
	ctx := &Context{
		MqClient: mq.NewMqClient(mqobj),
		mux:      mux,
	}
	go ctx.respLoop()
	return ctx
}

func (c *Context) Close() {}

func (c *Context) callback(p *rpc.Packet) {
	logex.Info("mqClient receive", p)
}

func (c *Context) respLoop() {
	var msg *mq.Msg
	var err error
	for {
		select {
		case msg = <-c.RespChan:
		case <-c.StopChan:
			return
		}

		logex.Info("ctx receive msg:", msg)
		err = c.mux.SendAsync(rpc.NewReqPacket(PathMsg, msg), c.callback)
		if err != nil {
			logex.Error(err)
		}
	}
}

func (c *Context) Stop() {
	c.MqClient.Stop()
}
