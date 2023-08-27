package trojanx

import (
	"context"
	"github.com/chwjbn/trojanx/internal/pipe"
	"github.com/chwjbn/trojanx/metadata"
	"github.com/chwjbn/trojanx/protocol"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

type (
	ConnectHandler        = func(ctx context.Context) bool
	AuthenticationHandler = func(ctx context.Context, hash string) bool
	RequestHandler        = func(ctx context.Context, request protocol.Request) bool
	ForwardHandler        = func(ctx context.Context, hash string, request protocol.Request) error
	ErrorHandler          = func(ctx context.Context, err error)
)

func DefaultConnectHandler(ctx context.Context) bool {
	return true
}

func DefaultAuthenticationHandler(ctx context.Context, hash string) bool {
	return false
}

func DefaultRequestHandler(ctx context.Context, request protocol.Request) bool {
	var remoteIP net.IP
	if request.AddressType == protocol.AddressTypeDomain {
		tcpAddr, err := net.ResolveTCPAddr("tcp", request.DescriptionAddress)
		if err != nil {
			logrus.Errorln(err)
			return false
		}
		remoteIP = tcpAddr.IP
	} else {
		remoteIP = net.ParseIP(request.DescriptionAddress)
	}
	if remoteIP.IsLoopback() || remoteIP.IsLinkLocalUnicast() || remoteIP.IsLinkLocalMulticast() || remoteIP.IsPrivate() {
		return false
	}
	return true
}

func DefaultForwardHandler(ctx context.Context, hash string, request protocol.Request) error  {

	xMetadata:=metadata.FromContext(ctx)

	dst, err := net.Dial("tcp", net.JoinHostPort(request.DescriptionAddress, strconv.Itoa(request.DescriptionPort)))
	if err != nil {
		return err
	}
	defer dst.Close()

	go pipe.Copy(dst, xMetadata.SrcConn)
	pipe.Copy(xMetadata.SrcConn, dst)

	return err
}

func DefaultErrorHandler(ctx context.Context, err error) {
	logrus.Errorln(err)
}
