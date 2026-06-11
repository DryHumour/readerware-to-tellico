package images

import "github.com/DryHumour/readerware-to-tellico/internal/config"

type Row [8]*ManifestEntry

func (r Row) Slot(slot config.Slot) *ManifestEntry {
	if slot < config.Slot1 || slot > config.Slot8 {
		return nil
	}
	if entry := r[slot]; entry != nil {
		return entry
	}
	return nil
}

func (r Row) Image(pos int) *ManifestEntry {
	if pos < 1 || pos > 4 {
		return nil
	}
	slot := config.Slot1 + config.Slot(pos-1)
	return r.Slot(slot)
}

func (r Row) LargeImage(pos int) *ManifestEntry {
	if pos < 1 || pos > 4 {
		return nil
	}
	slot := config.SlotLarge1 + config.Slot(pos-1)
	return r.Slot(slot)
}

func (r Row) Cover() *ManifestEntry {
	for _, entry := range r {
		if entry != nil {
			return entry
		}
	}
	return nil
}

func (r Row) LargeCover() *ManifestEntry {
	for n := range r {
		slot := config.Slot(n)
		if entry := r[slot.Invert()]; entry != nil {
			return entry
		}
	}
	return nil
}

func (r *Row) First() *ManifestEntry   { return r.Slot(config.Slot1) }
func (r *Row) Second() *ManifestEntry  { return r.Slot(config.Slot2) }
func (r *Row) Third() *ManifestEntry   { return r.Slot(config.Slot3) }
func (r *Row) Fourth() *ManifestEntry  { return r.Slot(config.Slot4) }
func (r *Row) Fifth() *ManifestEntry   { return r.Slot(config.Slot5) }
func (r *Row) Sixth() *ManifestEntry   { return r.Slot(config.Slot6) }
func (r *Row) Seventh() *ManifestEntry { return r.Slot(config.Slot7) }
func (r *Row) Eighth() *ManifestEntry  { return r.Slot(config.Slot8) }
