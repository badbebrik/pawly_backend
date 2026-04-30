package usecase

import (
	"acl/internal/application/ports"
)

type Dependencies struct {
	Memberships ports.MembershipRepository
	Roles       ports.RoleRepository
	Invites     ports.InviteRepository
	Options     Options
}

type Set struct {
	*ACL
}

func NewSet(deps Dependencies) *Set {
	return &Set{
		ACL: newACL(
			deps.Memberships,
			deps.Roles,
			deps.Invites,
			Options{
				InviteTTL:          deps.Options.InviteTTL,
				InviteDeeplinkBase: deps.Options.InviteDeeplinkBase,
			},
		),
	}
}
