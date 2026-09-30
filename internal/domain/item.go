package domain

type Rarity string
type ItemID int

const (
	Common    Rarity = "common"
	Uncommon  Rarity = "uncommon"
	Rare      Rarity = "rare"
	Epic      Rarity = "epic"
	Legendary Rarity = "legendary"
)

type Item struct {
	ID        ItemID `json:"id"`
	Name      string `json:"name"`
	Rarity    Rarity `json:"rarity"`
	BuyPrice  int    `json:"buy_price"`
	SellPrice int    `json:"sell_price"`
	Stackable bool   `json:"stackable"`
	MaxStack  int    `json:"max_stack"`
}

var Items = map[ItemID]Item{
	0: {ID: 0, Name: "Healing Potion", Rarity: Common, BuyPrice: 25, SellPrice: 10, Stackable: true, MaxStack: 20},
	1: {ID: 1, Name: "Mana Potion", Rarity: Common, BuyPrice: 25, SellPrice: 10, Stackable: true, MaxStack: 20},
	2: {ID: 2, Name: "Iron Sword", Rarity: Uncommon, BuyPrice: 150, SellPrice: 60},
	3: {ID: 3, Name: "Oak Staff", Rarity: Uncommon, BuyPrice: 150, SellPrice: 60},
	4: {ID: 4, Name: "Thunderfury, Blessed Blade", Rarity: Legendary, BuyPrice: 5000, SellPrice: 2000},
}
