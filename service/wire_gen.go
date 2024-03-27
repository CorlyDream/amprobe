// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package service

import (
	api3 "github.com/amuluze/amprobe/service/auth/api"
	repository3 "github.com/amuluze/amprobe/service/auth/repository"
	service3 "github.com/amuluze/amprobe/service/auth/service"
	"github.com/amuluze/amprobe/service/container/api"
	"github.com/amuluze/amprobe/service/container/repository"
	"github.com/amuluze/amprobe/service/container/service"
	api2 "github.com/amuluze/amprobe/service/host/api"
	repository2 "github.com/amuluze/amprobe/service/host/repository"
	service2 "github.com/amuluze/amprobe/service/host/service"
	"github.com/amuluze/amprobe/service/model"
)

// Injectors from wire.go:

func BuildInjector(configFile string) (*Injector, func(), error) {
	config, err := NewConfig(configFile)
	if err != nil {
		return nil, nil, err
	}
	store, cleanup, err := InitAuthStore(config)
	if err != nil {
		return nil, nil, err
	}
	auther, cleanup2, err := InitAuth(config, store)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	models := model.NewModels()
	db, err := NewDB(config, models)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	containerRepo := repository.NewContainerRepo(db)
	containerService := service.NewContainerService(containerRepo)
	containerAPI := api.NewContainerAPI(containerService)
	hostRepo := repository2.NewHostRepo(db)
	hostService := service2.NewHostService(hostRepo)
	hostAPI := api2.NewHostAPI(hostService)
	authRepo := repository3.NewAuthRepo(db)
	authService := service3.NewAuthService(auther, authRepo)
	authAPI := api3.NewLoginAPI(authService)
	loggerHandler := NewLoggerHandler()
	router := &Router{
		config:        config,
		auth:          auther,
		containerAPI:  containerAPI,
		hostAPI:       hostAPI,
		authAPI:       authAPI,
		loggerHandler: loggerHandler,
	}
	app := NewFiberApp(config, router)
	prepare := &Prepare{
		db: db,
	}
	timedTask := NewTimedTask(config, db)
	logger := NewLogger(config)
	injector, err := NewInjector(app, router, prepare, config, timedTask, logger)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return injector, func() {
		cleanup2()
		cleanup()
	}, nil
}
