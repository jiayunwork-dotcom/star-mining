package game

import (
	"fmt"

	"star-mining/internal/models"
)

const (
	InitialTotalShares = 1000
	InitialSharePrice  = 10.0
	DividendInterval   = 10
	TakeoverThreshold  = 0.51
)

func NewStock(playerID, issuerID string, shares int) *models.Stock {
	return &models.Stock{
		ID:          fmt.Sprintf("stock-%s-%s", playerID, issuerID),
		PlayerID:    playerID,
		IssuerID:    issuerID,
		Shares:      shares,
		SharePrice:  InitialSharePrice,
		TotalShares: InitialTotalShares,
		Dividend:    0,
	}
}

func InitializePlayerStock(player *models.Player) {
	stock := NewStock(player.ID, player.ID, InitialTotalShares)
	player.Stocks = []*models.Stock{stock}
}

func CalculateSharePrice(player *models.Player, exchanges []*models.Exchange) float64 {
	netWorth := CalculateNetWorth(player, exchanges)
	totalShares := InitialTotalShares

	if netWorth <= 0 {
		return 1.0
	}

	return netWorth / float64(totalShares)
}

func UpdateStockPrices(players []*models.Player, exchanges []*models.Exchange) {
	for _, player := range players {
		price := CalculateSharePrice(player, exchanges)

		for _, stock := range player.Stocks {
			stock.SharePrice = price
		}
	}
}

func BuyStock(buyer *models.Player, sellerPlayer *models.Player, stock *models.Stock, shares int, exchanges []*models.Exchange) error {
	if shares <= 0 {
		return fmt.Errorf("shares must be positive")
	}

	if stock.Shares < shares {
		return fmt.Errorf("not enough shares available")
	}

	price := CalculateSharePrice(sellerPlayer, exchanges)
	totalCost := float64(shares) * price

	if buyer.Credits < totalCost {
		return fmt.Errorf("not enough credits: have %f, need %f", buyer.Credits, totalCost)
	}

	buyer.Credits -= totalCost
	sellerPlayer.Credits += totalCost

	stock.Shares -= shares

	var buyerStock *models.Stock
	for _, s := range buyer.Stocks {
		if s.IssuerID == sellerPlayer.ID {
			buyerStock = s
			break
		}
	}

	if buyerStock == nil {
		buyerStock = NewStock(buyer.ID, sellerPlayer.ID, 0)
		buyer.Stocks = append(buyer.Stocks, buyerStock)
	}

	buyerStock.Shares += shares

	return nil
}

func SellStock(seller *models.Player, buyerPlayer *models.Player, stock *models.Stock, shares int, exchanges []*models.Exchange) error {
	if shares <= 0 {
		return fmt.Errorf("shares must be positive")
	}

	if stock.Shares < shares {
		return fmt.Errorf("not enough shares to sell")
	}

	price := CalculateSharePrice(buyerPlayer, exchanges)
	totalValue := float64(shares) * price

	seller.Credits += totalValue
	buyerPlayer.Credits -= totalValue

	stock.Shares -= shares

	var buyerStock *models.Stock
	for _, s := range buyerPlayer.Stocks {
		if s.IssuerID == buyerPlayer.ID {
			buyerStock = s
			break
		}
	}

	if buyerStock == nil {
		buyerStock = NewStock(buyerPlayer.ID, buyerPlayer.ID, 0)
		buyerPlayer.Stocks = append(buyerPlayer.Stocks, buyerStock)
	}

	buyerStock.Shares += shares

	return nil
}

func CalculateDividend(player *models.Player, turn int) float64 {
	if turn%DividendInterval != 0 {
		return 0
	}

	profitRatio := 0.05
	dividendPerShare := player.Credits * profitRatio / float64(InitialTotalShares)

	return dividendPerShare
}

func DistributeDividends(players []*models.Player, turn int) map[string]float64 {
	dividends := make(map[string]float64)

	if turn%DividendInterval != 0 {
		return dividends
	}

	for _, issuer := range players {
		dividendPerShare := CalculateDividend(issuer, turn)

		if dividendPerShare <= 0 {
			continue
		}

		for _, stock := range issuer.Stocks {
			if stock.IssuerID == issuer.ID {
				continue
			}

			totalDividend := float64(stock.Shares) * dividendPerShare

			holder := findPlayerByID(players, stock.PlayerID)
			if holder != nil {
				holder.Credits += totalDividend
				dividends[holder.ID] += totalDividend
			}
		}
	}

	return dividends
}

func findPlayerByID(players []*models.Player, playerID string) *models.Player {
	for _, player := range players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

func CheckTakeover(player *models.Player, targetPlayer *models.Player, exchanges []*models.Exchange) bool {
	var playerShares int

	for _, stock := range player.Stocks {
		if stock.IssuerID == targetPlayer.ID {
			playerShares += stock.Shares
		}
	}

	ownershipRatio := float64(playerShares) / float64(InitialTotalShares)

	return ownershipRatio >= TakeoverThreshold
}

func ExecuteTakeover(acquirer *models.Player, target *models.Player) {
	for _, station := range target.Stations {
		station.PlayerID = acquirer.ID
		acquirer.Stations = append(acquirer.Stations, station)
	}

	for _, refinery := range target.Refineries {
		refinery.PlayerID = acquirer.ID
		acquirer.Refineries = append(acquirer.Refineries, refinery)
	}

	for _, shipyard := range target.Shipyards {
		shipyard.PlayerID = acquirer.ID
		acquirer.Shipyards = append(acquirer.Shipyards, shipyard)
	}

	for _, fleet := range target.Fleets {
		fleet.PlayerID = acquirer.ID
		acquirer.Fleets = append(acquirer.Fleets, fleet)
	}

	for _, ship := range target.Ships {
		ship.PlayerID = acquirer.ID
		acquirer.Ships = append(acquirer.Ships, ship)
	}

	for resource, amount := range target.Resources {
		acquirer.Resources[resource] += amount
	}

	acquirer.Credits += target.Credits

	if target.TechTree != nil && acquirer.TechTree != nil {
		for techType, tech := range target.TechTree.Techs {
			if acquirerTech, exists := acquirer.TechTree.Techs[techType]; exists {
				if tech.Level > acquirerTech.Level {
					acquirerTech.Level = tech.Level
				}
			}
		}
	}

	target.IsDefeated = true
}

func GetOwnershipPercentage(player *models.Player, targetPlayerID string) float64 {
	var playerShares int

	for _, stock := range player.Stocks {
		if stock.IssuerID == targetPlayerID {
			playerShares += stock.Shares
		}
	}

	return float64(playerShares) / float64(InitialTotalShares) * 100
}

func CalculateMarketCap(player *models.Player, exchanges []*models.Exchange) float64 {
	price := CalculateSharePrice(player, exchanges)
	return price * float64(InitialTotalShares)
}
