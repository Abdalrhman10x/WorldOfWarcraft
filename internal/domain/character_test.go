package domain

import (
	"errors"
	"testing"
)

func TestLevelForTotalXPIsMonotonic(t *testing.T) {
	prev := 0
	for level := 1; level <= MaxLevel; level++ {
		got := levelForTotalXP(xpThresholds[level])
		if got != level {
			t.Fatalf("levelForTotalXP(threshold[%d]) = %d, want %d", level, got, level)
		}
		if level < MaxLevel {
			if got := levelForTotalXP(xpThresholds[level+1] - 1); got != level {
				t.Fatalf("one XP short of level %d resolved to %d", level+1, got)
			}
		}
		if level > 1 && level <= prev {
			t.Fatalf("thresholds not increasing at level %d", level)
		}
		prev = level
	}
}

func TestAddXpCrossesManyLevels(t *testing.T) {
	c, err := NewCharacter("thrall", Orc, Shaman)
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}

	grant := int(xpThresholds[10])
	gained, err := c.AddXp(grant)
	if err != nil {
		t.Fatalf("AddXp: %v", err)
	}
	if gained != 9 {
		t.Fatalf("gained = %d, want 9 (level 1 -> 10)", gained)
	}
	if c.Level != 10 {
		t.Fatalf("Level = %d, want 10", c.Level)
	}
	if want := c.TotalXp - xpThresholds[10]; int64(c.Xp) != want {
		t.Fatalf("Xp = %d, want %d (overflow lost)", c.Xp, want)
	}
	if c.MaxXp != c.xpForNextLevel() {
		t.Fatalf("MaxXp = %d, want %d", c.MaxXp, c.xpForNextLevel())
	}
}

func TestAddXpPartialProgressAccumulates(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)

	if _, err := c.AddXp(40); err != nil {
		t.Fatalf("AddXp: %v", err)
	}
	if _, err := c.AddXp(40); err != nil {
		t.Fatalf("AddXp: %v", err)
	}
	if c.Level != 1 {
		t.Fatalf("Level = %d, want 1 (two 40s must not each round down)", c.Level)
	}
	if c.Xp != 80 {
		t.Fatalf("Xp = %d, want 80", c.Xp)
	}
	if _, err := c.AddXp(20); err != nil {
		t.Fatalf("AddXp: %v", err)
	}
	if c.Level != 2 || c.Xp != 0 {
		t.Fatalf("Level, Xp = %d, %d; want 2, 0", c.Level, c.Xp)
	}
}

func TestAddXpAtMaxLevelIsRejected(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)
	c.TotalXp = xpThresholds[MaxLevel]
	c.setLevel(MaxLevel)

	if c.MaxXp != 0 || c.Xp != 0 {
		t.Fatalf("MaxXp, Xp = %d, %d; want 0, 0 at max level", c.MaxXp, c.Xp)
	}
	if _, err := c.AddXp(1000); err != ErrCharacterAlreadyReachedMaxLevel {
		t.Fatalf("AddXp at max level err = %v, want ErrCharacterAlreadyReachedMaxLevel", err)
	}
	if err := c.LevelUp(); err != ErrCharacterAlreadyReachedMaxLevel {
		t.Fatalf("LevelUp at max level err = %v, want ErrCharacterAlreadyReachedMaxLevel", err)
	}
	if c.Level != MaxLevel {
		t.Fatalf("Level = %d, want %d", c.Level, MaxLevel)
	}
}

func TestAddXpRejectsNonPositive(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)
	for _, amount := range []int{0, -1, -500} {
		if _, err := c.AddXp(amount); err != ErrInvalidXPAmount {
			t.Errorf("AddXp(%d) err = %v, want ErrInvalidXPAmount", amount, err)
		}
	}
}

func TestLevelUpGrowsStatsMonotonically(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)

	prevHP, prevMP := c.MaxHealth, c.MaxMana
	for c.Level < MaxLevel {
		if err := c.LevelUp(); err != nil {
			t.Fatalf("LevelUp: %v", err)
		}
		if c.MaxHealth <= prevHP {
			t.Fatalf("MaxHealth did not grow at level %d: %d -> %d", c.Level, prevHP, c.MaxHealth)
		}
		if c.MaxMana <= prevMP {
			t.Fatalf("MaxMana did not grow at level %d: %d -> %d", c.Level, prevMP, c.MaxMana)
		}
		if c.Health != c.MaxHealth || c.Mana != c.MaxMana {
			t.Fatalf("level %d did not refill: %d/%d hp, %d/%d mp",
				c.Level, c.Health, c.MaxHealth, c.Mana, c.MaxMana)
		}
		prevHP, prevMP = c.MaxHealth, c.MaxMana
	}
}

func TestNewCharacterInitializesDerivedStats(t *testing.T) {
	c, err := NewCharacter("thrall", Orc, Shaman)
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	if c.MaxHealth <= 0 || c.MaxMana <= 0 {
		t.Fatalf("stats not initialized: %d hp, %d mp", c.MaxHealth, c.MaxMana)
	}
	if c.Health != c.MaxHealth || c.Mana != c.MaxMana {
		t.Fatal("character did not spawn at full resources")
	}
	if c.MaxXp != c.xpForNextLevel() || c.Xp != 0 {
		t.Fatalf("xp bar = %d/%d, want 0/%d", c.Xp, c.MaxXp, c.xpForNextLevel())
	}
	if c.Inventory == nil || c.Inventory.Count() != DefaultInventorySize {
		t.Fatal("inventory not initialized")
	}
}

func TestNewCharacterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		playable Name
		race     Race
		class    Class
		want     error
	}{
		{name: "empty name", playable: "", race: Orc, class: Shaman, want: ErrInvalidCharacterName},
		{name: "whitespace only", playable: "   ", race: Orc, class: Shaman, want: ErrInvalidCharacterName},
		{name: "too long", playable: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa", race: Orc, class: Shaman, want: ErrInvalidCharacterName},
		{name: "illegal rune", playable: "thr@ll", race: Orc, class: Shaman, want: ErrInvalidCharacterName},
		{name: "unknown race", playable: "thrall", race: "gnome", class: Shaman, want: ErrUnknownRace},
		{name: "unknown class", playable: "thrall", race: Orc, class: "warlock", want: ErrUnknownClass},
		{name: "orc cannot be druid", playable: "thrall", race: Orc, class: Druid, want: ErrInvalidClassCombo},
		{name: "blood elf cannot be shaman", playable: "thrall", race: BloodElf, class: Shaman, want: ErrInvalidClassCombo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewCharacter(tt.playable, tt.race, tt.class); err != tt.want {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNewCharacterTrimsName(t *testing.T) {
	c, err := NewCharacter("  Thrall  ", Orc, Shaman)
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	if c.Name != "thrall" {
		t.Fatalf("Name = %q, want %q", c.Name, "thrall")
	}
}

func TestNewCharacterRejectsInvalidXP(t *testing.T) {
	for _, raw := range []string{"", "   ", "thr@ll", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		if _, err := NewCharacter(Name(raw), Orc, Shaman); !errors.Is(err, ErrInvalidCharacterName) {
			t.Errorf("NewCharacter(%q) err = %v, want ErrInvalidCharacterName", raw, err)
		}
	}
}

func TestNormalizeNameIsCanonical(t *testing.T) {
	for _, raw := range []string{"Thrall", "thrall", "  THRALL  ", "ThRaLl"} {
		got, err := NormalizeName(Name(raw))
		if err != nil {
			t.Fatalf("NormalizeName(%q): %v", raw, err)
		}
		if got != "thrall" {
			t.Errorf("NormalizeName(%q) = %q, want %q", raw, got, "thrall")
		}
	}
}

func TestIsValidCombo(t *testing.T) {
	if !IsValidCombo(NightElf, Druid) {
		t.Error("night elf druid should be valid")
	}
	if IsValidCombo(Orc, Paladin) {
		t.Error("orc paladin should be invalid")
	}
	if IsValidCombo("gnome", Mage) {
		t.Error("unknown race should be invalid")
	}
}

func TestGoldCannotGoNegative(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)
	if err := c.AddGold(100); err != nil {
		t.Fatalf("AddGold: %v", err)
	}
	if err := c.SpendGold(150); err != ErrNotEnoughGold {
		t.Fatalf("SpendGold(150) err = %v, want ErrNotEnoughGold", err)
	}
	if c.Gold != 100 {
		t.Fatalf("Gold = %d, want 100 (failed spend must not mutate)", c.Gold)
	}
	if err := c.SpendGold(100); err != nil || c.Gold != 0 {
		t.Fatalf("SpendGold(100) = %v, gold = %d; want nil, 0", err, c.Gold)
	}
	if err := c.AddGold(0); err != ErrInvalidAmount {
		t.Errorf("AddGold(0) err = %v, want ErrInvalidAmount", err)
	}
	if err := c.SpendGold(-1); err != ErrInvalidAmount {
		t.Errorf("SpendGold(-1) err = %v, want ErrInvalidAmount", err)
	}
}

func TestInventoryCloneIsIndependent(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)
	if err := c.Inventory.Add(0, 3); err != nil {
		t.Fatalf("Add: %v", err)
	}

	fork := c.Inventory.ClonePtr()
	if err := fork.Add(1, 2); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if len(c.Inventory.Slots) != 1 {
		t.Fatalf("original has %d slots, want 1: clone shares the backing array", len(c.Inventory.Slots))
	}
	if c.Inventory.Slots[0].ItemID != 0 || c.Inventory.Slots[0].Quantity != 3 {
		t.Fatalf("original slot mutated: %+v", c.Inventory.Slots[0])
	}
	if len(fork.Slots) != 2 {
		t.Fatalf("clone has %d slots, want 2", len(fork.Slots))
	}

	if err := c.Inventory.Remove(0, 3); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(fork.Slots) != 2 {
		t.Fatal("removing from the original emptied the clone")
	}
}

func TestActivateSuspendTransitions(t *testing.T) {
	c, _ := NewCharacter("thrall", Orc, Shaman)
	if err := c.Suspend(); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	if err := c.Suspend(); err != ErrIsAlreadySuspended {
		t.Errorf("second Suspend err = %v, want ErrIsAlreadySuspended", err)
	}
	if err := c.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := c.Activate(); err != ErrIsAlreadyActive {
		t.Errorf("second Activate err = %v, want ErrIsAlreadyActive", err)
	}
}
