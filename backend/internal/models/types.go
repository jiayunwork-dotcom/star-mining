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

type WarStatus string

const (
	WarActive    WarStatus = "active"
	WarSurrendered WarStatus = "surrendered"
)

type AllianceWar struct {
	ID                   string    `json:"id"`
	AttackerAllianceID   string    `json:"attacker_alliance_id"`
	DefenderAllianceID   string    `json:"defender_alliance_id"`
	DeclaredTurn         int       `json:"declared_turn"`
	Status               WarStatus `json:"status"`
	SurrenderedAllianceID string   `json:"surrendered_alliance_id,omitempty"`
	SurrenderTurn        int       `json:"surrender_turn,omitempty"`
	AttackerTotalAssets  float64   `json:"attacker_total_assets"`
	DefenderTotalAssets  float64   `json:"defender_total_assets"`
}

type SanctionProposalStatus string

const (
	SanctionProposalPending  SanctionProposalStatus = "pending"
	SanctionProposalActive   SanctionProposalStatus = "active"
	SanctionProposalExpired  SanctionProposalStatus = "expired"
	SanctionProposalRejected SanctionProposalStatus = "rejected"
)

type SanctionProposal struct {
	ID           string                 `json:"id"`
	InitiatorID  string                 `json:"initiator_id"`
	TargetID     string                 `json:"target_id"`
	SeconderIDs  []string               `json:"seconder_ids"`
	Status       SanctionProposalStatus `json:"status"`
	CreatedTurn  int                    `json:"created_turn"`
	ExpiryTurn   int                    `json:"expiry_turn"`
}

type Sanction struct {
	ID          string  `json:"id"`
	TargetID    string  `json:"target_id"`
	InitiatorID string  `json:"initiator_id"`
	CreatedTurn int     `json:"created_turn"`
	ExpiryTurn  int     `json:"expiry_turn"`
	TurnsLeft   int     `json:"turns_left"`
}

type WarEvent struct {
	WarID                string  `json:"war_id"`
	EventType            string  `json:"event_type"`
	AttackerAllianceName string  `json:"attacker_alliance_name,omitempty"`
	DefenderAllianceName string  `json:"defender_alliance_name,omitempty"`
	SurrenderAllianceName string `json:"surrender_alliance_name,omitempty"`
	Turn                 int     `json:"turn"`
	Reparations          []ReparationDetail `json:"reparations,omitempty"`
	SurrenderSuggested   bool    `json:"surrender_suggested,omitempty"`
	SurrenderAllianceID  string  `json:"surrender_alliance_id,omitempty"`
}

type ReparationDetail struct {
	PayerID   string  `json:"payer_id"`
	PayerName string  `json:"payer_name"`
	PayeeID   string  `json:"payee_id"`
	PayeeName string  `json:"payee_name"`
	Amount    float64 `json:"amount"`
}

type SanctionEvent struct {
	SanctionID    string  `json:"sanction_id"`
	EventType     string  `json:"event_type"`
	TargetID      string  `json:"target_id"`
	TargetName    string  `json:"target_name"`
	InitiatorID   string  `json:"initiator_id"`
	InitiatorName string  `json:"initiator_name"`
	Turn          int     `json:"turn"`
	MaintenanceFee float64 `json:"maintenance_fee,omitempty"`
}

type DiplomacySection struct {
	Changes       []*DiplomacyChange `json:"changes"`
	WarEvents     []*WarEvent        `json:"war_events"`
	SanctionEvents []*SanctionEvent  `json:"sanction_events"`
}

type SpyLevel string

const (
	SpyLevelJunior  SpyLevel = "junior"
	SpyLevelMiddle  SpyLevel = "middle"
	SpyLevelSenior  SpyLevel = "senior"
)

type SpyStatus string

const (
	SpyStatusIdle      SpyStatus = "idle"
	SpyStatusOnMission SpyStatus = "on_mission"
	SpyStatusCaptured  SpyStatus = "captured"
)

type SpyMissionType string

const (
	MissionStealTech    SpyMissionType = "steal_tech"
	MissionEconSabotage SpyMissionType = "econ_sabotage"
	MissionIntelGather  SpyMissionType = "intel_gather"
	MissionTurncoat     SpyMissionType = "turncoat"
	MissionDiploPressure SpyMissionType = "diplo_pressure"
)

type CounterSpyLevel string

const (
	CounterSpyLow    CounterSpyLevel = "low"
	CounterSpyMedium CounterSpyLevel = "medium"
	CounterSpyHigh   CounterSpyLevel = "high"
)

type Spy struct {
	ID               string        `json:"id"`
	PlayerID         string        `json:"player_id"`
	Level            SpyLevel      `json:"level"`
	Exposure         int           `json:"exposure"`
	Status           SpyStatus     `json:"status"`
	CompletedMissions int          `json:"completed_missions"`
	TurnRecruited    int           `json:"turn_recruited"`
	IsDoubleAgent    bool          `json:"is_double_agent"`
	DoubleAgentFor   string        `json:"double_agent_for,omitempty"`
}

type SpyMission struct {
	ID            string         `json:"id"`
	SpyID         string         `json:"spy_id"`
	OwnerPlayerID string         `json:"owner_player_id"`
	TargetPlayerID string        `json:"target_player_id"`
	ThirdPartyID  string         `json:"third_party_id,omitempty"`
	MissionType   SpyMissionType `json:"mission_type"`
	TurnSubmitted int            `json:"turn_submitted"`
	Success       bool           `json:"success"`
	Result        string         `json:"result,omitempty"`
	Resolved      bool           `json:"resolved"`
}

type Intelligence struct {
	ID              string   `json:"id"`
	OwnerPlayerID   string   `json:"owner_player_id"`
	SourcePlayerID  string   `json:"source_player_id"`
	SourceSpyID     string   `json:"source_spy_id,omitempty"`
	Content         string   `json:"content"`
	Summary         string   `json:"summary"`
	TurnAcquired    int      `json:"turn_acquired"`
	ExpiryTurn      int      `json:"expiry_turn"`
	IntelType       string   `json:"intel_type"`
	Expired         bool     `json:"expired"`
	TargetCredits   float64  `json:"target_credits,omitempty"`
	TargetFleetCount int     `json:"target_fleet_count,omitempty"`
	TargetTechs     []string `json:"target_techs,omitempty"`
	TargetAlliances []string `json:"target_alliances,omitempty"`
}

type IntelMarketListing struct {
	ID            string  `json:"id"`
	SellerID      string  `json:"seller_id"`
	IntelID       string  `json:"intel_id"`
	Price         float64 `json:"price"`
	BasePrice     float64 `json:"base_price"`
	CreatedTurn   int     `json:"created_turn"`
	RemainingTurns int    `json:"remaining_turns"`
	Active        bool    `json:"active"`
}

type SpyResult struct {
	SpyID              string  `json:"spy_id"`
	MissionType        SpyMissionType `json:"mission_type"`
	TargetPlayerID     string  `json:"target_player_id"`
	Success            bool    `json:"success"`
	Captured           bool    `json:"captured"`
	ExposureGain       int     `json:"exposure_gain"`
	Result             string  `json:"result"`
	StolenTech         string  `json:"stolen_tech,omitempty"`
	DamageAmount       float64 `json:"damage_amount,omitempty"`
	IntelGathered      bool    `json:"intel_gathered,omitempty"`
	TurncoatSpyID      string  `json:"turncoat_spy_id,omitempty"`
	DiploTarget        string  `json:"diplo_target,omitempty"`
	DiploThirdParty    string  `json:"diplo_third_party,omitempty"`
	IsDoubleAgentFail  bool    `json:"is_double_agent_fail,omitempty"`
}

type CounterSpyResult struct {
	Detected       bool   `json:"detected"`
	Identified     bool   `json:"identified"`
	CounterDone    bool   `json:"counter_done"`
	TargetPlayerID string `json:"target_player_id"`
	SourcePlayerID string `json:"source_player_id,omitempty"`
	SpyID          string `json:"spy_id,omitempty"`
	ExposureAdded  int    `json:"exposure_added,omitempty"`
	RemovedSpyID   string `json:"removed_spy_id,omitempty"`
}

type SpyNotification struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Turn    int     `json:"turn"`
	SpyID   string  `json:"spy_id,omitempty"`
	Details string  `json:"details,omitempty"`
}

type SpySection struct {
	MissionResults     []*SpyResult      `json:"mission_results"`
	CounterSpyResults  []*CounterSpyResult `json:"counter_spy_results"`
	ExpiredIntel       []string          `json:"expired_intel"`
	LevelUpNotifications []*SpyNotification `json:"level_up_notifications"`
	Notifications      []*SpyNotification `json:"notifications"`
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
	Spy               *SpySection            `json:"spy"`
	GeneratedAt       time.Time              `json:"generated_at"`
}

type CounterSpySetting struct {
	PlayerID string         `json:"player_id"`
	Level    CounterSpyLevel `json:"level"`
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
	Alliances           []*Alliance           `json:"alliances"`
	TradeAgreements     []*TradeAgreement      `json:"trade_agreements"`
	JointMilitaryActions []*JointMilitaryAction `json:"joint_military_actions"`
	DiplomacyRelations  []*DiplomacyRelation   `json:"diplomacy_relations"`
	PlayerCooldowns     []*PlayerCooldown      `json:"player_cooldowns"`
	AllianceWars        []*AllianceWar         `json:"alliance_wars"`
	SanctionProposals   []*SanctionProposal    `json:"sanction_proposals"`
	ActiveSanctions     []*Sanction            `json:"active_sanctions"`
	Spies               []*Spy                 `json:"spies"`
	SpyMissions         []*SpyMission          `json:"spy_missions"`
	Intelligences       []*Intelligence        `json:"intelligences"`
	IntelMarketListings []*IntelMarketListing  `json:"intel_market_listings"`
	CounterSpySettings  []*CounterSpySetting   `json:"counter_spy_settings"`
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
