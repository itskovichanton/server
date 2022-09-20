package pipeline

import (
	"context"
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"google.golang.org/grpc"
	"net"
)

type IGrpcController interface {
	Start() error
}

type GrpcControllerImpl struct {
	IGrpcController

	CheckSecurityAction   *CheckSecurityAction
	GetUserAction         *GetUserAction
	ValidateCallerAction  *ValidateCallerAction
	RegisterAccountAction *RegisterAccountAction
	GetFileAction         *GetFileAction

	Config                      *core.Config
	ActionRunner                IActionRunner
	EntityFromGRPCReaderService IEntityFromGRPCReaderService
	routerModifiers             []func(s *grpc.Server)
}

func (c *GrpcControllerImpl) AddRouterModifier(modifier func(e *grpc.Server)) {
	c.routerModifiers = append(c.routerModifiers, modifier)
}

func (c *GrpcControllerImpl) Start() error {

	if c.Config.Server == nil {
		return nil
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", c.Config.Server.GrpcPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	for _, modifier := range c.routerModifiers {
		modifier(s)
	}

	println(fmt.Sprintf("%v grpc server started on port %v", c.Config.App.Name, c.Config.Server.GrpcPort))
	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (c *GrpcControllerImpl) RunByErrorProvider(ctx context.Context, action IAction, errorProviderService IErrorProviderService) *Result {
	return c.ActionRunner.Run(
		action,
		func() (interface{}, error) {
			return c.EntityFromGRPCReaderService.ReadCallParams(ctx)
		},
		errorProviderService,
	)
}

func (c *GrpcControllerImpl) Run(ctx context.Context, action IAction) *Result {
	return c.RunByErrorProvider(ctx, action, nil)
}

func (c *GrpcControllerImpl) RunByActionAndErrorProvider(ctx context.Context, action func(args *core.CallParams) IAction, errorProviderService IErrorProviderService) *Result {
	return c.ActionRunner.RunByProvider(
		func(args interface{}) IAction {
			return action(args.(*core.CallParams))
		},
		func() (interface{}, error) {
			return c.EntityFromGRPCReaderService.ReadCallParams(ctx)
		},
		errorProviderService,
	)
}

func (c *GrpcControllerImpl) RunByActionProvider(ctx context.Context, action func(args *core.CallParams) IAction) *Result {
	return c.RunByActionAndErrorProvider(ctx, action, nil)
}
