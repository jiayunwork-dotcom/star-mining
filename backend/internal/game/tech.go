package game

import (
	"fmt"

	"star-mining/internal/models"
)

const (
	MaxTechLevel    = 10
	BaseResearchCost = 100.0
	ResearchCostMult = 1.8
)

func NewTechTree(playerID string) *models.TechTree {
	techs := make(map[models.TechType]*models.TechLevel)

	techs[models.MiningEfficiency] = &models.TechLevel{
		Type:     models.MiningEfficiency,
		Level:    1,
		MaxLevel: MaxTechLevel,
	}

	techs[models.RefiningTech] = &models.TechLevel{
		Type:     models.RefiningTech,
		Level:    1,
		MaxLevel: MaxTechLevel,
	}

	techs[models.EngineImprovement] = &models.TechLevel{
		Type:     models.EngineImprovement,
		Level:    1,
		MaxLevel: MaxTechLevel,
	}

	techs[models.WeaponUpgrade] = &models.TechLevel{
		Type:     models.WeaponUpgrade,
		Level:    1,
		MaxLevel: MaxTechLevel,
	}

	return &models.TechTree{
		PlayerID:    playerID,
		Techs:       techs,
		Researching: "",
		Progress:    0,
		ResearchCost: make(models.Resources),
	}
}

func GetTechBonus(techTree *models.TechTree, techType models.TechType) float64 {
	if techTree == nil {
		return 1.0
	}

	tech, exists := techTree.Techs[techType]
	if !exists {
		return 1.0
	}

	switch techType {
	case models.MiningEfficiency:
		return 1.0 + float64(tech.Level-1)*0.15
	case models.RefiningTech:
		return 1.0 + float64(tech.Level-1)*0.12
	case models.EngineImprovement:
		return 1.0 + float64(tech.Level-1)*0.1
	case models.WeaponUpgrade:
		return 1.0 + float64(tech.Level-1)*0.2
	default:
		return 1.0
	}
}

func GetResearchCost(techType models.TechType, currentLevel int) models.Resources {
	cost := make(models.Resources)
	baseCost := BaseResearchCost

	switch techType {
	case models.MiningEfficiency:
		cost[models.Credits] = baseCost * ResearchCostMult * float64(currentLevel)
		cost[models.IronOre] = baseCost * 0.5 * ResearchCostMult * float64(currentLevel)
	case models.RefiningTech:
		cost[models.Credits] = baseCost * 1.2 * ResearchCostMult * float64(currentLevel)
		cost[models.Titanium] = baseCost * 0.3 * ResearchCostMult * float64(currentLevel)
	case models.EngineImprovement:
		cost[models.Credits] = baseCost * 1.5 * ResearchCostMult * float64(currentLevel)
		cost[models.Fuel] = baseCost * 0.5 * ResearchCostMult * float64(currentLevel)
	case models.WeaponUpgrade:
		cost[models.Credits] = baseCost * 2.0 * ResearchCostMult * float64(currentLevel)
		cost[models.RareEarth] = baseCost * 0.2 * ResearchCostMult * float64(currentLevel)
	}

	return cost
}

func GetResearchTime(techType models.TechType, currentLevel int) float64 {
	baseTime := 3.0

	switch techType {
	case models.MiningEfficiency:
		return baseTime + float64(currentLevel)*0.5
	case models.RefiningTech:
		return baseTime*1.2 + float64(currentLevel)*0.6
	case models.EngineImprovement:
		return baseTime*1.5 + float64(currentLevel)*0.7
	case models.WeaponUpgrade:
		return baseTime*2.0 + float64(currentLevel)*0.8
	default:
		return baseTime
	}
}

func StartResearch(techTree *models.TechTree, techType models.TechType, player *models.Player) error {
	if techTree.Researching != "" {
		return fmt.Errorf("already researching %s", techTree.Researching)
	}

	tech, exists := techTree.Techs[techType]
	if !exists {
		return fmt.Errorf("unknown tech type: %s", techType)
	}

	if tech.Level >= tech.MaxLevel {
		return fmt.Errorf("tech %s already at max level", techType)
	}

	cost := GetResearchCost(techType, tech.Level)

	for resource, amount := range cost {
		if resource == models.Credits {
			if player.Credits < amount {
				return fmt.Errorf("not enough credits: have %f, need %f", player.Credits, amount)
			}
		} else {
			if player.Resources[resource] < amount {
				return fmt.Errorf("not enough %s: have %f, need %f", resource, player.Resources[resource], amount)
			}
		}
	}

	for resource, amount := range cost {
		if resource == models.Credits {
			player.Credits -= amount
		} else {
			player.Resources[resource] -= amount
		}
	}

	techTree.Researching = techType
	techTree.Progress = 0
	techTree.ResearchCost = cost

	return nil
}

func ProcessResearch(techTree *models.TechTree) (models.TechType, bool) {
	if techTree.Researching == "" {
		return "", false
	}

	techType := techTree.Researching
	tech := techTree.Techs[techType]
	researchTime := GetResearchTime(techType, tech.Level)

	techTree.Progress += 1.0 / researchTime

	if techTree.Progress >= 1.0 {
		tech.Level++
		techTree.Researching = ""
		techTree.Progress = 0
		techTree.ResearchCost = make(models.Resources)
		return techType, true
	}

	return "", false
}

func ProcessAllResearch(players []*models.Player) map[string]models.TechType {
	completed := make(map[string]models.TechType)

	for _, player := range players {
		if player.TechTree != nil {
			if techType, success := ProcessResearch(player.TechTree); success {
				completed[player.ID] = techType
			}
		}
	}

	return completed
}

func GetMiningBonus(techTree *models.TechTree) float64 {
	return GetTechBonus(techTree, models.MiningEfficiency)
}

func GetRefiningBonus(techTree *models.TechTree) float64 {
	return GetTechBonus(techTree, models.RefiningTech)
}

func GetEngineBonus(techTree *models.TechTree) float64 {
	return GetTechBonus(techTree, models.EngineImprovement)
}

func GetCombatBonus(techTree *models.TechTree) float64 {
	return GetTechBonus(techTree, models.WeaponUpgrade)
}

func CanResearch(techTree *models.TechTree, techType models.TechType) bool {
	if techTree == nil {
		return false
	}

	if techTree.Researching != "" {
		return false
	}

	tech, exists := techTree.Techs[techType]
	if !exists {
		return false
	}

	return tech.Level < tech.MaxLevel
}

func GetTechDescription(techType models.TechType) string {
	switch techType {
	case models.MiningEfficiency:
		return "Increases mining output by 15% per level"
	case models.RefiningTech:
		return "Increases refining speed by 12% per level"
	case models.EngineImprovement:
		return "Increases ship speed by 10% per level"
	case models.WeaponUpgrade:
		return "Increases weapon damage by 20% per level"
	default:
		return "Unknown technology"
	}
}
