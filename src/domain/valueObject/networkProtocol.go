package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	NetworkProtocolHttp    NetworkProtocol = "http"
	NetworkProtocolHttps   NetworkProtocol = "https"
	NetworkProtocolWs      NetworkProtocol = "ws"
	NetworkProtocolWss     NetworkProtocol = "wss"
	NetworkProtocolGrpc    NetworkProtocol = "grpc"
	NetworkProtocolGrpcs   NetworkProtocol = "grpcs"
	NetworkProtocolTcp     NetworkProtocol = "tcp"
	NetworkProtocolUdp     NetworkProtocol = "udp"
	NetworkProtocolDefault NetworkProtocol = "http"
)

type NetworkProtocol string

func NewNetworkProtocol(value any) (networkProtocol NetworkProtocol, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return networkProtocol, errors.New("NetworkProtocolMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	networkProtocol = NetworkProtocol(stringValue)
	switch networkProtocol {
	case NetworkProtocolHttp, NetworkProtocolHttps, NetworkProtocolWs,
		NetworkProtocolWss, NetworkProtocolGrpc, NetworkProtocolGrpcs,
		NetworkProtocolTcp, NetworkProtocolUdp:
		return networkProtocol, nil
	default:
		return networkProtocol, errors.New("UnknownNetworkProtocol")
	}
}

func (vo NetworkProtocol) String() string {
	return string(vo)
}
