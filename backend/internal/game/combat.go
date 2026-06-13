package game

import (
	"fmt"
	"math/rand"

	"star-mining/internal/models"
)

type CombatResult struct {
	Winner         string
	AttackerLosses []*models.Ship
	DefenderLosses []*models.Ship
	AttackerDamage float64
	DefenderDamage float64
	Loot           models.Resources
}

type PirateFleet struct {
	ID     string
	Ships  []*models.Ship
	Power  float64
	Loot   models.Resources
}

func CalculateFleetPower(fleet *models.Fleet, techBonus float64) float64 {
	if fleet == nil {
		return 0
	}

	totalAttack := fleet.TotalAttack * techBonus
	totalDefense := fleet.TotalDefense
	totalHealth := 0.0

	for _, ship := range fleet.Ships {
		totalHealth += ship.Health
	}

	power := totalAttack + totalDefense*0.5 + totalHealth*0.1
	return power
}

func SimulateCombat(attacker, defender *models.Fleet, attackerTechBonus, defenderTechBonus float64, rng *rand.Rand) *CombatResult {
	result := &CombatResult{
		AttackerLosses: make([]*models.Ship, 0),
		DefenderLosses: make([]*models.Ship, 0),
		Loot:           make(models.Resources),
	}

	if len(attacker.Ships) == 0 {
		result.Winner = "defender"
		return result
	}
	if len(defender.Ships) == 0 {
		result.Winner = "attacker"
		result.Loot = defender.TotalCargo
		return result
	}

	attackerPower := CalculateFleetPower(attacker, attackerTechBonus)
	defenderPower := CalculateFleetPower(defender, defenderTechBonus)

	attackerEffectiveAttack := attacker.TotalAttack * attackerTechBonus
	defenderEffectiveAttack := defender.TotalAttack * defenderTechBonus

	attackerDamage := attackerEffectiveAttack * (0.8 + rng.Float64()*0.4)
	defenderDamage := defenderEffectiveAttack * (0.8 + rng.Float64()*0.4)

	result.AttackerDamage = attackerDamage
	result.DefenderDamage = defenderDamage

	distributeDamage(defender, attackerDamage, result.DefenderLosses)
	distributeDamage(attacker, defenderDamage, result.AttackerLosses)

	attackerAlive := len(attacker.Ships) > 0
	defenderAlive := len(defender.Ships) > 0

	if attackerAlive && !defenderAlive {
		result.Winner = "attacker"
		result.Loot = defender.TotalCargo
	} else if !attackerAlive && defenderAlive {
		result.Winner = "defender"
	} else if attackerAlive && defenderAlive {
		if attackerPower > defenderPower*1.5 {
			result.Winner = "attacker"
		} else if defenderPower > attackerPower*1.5 {
			result.Winner = "defender"
		} else {
			result.Winner = "draw"
		}
	} else {
		result.Winner = "draw"
	}

	UpdateFleetStats(attacker)
	UpdateFleetStats(defender)

	return result
}

func distributeDamage(fleet *models.Fleet, totalDamage float64, losses *[]*models.Ship) {
	if len(fleet.Ships) == 0 {
		return
	}

	damagePerShip := totalDamage / float64(len(fleet.Ships))

	survivingShips := make([]*models.Ship, 0, len(fleet.Ships))

	for _, ship := range fleet.Ships {
		ship.Health -= damagePerShip

		if ship.Health <= 0 {
			*losses = append(*losses, ship)
		} else {
			survivingShips = append(survivingShips, ship)
		}
	}

	fleet.Ships = survivingShips
}

func GeneratePirateFleet(difficulty int, rng *rand.Rand) *PirateFleet {
	numShips := difficulty + rng.Intn(difficulty+1)
	if numShips < 1 {
		numShips = 1
	}

	pirate := &PirateFleet{
		ID:    fmt.Sprintf("pirate-%d", rng.Intn(100000)),
		Ships: make([]*models.Ship, 0, numShips),
		Loot:  make(models.Resources),
	}

	for i := 0; i < numShips; i++ {
		shipType := models.Frigate
		if rng.Float64() > 0.7 {
			shipType = models.CargoShip
		}

		ship := NewShip(shipType, "pirate")
		ship.ID = fmt.Sprintf("pirate-ship-%d", i)

		healthMod := 0.6 + rng.Float64()*0.4
		attackMod := 0.7 + rng.Float64()*0.6

		ship.MaxHealth *= healthMod
		ship.Health = ship.MaxHealth
		ship.Attack *= attackMod

		pirate.Ships = append(pirate.Ships, ship)
	}

	pirate.Power = 0
	for _, ship := range pirate.Ships {
		pirate.Power += ship.Attack + ship.Defense*0.5
	}

	lootValue := float64(numShips) * 100 * (0.5 + rng.Float64())
	pirate.Loot[models.IronOre] = lootValue * 0.4
	pirate.Loot[models.Titanium] = lootValue * 0.2

	return pirate
}

func PirateAttack(fleet *models.Fleet, difficulty int, techBonus float64, rng *rand.Rand) (*CombatResult, bool) {
	pirate := GeneratePirateFleet(difficulty, rng)

	pirateFleet := &models.Fleet{
		ID:         pirate.ID,
		Name:       "Pirate Fleet",
		PlayerID:   "pirate",
		Ships:      pirate.Ships,
		TotalCargo: pirate.Loot,
	}
	UpdateFleetStats(pirateFleet)

	result := SimulateCombat(fleet, pirateFleet, techBonus, 1.0, rng)

	if result.Winner == "attacker" {
		for resource, amount := range pirate.Loot {
			fleet.TotalCargo[resource] += amount
		}
	}

	attacked := result.Winner != "attacker" || len(result.AttackerLosses) > 0

	return result, attacked
}

func CheckForPirateAttack(fleet *models.Fleet, lane *models.Lane, rng *rand.Rand) bool {
	if lane == nil {
		return false
	}

	if lane.Safe {
		return rng.Float64() < 0.05
	}

	return rng.Float64() < 0.2
}

func NewBlockade(playerID, bodyID, fleetID string, tollRate float64) *models.Blockade {
	return &models.Blockade{
		ID:             fmt.Sprintf("blockade-%s-%s", playerID, bodyID),
		PlayerID:       playerID,
		TargetBodyID:   bodyID,
		FleetID:        fleetID,
		TurnsRemaining: 10,
		TollRate:       tollRate,
		Active:         true,
	}
}

func ProcessBlockade(blockade *models.Blockade, passingFleet *models.Fleet, player *models.Player, tollPlayer *models.Player) (bool, float64) {
	if !blockade.Active {
		return false, 0
	}

	toll := CalculateToll(passingFleet, blockade.TollRate)

	if player.Credits >= toll {
		player.Credits -= toll
		tollPlayer.Credits += toll
		return true, toll
	}

	return false, 0
}

func CalculateToll(fleet *models.Fleet, tollRate float64) float64 {
	cargoValue := 0.0
	for _, amount := range fleet.TotalCargo {
		cargoValue += amount
	}

	return cargoValue * tollRate
}

func UpdateBlockades(blockades []*models.Blockade) []*models.Blockade {
	activeBlockades := make([]*models.Blockade, 0)

	for _, blockade := range blockades {
		if !blockade.Active {
			continue
		}

		blockade.TurnsRemaining--
		if blockade.TurnsRemaining > 0 {
			activeBlockades = append(activeBlockades, blockade)
		}
	}

	return activeBlockades
}

func FindBlockadeAtBody(blockades []*models.Blockade, bodyID string) *models.Blockade {
	for _, blockade := range blockades {
		if blockade.TargetBodyID == bodyID && blockade.Active {
			return blockade
		}
	}
	return nil
}
