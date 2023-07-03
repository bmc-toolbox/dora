package model

import "time"

type DiscoverHint struct {
	IP        string `gorm:"primary_key;column:ip"`
	Hint      string `gorm:"column:hint"`
	UpdatedAt time.Time
}
