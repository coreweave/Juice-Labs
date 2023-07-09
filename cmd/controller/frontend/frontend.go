/*
 *  Copyright (c) 2023 Juice Technologies, Inc. All Rights Reserved.
 */
package frontend

import (
	"crypto/tls"
	"flag"
	"os"
	"time"

	"github.com/Juice-Labs/Juice-Labs/cmd/controller/storage"
	"github.com/Juice-Labs/Juice-Labs/pkg/logger"
	"github.com/Juice-Labs/Juice-Labs/pkg/restapi"
	"github.com/Juice-Labs/Juice-Labs/pkg/server"
	"github.com/Juice-Labs/Juice-Labs/pkg/task"
)

var (
	address = flag.String("address", "0.0.0.0:43210", "The IP address and port to use for listening for client connections")
)

type Frontend struct {
	startTime time.Time

	hostname string

	server  *server.Server
	storage storage.Storage
}

func NewFrontend(tlsConfig *tls.Config, storage storage.Storage) (*Frontend, error) {
	if tlsConfig == nil {
		logger.Warning("TLS is disabled, data will be unencrypted")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	server, err := server.NewServer(*address, tlsConfig)
	if err != nil {
		return nil, err
	}

	frontend := &Frontend{
		startTime: time.Now(),
		hostname:  hostname,
		server:    server,
		storage:   storage,
	}

	frontend.initializeEndpoints()

	return frontend, nil
}

func (frontend *Frontend) Run(group task.Group) error {
	group.Go("Frontend Server", frontend.server)
	return nil
}

func (frontend *Frontend) registerAgent(agent restapi.Agent) (string, error) {
	agent.State = restapi.AgentActive
	return frontend.storage.RegisterAgent(agent)
}

func (frontend *Frontend) getAgentById(id string) (restapi.Agent, error) {
	return frontend.storage.GetAgentById(id)
}

func (frontend *Frontend) updateAgent(agent restapi.Agent) error {
	update := storage.AgentUpdate{
		Id:       agent.Id,
		State:    agent.State,
		Sessions: make([]storage.SessionUpdate, len(agent.Sessions)),
	}

	for index, session := range agent.Sessions {
		update.Sessions[index] = storage.SessionUpdate{
			Id:    session.Id,
			State: session.State,
		}
	}

	return frontend.storage.UpdateAgent(update)
}

func (frontend *Frontend) requestSession(sessionRequirements restapi.SessionRequirements) (string, error) {
	return frontend.storage.RequestSession(sessionRequirements)
}

func (frontend *Frontend) getSessionById(id string) (restapi.Session, error) {
	return frontend.storage.GetSessionById(id)
}
