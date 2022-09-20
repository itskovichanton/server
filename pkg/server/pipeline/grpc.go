package pipeline

import (
	"context"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/httputils"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"strconv"
)

func (c *EntityFromGRPCReaderServiceImpl) ReadCaller(md metadata.MD, peerInfo *peer.Peer) *core.Caller {

	r := &core.Caller{
		IP:       peerInfo.Addr.String(),
		Version:  c.readVersion(md),
		Type:     c.readCallerType(md),
		Language: c.readLanguage(md),
		AuthArgs: c.readAuthArgs(md),
	}

	if r.AuthArgs != nil {
		r.Session = &core.Session{
			Token: r.AuthArgs.SessionToken,
		}
		r.Session.Account = &core.Account{
			Username:     r.AuthArgs.Username,
			Lang:         r.Language,
			SessionToken: r.AuthArgs.SessionToken,
			Role:         core.RoleUser,
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

func (c *EntityFromGRPCReaderServiceImpl) readAuthArgs(r metadata.MD) *core.AuthArgs {

	sessionToken := utils.GetFirstElementStr(r.Get("sessiontoken"))
	username, password, authOK := httputils.ParseBasicAuth(utils.GetFirstElementStr(r.Get("authorization")))

	if !authOK && len(sessionToken) == 0 {
		return nil
	}

	return &core.AuthArgs{
		Username:     username,
		Password:     password,
		SessionToken: sessionToken,
	}

}

type IEntityFromGRPCReaderService interface {
	ReadCallParams(r context.Context) (*core.CallParams, error)
}

type EntityFromGRPCReaderServiceImpl struct {
	IEntityFromGRPCReaderService

	Config *core.Config
}

func (c *EntityFromGRPCReaderServiceImpl) readLanguage(r metadata.MD) string {
	lang := utils.GetFirstElementStr(r.Get("lang"))
	if len(lang) == 0 {
		lang = c.Config.Actions.DefaultLang
	}
	return lang
}

func (c *EntityFromGRPCReaderServiceImpl) readVersion(r metadata.MD) *core.Version {
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
	return &core.Version{
		Code: code,
		Name: name,
	}
}

func (c *EntityFromGRPCReaderServiceImpl) ReadCallParams(ctx context.Context) (*core.CallParams, error) {

	peerInfo, _ := peer.FromContext(ctx)
	md, _ := metadata.FromIncomingContext(ctx)

	return &core.CallParams{
		Request: ctx,
		URL:     peerInfo.Addr.String(),
		Caller:  c.ReadCaller(md, peerInfo),
	}, nil
}
