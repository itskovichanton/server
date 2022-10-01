package pipeline

import (
	"context"
	"github.com/itskovichanton/goava/pkg/goava/httputils"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server"
	"github.com/itskovichanton/server/pkg/server/entities"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"strconv"
)

func (c *EntityFromGRPCReaderServiceImpl) ReadCaller(md metadata.MD, peerInfo *peer.Peer) *entities.Caller {

	r := &entities.Caller{
		IP:       peerInfo.Addr.String(),
		Version:  c.readVersion(md),
		Type:     c.readCallerType(md),
		Language: c.readLanguage(md),
		AuthArgs: c.readAuthArgs(md),
	}

	if r.AuthArgs != nil {
		r.Session = &entities.Session{
			Token: r.AuthArgs.SessionToken,
		}
		r.Session.Account = &entities.Account{
			Username:     r.AuthArgs.Username,
			Lang:         r.Language,
			SessionToken: r.AuthArgs.SessionToken,
			Role:         entities.RoleUser,
			Password:     r.AuthArgs.Password,
			IP:           r.IP,
		}
	}
	return r
}

func (c *EntityFromGRPCReaderServiceImpl) readCallerType(r metadata.MD) string {
	t := r.Get("caller-type")
	if len(t) == 0 {
		t = r.Get("user-agent")
	}
	return utils.GetFirstElementStr(t)
}

func (c *EntityFromGRPCReaderServiceImpl) readAuthArgs(r metadata.MD) *entities.AuthArgs {

	sessionToken := utils.GetFirstElementStr(r.Get("sessiontoken"))
	username, password, authOK := httputils.ParseBasicAuth(utils.GetFirstElementStr(r.Get("authorization")))

	if !authOK && len(sessionToken) == 0 {
		return nil
	}

	return &entities.AuthArgs{
		Username:     username,
		Password:     password,
		SessionToken: sessionToken,
	}

}

type IEntityFromGRPCReaderService interface {
	ReadCallParams(r context.Context) (*entities.CallParams, error)
}

type EntityFromGRPCReaderServiceImpl struct {
	IEntityFromGRPCReaderService

	Config *server.Config
}

func (c *EntityFromGRPCReaderServiceImpl) readLanguage(r metadata.MD) string {
	lang := utils.GetFirstElementStr(r.Get("lang"))
	if len(lang) == 0 {
		lang = c.Config.Server.DefaultLang
	}
	return lang
}

func (c *EntityFromGRPCReaderServiceImpl) readVersion(r metadata.MD) *entities.Version {
	vc := r.Get("caller-version-code")
	if len(vc) == 0 || len(vc[0]) == 0 {
		return nil
	}

	code, _ := strconv.Atoi(vc[0])
	name := ""
	nameFromCtx := r.Get("caller-version-name")
	if len(nameFromCtx) != 0 {
		name = nameFromCtx[0]
	}
	return &entities.Version{
		Code: code,
		Name: name,
	}
}

func (c *EntityFromGRPCReaderServiceImpl) ReadCallParams(ctx context.Context) (*entities.CallParams, error) {

	peerInfo, _ := peer.FromContext(ctx)
	md, _ := metadata.FromIncomingContext(ctx)

	return &entities.CallParams{
		Request: ctx,
		URL:     peerInfo.Addr.String(),
		Caller:  c.ReadCaller(md, peerInfo),
	}, nil
}
