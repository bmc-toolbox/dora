package connectors

import "gitlab.booking.com/infra/dora/model"

// Bmc represents the requirement of items to be collected a server
type Bmc interface {
	Serial() (string, error)
	Model() (string, error)
	BmcType() (string, error)
	BmcVersion() (string, error)
	Name() (string, error)
	Status() (string, error)
	Memory() (int, error)
	CPU() (string, int, int, int, error)
	BiosVersion() (string, error)
	PowerKw() (float64, error)
	TempC() (int, error)
	Nics() ([]*model.Nic, error)
	License() (string, string, error)
	Login() error
	Logout() error
}

// BmcChassis represents the requirement of items to be collected from a chassis
type BmcChassis interface {
	Name() (string, error)
	Model() (string, error)
	Serial() (string, error)
	PowerKw() (float64, error)
	TempC() (int, error)
	Status() (string, error)
	FwVersion() (string, error)
	PassThru() (string, error)
	PowerSupplyCount() (int, error)
	Blades() ([]*model.Blade, error)
	StorageBlades() ([]*model.StorageBlade, error)
}
