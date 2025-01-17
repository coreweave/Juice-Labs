/*
 *  Copyright (c) 2023 Juice Technologies, Inc. All Rights Reserved.
 */
package main

import (
	"crypto/tls"
	"errors"
	"flag"

	"github.com/Juice-Labs/Juice-Labs/cmd/agent/app"
	"github.com/Juice-Labs/Juice-Labs/cmd/agent/playnite"
	"github.com/Juice-Labs/Juice-Labs/cmd/agent/prometheus"
	"github.com/Juice-Labs/Juice-Labs/cmd/internal/build"
	"github.com/Juice-Labs/Juice-Labs/pkg/appmain"
	"github.com/Juice-Labs/Juice-Labs/pkg/crypto"
	"github.com/Juice-Labs/Juice-Labs/pkg/logger"
	"github.com/Juice-Labs/Juice-Labs/pkg/task"
)

var (
	certFile     = flag.String("cert-file", "", "")
	keyFile      = flag.String("key-file", "", "")
	generateCert = flag.Bool("generate-cert", false, "Generates a certificate for https")
	disableTls   = flag.Bool("disable-tls", true, "")
)

func main() {
	appmain.Run("Juice Agent", build.Version, func(group task.Group) error {
		var tlsConfig *tls.Config

		var err error
		if !*disableTls {
			var certificate tls.Certificate
			if *certFile != "" && *keyFile != "" {
				certificate, err = tls.LoadX509KeyPair(*certFile, *keyFile)
			} else if *generateCert {
				certificate, err = crypto.GenerateCertificate()
			} else {
				err = errors.New("https is required, use both --cert-file and --key-file or --generate-cert")
			}

			if err == nil {
				tlsConfig = &tls.Config{
					Certificates: []tls.Certificate{certificate},
				}
			}
		}

		if err == nil {
			agent, err := app.NewAgent(tlsConfig)
			if err == nil {
				consumer, err_ := playnite.NewGpuMetricsConsumer(agent)
				err = err_
				if err != nil {
					logger.Warning(err)
				}

				agent.GpuMetricsProvider.AddConsumer(consumer)
				agent.GpuMetricsProvider.AddConsumer(prometheus.NewGpuMetricsConsumer())

				err = agent.ConnectToController(group)
				if err == nil {
					group.Go("Agent", agent)
				} else {
					group.Cancel()
				}
			}

			return err
		}

		return nil
	})
}
