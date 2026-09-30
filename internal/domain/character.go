package domain

import (
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Name string

const (
	MaxLevel       = 60
	MaxNameLength  = 24
	baseXPPerLevel = 100
)

var (
	ErrInvalidClassCombo               = errors.New("invalid class combo")
	ErrInvalidCharacterName            = errors.New("invalid character name")
	ErrUnknownRace                     = errors.New("unknown race")
	ErrUnknownClass                    = errors.New("unknown class")
	ErrIsAlreadyActive                 = errors.New("character is already active")
	ErrIsAlreadySuspended              = errors.New("character is already suspended")
	ErrCharacterAlreadyReachedMaxLevel = errors.New("character already reached maximum level")
	ErrInvalidXPAmount                 = errors.New("invalid xp amount")
	ErrInvalidAmount                   = errors.New("invalid amount")
)

type raceClass struct {
	race  Race
	class Class
}

var classCombos = map[raceClass]struct{}{
	{race: Human, class: Warrior}:    {},
	{race: Human, class: Mage}:       {},
	{race: Human, class: Paladin}:    {},
	{race: Orc, class: Warrior}:      {},
	{race: Orc, class: Mage}:         {},
	{race: Orc, class: Shaman}:       {},
	{race: NightElf, class: Warrior}: {},
	{race: NightElf, class: Druid}:   {},
	{race: BloodElf, class: Warrior}: {},
	{race: BloodElf, class: Mage}:    {},
	{race: BloodElf, class: Paladin}: {},
}

func IsValidCombo(race Race, class Class) bool {
	_, ok := classCombos[raceClass{race: race, class: class}]
	return ok
}

type Character struct {
	Name      Name       `json:"name"`
	Race      Race       `json:"race"`
	Class     Class      `json:"class"`
	Level     int        `json:"level"`
	Gold      int        `json:"gold"`
	Inventory *Inventory `json:"inventory"`
	IsActive  bool       `json:"is_active"`

	Health    int   `json:"health"`
	MaxHealth int   `json:"max_health"`
	Mana      int   `json:"mana"`
	MaxMana   int   `json:"max_mana"`
	Xp        int   `json:"xp"`
	MaxXp     int   `json:"max_xp"`
	TotalXp   int64 `json:"total_xp"`
}

var xpThresholds = buildXPThresholds()

func buildXPThresholds() []int64 {
	thresholds := make([]int64, MaxLevel+1)
	var total int64
	for level := 2; level <= MaxLevel; level++ {
		total += int64(baseXPPerLevel * (level - 1))
		thresholds[level] = total
	}
	return thresholds
}

func levelForTotalXP(total int64) int {
	idx := sort.Search(len(xpThresholds), func(i int) bool {
		return xpThresholds[i] > total
	})
	return min(max(idx-1, 1), MaxLevel)
}

func NewCharacter(name Name, race Race, class Class) (*Character, error) {
	name, err := normalizeName(name)
	if err != nil {
		return nil, err
	}
	if _, ok := Races[race]; !ok {
		return nil, ErrUnknownRace
	}
	if _, ok := Classes[class]; !ok {
		return nil, ErrUnknownClass
	}
	if !IsValidCombo(race, class) {
		return nil, ErrInvalidClassCombo
	}

	c := &Character{
		Name:      name,
		Race:      race,
		Class:     class,
		Level:     1,
		Gold:      0,
		Inventory: NewInventory(DefaultInventorySize),
		IsActive:  true,
	}
	c.MaxHealth = c.maxHealthFor(1)
	c.MaxMana = c.maxManaFor(1)
	c.Health, c.Mana = c.MaxHealth, c.MaxMana
	c.MaxXp = c.xpForNextLevel()

	return c, nil
}

func (c *Character) maxHealthFor(level int) int {
	total := Races[c.Race].BonusHP + Classes[c.Class].BaseHP +
		Classes[c.Class].HPPerLevel*(level-1)
	return max(total, 1)
}

func (c *Character) maxManaFor(level int) int {
	total := Races[c.Race].BonusMana + Classes[c.Class].BaseMana +
		Classes[c.Class].ManaPerLevel*(level-1)
	return max(total, 0)
}

func (c *Character) xpForNextLevel() int {
	if c.Level >= MaxLevel {
		return 0
	}
	return int(xpThresholds[c.Level+1] - xpThresholds[c.Level])
}

func (c *Character) syncXp() {
	if c.Level >= MaxLevel {
		c.Xp, c.MaxXp = 0, 0
		return
	}
	c.Xp = int(c.TotalXp - xpThresholds[c.Level])
	c.MaxXp = c.xpForNextLevel()
}

func (c *Character) setLevel(level int) {
	level = min(max(level, 1), MaxLevel)
	if level == c.Level {
		c.syncXp()
		return
	}

	c.Level = level
	c.syncXp()
	c.MaxHealth = c.maxHealthFor(level)
	c.MaxMana = c.maxManaFor(level)
	c.Health, c.Mana = c.MaxHealth, c.MaxMana
}

func (c *Character) AddXp(xpAmount int) (int, error) {
	if xpAmount <= 0 {
		return 0, ErrInvalidXPAmount
	}
	if c.Level >= MaxLevel {
		return 0, ErrCharacterAlreadyReachedMaxLevel
	}

	before := c.Level
	c.TotalXp += int64(xpAmount)
	c.setLevel(levelForTotalXP(c.TotalXp))

	return c.Level - before, nil
}

func (c *Character) LevelUp() error {
	if c.Level >= MaxLevel {
		return ErrCharacterAlreadyReachedMaxLevel
	}
	c.TotalXp = xpThresholds[c.Level+1]
	c.setLevel(c.Level + 1)
	return nil
}

func (c *Character) AddGold(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	c.Gold += amount
	return nil
}

func (c *Character) SpendGold(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if c.Gold < amount {
		return ErrNotEnoughGold
	}
	c.Gold -= amount
	return nil
}

func (c *Character) Suspend() error {
	if !c.IsActive {
		return ErrIsAlreadySuspended
	}
	c.IsActive = false
	return nil
}

func (c *Character) Activate() error {
	if c.IsActive {
		return ErrIsAlreadyActive
	}
	c.IsActive = true
	return nil
}

func normalizeName(name Name) (Name, error) {
	trimmed := Name(strings.ToLower(strings.TrimSpace(string(name))))
	if trimmed == "" {
		return "", ErrInvalidCharacterName
	}
	if utf8.RuneCountInString(string(trimmed)) > MaxNameLength {
		return "", ErrInvalidCharacterName
	}
	for _, r := range string(trimmed) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if r == ' ' || r == '-' || r == '_' {
			continue
		}
		return "", ErrInvalidCharacterName
	}
	return trimmed, nil
}

func NormalizeName(name Name) (Name, error) {
	return normalizeName(name)
}
