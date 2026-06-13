package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"star-mining/internal/models"
)

const (
	DefaultMaxTurns     = 60
	DefaultStartCredits = 5000
	DefaultWinCredits   = 100000
	InterestRate        = 0.005
)

type GameInstance struct {
	ID       string
	State    *models.GameState
	players  map[string]*models.Player
	seed     int64
	started  bool
	gameOver bool
	maxTurns int
	winnerID string
	rng      *rand.Rand
}

func NewGameInstance(roomID string, seed int64) *GameInstance {
	return &GameInstance{
		ID:       roomID,
		State:    nil,
		players:  make(map[string]*models.Player),
		seed:     seed,
		started:  false,
		gameOver: false,
		maxTurns: DefaultMaxTurns,
	}
}

func (gi *GameInstance) Initialize(playerIDs []string, playerNames map[string]string) error {
	seed := gi.seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	gi.rng = rand.New(rand.NewSource(seed))

	mapGen := NewMapGenerator(seed)
	gameMap := mapGen.GenerateMap()

	var exchanges []*models.Exchange
	for _, galaxy := range gameMap.Galaxies {
		for _, body := range galaxy.CelestialBodies {
			if body.HasExchange {
				exchange := NewExchange(body.ID)
				exchanges = append(exchanges, exchange)
			}
		}
	}

	if len(exchanges) == 0 && len(gameMap.Galaxies) > 0 {
		firstGalaxy := gameMap.Galaxies[0]
		if len(firstGalaxy.CelestialBodies) > 0 {
			body := firstGalaxy.CelestialBodies[0]
			body.HasExchange = true
			exchange := NewExchange(body.ID)
			exchanges = append(exchanges, exchange)
		}
	}

	players := make([]*models.Player, 0, len(playerIDs))
	for i, playerID := range playerIDs {
		name := playerNames[playerID]
		if name == "" {
			name = fmt.Sprintf("Player %s", playerID[:8])
		}

		player := &models.Player{
			ID:          playerID,
			Name:        name,
			CompanyName: fmt.Sprintf("%s Corp", name),
			Credits:     DefaultStartCredits,
			Resources:   make(models.Resources),
			Stations:    make([]*models.MiningStation, 0),
			Refineries:  make([]*models.Refinery, 0),
			Shipyards:   make([]*models.Shipyard, 0),
			Fleets:      make([]*models.Fleet, 0),
			Ships:       make([]*models.Ship, 0),
			TechTree:    NewTechTree(playerID),
			Stocks:      make([]*models.Stock, 0),
			IsAI:        false,
			IsBankrupt:  false,
			IsDefeated:  false,
			Reputation:  50.0,
			DailyIncome: 0,
			DailyExpense: 0,
		}

		player.Resources[models.Fuel] = 200

		InitializePlayerStock(player)

		gi.players[playerID] = player
		players = append(players, player)

		if len(gameMap.Galaxies) > 0 && len(gameMap.Galaxies[0].CelestialBodies) > 0 {
			ship := NewShip(models.MiningShip, playerID)
			ship.ID = fmt.Sprintf("ship-%s-%d", playerID, i)
			player.Ships = append(player.Ships, ship)
		}
	}

	gi.State = &models.GameState{
		ID:            gi.ID,
		Turn:          1,
		Phase:         models.PhasePlanning,
		Players:       players,
		GameMap:       gameMap,
		Exchanges:     exchanges,
		RandomEvents:  make([]*models.RandomEvent, 0),
		Blockades:     make([]*models.Blockade, 0),
		Bids:          make([]*models.Bid, 0),
		Started:       false,
		GameOver:      false,
		WinnerID:      "",
		MaxTurns:      gi.maxTurns,
		Seed:          seed,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		WinCondition:  "net_worth",
		TargetCredits: DefaultWinCredits,
	}

	gi.started = false

	return nil
}

func (gi *GameInstance) Start() error {
	if gi.started {
		return fmt.Errorf("game already started")
	}
	gi.State.Started = true
	gi.started = true
	gi.State.UpdatedAt = time.Now()
	return nil
}

func (gi *GameInstance) ProcessTurn() error {
	if !gi.started {
		return fmt.Errorf("game not started")
	}
	if gi.gameOver {
		return fmt.Errorf("game is over")
	}

	playerMap := make(map[string]*models.Player)
	for _, player := range gi.State.Players {
		playerMap[player.ID] = player
	}

	for _, player := range gi.State.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		player.DailyIncome = 0
		player.DailyExpense = 0

		miningBonus := GetMiningBonus(player.TechTree)
		_ = ProcessAllMining(player, gi.State.GameMap, miningBonus)

		refiningBonus := GetRefiningBonus(player.TechTree)
		_ = ProcessAllRefining(player, refiningBonus)

		_ = ProcessAllShipyards(player)

		_, _ = ProcessResearch(player.TechTree)

		_ = ProcessAllFleetMoves(player)

		maintenance := PayMaintenance(player)
		player.DailyExpense = maintenance
	}

	for _, exchange := range gi.State.Exchanges {
		_ = MatchOrders(exchange, playerMap)
	}

	for _, player := range gi.State.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}
		player.MilitaryStrength = CalculateMilitaryStrength(player)
	}

	gi.State.RandomEvents = ProcessTurnEvents(gi.rng, gi.State.GameMap, gi.State.Players, gi.State.RandomEvents)

	gi.State.Blockades = UpdateBlockades(gi.State.Blockades)

	_ = DistributeDividends(gi.State.Players, gi.State.Turn)

	UpdateStockPrices(gi.State.Players, gi.State.Exchanges)

	gi.checkTakeovers()

	gi.applyInterest()

	for _, player := range gi.State.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		if player.Credits < 0 {
			player.NegativeTurns++
			if player.NegativeTurns >= 3 {
				player.IsBankrupt = true
			}
		} else {
			player.NegativeTurns = 0
		}
	}

	gi.State.Turn++
	gi.State.UpdatedAt = time.Now()

	gi.checkWinConditions()

	return nil
}

func (gi *GameInstance) checkTakeovers() {
	for _, acquirer := range gi.State.Players {
		if acquirer.IsDefeated || acquirer.IsBankrupt {
			continue
		}

		for _, target := range gi.State.Players {
			if target.ID == acquirer.ID {
				continue
			}
			if target.IsDefeated || target.IsBankrupt {
				continue
			}

			if CheckTakeover(acquirer, target, gi.State.Exchanges) {
				ExecuteTakeover(acquirer, target)
			}
		}
	}
}

func (gi *GameInstance) applyInterest() {
	for _, player := range gi.State.Players {
		if player.IsDefeated || player.IsBankrupt {
			continue
		}

		if player.Credits > 0 {
			interest := player.Credits * InterestRate
			player.Credits += interest
			player.DailyIncome += interest
		}
	}
}

func (gi *GameInstance) checkWinConditions() {
	if gi.gameOver {
		return
	}

	if gi.State.Turn > gi.State.MaxTurns {
		gi.State.WinCondition = "turn_limit"
		gi.EndGame()
		return
	}

	if gi.checkTakeoverVictory() {
		gi.State.WinCondition = "takeover"
		gi.EndGame()
		return
	}

	activePlayers := 0
	for _, player := range gi.State.Players {
		if !player.IsDefeated && !player.IsBankrupt {
			activePlayers++
		}
	}

	if activePlayers == 1 {
		gi.State.WinCondition = "last_standing"
		gi.EndGame()
		return
	}

	for _, player := range gi.State.Players {
		if player.Credits >= gi.State.TargetCredits {
			gi.State.WinCondition = "target_credits"
			gi.EndGame()
			return
		}
	}
}

func (gi *GameInstance) EndGame() {
	gi.gameOver = true
	gi.State.GameOver = true

	rankings := gi.calculateRankings()
	gi.State.FinalRankings = rankings

	if len(rankings) > 0 {
		var winnerID string
		for _, ranking := range rankings {
			if !ranking.IsDefeated && !ranking.IsBankrupt {
				winnerID = ranking.PlayerID
				break
			}
		}
		if winnerID == "" {
			winnerID = rankings[0].PlayerID
		}
		gi.winnerID = winnerID
		gi.State.WinnerID = winnerID
	}
}

func (gi *GameInstance) GetPlayer(playerID string) (*models.Player, bool) {
	player, exists := gi.players[playerID]
	return player, exists
}

func (gi *GameInstance) GetGameState() *models.GameState {
	return gi.State
}

func (gi *GameInstance) IsStarted() bool {
	return gi.started
}

func (gi *GameInstance) IsGameOver() bool {
	return gi.gameOver
}

func (gi *GameInstance) getPlayer(playerID string) *models.Player {
	player, exists := gi.players[playerID]
	if !exists {
		return nil
	}
	return player
}

func (gi *GameInstance) BuildStation(playerID string, bodyID string, resourceType models.ResourceType) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	body := FindBodyByID(gi.State.GameMap, bodyID)
	if body == nil {
		return fmt.Errorf("celestial body not found")
	}

	if _, exists := body.Resources[resourceType]; !exists {
		return fmt.Errorf("resource %s not found on this body", resourceType)
	}

	cost := CalculateStationCost(1)
	if player.Credits < cost {
		return fmt.Errorf("not enough credits")
	}

	player.Credits -= cost

	station := NewMiningStation(playerID, bodyID, resourceType)
	station.ID = fmt.Sprintf("station-%s-%d", playerID, len(player.Stations))

	player.Stations = append(player.Stations, station)
	body.MiningStations = append(body.MiningStations, station)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) BuildRefinery(playerID string, bodyID string, inputResource models.ResourceType, outputResource models.ResourceType) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	body := FindBodyByID(gi.State.GameMap, bodyID)
	if body == nil {
		return fmt.Errorf("celestial body not found")
	}

	recipes := GetRefineryRecipes()
	if validOutput, ok := recipes[inputResource]; !ok || validOutput != outputResource {
		return fmt.Errorf("invalid recipe: %s -> %s", inputResource, outputResource)
	}

	cost := CalculateRefineryCost(1)
	if player.Credits < cost {
		return fmt.Errorf("not enough credits")
	}

	player.Credits -= cost

	refinery := NewRefinery(playerID, bodyID, inputResource, outputResource)
	refinery.ID = fmt.Sprintf("refinery-%s-%d", playerID, len(player.Refineries))

	player.Refineries = append(player.Refineries, refinery)
	body.Refineries = append(body.Refineries, refinery)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) BuildShipyard(playerID, bodyID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	body := FindBodyByID(gi.State.GameMap, bodyID)
	if body == nil {
		return fmt.Errorf("celestial body not found")
	}

	cost := BaseShipyardCost
	if player.Credits < cost {
		return fmt.Errorf("not enough credits")
	}

	player.Credits -= cost

	shipyard := NewShipyard(playerID, bodyID)
	shipyard.ID = fmt.Sprintf("shipyard-%s-%d", playerID, len(player.Shipyards))

	player.Shipyards = append(player.Shipyards, shipyard)
	body.Shipyards = append(body.Shipyards, shipyard)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) BuildShip(playerID, shipyardID string, shipType models.ShipType) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var shipyard *models.Shipyard
	for _, sy := range player.Shipyards {
		if sy.ID == shipyardID {
			shipyard = sy
			break
		}
	}

	if shipyard == nil {
		return fmt.Errorf("shipyard not found")
	}

	err := BuildShip(shipyard, shipType, player)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) CreateFleet(playerID, name string, shipIDs []string) (*models.Fleet, error) {
	player := gi.getPlayer(playerID)
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return nil, fmt.Errorf("player is defeated or bankrupt")
	}

	fleet := NewFleet(playerID, name)
	fleet.ID = fmt.Sprintf("fleet-%s-%d", playerID, len(player.Fleets))

	var currentBodyID string
	for _, shipID := range shipIDs {
		for _, ship := range player.Ships {
			if ship.ID == shipID && ship.FleetID == "" {
				AddShipToFleet(fleet, ship)
				if currentBodyID == "" {
					for _, f := range player.Fleets {
						for _, s := range f.Ships {
							if s.ID == shipID {
								currentBodyID = f.CurrentBodyID
								break
							}
						}
					}
				}
				break
			}
		}
	}

	if currentBodyID == "" && len(player.Ships) > 0 && len(fleet.Ships) > 0 {
		if len(gi.State.GameMap.Galaxies) > 0 && len(gi.State.GameMap.Galaxies[0].CelestialBodies) > 0 {
			currentBodyID = gi.State.GameMap.Galaxies[0].CelestialBodies[0].ID
		}
	}
	fleet.CurrentBodyID = currentBodyID

	player.Fleets = append(player.Fleets, fleet)

	gi.State.UpdatedAt = time.Now()

	return fleet, nil
}

func (gi *GameInstance) MoveFleet(playerID, fleetID, targetBodyID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var fleet *models.Fleet
	for _, f := range player.Fleets {
		if f.ID == fleetID {
			fleet = f
			break
		}
	}

	if fleet == nil {
		return fmt.Errorf("fleet not found")
	}

	engineBonus := GetEngineBonus(player.TechTree)
	err := StartFleetMove(fleet, targetBodyID, gi.State.GameMap, engineBonus)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) ResearchTech(playerID string, techType models.TechType) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	if player.TechTree == nil {
		return fmt.Errorf("player has no tech tree")
	}

	err := StartResearch(player.TechTree, techType, player)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) PlaceBid(playerID string, bodyID string, amount float64) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	body := FindBodyByID(gi.State.GameMap, bodyID)
	if body == nil {
		return fmt.Errorf("celestial body not found")
	}

	if player.Credits < amount {
		return fmt.Errorf("not enough credits")
	}

	player.Credits -= amount

	bid := &models.Bid{
		ID:            fmt.Sprintf("bid-%s-%d", playerID, len(gi.State.Bids)),
		AuctionID:     bodyID,
		BidderID:      playerID,
		Amount:        amount,
		TurnSubmitted: gi.State.Turn,
	}

	gi.State.Bids = append(gi.State.Bids, bid)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) BlockLane(playerID string, laneID string, toll float64) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var lane *models.Lane
	for _, galaxy := range gi.State.GameMap.Galaxies {
		for _, l := range galaxy.Lanes {
			if l.ID == laneID {
				lane = l
				break
			}
		}
	}

	if lane == nil {
		return fmt.Errorf("lane not found")
	}

	if len(player.Fleets) == 0 {
		return fmt.Errorf("player has no fleets")
	}

	var fleet *models.Fleet
	for _, f := range player.Fleets {
		if f.CurrentBodyID == lane.FromBodyID || f.CurrentBodyID == lane.ToBodyID {
			fleet = f
			break
		}
	}

	if fleet == nil {
		return fmt.Errorf("no fleet available at the lane")
	}

	blockade := NewBlockade(playerID, lane.ToBodyID, fleet.ID, toll)
	gi.State.Blockades = append(gi.State.Blockades, blockade)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) HirePirates(playerID string, targetPlayerID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	targetPlayer := gi.getPlayer(targetPlayerID)
	if targetPlayer == nil {
		return fmt.Errorf("target player not found")
	}

	if targetPlayer.IsDefeated || targetPlayer.IsBankrupt {
		return fmt.Errorf("target player is defeated or bankrupt")
	}

	cost := 500.0
	if player.Credits < cost {
		return fmt.Errorf("not enough credits")
	}

	player.Credits -= cost
	player.Reputation -= 10

	if len(targetPlayer.Fleets) > 0 {
		fleet := targetPlayer.Fleets[0]
		combatBonus := GetCombatBonus(targetPlayer.TechTree)
		_, _ = PirateAttack(fleet, 2, combatBonus, gi.rng)
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) BuyStock(playerID string, targetPlayerID string, shares int) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	targetPlayer := gi.getPlayer(targetPlayerID)
	if targetPlayer == nil {
		return fmt.Errorf("target player not found")
	}

	var sellerStock *models.Stock
	for _, stock := range targetPlayer.Stocks {
		if stock.IssuerID == targetPlayerID {
			sellerStock = stock
			break
		}
	}

	if sellerStock == nil {
		return fmt.Errorf("target player has no stock")
	}

	err := BuyStock(player, targetPlayer, sellerStock, shares, gi.State.Exchanges)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) SellStock(playerID string, targetPlayerID string, shares int) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	targetPlayer := gi.getPlayer(targetPlayerID)
	if targetPlayer == nil {
		return fmt.Errorf("target player not found")
	}

	var sellerStock *models.Stock
	for _, stock := range player.Stocks {
		if stock.IssuerID == targetPlayerID {
			sellerStock = stock
			break
		}
	}

	if sellerStock == nil {
		return fmt.Errorf("player has no stock of target player")
	}

	err := SellStock(player, targetPlayer, sellerStock, shares, gi.State.Exchanges)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) ProposeTakeover(playerID string, targetPlayerID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	targetPlayer := gi.getPlayer(targetPlayerID)
	if targetPlayer == nil {
		return fmt.Errorf("target player not found")
	}

	if targetPlayer.IsDefeated || targetPlayer.IsBankrupt {
		return fmt.Errorf("target player is already defeated or bankrupt")
	}

	if CheckTakeover(player, targetPlayer, gi.State.Exchanges) {
		ExecuteTakeover(player, targetPlayer)
	} else {
		return fmt.Errorf("insufficient shares for takeover")
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) LoadCargo(playerID string, fleetID string, resource models.ResourceType, amount float64) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var fleet *models.Fleet
	for _, f := range player.Fleets {
		if f.ID == fleetID {
			fleet = f
			break
		}
	}

	if fleet == nil {
		return fmt.Errorf("fleet not found")
	}

	if fleet.IsMoving {
		return fmt.Errorf("fleet is moving")
	}

	for _, ship := range fleet.Ships {
		if amount <= 0 {
			break
		}

		currentLoad := 0.0
		for _, amt := range ship.Cargo {
			currentLoad += amt
		}

		availableSpace := ship.CargoCapacity - currentLoad
		if availableSpace <= 0 {
			continue
		}

		loadAmount := amount
		if loadAmount > availableSpace {
			loadAmount = availableSpace
		}

		if player.Resources[resource] < loadAmount {
			loadAmount = player.Resources[resource]
		}

		err := LoadCargo(ship, resource, loadAmount, player)
		if err != nil {
			return err
		}

		amount -= loadAmount
	}

	UpdateFleetStats(fleet)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) UnloadCargo(playerID string, fleetID string, resource models.ResourceType, amount float64) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var fleet *models.Fleet
	for _, f := range player.Fleets {
		if f.ID == fleetID {
			fleet = f
			break
		}
	}

	if fleet == nil {
		return fmt.Errorf("fleet not found")
	}

	if fleet.IsMoving {
		return fmt.Errorf("fleet is moving")
	}

	for _, ship := range fleet.Ships {
		if amount <= 0 {
			break
		}

		shipCargo := ship.Cargo[resource]
		if shipCargo <= 0 {
			continue
		}

		unloadAmount := amount
		if unloadAmount > shipCargo {
			unloadAmount = shipCargo
		}

		err := UnloadCargo(ship, resource, unloadAmount, player)
		if err != nil {
			return err
		}

		amount -= unloadAmount
	}

	UpdateFleetStats(fleet)

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) UpgradeStation(playerID string, stationID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var station *models.MiningStation
	for _, s := range player.Stations {
		if s.ID == stationID {
			station = s
			break
		}
	}

	if station == nil {
		return fmt.Errorf("station not found")
	}

	err := UpgradeStation(station, player)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) UpgradeRefinery(playerID string, refineryID string) error {
	player := gi.getPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var refinery *models.Refinery
	for _, r := range player.Refineries {
		if r.ID == refineryID {
			refinery = r
			break
		}
	}

	if refinery == nil {
		return fmt.Errorf("refinery not found")
	}

	err := UpgradeRefinery(refinery, player)
	if err != nil {
		return err
	}

	gi.State.UpdatedAt = time.Now()

	return nil
}

func (gi *GameInstance) PlaceBuyOrder(playerID string, exchangeID string, resource models.ResourceType, quantity float64, price float64) error {
	player, exists := gi.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var exchange *models.Exchange
	for _, ex := range gi.State.Exchanges {
		if ex.ID == exchangeID {
			exchange = ex
			break
		}
	}
	if exchange == nil {
		return fmt.Errorf("exchange not found")
	}

	totalCost := quantity * price
	fee := CalculateFee(exchange, totalCost)
	if player.Credits < totalCost+fee {
		return fmt.Errorf("not enough credits")
	}

	order := &models.Order{
		ID:          fmt.Sprintf("order-%s-%d", playerID, time.Now().UnixNano()),
		PlayerID:    playerID,
		ExchangeID:  exchangeID,
		Type:        models.BuyOrder,
		Resource:    resource,
		Quantity:    quantity,
		Price:       price,
		Status:      models.OrderPending,
		FilledQty:   0,
		CreatedTurn: gi.State.Turn,
	}

	player.Credits -= totalCost + fee

	return PlaceOrder(exchange, order)
}

func (gi *GameInstance) PlaceSellOrder(playerID string, exchangeID string, resource models.ResourceType, quantity float64, price float64) error {
	player, exists := gi.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found")
	}

	if player.IsDefeated || player.IsBankrupt {
		return fmt.Errorf("player is defeated or bankrupt")
	}

	var exchange *models.Exchange
	for _, ex := range gi.State.Exchanges {
		if ex.ID == exchangeID {
			exchange = ex
			break
		}
	}
	if exchange == nil {
		return fmt.Errorf("exchange not found")
	}

	if player.Resources[resource] < quantity {
		return fmt.Errorf("not enough resources")
	}

	order := &models.Order{
		ID:          fmt.Sprintf("order-%s-%d", playerID, time.Now().UnixNano()),
		PlayerID:    playerID,
		ExchangeID:  exchangeID,
		Type:        models.SellOrder,
		Resource:    resource,
		Quantity:    quantity,
		Price:       price,
		Status:      models.OrderPending,
		FilledQty:   0,
		CreatedTurn: gi.State.Turn,
	}

	player.Resources[resource] -= quantity

	return PlaceOrder(exchange, order)
}

func (gi *GameInstance) CancelOrder(playerID string, exchangeID string, orderID string) error {
	player, exists := gi.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found")
	}

	var exchange *models.Exchange
	for _, ex := range gi.State.Exchanges {
		if ex.ID == exchangeID {
			exchange = ex
			break
		}
	}
	if exchange == nil {
		return fmt.Errorf("exchange not found")
	}

	var order *models.Order
	if o, ok := exchange.BuyOrders[orderID]; ok && o.PlayerID == playerID {
		order = o
	} else if o, ok := exchange.SellOrders[orderID]; ok && o.PlayerID == playerID {
		order = o
	}

	if order == nil {
		return fmt.Errorf("order not found or not owned by player")
	}

	if order.Type == models.BuyOrder {
		remaining := order.Quantity - order.FilledQty
		refund := remaining * order.Price
		feeRefund := CalculateFee(exchange, refund)
		player.Credits += refund + feeRefund
	} else {
		remaining := order.Quantity - order.FilledQty
		player.Resources[order.Resource] += remaining
	}

	return CancelOrder(exchange, orderID)
}

func (gi *GameInstance) GetWinner() *models.Player {
	if gi.winnerID == "" {
		return nil
	}
	return gi.getPlayer(gi.winnerID)
}

func CalculateMilitaryStrength(player *models.Player) float64 {
	if player == nil {
		return 0
	}

	totalAttack := 0.0
	for _, ship := range player.Ships {
		totalAttack += ship.Attack
	}

	weaponBonus := GetCombatBonus(player.TechTree)

	return totalAttack * weaponBonus
}

func GetTotalTechLevel(player *models.Player) int {
	if player == nil || player.TechTree == nil {
		return 0
	}

	total := 0
	for _, tech := range player.TechTree.Techs {
		total += tech.Level
	}
	return total
}

func CalculateScore(player *models.Player, exchanges []*models.Exchange) float64 {
	if player == nil {
		return 0
	}

	netWorth := CalculateNetWorth(player, exchanges)
	totalTradeProfit := player.TotalTradeProfit
	techLevelSum := float64(GetTotalTechLevel(player))
	militaryStrength := player.MilitaryStrength

	score := netWorth*0.4 + totalTradeProfit*0.25 + techLevelSum*0.2 + militaryStrength*0.15

	return score
}

func (gi *GameInstance) checkTakeoverVictory() bool {
	for _, acquirer := range gi.State.Players {
		if acquirer.IsDefeated || acquirer.IsBankrupt {
			continue
		}

		allTakenOver := true
		hasOtherPlayers := false

		for _, target := range gi.State.Players {
			if target.ID == acquirer.ID {
				continue
			}
			hasOtherPlayers = true

			if !CheckTakeover(acquirer, target, gi.State.Exchanges) {
				allTakenOver = false
				break
			}
		}

		if hasOtherPlayers && allTakenOver {
			return true
		}
	}
	return false
}

func (gi *GameInstance) calculateRankings() []*models.PlayerRanking {
	rankings := make([]*models.PlayerRanking, 0, len(gi.State.Players))

	for _, player := range gi.State.Players {
		netWorth := CalculateNetWorth(player, gi.State.Exchanges)
		score := CalculateScore(player, gi.State.Exchanges)

		ranking := &models.PlayerRanking{
			PlayerID:         player.ID,
			PlayerName:       player.Name,
			CompanyName:      player.CompanyName,
			Score:            score,
			NetWorth:         netWorth,
			TotalTradeProfit: player.TotalTradeProfit,
			TechLevelSum:     GetTotalTechLevel(player),
			MilitaryStrength: player.MilitaryStrength,
			IsBankrupt:       player.IsBankrupt,
			IsDefeated:       player.IsDefeated,
		}
		rankings = append(rankings, ranking)
	}

	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].IsDefeated || rankings[i].IsBankrupt {
			return false
		}
		if rankings[j].IsDefeated || rankings[j].IsBankrupt {
			return true
		}
		return rankings[i].Score > rankings[j].Score
	})

	for i, ranking := range rankings {
		ranking.Rank = i + 1
	}

	return rankings
}
