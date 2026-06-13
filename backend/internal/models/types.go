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

type ResourceChange struct {
	ResourceType ResourceType `json:"resource_type"`
	ResourceName string       `json:"resource_name"`
	Produced     float64      `json:"produced"`
	Consumed     float64      `json:"consumed"`
	Traded       float64      `json:"traded"`
	NetChange    float64      `json:"net_change"`
}

type FinanceRecord struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Label    string  `json:"label"`
}

type FinanceSummary struct {
	IncomeItems  []*FinanceRecord `json:"income_items"`
	ExpenseItems []*FinanceRecord `json:"expense_items"`
	TotalIncome  float64          `json:"total_income"`
	TotalExpense float64          `json:"total_expense"`
	NetIncome    float64          `json:"net_income"`
	CurrentBalance float64        `json:"current_balance"`
}

type FleetMovement struct {
	FleetID        string `json:"fleet_id"`
	FleetName      string `json:"fleet_name"`
	FromBodyID     string `json:"from_body_id"`
	FromBodyName   string `json:"from_body_name"`
	ToBodyID       string `json:"to_body_id"`
	ToBodyName     string `json:"to_body_name"`
	TurnsRemaining int    `json:"turns_remaining"`
	Arrived        bool   `json:"arrived"`
}

type CombatRecord struct {
	CombatID        string  `json:"combat_id"`
	AttackerID      string  `json:"attacker_id"`
	AttackerName    string  `json:"attacker_name"`
	DefenderID      string  `json:"defender_id"`
	DefenderName    string  `json:"defender_name"`
	Winner          string  `json:"winner"`
	AttackerLosses  int     `json:"attacker_losses"`
	DefenderLosses  int     `json:"defender_losses"`
	AttackerDamage  float64 `json:"attacker_damage"`
	DefenderDamage  float64 `json:"defender_damage"`
	LootSummary     string  `json:"loot_summary"`
}

type PirateAttackRecord struct {
	FleetID       string  `json:"fleet_id"`
	FleetName     string  `json:"fleet_name"`
	LocationName  string  `json:"location_name"`
	PlayerLosses  int     `json:"player_losses"`
	PirateLosses  int     `json:"pirate_losses"`
	Defended      bool    `json:"defended"`
	DamageTaken   float64 `json:"damage_taken"`
}

type FleetActivity struct {
	Movements     []*FleetMovement     `json:"movements"`
	Combats       []*CombatRecord      `json:"combats"`
	PirateAttacks []*PirateAttackRecord `json:"pirate_attacks"`
}

type MarketPriceChange struct {
	ResourceType ResourceType `json:"resource_type"`
	ResourceName string       `json:"resource_name"`
	OldPrice     float64      `json:"old_price"`
	NewPrice     float64      `json:"new_price"`
	ChangePercent float64     `json:"change_percent"`
}

type RandomEventRecord struct {
	EventID     string    `json:"event_id"`
	EventType   EventType `json:"event_type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsGlobal    bool      `json:"is_global"`
	TargetID    string    `json:"target_id"`
	AffectsMe   bool      `json:"affects_me"`
}

type RankingChangeEntry struct {
	PlayerID       string  `json:"player_id"`
	PlayerName     string  `json:"player_name"`
	CompanyName    string  `json:"company_name"`
	Score          float64 `json:"score"`
	Rank           int     `json:"rank"`
	PreviousRank   int     `json:"previous_rank"`
	RankChange     int     `json:"rank_change"`
	IsMe           bool    `json:"is_me"`
	IsBankrupt     bool    `json:"is_bankrupt"`
	IsDefeated     bool    `json:"is_defeated"`
}

type AllianceStatus string

const (
	AllianceActive  AllianceStatus = "active"
	AlliancePending AllianceStatus = "pending"
	AllianceDisbanded AllianceStatus = "disbanded"
)

type AllianceColor string

const (
	AllianceColorRed    AllianceColor = "#FF4444"
	AllianceColorBlue   AllianceColor = "#4488FF"
	AllianceColorGreen  AllianceColor = "#44FF44"
	AllianceColorYellow AllianceColor = "#FFFF44"
	AllianceColorPurple AllianceColor = "#FF44FF"
	AllianceColorCyan   AllianceColor = "#44FFFF"
)

type TradeAgreementStatus string

const (
	TradeAgreementActive  TradeAgreementStatus = "active"
	TradeAgreementExpired TradeAgreementStatus = "expired"
)

type MilitaryActionStatus string

const (
	MilitaryActionRecruiting MilitaryActionStatus = "recruiting"
	MilitaryActionInProgress MilitaryActionStatus = "in_progress"
	MilitaryActionCompleted  MilitaryActionStatus = "completed"
	MilitaryActionCancelled  MilitaryActionStatus = "cancelled"
)

type Alliance struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	LeaderID     string         `json:"leader_id"`
	MemberIDs    []string       `json:"member_ids"`
	Color        AllianceColor  `json:"color"`
	Status       AllianceStatus `json:"status"`
	CreatedTurn  int            `json:"created_turn"`
	InviteExpiry map[string]int `json:"invite_expiry"`
}

type TradeAgreement struct {
	ID           string                `json:"id"`
	PlayerID1    string                `json:"player_id_1"`
	PlayerID2    string                `json:"player_id_2"`
	AllianceID   string                `json:"alliance_id"`
	Status       TradeAgreementStatus  `json:"status"`
	CreatedTurn  int                   `json:"created_turn"`
	ExpiryTurn   int                   `json:"expiry_turn"`
}

type MilitaryParticipant struct {
	PlayerID string `json:"player_id"`
	FleetID  string `json:"fleet_id"`
	Joined   bool   `json:"joined"`
}

type JointMilitaryAction struct {
	ID            string                 `json:"id"`
	AllianceID    string                 `json:"alliance_id"`
	InitiatorID   string                 `json:"initiator_id"`
	TargetPlayerID string                `json:"target_player_id"`
	TargetBodyID  string                 `json:"target_body_id"`
	Status        MilitaryActionStatus   `json:"status"`
	Participants  []*MilitaryParticipant `json:"participants"`
	CreatedTurn   int                    `json:"created_turn"`
	DeadlineTurn  int                    `json:"deadline_turn"`
	ArrivalTurn   int                    `json:"arrival_turn"`
	TotalAttack   float64                `json:"total_attack"`
}

type DiplomacyRelation struct {
	Player1ID string  `json:"player1_id"`
	Player2ID string  `json:"player2_id"`
	Value     float64 `json:"value"`
}

type DiplomacyChange struct {
	PlayerID  string  `json:"player_id"`
	OldValue  float64 `json:"old_value"`
	NewValue  float64 `json:"new_value"`
	Change    float64 `json:"change"`
	Reason    string  `json:"reason"`
}

type DiplomacySection struct {
	Changes []*DiplomacyChange `json:"changes"`
}

type PlayerCooldown struct {
	PlayerID      string `json:"player_id"`
	CooldownTurns int    `json:"cooldown_turns"`
	LeftTurn      int    `json:"left_turn"`
}

type TurnReport struct {
	Turn              int                    `json:"turn"`
	PlayerID          string                 `json:"player_id"`
	PlayerName        string                 `json:"player_name"`
	ResourceChanges   []*ResourceChange      `json:"resource_changes"`
	Finance           *FinanceSummary        `json:"finance"`
	FleetActivity     *FleetActivity         `json:"fleet_activity"`
	MarketChanges     []*MarketPriceChange   `json:"market_changes"`
	RandomEvents      []*RandomEventRecord   `json:"random_events"`
	Rankings          []*RankingChangeEntry  `json:"rankings"`
	Diplomacy         *DiplomacySection      `json:"diplomacy"`
	GeneratedAt       time.Time              `json:"generated_at"`
}

type GameState struct {
	ID                  string               `json:"id"`
	Turn                int                  `json:"turn"`
	Phase               GamePhase            `json:"phase"`
	Players             []*Player            `json:"players"`
	GameMap             *GameMap             `json:"game_map"`
	Exchanges           []*Exchange          `json:"exchanges"`
	RandomEvents        []*RandomEvent       `json:"random_events"`
	Blockades           []*Blockade          `json:"blockades"`
	Bids                []*Bid               `json:"bids"`
	Alliances           []*Alliance          `json:"alliances"`
	TradeAgreements     []*TradeAgreement    `json:"trade_agreements"`
	JointMilitaryActions []*JointMilitaryAction `json:"joint_military_actions"`
	DiplomacyRelations  []*DiplomacyRelation `json:"diplomacy_relations"`
	PlayerCooldowns     []*PlayerCooldown    `json:"player_cooldowns"`
	Started             bool                 `json:"started"`
	GameOver            bool                 `json:"game_over"`
	WinnerID            string               `json:"winner_id"`
	MaxTurns            int                  `json:"max_turns"`
	Seed                int64                `json:"seed"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	WinCondition        string               `json:"win_condition"`
	TargetCredits       float64              `json:"target_credits"`
	FinalRankings       []*PlayerRanking     `json:"final_rankings"`
}
