package models

import "time"

type ResourceType string

const (
	IronOre    ResourceType = "iron_ore"
	Titanium   ResourceType = "titanium"
	Helium3    ResourceType = "helium_3"
	RareEarth  ResourceType = "rare_earth"
	IceCrystal ResourceType = "ice_crystal"
	Credits    ResourceType = "credits"
	Fuel       ResourceType = "fuel"
)

type CelestialType string

const (
	Star          CelestialType = "star"
	Planet        CelestialType = "planet"
	AsteroidBelt  CelestialType = "asteroid_belt"
	GasGiant      CelestialType = "gas_giant"
	Terrestrial   CelestialType = "terrestrial"
)

type ShipType string

const (
	CargoShip   ShipType = "cargo"
	Frigate     ShipType = "frigate"
	MiningShip  ShipType = "mining"
)

type TechType string

const (
	MiningEfficiency  TechType = "mining_efficiency"
	RefiningTech      TechType = "refining_tech"
	EngineImprovement TechType = "engine_improvement"
	WeaponUpgrade     TechType = "weapon_upgrade"
)

type OrderType string

const (
	BuyOrder  OrderType = "buy"
	SellOrder OrderType = "sell"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
	OrderPartial   OrderStatus = "partial"
)

type EventType string

const (
	AsteroidStorm    EventType = "asteroid_storm"
	NewVeinDiscovery EventType = "new_vein_discovery"
	PirateInvasion   EventType = "pirate_invasion"
	TradeEmbargo     EventType = "trade_embargo"
	TechBreakthrough EventType = "tech_breakthrough"
)

type GamePhase string

const (
	PhasePlanning GamePhase = "planning"
	PhaseAction   GamePhase = "action"
	PhaseResult   GamePhase = "result"
)

type Resources map[ResourceType]float64

type CelestialBody struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Type           CelestialType     `json:"type"`
	GalaxyID       string            `json:"galaxy_id"`
	PositionX      float64           `json:"position_x"`
	PositionY      float64           `json:"position_y"`
	Resources      Resources         `json:"resources"`
	MaxResources   Resources         `json:"max_resources"`
	MiningStations []*MiningStation  `json:"mining_stations"`
	Refineries     []*Refinery       `json:"refineries"`
	Shipyards      []*Shipyard       `json:"shipyards"`
	OwnerPlayerID  string            `json:"owner_player_id"`
	HasExchange    bool              `json:"has_exchange"`
}

type Lane struct {
	ID         string  `json:"id"`
	FromBodyID string  `json:"from_body_id"`
	ToBodyID   string  `json:"to_body_id"`
	Distance   float64 `json:"distance"`
	Safe       bool    `json:"safe"`
}

type Galaxy struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	CelestialBodies []*CelestialBody `json:"celestial_bodies"`
	Lanes           []*Lane          `json:"lanes"`
}

type GameMap struct {
	Galaxies []*Galaxy `json:"galaxies"`
}

type MiningStation struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	BodyID         string       `json:"body_id"`
	PlayerID       string       `json:"player_id"`
	ResourceType   ResourceType `json:"resource_type"`
	Efficiency     float64      `json:"efficiency"`
	Level          int          `json:"level"`
	Workers        int          `json:"workers"`
	Maintenance    float64      `json:"maintenance"`
	OutputPerTurn  float64      `json:"output_per_turn"`
}

type Refinery struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	BodyID          string       `json:"body_id"`
	PlayerID        string       `json:"player_id"`
	InputResource   ResourceType `json:"input_resource"`
	OutputResource  ResourceType `json:"output_resource"`
	ConversionRate  float64      `json:"conversion_rate"`
	Level           int          `json:"level"`
	Maintenance     float64      `json:"maintenance"`
	ProcessingSpeed float64      `json:"processing_speed"`
	InputInventory  float64      `json:"input_inventory"`
	OutputInventory float64      `json:"output_inventory"`
}

type Shipyard struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	BodyID       string  `json:"body_id"`
	PlayerID     string  `json:"player_id"`
	Level        int     `json:"level"`
	BuildQueue   []*Ship `json:"build_queue"`
	Maintenance  float64 `json:"maintenance"`
}

type Ship struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          ShipType  `json:"type"`
	PlayerID      string    `json:"player_id"`
	FleetID       string    `json:"fleet_id"`
	Health        float64   `json:"health"`
	MaxHealth     float64   `json:"max_health"`
	Attack        float64   `json:"attack"`
	Defense       float64   `json:"defense"`
	CargoCapacity float64   `json:"cargo_capacity"`
	Cargo         Resources `json:"cargo"`
	Speed         float64   `json:"speed"`
	Fuel          float64   `json:"fuel"`
	MaxFuel       float64   `json:"max_fuel"`
	FuelPerTurn   float64   `json:"fuel_per_turn"`
	BuildCost     Resources `json:"build_cost"`
	BuildTime     int       `json:"build_time"`
	TurnsLeft     int       `json:"turns_left"`
}

type Fleet struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PlayerID        string    `json:"player_id"`
	Ships           []*Ship   `json:"ships"`
	CurrentBodyID   string    `json:"current_body_id"`
	DestinationID   string    `json:"destination_id"`
	IsMoving        bool      `json:"is_moving"`
	TurnsRemaining  int       `json:"turns_remaining"`
	TotalCargo      Resources `json:"total_cargo"`
	TotalCargoCap   float64   `json:"total_cargo_cap"`
	TotalAttack     float64   `json:"total_attack"`
	TotalDefense    float64   `json:"total_defense"`
	CommandingOfficer string  `json:"commanding_officer"`
}

type Order struct {
	ID         string       `json:"id"`
	PlayerID   string       `json:"player_id"`
	ExchangeID string       `json:"exchange_id"`
	Type       OrderType    `json:"type"`
	Resource   ResourceType `json:"resource"`
	Quantity   float64      `json:"quantity"`
	Price      float64      `json:"price"`
	Status     OrderStatus  `json:"status"`
	FilledQty  float64      `json:"filled_qty"`
	CreatedTurn int         `json:"created_turn"`
}

type Exchange struct {
	ID        string            `json:"id"`
	BodyID    string            `json:"body_id"`
	Name      string            `json:"name"`
	BuyOrders map[string]*Order `json:"buy_orders"`
	SellOrders map[string]*Order `json:"sell_orders"`
	Prices    map[ResourceType]float64 `json:"prices"`
	FeeRate   float64           `json:"fee_rate"`
}

type Player struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	CompanyName       string              `json:"company_name"`
	Credits           float64             `json:"credits"`
	Resources         Resources           `json:"resources"`
	Stations          []*MiningStation    `json:"stations"`
	Refineries        []*Refinery         `json:"refineries"`
	Shipyards         []*Shipyard         `json:"shipyards"`
	Fleets            []*Fleet            `json:"fleets"`
	Ships             []*Ship             `json:"ships"`
	TechTree          *TechTree           `json:"tech_tree"`
	Stocks            []*Stock            `json:"stocks"`
	IsAI              bool                `json:"is_ai"`
	IsBankrupt        bool                `json:"is_bankrupt"`
	IsDefeated        bool                `json:"is_defeated"`
	Reputation        float64             `json:"reputation"`
	DailyIncome       float64             `json:"daily_income"`
	DailyExpense      float64             `json:"daily_expense"`
	TotalTradeProfit  float64             `json:"total_trade_profit"`
	MilitaryStrength  float64             `json:"military_strength"`
	NegativeTurns     int                 `json:"negative_turns"`
}

type Stock struct {
	ID         string  `json:"id"`
	PlayerID   string  `json:"player_id"`
	IssuerID   string  `json:"issuer_id"`
	Shares     int     `json:"shares"`
	SharePrice float64 `json:"share_price"`
	TotalShares int    `json:"total_shares"`
	Dividend   float64 `json:"dividend"`
}

type TechLevel struct {
	Type  TechType `json:"type"`
	Level int      `json:"level"`
	MaxLevel int   `json:"max_level"`
}

type TechTree struct {
	PlayerID     string                `json:"player_id"`
	Techs        map[TechType]*TechLevel `json:"techs"`
	Researching  TechType              `json:"researching"`
	Progress     float64               `json:"progress"`
	ResearchCost Resources             `json:"research_cost"`
}

type RandomEvent struct {
	ID          string    `json:"id"`
	Type        EventType `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
	TurnsLeft   int       `json:"turns_left"`
	Active      bool      `json:"active"`
	Global      bool      `json:"global"`
	TargetID    string    `json:"target_id"`
	Effect      string    `json:"effect"`
}

type PlayerRanking struct {
	PlayerID         string  `json:"player_id"`
	PlayerName       string  `json:"player_name"`
	CompanyName      string  `json:"company_name"`
	Score            float64 `json:"score"`
	NetWorth         float64 `json:"net_worth"`
	TotalTradeProfit float64 `json:"total_trade_profit"`
	TechLevelSum     int     `json:"tech_level_sum"`
	MilitaryStrength float64 `json:"military_strength"`
	Rank             int     `json:"rank"`
	IsBankrupt       bool    `json:"is_bankrupt"`
	IsDefeated       bool    `json:"is_defeated"`
}

type Bid struct {
	ID         string  `json:"id"`
	AuctionID  string  `json:"auction_id"`
	BidderID   string  `json:"bidder_id"`
	Amount     float64 `json:"amount"`
	TurnSubmitted int `json:"turn_submitted"`
}

type Blockade struct {
	ID              string   `json:"id"`
	PlayerID        string   `json:"player_id"`
	TargetBodyID    string   `json:"target_body_id"`
	FleetID         string   `json:"fleet_id"`
	TurnsRemaining  int      `json:"turns_remaining"`
	TollRate        float64  `json:"toll_rate"`
	Active          bool     `json:"active"`
}

type GameState struct {
	ID             string            `json:"id"`
	Turn           int               `json:"turn"`
	Phase          GamePhase         `json:"phase"`
	Players        []*Player         `json:"players"`
	GameMap        *GameMap          `json:"game_map"`
	Exchanges      []*Exchange       `json:"exchanges"`
	RandomEvents   []*RandomEvent    `json:"random_events"`
	Blockades      []*Blockade       `json:"blockades"`
	Bids           []*Bid            `json:"bids"`
	Started        bool              `json:"started"`
	GameOver       bool              `json:"game_over"`
	WinnerID       string            `json:"winner_id"`
	MaxTurns       int               `json:"max_turns"`
	Seed           int64             `json:"seed"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	WinCondition   string            `json:"win_condition"`
	TargetCredits  float64           `json:"target_credits"`
	FinalRankings  []*PlayerRanking  `json:"final_rankings"`
}
