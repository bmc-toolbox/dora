package resource

import (
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/manyminds/api2go"
	"gitlab.booking.com/infra/dora/filter"
	"gitlab.booking.com/infra/dora/model"
	"gitlab.booking.com/infra/dora/storage"
)

// ScannedNetworkResource for api2go routes
type ScannedNetworkResource struct {
	ScannedNetworkStorage *storage.ScannedNetworkStorage
}

// FindAll Scans
func (s ScannedNetworkResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	_, scans, err := s.queryAndCountAllWrapper(r)
	return &Response{Res: scans}, err
}

// FindOne Scanner
func (s ScannedNetworkResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := s.ScannedNetworkStorage.GetOne(strings.Replace(ID, "-", "/", -1))
	if err == gorm.ErrRecordNotFound {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	return &Response{Res: res}, err
}

// PaginatedFindAll can be used to load Scans in chunks
func (s ScannedNetworkResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	count, scans, err := s.queryAndCountAllWrapper(r)
	return uint(count), &Response{Res: scans}, err
}

// queryAndCountAllWrapper retrieve the data to be used for FindAll and PaginatedFindAll in a standard way
func (s ScannedNetworkResource) queryAndCountAllWrapper(r api2go.Request) (count int, scans []model.ScannedNetwork, err error) {
	for _, invalidQuery := range []string{"page[number]", "page[size]"} {
		_, invalid := r.QueryParams[invalidQuery]
		if invalid {
			return count, scans, ErrPageSizeAndNumber
		}
	}

	offset, limit := filter.OffSetAndLimitParse(&r)

	include, hasInclude := r.QueryParams["include"]
	if hasInclude && include[0] == "scanned_hosts" {
		if len(scans) == 0 {
			count, scans, err = s.ScannedNetworkStorage.GetAllWithAssociations(offset, limit)
		} else {
			var scannedNetworksWithInclude []model.ScannedNetwork
			for _, sn := range scans {
				snWithInclude, err := s.ScannedNetworkStorage.GetOne(sn.CIDR)
				if err != nil {
					return count, scans, err
				}
				scannedNetworksWithInclude = append(scannedNetworksWithInclude, snWithInclude)
			}
			scans = scannedNetworksWithInclude
		}
	}

	if !hasInclude {
		count, scans, err = s.ScannedNetworkStorage.GetAll(offset, limit)
		if err != nil {
			return count, scans, err
		}
	}

	return count, scans, err
}
