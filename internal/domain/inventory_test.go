package domain

import "testing"

func TestNewInventory(t *testing.T) {
	inventory := NewInventory(DefaultInventorySize)

	if inventory == nil {
		t.Fatalf("NewInventory() = nil, want non-nil")
	}
	if inventory.Size != DefaultInventorySize {
		t.Errorf("Size = %d, want %d", inventory.Size, DefaultInventorySize)
	}

	if len(inventory.Slots) != 0 {
		t.Errorf("len(Slots) = %d, want 0", len(inventory.Slots))
	}

	if cap(inventory.Slots) != DefaultInventorySize {
		t.Errorf("cap(Slots) = %d, want %d", cap(inventory.Slots), DefaultInventorySize)
	}
}

func TestAdd(t *testing.T) {
	inv := NewInventory(DefaultInventorySize)

	err := inv.Add(0, 1)
	if err != nil {
		t.Fatalf("Add() error = %v, want nil", err)
	}

	if len(inv.Slots) != 1 {
		t.Fatalf("len(Slots) = %d, want 1", len(inv.Slots))
	}

	if inv.Slots[0].ItemID != 0 {
		t.Errorf("Slots[0].ItemID = %d, want 0", inv.Slots[0].ItemID)
	}

	if inv.Slots[0].Quantity != 1 {
		t.Errorf("Slots[0].Quantity = %d, want 1", inv.Slots[0].Quantity)
	}
}
