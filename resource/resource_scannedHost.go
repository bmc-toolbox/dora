package resource

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// ScannedHostResource for api2go routes
type ScannedHostResource struct {
	ScannedHostStorage *storage.ScannedHostStorage
}

// FindAll Scans
func (s ScannedHostResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, scans, err := s.queryAndCountAllWrapper(r)
	return &Response{Res: scans}, err
}

// FindOne Scanner
func (s ScannedHostResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := s.ScannedHostStorage.GetOne(ID)
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load Scans in chunks
func (s ScannedHostResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, scans, err := s.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: scans}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (s ScannedHostResource) queryAndCountAllWrapper(r api2go.Request) (count int, scans []model.ScannedHost, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, scans, ErrPageSizeAndNumber
		}
	}

	filters, hasFilters := filter.NewFilterSet(&r)
	offset, limit := filter.OffSetAndLimitParse(&r)

	if hasFilters {
		count, scans, err = s.ScannedHostStorage.GetAllByFilters(offset, limit, filters)
		filters.Clean()
		if err != nil {
			return count, scans, err
		}
	}

	include, hasInclude := r.QueryParams["include"]
	if hasInclude && include[0] == "scanned_hosts" {
		if len(scans) == 0 {
			count, scans, err = s.ScannedHostStorage.GetAllWithAssociations(offset, limit)
		} else {
			var scannedHostsWithInclude []model.ScannedHost
			for _, sn := range scans {
				snWithInclude, err := s.ScannedHostStorage.GetOne(sn.CIDR)
				if err != nil {
					return count, scans, err
				}
				scannedHostsWithInclude = append(scannedHostsWithInclude, snWithInclude)
			}
			scans = scannedHostsWithInclude
		}
	}

	scannedPortsID, hasScannedPorts := r.QueryParams["storage_hostsID"]
	if hasScannedHosts {
		count, scans, err = s.ScannedHostStorage.GetAllByP(offset, limit, scannedHostsID)
		if err != nil {
			return count, scans, err
		}
	}

	if !hasInclude && !hasFilters {
		count, scans, err = s.ScannedHostStorage.GetAll(offset, limit)
		if err != nil {
			return count, scans, err
		}
	}

	return count, scans, err
}
