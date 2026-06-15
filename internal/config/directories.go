package config

import (
	"os"
	"path/filepath"
)

type Directories struct {
	First       string `mapstructure:"first-images-dir" validate:"omitempty,dir"`
	Second      string `mapstructure:"second-images-dir" validate:"omitempty,dir"`
	Third       string `mapstructure:"third-images-dir" validate:"omitempty,dir"`
	Fourth      string `mapstructure:"fourth-images-dir" validate:"omitempty,dir"`
	FirstLarge  string `mapstructure:"first-large-images-dir" validate:"omitempty,dir"`
	SecondLarge string `mapstructure:"second-large-images-dir" validate:"omitempty,dir"`
	ThirdLarge  string `mapstructure:"third-large-images-dir" validate:"omitempty,dir"`
	FourthLarge string `mapstructure:"fourth-large-images-dir" validate:"omitempty,dir"`
}

func (d Directories) Get(slot Slot) string {
	switch slot {
	case Slot1:
		return d.First
	case Slot2:
		return d.Second
	case Slot3:
		return d.Third
	case Slot4:
		return d.Fourth
	case SlotLarge1:
		return d.FirstLarge
	case SlotLarge2:
		return d.SecondLarge
	case SlotLarge3:
		return d.ThirdLarge
	case SlotLarge4:
		return d.FourthLarge
	default:
		return ""
	}
}

func (d Directories) DefaultToExtracted(path string) Directories {
	if path == "" {
		return d
	}
	if d.First == "" {
		d.First = extractionDir(path, "rw_images1")
	}
	if d.Second == "" {
		d.Second = extractionDir(path, "rw_images2")
	}
	if d.Third == "" {
		d.Third = extractionDir(path, "rw_images3")
	}
	if d.Fourth == "" {
		d.Fourth = extractionDir(path, "rw_images4")
	}
	if d.FirstLarge == "" {
		d.FirstLarge = extractionDir(path, "rw_large1")
	}
	if d.SecondLarge == "" {
		d.SecondLarge = extractionDir(path, "rw_large2")
	}
	if d.ThirdLarge == "" {
		d.ThirdLarge = extractionDir(path, "rw_large3")
	}
	if d.FourthLarge == "" {
		d.FourthLarge = extractionDir(path, "rw_large4")
	}
	return d
}

func extractionDir(path string, name string) string {
	dir := filepath.Join(path, name)
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return dir
	}
	return ""
}
