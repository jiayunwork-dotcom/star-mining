package game

import (
	"fmt"

	"star-mining/internal/models"
)

const (
	BaseMiningOutput    = 10.0
	BaseRefiningSpeed   = 5.0
	BaseStationCost     = 500.0
	BaseRefineryCost    = 800.0
	StationUpgradeMult  = 1.5
	RefineryUpgradeMult = 1.6
)

func NewMiningStation(playerID, bodyID string, resourceType models.ResourceType) *models.MiningStation {
	return &models.MiningStation{
		ID:           fmt.Sprintf("station-%s-%s", playerID, resourceType),
		Name:         fmt.Sprintf("%s Mining Station", resourceType),
		BodyID:       bodyID,
		PlayerID:     playerID,
		ResourceType: resourceType,
		Efficiency:   1.0,
		Level:        1,
		Workers:      10,
		Maintenance:  50.0,
		OutputPerTurn: BaseMiningOutput,
	}
}

func NewRefinery(playerID, bodyID string, inputResource, outputResource models.ResourceType) *models.Refinery {
	return &models.Refinery{
		ID:              fmt.Sprintf("refinery-%s-%s", playerID, inputResource),
		Name:            fmt.Sprintf("%s Refinery", inputResource),
		BodyID:          bodyID,
		PlayerID:        playerID,
		InputResource:   inputResource,
		OutputResource:  outputResource,
		ConversionRate:  0.8,
		Level:           1,
		Maintenance:     80.0,
		ProcessingSpeed: BaseRefiningSpeed,
		InputInventory:  0,
		OutputInventory: 0,
	}
}

func CalculateMiningOutput(station *models.MiningStation, body *models.CelestialBody, techBonus float64) float64 {
	if body == nil {
		return 0
	}

	resourceAvailable, exists := body.Resources[station.ResourceType]
	if !exists || resourceAvailable <= 0 {
		return 0
	}

	levelBonus := 1.0 + float64(station.Level-1)*0.3
	workerBonus := 1.0 + float64(station.Workers-10)*0.02

	baseOutput := station.OutputPerTurn * station.Efficiency * levelBonus * workerBonus * techBonus

	if baseOutput > resourceAvailable {
		baseOutput = resourceAvailable
	}

	return baseOutput
}

func ProcessMining(station *models.MiningStation, body *models.CelestialBody, player *models.Player, techBonus float64) float64 {
	output := CalculateMiningOutput(station, body, techBonus)

	if output <= 0 {
		return 0
	}

	body.Resources[station.ResourceType] -= output

	if player.Resources == nil {
		player.Resources = make(models.Resources)
	}
	player.Resources[station.ResourceType] += output

	return output
}

func ProcessAllMining(player *models.Player, gameMap *models.GameMap, techBonus float64) map[string]float64 {
	results := make(map[string]float64)

	for _, station := range player.Stations {
		body := FindBodyByID(gameMap, station.BodyID)
		if body == nil {
			continue
		}

		output := ProcessMining(station, body, player, techBonus)
		results[station.ID] = output
	}

	return results
}

func CalculateRefiningOutput(refinery *models.Refinery, techBonus float64) float64 {
	if refinery.InputInventory <= 0 {
		return 0
	}

	levelBonus := 1.0 + float64(refinery.Level-1)*0.25
	processingAmount := refinery.ProcessingSpeed * levelBonus * techBonus

	if processingAmount > refinery.InputInventory {
		processingAmount = refinery.InputInventory
	}

	output := processingAmount * refinery.ConversionRate

	return output
}

func ProcessRefining(refinery *models.Refinery, player *models.Player, techBonus float64) float64 {
	if refinery.InputInventory <= 0 {
		return 0
	}

	levelBonus := 1.0 + float64(refinery.Level-1)*0.25
	processingAmount := refinery.ProcessingSpeed * levelBonus * techBonus

	if processingAmount > refinery.InputInventory {
		processingAmount = refinery.InputInventory
	}

	outputAmount := processingAmount * refinery.ConversionRate

	refinery.InputInventory -= processingAmount
	refinery.OutputInventory += outputAmount

	if player.Resources == nil {
		player.Resources = make(models.Resources)
	}
	player.Resources[refinery.OutputResource] += outputAmount

	return outputAmount
}

func ProcessAllRefining(player *models.Player, techBonus float64) map[string]float64 {
	results := make(map[string]float64)

	for _, refinery := range player.Refineries {
		output := ProcessRefining(refinery, player, techBonus)
		results[refinery.ID] = output
	}

	return results
}

func TransferToRefinery(player *models.Player, refinery *models.Refinery, amount float64) error {
	if player.Resources == nil {
		return fmt.Errorf("player has no resources")
	}

	available := player.Resources[refinery.InputResource]
	if available < amount {
		return fmt.Errorf("not enough resources: have %f, need %f", available, amount)
	}

	player.Resources[refinery.InputResource] -= amount
	refinery.InputInventory += amount

	return nil
}

func TransferFromRefinery(player *models.Player, refinery *models.Refinery, amount float64) error {
	if refinery.OutputInventory < amount {
		return fmt.Errorf("not enough output: have %f, need %f", refinery.OutputInventory, amount)
	}

	refinery.OutputInventory -= amount

	if player.Resources == nil {
		player.Resources = make(models.Resources)
	}
	player.Resources[refinery.OutputResource] += amount

	return nil
}

func CalculateStationCost(level int) float64 {
	return BaseStationCost * StationUpgradeMult * float64(level-1)
}

func CalculateRefineryCost(level int) float64 {
	return BaseRefineryCost * RefineryUpgradeMult * float64(level-1)
}

func UpgradeStation(station *models.MiningStation, player *models.Player) error {
	cost := CalculateStationCost(station.Level + 1)

	if player.Credits < cost {
		return fmt.Errorf("not enough credits: have %f, need %f", player.Credits, cost)
	}

	player.Credits -= cost
	station.Level++
	station.Maintenance *= 1.2
	station.OutputPerTurn *= 1.3

	return nil
}

func UpgradeRefinery(refinery *models.Refinery, player *models.Player) error {
	cost := CalculateRefineryCost(refinery.Level + 1)

	if player.Credits < cost {
		return fmt.Errorf("not enough credits: have %f, need %f", player.Credits, cost)
	}

	player.Credits -= cost
	refinery.Level++
	refinery.Maintenance *= 1.25
	refinery.ProcessingSpeed *= 1.3

	return nil
}

func CalculateTotalMaintenance(player *models.Player) float64 {
	total := 0.0

	for _, station := range player.Stations {
		total += station.Maintenance
	}

	for _, refinery := range player.Refineries {
		total += refinery.Maintenance
	}

	for _, shipyard := range player.Shipyards {
		total += shipyard.Maintenance
	}

	return total
}

func PayMaintenance(player *models.Player) float64 {
	maintenance := CalculateTotalMaintenance(player)
	player.Credits -= maintenance
	return maintenance
}

func GetRefineryRecipes() map[models.ResourceType]models.ResourceType {
	return map[models.ResourceType]models.ResourceType{
		models.IronOre:   models.Titanium,
		models.Titanium:  models.RareEarth,
		models.IceCrystal: models.Fuel,
	}
}
