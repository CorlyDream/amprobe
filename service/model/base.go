
package model

import "time"

type SeriesModel struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
}