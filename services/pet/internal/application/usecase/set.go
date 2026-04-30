package usecase

import "pet/internal/application/ports"

type Set struct {
	*Pet
}

func NewSet(pets ports.PetRepository, acl ports.ACLClient, file ports.FileClient) *Set {
	return &Set{
		Pet: New(pets, acl, file),
	}
}
