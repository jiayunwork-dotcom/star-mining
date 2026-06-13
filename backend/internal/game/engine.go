package game

import (
	"fmt"
	"time"

	"star-mining/internal/models"
)

type GameEngine struct {
	instance *GameInstance
}

func NewGameEngine(seed int64) *GameEngine {
	return &GameEngine{
		instance: NewGameInstance(fmt.Sprintf("game-%d", time.Now().Unix()), seed),
	}
}

func (ge *GameEngine) InitializeGame(numPlayers int, maxTurns int) *models.GameState {
	playerIDs := make([]string, 0, numPlayers)
	playerNames := make(map[string]string)
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("player-%d", i)
		playerIDs = append(playerIDs, playerID)
		playerNames[playerID] = fmt.Sprintf("Player %d", i+1)
	}

	ge.instance.maxTurns = maxTurns
	ge.instance.Initialize(playerIDs, playerNames)
	ge.instance.Start()

	for i, player := range ge.instance.State.Players {
		if i > 0 {
			player.IsAI = true
		}
	}

	return ge.instance.State
}

func (ge *GameEngine) NextTurn() {
	ge.instance.ProcessTurn()
}

func (ge *GameEngine) BuildStation(playerID, bodyID string, resourceType models.ResourceType) error {
	return ge.instance.BuildStation(playerID, bodyID, resourceType)
}

func (ge *GameEngine) BuildRefinery(playerID, bodyID string, inputResource, outputResource models.ResourceType) error {
	return ge.instance.BuildRefinery(playerID, bodyID, inputResource, outputResource)
}

func (ge *GameEngine) BuildShipyard(playerID, bodyID string) error {
	return ge.instance.BuildShipyard(playerID, bodyID)
}

func (ge *GameEngine) BuildShip(playerID, shipyardID string, shipType models.ShipType) error {
	return ge.instance.BuildShip(playerID, shipyardID, shipType)
}

func (ge *GameEngine) PlaceOrder(playerID, exchangeID string, orderType models.OrderType, resource models.ResourceType, quantity, price float64) error {
	if orderType == models.BuyOrder {
		return ge.instance.PlaceBuyOrder(playerID, exchangeID, resource, quantity, price)
	}
	return ge.instance.PlaceSellOrder(playerID, exchangeID, resource, quantity, price)
}

func (ge *GameEngine) MoveFleet(playerID, fleetID, destinationID string) error {
	return ge.instance.MoveFleet(playerID, fleetID, destinationID)
}

func (ge *GameEngine) StartResearch(playerID string, techType models.TechType) error {
	return ge.instance.ResearchTech(playerID, techType)
}

func (ge *GameEngine) CreateFleet(playerID, name string, shipIDs []string) (*models.Fleet, error) {
	return ge.instance.CreateFleet(playerID, name, shipIDs)
}

func (ge *GameEngine) getPlayer(playerID string) *models.Player {
	player, _ := ge.instance.GetPlayer(playerID)
	return player
}

func (ge *GameEngine) GetPlayer(playerID string) *models.Player {
	return ge.getPlayer(playerID)
}

func (ge *GameEngine) GetWinner() *models.Player {
	return ge.instance.GetWinner()
}
