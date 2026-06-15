package config

type Slot int

const (
	Slot1 Slot = iota
	Slot2
	Slot3
	Slot4
	Slot5
	Slot6
	Slot7
	Slot8

	SlotLarge1 = Slot5
	SlotLarge2 = Slot6
	SlotLarge3 = Slot7
	SlotLarge4 = Slot8
)

func (s Slot) String() string {
	switch s {
	case Slot1:
		return "first"
	case Slot2:
		return "second"
	case Slot3:
		return "third"
	case Slot4:
		return "fourth"
	case SlotLarge1:
		return "first large"
	case SlotLarge2:
		return "second large"
	case SlotLarge3:
		return "third large"
	case SlotLarge4:
		return "fourth large"
	default:
		return ""
	}
}

func (s Slot) Position() int {
	if s.IsLarge() {
		return int(s) - int(SlotLarge1) + 1
	}
	return int(s) + 1
}

func (s Slot) IsLarge() bool {
	return s >= SlotLarge1
}

func (s Slot) Invert() Slot {
	if s.IsLarge() {
		return Slot(s - SlotLarge1 + Slot1)
	}
	return Slot(s + SlotLarge1 - Slot1)
}
