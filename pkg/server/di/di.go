package di

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/di"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/echo-http"
	"github.com/itskovichanton/server/pkg/server"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/itskovichanton/server/pkg/server/users"
	"go.uber.org/dig"
)

type DI struct {
	di.DI
}

func (c *DI) InitDI(container *dig.Container) {
	c.buildContainer(container)
	c.DI.InitDI(container)
}

func (c *DI) buildContainer(container *dig.Container) *dig.Container {

	container.Provide(c.NewHttpController)
	container.Provide(c.NewGrpcController)
	container.Provide(c.NewJsonPresenter)
	container.Provide(c.NewErrorProviderService)
	container.Provide(c.NewActionRunner)
	container.Provide(c.NewDefaultErrorProviderService)
	container.Provide(c.NewEntityFromHTTPReaderService)
	container.Provide(c.NewEntityFromGRPCReaderService)
	container.Provide(c.NewSessionStorageService)
	container.Provide(c.NewUserRepo)
	container.Provide(c.NewAuthService)
	container.Provide(c.NewGetUserAction)
	container.Provide(c.NewValidateActiveUserAction)
	container.Provide(c.NewValidateCallerAction)
	container.Provide(c.NewGetSessionAction)
	container.Provide(c.NewServerSettingsProviderService)
	container.Provide(c.NewGetFileAction)
	container.Provide(c.NewFileStorageService)
	container.Provide(c.NewRegisterAccountAction)
	container.Provide(c.NewCallerValidatorService)
	container.Provide(c.NewDefaultFilePresenter)
	container.Provide(c.NewResponsePresenter)
	container.Provide(c.NewServerConfigService)

	return container
}

func (c *DI) NewServerConfigService(config *core.Config) (*server.Config, error) {
	return (&server.ConfigServiceImpl{Config: config}).LoadConfig()
}

func (c *DI) NewResponsePresenter() pipeline.IResponsePresenter {
	return &pipeline.JSONResponsePresenterImpl{
		ResponsePresenterImpl: pipeline.ResponsePresenterImpl{
			ResponseModelProvider: &pipeline.ResponseModelProviderImpl{},
		},
	}
}

func (c *DI) NewDefaultFilePresenter(jsonPresenter *pipeline.JSONResponsePresenterImpl) *pipeline.FileResponsePresenterImpl {
	return &pipeline.FileResponsePresenterImpl{
		JSONResponsePresenterImpl: *jsonPresenter,
	}
}

func (c *DI) NewJsonPresenter() *pipeline.JSONResponsePresenterImpl {
	return &pipeline.JSONResponsePresenterImpl{
		ResponsePresenterImpl: pipeline.ResponsePresenterImpl{
			ResponseModelProvider: &pipeline.ResponseModelProviderImpl{},
		},
	}
}

func (c *DI) NewErrorProviderService(config *core.Config) *pipeline.ErrorProviderServiceImpl {
	return &pipeline.ErrorProviderServiceImpl{
		Config: config,
	}
}

func (c *DI) NewDefaultErrorProviderService(config *core.Config) pipeline.IErrorProviderService {
	return c.NewErrorProviderService(config)
}

func (c *DI) NewActionRunner(loggerService logger.ILoggerService, errorHandler core.IErrorHandler, errorProviderService pipeline.IErrorProviderService) pipeline.IActionRunner {
	return &pipeline.ActionRunnerImpl{
		LoggerService:               loggerService,
		ErrorHandler:                errorHandler,
		DefaultErrorProviderService: errorProviderService,
	}
}

func (c *DI) NewEntityFromHTTPReaderService(config *server.Config) pipeline.IEntityFromHTTPReaderService {
	return &pipeline.EntityFromHTTPReaderServiceImpl{
		Config: config,
	}
}

func (c *DI) NewEntityFromGRPCReaderService(config *server.Config) pipeline.IEntityFromGRPCReaderService {
	return &pipeline.EntityFromGRPCReaderServiceImpl{
		Config: config,
	}
}

func (c *DI) NewUserRepo() users.IUserRepoService {
	r := users.UserRepoServiceImpl{}
	r.Init()
	return &r
}

func (c *DI) NewSessionStorageService() users.ISessionStorageService {
	r := users.SessionStorageServiceImpl{}
	r.Clear()
	return &r
}

func (c *DI) NewAuthService(userRepoService users.IUserRepoService, sessionStorageService users.ISessionStorageService) users.IAuthService {
	return &users.AuthServiceImpl{
		UserRepo:              userRepoService,
		SessionStorageService: sessionStorageService,
	}
}

func (c *DI) NewGetUserAction(authService users.IAuthService) *pipeline.GetUserAction {
	return &pipeline.GetUserAction{
		AuthService: authService,
	}
}

func (c *DI) NewValidateActiveUserAction() *pipeline.ValidateActiveUserAction {
	return &pipeline.ValidateActiveUserAction{}
}

func (c *DI) NewValidateCallerAction(callerValidatorService pipeline.ICallerValidatorService) *pipeline.ValidateCallerAction {
	return &pipeline.ValidateCallerAction{
		CallerValidatorService: callerValidatorService,
	}
}

func (c *DI) NewServerSettingsProviderService(config *core.Config) (pipeline.IServerSettingsProviderService, error) {
	r := &pipeline.ServerSettingsProviderServiceImpl{
		Config: config,
	}
	err := r.Reload()
	if err != nil {
		return r, err
	}
	return r, nil
}

func (c *DI) NewCallerValidatorService(serverSettingsProviderService pipeline.IServerSettingsProviderService) pipeline.ICallerValidatorService {
	return &pipeline.CallerValidatorServiceImpl{
		ServerSettingsProviderService: serverSettingsProviderService,
	}
}

func (c *DI) NewRegisterAccountAction(authService users.IAuthService) *pipeline.RegisterAccountAction {
	return &pipeline.RegisterAccountAction{
		AuthService: authService,
	}
}

func (c *DI) NewFileStorageService(config *core.Config) filestorage.IFileStorageService {
	return &filestorage.FileStorageService{
		Config: config,
	}
}

func (c *DI) NewGetFileAction(fileStorageService filestorage.IFileStorageService) *pipeline.GetFileAction {
	return &pipeline.GetFileAction{
		FileStorageService: fileStorageService,
	}
}

func (c *DI) NewGetSessionAction() *pipeline.GetSessionAction {
	return &pipeline.GetSessionAction{}
}

func (c *DI) NewHttpController(config *server.Config, getSessionAction *pipeline.GetSessionAction, responsePresenter pipeline.IResponsePresenter, filePresenter *pipeline.FileResponsePresenterImpl, getFileAction *pipeline.GetFileAction, registerAccountAction *pipeline.RegisterAccountAction, validateCallerAction *pipeline.ValidateCallerAction, getUserAction *pipeline.GetUserAction, actionRunner pipeline.IActionRunner, entityFromHTTPReaderService pipeline.IEntityFromHTTPReaderService) *pipeline.HttpControllerImpl {
	return &pipeline.HttpControllerImpl{
		NopAction:                   &pipeline.NopActionImpl{},
		GetSessionAction:            getSessionAction,
		GetUserAction:               getUserAction,
		ValidateCallerAction:        validateCallerAction,
		RegisterAccountAction:       registerAccountAction,
		GetFileAction:               getFileAction,
		Config:                      config,
		ActionRunner:                actionRunner,
		EntityFromHTTPReaderService: entityFromHTTPReaderService,
		DefaultResponsePresenter:    responsePresenter,
		FileResponsePresenter:       filePresenter,
		EchoEngine:                  echo.New(),
	}
}

func (c *DI) NewGrpcController(registerAccountAction *pipeline.RegisterAccountAction, validateCallerAction *pipeline.ValidateCallerAction, getUserAction *pipeline.GetUserAction, config *server.Config, actionRunner pipeline.IActionRunner, entityFromGRPCReaderService pipeline.IEntityFromGRPCReaderService) *pipeline.GrpcControllerImpl {
	return &pipeline.GrpcControllerImpl{
		GetUserAction:               getUserAction,
		ValidateCallerAction:        validateCallerAction,
		RegisterAccountAction:       registerAccountAction,
		Config:                      config,
		ActionRunner:                actionRunner,
		EntityFromGRPCReaderService: entityFromGRPCReaderService,
	}
}
