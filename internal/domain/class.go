package domain

type Class string

const (
	Warrior Class = "warrior"
	Mage    Class = "mage"
	Shaman  Class = "shaman"
	Paladin Class = "paladin"
	Druid   Class = "druid"
)

type ClassSpec struct {
	BaseHP       int
	BaseMana     int
	HPPerLevel   int
	ManaPerLevel int
}

var Classes = map[Class]ClassSpec{
	Warrior: {BaseHP: 120, BaseMana: 20, HPPerLevel: 15, ManaPerLevel: 2},
	Mage:    {BaseHP: 60, BaseMana: 120, HPPerLevel: 6, ManaPerLevel: 18},
	Shaman:  {BaseHP: 90, BaseMana: 80, HPPerLevel: 10, ManaPerLevel: 12},
	Paladin: {BaseHP: 110, BaseMana: 60, HPPerLevel: 13, ManaPerLevel: 9},
	Druid:   {BaseHP: 85, BaseMana: 90, HPPerLevel: 19, ManaPerLevel: 13},
}
