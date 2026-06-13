package game

import (
	"fmt"
	"math/rand"

	"star-mining/internal/models"
)

type MapGenerator struct {
	seed int64
	rng  *rand.Rand
}

func NewMapGenerator(seed int64) *MapGenerator {
	return &MapGenerator{
		seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

func (mg *MapGenerator) GenerateMap() *models.GameMap {
	numGalaxies := mg.rng.Intn(4) + 5

	gameMap := &models.GameMap{
		Galaxies: make([]*models.Galaxy, 0, numGalaxies),
	}

	for i := 0; i < numGalaxies; i++ {
		galaxy := mg.generateGalaxy(i)
		gameMap.Galaxies = append(gameMap.Galaxies, galaxy)
	}

	return gameMap
}

func (mg *MapGenerator) generateGalaxy(index int) *models.Galaxy {
	galaxy := &models.Galaxy{
		ID:              fmt.Sprintf("galaxy-%d", index),
		Name:            fmt.Sprintf("Galaxy %s", string(rune('A'+index))),
		CelestialBodies: make([]*models.CelestialBody, 0),
		Lanes:           make([]*models.Lane, 0),
	}

	numBodies := mg.rng.Intn(4) + 3

	for i := 0; i < numBodies; i++ {
		body := mg.generateCelestialBody(i, galaxy.ID)
		galaxy.CelestialBodies = append(galaxy.CelestialBodies, body)
	}

	mg.generateLanes(galaxy)

	return galaxy
}

func (mg *MapGenerator) generateCelestialBody(index int, galaxyID string) *models.CelestialBody {
	types := []models.CelestialType{
		models.Star,
		models.Planet,
		models.AsteroidBelt,
		models.GasGiant,
		models.Terrestrial,
	}

	bodyType := types[mg.rng.Intn(len(types))]

	body := &models.CelestialBody{
		ID:             fmt.Sprintf("%s-body-%d", galaxyID, index),
		Name:           generateBodyName(bodyType, index),
		Type:           bodyType,
		GalaxyID:       galaxyID,
		PositionX:      mg.rng.Float64()*200 - 100,
		PositionY:      mg.rng.Float64()*200 - 100,
		Resources:      make(models.Resources),
		MaxResources:   make(models.Resources),
		MiningStations: make([]*models.MiningStation, 0),
		Refineries:     make([]*models.Refinery, 0),
		Shipyards:      make([]*models.Shipyard, 0),
		OwnerPlayerID:  "",
		HasExchange:    false,
	}

	mg.generateResources(body)

	if bodyType == models.Planet || bodyType == models.Terrestrial {
		body.HasExchange = mg.rng.Float64() > 0.5
	}

	return body
}

func (mg *MapGenerator) generateResources(body *models.CelestialBody) {
	resourceTypes := []models.ResourceType{
		models.IronOre,
		models.Titanium,
		models.Helium3,
		models.RareEarth,
		models.IceCrystal,
	}

	numResources := mg.rng.Intn(3) + 1

	shuffled := make([]models.ResourceType, len(resourceTypes))
	copy(shuffled, resourceTypes)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := mg.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	baseAmount := 1000.0
	switch body.Type {
	case models.AsteroidBelt:
		baseAmount = 5000.0
	case models.Planet:
		baseAmount = 3000.0
	case models.GasGiant:
		baseAmount = 8000.0
	case models.Terrestrial:
		baseAmount = 2000.0
	case models.Star:
		baseAmount = 10000.0
	}

	for i := 0; i < numResources && i < len(shuffled); i++ {
		resourceType := shuffled[i]
		abundance := mg.rng.Float64()*0.7 + 0.3
		maxAmount := baseAmount * abundance

		body.Resources[resourceType] = maxAmount
		body.MaxResources[resourceType] = maxAmount
	}
}

func (mg *MapGenerator) generateLanes(galaxy *models.Galaxy) {
	bodies := galaxy.CelestialBodies
	numBodies := len(bodies)

	if numBodies < 2 {
		return
	}

	for i := 0; i < numBodies; i++ {
		numConnections := mg.rng.Intn(2) + 1

		for j := 0; j < numConnections; j++ {
			target := mg.rng.Intn(numBodies)
			if target == i {
				continue
			}

			alreadyConnected := false
			for _, lane := range galaxy.Lanes {
				if (lane.FromBodyID == bodies[i].ID && lane.ToBodyID == bodies[target].ID) ||
				   (lane.FromBodyID == bodies[target].ID && lane.ToBodyID == bodies[i].ID) {
					alreadyConnected = true
					break
				}
			}

			if alreadyConnected {
				continue
			}

			distance := calculateDistance(bodies[i], bodies[target])

			lane := &models.Lane{
				ID:         fmt.Sprintf("lane-%s-%s", bodies[i].ID, bodies[target].ID),
				FromBodyID: bodies[i].ID,
				ToBodyID:   bodies[target].ID,
				Distance:   distance,
				Safe:       mg.rng.Float64() > 0.3,
			}

			galaxy.Lanes = append(galaxy.Lanes, lane)
		}
	}

	mg.ensureConnectivity(galaxy)
}

func (mg *MapGenerator) ensureConnectivity(galaxy *models.Galaxy) {
	bodies := galaxy.CelestialBodies
	if len(bodies) < 2 {
		return
	}

	visited := make(map[string]bool)
	queue := []string{bodies[0].ID}
	visited[bodies[0].ID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, lane := range galaxy.Lanes {
			var neighbor string
			if lane.FromBodyID == current {
				neighbor = lane.ToBodyID
			} else if lane.ToBodyID == current {
				neighbor = lane.FromBodyID
			} else {
				continue
			}

			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	for _, body := range bodies {
		if !visited[body.ID] {
			closestBody := bodies[0]
			minDist := calculateDistance(body, closestBody)

			for _, other := range bodies {
				if other.ID == body.ID {
					continue
				}
				if visited[other.ID] {
					dist := calculateDistance(body, other)
					if dist < minDist {
						minDist = dist
						closestBody = other
					}
				}
			}

			lane := &models.Lane{
				ID:         fmt.Sprintf("lane-%s-%s", body.ID, closestBody.ID),
				FromBodyID: body.ID,
				ToBodyID:   closestBody.ID,
				Distance:   minDist,
				Safe:       mg.rng.Float64() > 0.5,
			}

			galaxy.Lanes = append(galaxy.Lanes, lane)
			visited[body.ID] = true
		}
	}
}

func calculateDistance(a, b *models.CelestialBody) float64 {
	dx := a.PositionX - b.PositionX
	dy := a.PositionY - b.PositionY
	return float64(int((dx*dx + dy*dy) * 100)) / 100
}

func generateBodyName(bodyType models.CelestialType, index int) string {
	prefixes := map[models.CelestialType]string{
		models.Star:         "Star",
		models.Planet:       "Planet",
		models.AsteroidBelt: "Belt",
		models.GasGiant:     "Giant",
		models.Terrestrial:  "Terra",
	}

	suffixes := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
	prefix := prefixes[bodyType]
	suffix := suffixes[index%len(suffixes)]

	return fmt.Sprintf("%s-%s-%d", prefix, suffix, index)
}

func FindBodyByID(gameMap *models.GameMap, bodyID string) *models.CelestialBody {
	for _, galaxy := range gameMap.Galaxies {
		for _, body := range galaxy.CelestialBodies {
			if body.ID == bodyID {
				return body
			}
		}
	}
	return nil
}

func FindLaneBetween(gameMap *models.GameMap, fromID, toID string) *models.Lane {
	for _, galaxy := range gameMap.Galaxies {
		for _, lane := range galaxy.Lanes {
			if (lane.FromBodyID == fromID && lane.ToBodyID == toID) ||
			   (lane.FromBodyID == toID && lane.ToBodyID == fromID) {
				return lane
			}
		}
	}
	return nil
}

func GetAdjacentBodies(gameMap *models.GameMap, bodyID string) []*models.CelestialBody {
	var adjacent []*models.CelestialBody

	for _, galaxy := range gameMap.Galaxies {
		for _, lane := range galaxy.Lanes {
			var otherID string
			if lane.FromBodyID == bodyID {
				otherID = lane.ToBodyID
			} else if lane.ToBodyID == bodyID {
				otherID = lane.FromBodyID
			} else {
				continue
			}

			body := FindBodyByID(gameMap, otherID)
			if body != nil {
				adjacent = append(adjacent, body)
			}
		}
	}

	return adjacent
}
