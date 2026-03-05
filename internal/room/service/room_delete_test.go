package service

import (
	"testing"

	"github.com/arcane/arcanelink/internal/room/repository"
	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/models"
)

func TestDeleteRoom(t *testing.T) {
	// This is a basic test structure
	// In a real scenario, you would set up a test database
	t.Skip("Requires database setup")

	db := &database.DB{} // Mock or test database
	repo := repository.NewRoomRepository(db)
	service := NewRoomService(repo)

	// Create a test room
	room, err := service.CreateRoom("@creator:example.com", "Test Room", "Test Topic", nil)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test: Creator can delete the room
	err = service.DeleteRoom(room.RoomID, "@creator:example.com")
	if err != nil {
		t.Errorf("Creator should be able to delete room: %v", err)
	}

	// Test: Non-creator cannot delete the room
	room2, _ := service.CreateRoom("@creator:example.com", "Test Room 2", "Test Topic", nil)
	err = service.DeleteRoom(room2.RoomID, "@other:example.com")
	if err == nil {
		t.Error("Non-creator should not be able to delete room")
	}
}
