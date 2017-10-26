package connectors

import "gitlab.booking.com/infra/dora/model"

// Bmc represents the requirement of items to be collected a server
type Bmc interface {
	BiosVersion() (string, error)
	BmcType() (string, error)
	BmcVersion() (string, error)
	CPU() (string, int, int, int, error)
	License() (string, string, error)
	Login() error
	Logout() error
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*model.Nic, error)
	PowerKw() (float64, error)
	Serial() (string, error)
	Status() (string, error)
	TempC() (int, error)
}

// BmcChassis represents the requirement of items to be collected from a chassis
type BmcChassis interface {
	Blades() ([]*model.Blade, error)
	FwVersion() (string, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*model.Nic, error)
	PassThru() (string, error)
	PowerKw() (float64, error)
	PowerSupplyCount() (int, error)
	Serial() (string, error)
	Status() (string, error)
	StorageBlades() ([]*model.StorageBlade, error)
	TempC() (int, error)
}
