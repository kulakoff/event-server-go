package draft

import (
	"github.com/kulakoff/event-server-go/internal/utils"
	"net"
)

type BaseService struct {
	ClickhouseClient *utils.ClickhouseClient
}

func New(clickhouseClient *utils.ClickhouseClient) *BaseService {
	return &BaseService{
		ClickhouseClient: clickhouseClient,
	}
}

func (b *BaseService) ProcessSyslogMessage(message string, addr *net.UDPAddr) error {
	return nil
}
