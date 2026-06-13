package game

import (
	"fmt"
	"math/rand"

	"star-mining/internal/models"
)

const (
	GlobalEventChance = 0.05
	DefaultEventDuration = 5
)

type EventEffect struct {
	MiningModifier   float64
	RefiningModifier float64
	CombatModifier   float64
	TradeModifier    float64
	PriceModifier    map[models.ResourceType]float64
}

func NewRandomEvent(eventType models.EventType, global bool, targetID string) *models.RandomEvent {
	event := &models.RandomEvent{
		ID:          fmt.Sprintf("event-%d", rand.Int63()),
		Type:        eventType,
		Duration:    DefaultEventDuration,
		TurnsLeft:   DefaultEventDuration,
		Active:      true,
		Global:      global,
		TargetID:    targetID,
	}

	switch eventType {
	case models.AsteroidStorm:
		event.Name = "Asteroid Storm"
		event.Description = "An asteroid storm reduces mining efficiency in the area"
		event.Effect = "mining_reduction"

	case models.NewVeinDiscovery:
		event.Name = "New Vein Discovery"
		event.Description = "A rich new mineral vein has been discovered"
		event.Effect = "resource_boost"

	case models.PirateInvasion:
		event.Name = "Pirate Invasion"
		event.Description = "Pirate activity has increased in this region"
		event.Effect = "pirate_risk_increase"

	case models.TradeEmbargo:
		event.Name = "Trade Embargo"
		event.Description = "Trade restrictions have been imposed"
		event.Effect = "trade_cost_increase"

	case models.TechBreakthrough:
		event.Name = "Technological Breakthrough"
		event.Description = "A major technological advancement has been made"
		event.Effect = "research_boost"
	}

	return event
}

func CheckForRandomEvent(rng *rand.Rand) bool {
	return rng.Float64() < GlobalEventChance
}

func GenerateRandomEvent(rng *rand.Rand, gameMap *models.GameMap, players []*models.Player) *models.RandomEvent {
	eventTypes := []models.EventType{
		models.AsteroidStorm,
		models.NewVeinDiscovery,
		models.PirateInvasion,
		models.TradeEmbargo,
		models.TechBreakthrough,
	}

	eventType := eventTypes[rng.Intn(len(eventTypes))]

	isGlobal := rng.Float64() > 0.5

	var targetID string
	if !isGlobal {
		if rng.Float64() > 0.5 && len(players) > 0 {
			targetID = players[rng.Intn(len(players))].ID
		} else {
			allBodies := getAllBodies(gameMap)
			if len(allBodies) > 0 {
				targetID = allBodies[rng.Intn(len(allBodies))].ID
			}
		}
	}

	return NewRandomEvent(eventType, isGlobal, targetID)
}

func getAllBodies(gameMap *models.GameMap) []*models.CelestialBody {
	var bodies []*models.CelestialBody
	for _, galaxy := range gameMap.Galaxies {
		bodies = append(bodies, galaxy.CelestialBodies...)
	}
	return bodies
}

func ApplyEventEffect(event *models.RandomEvent, gameMap *models.GameMap, players []*models.Player) {
	if !event.Active {
		return
	}

	switch event.Type {
	case models.AsteroidStorm:
		applyAsteroidStorm(event, gameMap)

	case models.NewVeinDiscovery:
		applyNewVeinDiscovery(event, gameMap)

	case models.PirateInvasion:
		applyPirateInvasion(event, players)

	case models.TradeEmbargo:
		applyTradeEmbargo(event, players)

	case models.TechBreakthrough:
		applyTechBreakthrough(event, players)
	}
}

func applyAsteroidStorm(event *models.RandomEvent, gameMap *models.GameMap) {
	if event.Global {
		for _, galaxy := range gameMap.Galaxies {
			for _, body := range galaxy.CelestialBodies {
				for _, station := range body.MiningStations {
					station.Efficiency *= 0.5
				}
			}
		}
	} else {
		body := FindBodyByID(gameMap, event.TargetID)
		if body != nil {
			for _, station := range body.MiningStations {
				station.Efficiency *= 0.5
			}
		}
	}
}

func applyNewVeinDiscovery(event *models.RandomEvent, gameMap *models.GameMap) {
	resourceTypes := []models.ResourceType{
		models.IronOre,
		models.Titanium,
		models.Helium3,
		models.RareEarth,
		models.IceCrystal,
	}

	if event.Global {
		for _, galaxy := range gameMap.Galaxies {
			for _, body := range galaxy.CelestialBodies {
				resourceType := resourceTypes[rand.Intn(len(resourceTypes))]
				boostAmount := body.MaxResources[resourceType] * 0.3
				body.Resources[resourceType] += boostAmount
				body.MaxResources[resourceType] += boostAmount
			}
		}
	} else {
		body := FindBodyByID(gameMap, event.TargetID)
		if body != nil {
			resourceType := resourceTypes[rand.Intn(len(resourceTypes))]
			boostAmount := body.MaxResources[resourceType] * 0.5
			body.Resources[resourceType] += boostAmount
			body.MaxResources[resourceType] += boostAmount
		}
	}
}

func applyPirateInvasion(event *models.RandomEvent, players []*models.Player) {
	damage := 10.0

	for _, player := range players {
		if !event.Global && player.ID != event.TargetID {
			continue
		}

		for _, fleet := range player.Fleets {
			for _, ship := range fleet.Ships {
				ship.Health -= damage
				if ship.Health < 0 {
					ship.Health = 1
				}
			}
			UpdateFleetStats(fleet)
		}

		for _, ship := range player.Ships {
			ship.Health -= damage
			if ship.Health < 0 {
				ship.Health = 1
			}
		}
	}
}

func applyTradeEmbargo(event *models.RandomEvent, players []*models.Player) {
	embargoPenalty := 0.2

	for _, player := range players {
		if !event.Global && player.ID != event.TargetID {
			continue
		}

		player.Reputation *= (1 - embargoPenalty)
		if player.Reputation < 0 {
			player.Reputation = 0
		}
	}
}

func applyTechBreakthrough(event *models.RandomEvent, players []*models.Player) {
	progressBoost := 0.3

	for _, player := range players {
		if !event.Global && player.ID != event.TargetID {
			continue
		}

		if player.TechTree != nil && player.TechTree.Researching != "" {
			player.TechTree.Progress += progressBoost
			if player.TechTree.Progress > 1.0 {
				player.TechTree.Progress = 1.0
			}
		}
	}
}

func UpdateEvents(events []*models.RandomEvent, gameMap *models.GameMap, players []*models.Player) []*models.RandomEvent {
	activeEvents := make([]*models.RandomEvent, 0)

	for _, event := range events {
		if !event.Active {
			continue
		}

		event.TurnsLeft--

		if event.TurnsLeft <= 0 {
			event.Active = false
			RemoveEventEffect(event, gameMap, players)
		} else {
			activeEvents = append(activeEvents, event)
		}
	}

	return activeEvents
}

func RemoveEventEffect(event *models.RandomEvent, gameMap *models.GameMap, players []*models.Player) {
	switch event.Type {
	case models.AsteroidStorm:
		removeAsteroidStorm(event, gameMap)
	}
}

func removeAsteroidStorm(event *models.RandomEvent, gameMap *models.GameMap) {
	if event.Global {
		for _, galaxy := range gameMap.Galaxies {
			for _, body := range galaxy.CelestialBodies {
				for _, station := range body.MiningStations {
					station.Efficiency /= 0.5
					if station.Efficiency > 1.0 {
						station.Efficiency = 1.0
					}
				}
			}
		}
	} else {
		body := FindBodyByID(gameMap, event.TargetID)
		if body != nil {
			for _, station := range body.MiningStations {
				station.Efficiency /= 0.5
				if station.Efficiency > 1.0 {
					station.Efficiency = 1.0
				}
			}
		}
	}
}

func GetEventEffectModifier(events []*models.RandomEvent, playerID string, bodyID string) *EventEffect {
	effect := &EventEffect{
		MiningModifier:   1.0,
		RefiningModifier: 1.0,
		CombatModifier:   1.0,
		TradeModifier:    1.0,
		PriceModifier:    make(map[models.ResourceType]float64),
	}

	for _, event := range events {
		if !event.Active {
			continue
		}

		if !event.Global && event.TargetID != playerID && event.TargetID != bodyID {
			continue
		}

		switch event.Type {
		case models.AsteroidStorm:
			effect.MiningModifier *= 0.5
		case models.PirateInvasion:
			effect.CombatModifier *= 0.8
		case models.TradeEmbargo:
			effect.TradeModifier *= 1.2
		}
	}

	return effect
}

func ProcessTurnEvents(rng *rand.Rand, gameMap *models.GameMap, players []*models.Player, currentEvents []*models.RandomEvent) []*models.RandomEvent {
	events := UpdateEvents(currentEvents, gameMap, players)

	if CheckForRandomEvent(rng) {
		newEvent := GenerateRandomEvent(rng, gameMap, players)
		ApplyEventEffect(newEvent, gameMap, players)
		events = append(events, newEvent)
	}

	return events
}
