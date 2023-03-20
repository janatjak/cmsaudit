package model

import "gorm.io/gorm"

type ProjectEntry struct {
	gorm.Model
	Name      string `gorm:"size:255;not null;"`
	GitlabUrl string `gorm:"size:255;not null;"`
	WebUrl    string `gorm:"size:255;not null;"`
}
