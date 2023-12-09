package util

type Role string

var (
	DepositorRole = Role("depositor")
	BankerRole    = Role("banker")
)

func (role Role) String() string {
	return string(role)
}
