package domain

type Race string

const (
	Human    Race = "human"
	Orc      Race = "orc"
	BloodElf Race = "blood elf"
	NightElf Race = "night elf"
)

type RaceSpec struct {
	BonusHP   int
	BonusMana int
}

var Races = map[Race]RaceSpec{
	Human:    {BonusHP: 5, BonusMana: 5},
	Orc:      {BonusHP: 15, BonusMana: -5},
	BloodElf: {BonusHP: -5, BonusMana: 15},
	NightElf: {BonusHP: 0, BonusMana: 10},
}
