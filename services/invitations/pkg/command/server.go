package command

import (
	"context"
	"fmt"

	"github.com/oklog/run"
	"github.com/opencloud-eu/opencloud/pkg/config/configlog"
	"github.com/opencloud-eu/opencloud/pkg/tracing"
	"github.com/opencloud-eu/opencloud/pkg/version"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/config"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/config/parser"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/logging"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/metrics"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/server/debug"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/server/http"
	"github.com/opencloud-eu/opencloud/services/invitations/pkg/service/v0"
	"github.com/urfave/cli/v2"
)

// Server is the entrypoint for the server command.
func Server(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:     "server",
		Usage:    fmt.Sprintf("start the %s service without runtime (unsupervised mode)", cfg.Service.Name),
		Category: "server",
		Before: func(c *cli.Context) error {
			return configlog.ReturnFatal(parser.ParseConfig(cfg))
		},
		Action: func(c *cli.Context) error {
			logger := logging.Configure(cfg.Service.Name, cfg.Log)
			traceProvider, err := tracing.GetServiceTraceProvider(cfg.Tracing, cfg.Service.Name)
			if err != nil {
				return err
			}

			var (
				gr          = run.Group{}
				ctx, cancel = context.WithCancel(c.Context)
				metrics     = metrics.New(metrics.Logger(logger))
			)

			defer cancel()

			metrics.BuildInfo.WithLabelValues(version.GetString()).Set(1)

			{

				svc, err := service.New(
					service.Logger(logger),
					service.Config(cfg),
					// service.WithRelationProviders(relationProviders),
				)
				if err != nil {
					logger.Error().Err(err).Msg("handler init")
					return err
				}
				svc = service.NewInstrument(svc, metrics)
				svc = service.NewLogging(svc, logger) // this logs service specific data
				svc = service.NewTracing(svc, traceProvider)

				server, err := http.Server(
					http.Logger(logger),
					http.Context(ctx),
					http.Config(cfg),
					http.Service(svc),
				)
				if err != nil {
					logger.Info().
						Err(err).
						Str("transport", "http").
						Msg("Failed to initialize server")

					return err
				}

				gr.Add(func() error {
					return server.Run()
				}, func(err error) {
					if err != nil {
						logger.Info().
							Str("transport", "http").
							Str("server", cfg.Service.Name).
							Msg("Shutting down server")
					} else {
						logger.Error().Err(err).
							Str("transport", "http").
							Str("server", cfg.Service.Name).
							Msg("Shutting down server")
					}

					cancel()
				})
			}

			{
				debugServer, err := debug.Server(
					debug.Logger(logger),
					debug.Context(ctx),
					debug.Config(cfg),
				)
				if err != nil {
					logger.Info().Err(err).Str("transport", "debug").Msg("Failed to initialize server")
					return err
				}

				gr.Add(debugServer.ListenAndServe, func(_ error) {
					_ = debugServer.Shutdown(ctx)
					cancel()
				})
			}

			return gr.Run()
		},
	}
}
