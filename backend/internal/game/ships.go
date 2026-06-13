package game

import (
	"fmt"

	"star-mining/internal/models"
)

const (
	BaseSpeed        = 10.0
	FuelPerDistance  = 0.1
	BaseShipyardCost = 1000.0
)

func NewShip(shipType models.ShipType, playerID string) *models.Ship {
	ship := &models.Ship{
		ID:       fmt.Sprintf("ship-%s-%d", playerID, 0),
		Name:     fmt.Sprintf("%s Ship", shipType),
		Type:     shipType,
		PlayerID: playerID,
		FleetID:  "",
		Cargo:    make(models.Resources),
	}

	switch shipType {
	case models.CargoShip:
		ship.MaxHealth = 100
		ship.Health = 100
		ship.Attack = 5
		ship.Defense = 10
		ship.CargoCapacity = 500
		ship.Speed = 12
		ship.MaxFuel = 200
		ship.Fuel = 200
		ship.FuelPerTurn = 2
		ship.BuildCost = models.Resources{
			models.IronOre:  100,
			models.Titanium: 50,
			models.Fuel:     50,
		}
		ship.BuildTime = 3

	case models.Frigate:
		ship.MaxHealth = 150
		ship.Health = 150
		ship.Attack = 50
		ship.Defense = 30
		ship.CargoCapacity = 50
		ship.Speed = 15
		ship.MaxFuel = 300
		ship.Fuel = 300
		ship.FuelPerTurn = 3
		ship.BuildCost = models.Resources{
			models.IronOre:   150,
			models.Titanium:  100,
			models.RareEarth: 30,
			models.Fuel:      80,
		}
		ship.BuildTime = 5

	case models.MiningShip:
		ship.MaxHealth = 80
		ship.Health = 80
		ship.Attack = 10
		ship.Defense = 5
		ship.CargoCapacity = 200
		ship.Speed = 8
		ship.MaxFuel = 150
		ship.Fuel = 150
		ship.FuelPerTurn = 1.5
		ship.BuildCost = models.Resources{
			models.IronOre:  80,
			models.Titanium: 30,
			models.Fuel:     40,
		}
		ship.BuildTime = 2
	}

	return ship
}

func NewFleet(playerID, name string) *models.Fleet {
	return &models.Fleet{
		ID:                fmt.Sprintf("fleet-%s-%d", playerID, 0),
		Name:              name,
		PlayerID:          playerID,
		Ships:             make([]*models.Ship, 0),
		CurrentBodyID:     "",
		DestinationID:     "",
		IsMoving:          false,
		TurnsRemaining:    0,
		TotalCargo:        make(models.Resources),
		TotalCargoCap:     0,
		TotalAttack:       0,
		TotalDefense:      0,
		CommandingOfficer: "",
	}
}

func NewShipyard(playerID, bodyID string) *models.Shipyard {
	return &models.Shipyard{
		ID:          fmt.Sprintf("shipyard-%s-%s", playerID, bodyID),
		Name:        fmt.Sprintf("Shipyard at %s", bodyID),
		BodyID:      bodyID,
		PlayerID:    playerID,
		Level:       1,
		BuildQueue:  make([]*models.Ship, 0),
		Maintenance: 100.0,
	}
}

func AddShipToFleet(fleet *models.Fleet, ship *models.Ship) {
	if ship.FleetID != "" && ship.FleetID != fleet.ID {
		return
	}

	ship.FleetID = fleet.ID
	fleet.Ships = append(fleet.Ships, ship)
	UpdateFleetStats(fleet)
}

func RemoveShipFromFleet(fleet *models.Fleet, shipID string) bool {
	for i, ship := range fleet.Ships {
		if ship.ID == shipID {
			ship.FleetID = ""
			fleet.Ships = append(fleet.Ships[:i], fleet.Ships[i+1:]...)
			UpdateFleetStats(fleet)
			return true
		}
	}
	return false
}

func UpdateFleetStats(fleet *models.Fleet) {
	fleet.TotalCargoCap = 0
	fleet.TotalAttack = 0
	fleet.TotalDefense = 0
	fleet.TotalCargo = make(models.Resources)

	for _, ship := range fleet.Ships {
		fleet.TotalCargoCap += ship.CargoCapacity
		fleet.TotalAttack += ship.Attack
		fleet.TotalDefense += ship.Defense

		for resource, amount := range ship.Cargo {
			fleet.TotalCargo[resource] += amount
		}
	}
}

func CalculateTravelTime(fleet *models.Fleet, distance float64, techBonus float64) int {
	if len(fleet.Ships) == 0 {
		return 0
	}

	minSpeed := fleet.Ships[0].Speed
	for _, ship := range fleet.Ships {
		if ship.Speed < minSpeed {
			minSpeed = ship.Speed
		}
	}

	effectiveSpeed := minSpeed * techBonus
	if effectiveSpeed <= 0 {
		return 999
	}

	turns := int(distance / effectiveSpeed)
	if turns < 1 {
		turns = 1
	}

	return turns
}

func CalculateFuelCost(fleet *models.Fleet, distance float64) float64 {
	totalFuel := 0.0
	for _, ship := range fleet.Ships {
		totalFuel += ship.FuelPerTurn
	}

	turns := float64(int(distance / BaseSpeed))
	if turns < 1 {
		turns = 1
	}

	return totalFuel * turns
}

func StartFleetMove(fleet *models.Fleet, destinationID string, gameMap *models.GameMap, techBonus float64) error {
	if fleet.IsMoving {
		return fmt.Errorf("fleet is already moving")
	}

	if len(fleet.Ships) == 0 {
		return fmt.Errorf("fleet has no ships")
	}

	lane := FindLaneBetween(gameMap, fleet.CurrentBodyID, destinationID)
	if lane == nil {
		return fmt.Errorf("no lane between bodies")
	}

	distance := lane.Distance
	fuelCost := CalculateFuelCost(fleet, distance)

	totalFuel := 0.0
	for _, ship := range fleet.Ships {
		totalFuel += ship.Fuel
	}

	if totalFuel < fuelCost {
		return fmt.Errorf("not enough fuel: need %f, have %f", fuelCost, totalFuel)
	}

	fuelPerShip := fuelCost / float64(len(fleet.Ships))
	for _, ship := range fleet.Ships {
		ship.Fuel -= fuelPerShip
	}

	fleet.DestinationID = destinationID
	fleet.IsMoving = true
	fleet.TurnsRemaining = CalculateTravelTime(fleet, distance, techBonus)

	return nil
}

func ProcessFleetMove(fleet *models.Fleet) bool {
	if !fleet.IsMoving {
		return false
	}

	fleet.TurnsRemaining--

	if fleet.TurnsRemaining <= 0 {
		fleet.CurrentBodyID = fleet.DestinationID
		fleet.DestinationID = ""
		fleet.IsMoving = false
		fleet.TurnsRemaining = 0
		return true
	}

	return false
}

func ProcessAllFleetMoves(player *models.Player) []*models.Fleet {
	var arrivedFleets []*models.Fleet

	for _, fleet := range player.Fleets {
		if fleet.IsMoving {
			arrived := ProcessFleetMove(fleet)
			if arrived {
				arrivedFleets = append(arrivedFleets, fleet)
			}
		}
	}

	return arrivedFleets
}

func BuildShip(shipyard *models.Shipyard, shipType models.ShipType, player *models.Player) error {
	ship := NewShip(shipType, player.ID)

	for resource, cost := range ship.BuildCost {
		if player.Resources[resource] < cost {
			return fmt.Errorf("not enough %s: have %f, need %f", resource, player.Resources[resource], cost)
		}
	}

	for resource, cost := range ship.BuildCost {
		player.Resources[resource] -= cost
	}

	ship.TurnsLeft = ship.BuildTime
	shipyard.BuildQueue = append(shipyard.BuildQueue, ship)

	return nil
}

func ProcessShipyard(shipyard *models.Shipyard, player *models.Player) []*models.Ship {
	var completedShips []*models.Ship

	if len(shipyard.BuildQueue) == 0 {
		return completedShips
	}

	currentShip := shipyard.BuildQueue[0]
	currentShip.TurnsLeft--

	if currentShip.TurnsLeft <= 0 {
		currentShip.ID = fmt.Sprintf("ship-%s-%d", player.ID, len(player.Ships))
		player.Ships = append(player.Ships, currentShip)
		shipyard.BuildQueue = shipyard.BuildQueue[1:]
		completedShips = append(completedShips, currentShip)
	}

	return completedShips
}

func ProcessAllShipyards(player *models.Player) []*models.Ship {
	var allCompleted []*models.Ship

	for _, shipyard := range player.Shipyards {
		completed := ProcessShipyard(shipyard, player)
		allCompleted = append(allCompleted, completed...)
	}

	return allCompleted
}

func LoadCargo(ship *models.Ship, resource models.ResourceType, amount float64, player *models.Player) error {
	currentLoad := 0.0
	for _, amt := range ship.Cargo {
		currentLoad += amt
	}

	if currentLoad+amount > ship.CargoCapacity {
		return fmt.Errorf("cargo capacity exceeded: current %f, adding %f, capacity %f", currentLoad, amount, ship.CargoCapacity)
	}

	if player.Resources[resource] < amount {
		return fmt.Errorf("not enough %s: have %f, need %f", resource, player.Resources[resource], amount)
	}

	player.Resources[resource] -= amount
	ship.Cargo[resource] += amount

	return nil
}

func UnloadCargo(ship *models.Ship, resource models.ResourceType, amount float64, player *models.Player) error {
	if ship.Cargo[resource] < amount {
		return fmt.Errorf("not enough cargo: have %f, need %f", ship.Cargo[resource], amount)
	}

	ship.Cargo[resource] -= amount

	if player.Resources == nil {
		player.Resources = make(models.Resources)
	}
	player.Resources[resource] += amount

	return nil
}

func RefuelShip(ship *models.Ship, amount float64, player *models.Player) error {
	fuelNeeded := ship.MaxFuel - ship.Fuel
	if amount > fuelNeeded {
		amount = fuelNeeded
	}

	if player.Resources[models.Fuel] < amount {
		return fmt.Errorf("not enough fuel: have %f, need %f", player.Resources[models.Fuel], amount)
	}

	player.Resources[models.Fuel] -= amount
	ship.Fuel += amount

	return nil
}
