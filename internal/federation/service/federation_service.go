package service

import (
	"github.com/arcane/arcanelink/internal/federation/discovery"
	"github.com/arcane/arcanelink/internal/federation/forwarder"
	"github.com/arcane/arcanelink/pkg/models"
)

type FederationService struct {
	resolver  *discovery.ServerResolver
	forwarder *forwarder.MessageForwarder
}

func NewFederationService() *FederationService {
	resolver := discovery.NewServerResolver()
	forwarder := forwarder.NewMessageForwarder(resolver)

	return &FederationService{
		resolver:  resolver,
		forwarder: forwarder,
	}
}

// ForwardDirectMessage forwards a direct message to a remote server
func (s *FederationService) ForwardDirectMessage(msg *models.DirectMessage) error {
	return s.forwarder.ForwardDirectMessage(msg)
}

// ForwardRoomEvent forwards a room event to remote servers
func (s *FederationService) ForwardRoomEvent(event *models.RoomEvent, memberIDs []string) error {
	return s.forwarder.ForwardRoomEvent(event, memberIDs)
}

// ResolveServer resolves a domain to server information
func (s *FederationService) ResolveServer(domain string) (*models.ServerInfo, error) {
	return s.resolver.Resolve(domain)
}
