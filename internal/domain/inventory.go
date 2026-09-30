package domain

import "errors"

var (
	ErrInventoryFull   = errors.New("inventory is full")
	ErrNotEnoughGold   = errors.New("not enough gold")
	ErrNotEnoughItems  = errors.New("not enough items")
	ErrUnknownItem     = errors.New("unknown item")
	ErrInvalidQuantity = errors.New("invalid quanitity")
)

const DefaultInventorySize = 16

type Slot struct {
	ItemID   ItemID `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type Inventory struct {
	Slots []Slot `json:"slots"`
	Size  int    `json:"size"`
}

func NewInventory(size int) *Inventory {
	return &Inventory{
		Slots: make([]Slot, 0, size),
		Size:  size,
	}
}

func (inv Inventory) Count() int {
	return inv.Size
}

func (inv Inventory) Clone() Inventory {
	slots := make([]Slot, len(inv.Slots), cap(inv.Slots))
	copy(slots, inv.Slots)
	return Inventory{Slots: slots, Size: inv.Size}
}

func (inv *Inventory) ClonePtr() *Inventory {
	if inv == nil {
		return nil
	}
	clone := inv.Clone()
	return &clone
}

func (inv *Inventory) Add(itemID ItemID, quantity int) error {
	item, ok := Items[itemID]
	if !ok {
		return ErrUnknownItem
	}

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if item.Stackable {
		for i := range inv.Slots {
			if inv.Slots[i].ItemID != itemID {
				continue
			}

			space := item.MaxStack - inv.Slots[i].Quantity
			add := min(space, quantity)
			inv.Slots[i].Quantity += add
			quantity -= add

			if quantity == 0 {
				return nil
			}
		}
	}

	for quantity > 0 {
		if len(inv.Slots) >= inv.Size {
			return ErrInventoryFull
		}

		add := quantity
		if item.Stackable {
			add = min(add, item.MaxStack)
		} else {
			add = 1
		}
		inv.Slots = append(inv.Slots, Slot{ItemID: itemID, Quantity: add})
		quantity -= add
	}

	return nil
}

func (inv *Inventory) Remove(itemID ItemID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	total := 0
	for _, slot := range inv.Slots {
		if slot.ItemID == itemID {
			total += slot.Quantity
		}
	}
	if total < quantity {
		return ErrNotEnoughItems
	}

	kept := inv.Slots[:0]
	for i := range inv.Slots {
		slot := inv.Slots[i]

		if slot.ItemID != itemID {
			kept = append(kept, slot)
			continue
		}

		remove := min(slot.Quantity, quantity)
		slot.Quantity -= remove
		quantity -= remove

		if slot.Quantity > 0 {
			kept = append(kept, slot)
		}
	}

	inv.Slots = kept

	return nil
}
