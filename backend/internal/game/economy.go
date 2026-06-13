package game

import (
	"fmt"
	"sort"

	"star-mining/internal/models"
)

const (
	DefaultFeeRate = 0.02
	MaxPriceChange = 0.1
)

func NewExchange(bodyID string) *models.Exchange {
	return &models.Exchange{
		ID:         fmt.Sprintf("exchange-%s", bodyID),
		BodyID:     bodyID,
		Name:       fmt.Sprintf("Exchange at %s", bodyID),
		BuyOrders:  make(map[string]*models.Order),
		SellOrders: make(map[string]*models.Order),
		Prices: map[models.ResourceType]float64{
			models.IronOre:    10.0,
			models.Titanium:   25.0,
			models.Helium3:    50.0,
			models.RareEarth:  100.0,
			models.IceCrystal: 30.0,
			models.Fuel:       15.0,
		},
		FeeRate: DefaultFeeRate,
	}
}

func PlaceOrder(exchange *models.Exchange, order *models.Order) error {
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if order.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}

	order.Status = models.OrderPending
	order.FilledQty = 0

	if order.Type == models.BuyOrder {
		exchange.BuyOrders[order.ID] = order
	} else {
		exchange.SellOrders[order.ID] = order
	}

	return nil
}

func CancelOrder(exchange *models.Exchange, orderID string) error {
	if order, exists := exchange.BuyOrders[orderID]; exists {
		order.Status = models.OrderCancelled
		delete(exchange.BuyOrders, orderID)
		return nil
	}

	if order, exists := exchange.SellOrders[orderID]; exists {
		order.Status = models.OrderCancelled
		delete(exchange.SellOrders, orderID)
		return nil
	}

	return fmt.Errorf("order not found")
}

func MatchOrders(exchange *models.Exchange, players map[string]*models.Player) []*models.Order {
	var filledOrders []*models.Order

	resourceTypes := []models.ResourceType{
		models.IronOre,
		models.Titanium,
		models.Helium3,
		models.RareEarth,
		models.IceCrystal,
		models.Fuel,
	}

	for _, resource := range resourceTypes {
		filled := matchResourceOrders(exchange, resource, players)
		filledOrders = append(filledOrders, filled...)
	}

	UpdatePrices(exchange)

	return filledOrders
}

func matchResourceOrders(exchange *models.Exchange, resource models.ResourceType, players map[string]*models.Player) []*models.Order {
	var filledOrders []*models.Order

	buyOrders := getOrdersByResource(exchange.BuyOrders, resource)
	sellOrders := getOrdersByResource(exchange.SellOrders, resource)

	if len(buyOrders) == 0 || len(sellOrders) == 0 {
		return filledOrders
	}

	sort.Slice(buyOrders, func(i, j int) bool {
		return buyOrders[i].Price > buyOrders[j].Price
	})

	sort.Slice(sellOrders, func(i, j int) bool {
		return sellOrders[i].Price < sellOrders[j].Price
	})

	buyIdx := 0
	sellIdx := 0

	for buyIdx < len(buyOrders) && sellIdx < len(sellOrders) {
		buyOrder := buyOrders[buyIdx]
		sellOrder := sellOrders[sellIdx]

		if buyOrder.Price < sellOrder.Price {
			break
		}

		tradePrice := (buyOrder.Price + sellOrder.Price) / 2
		tradeQty := min(buyOrder.Quantity-buyOrder.FilledQty, sellOrder.Quantity-sellOrder.FilledQty)

		if tradeQty <= 0 {
			break
		}

		buyer := players[buyOrder.PlayerID]
		seller := players[sellOrder.PlayerID]

		if buyer == nil || seller == nil {
			sellIdx++
			continue
		}

		totalCost := tradeQty * tradePrice
		buyerFee := totalCost * exchange.FeeRate
		sellerFee := totalCost * exchange.FeeRate

		if buyer.Credits < totalCost+buyerFee {
			buyIdx++
			continue
		}

		if seller.Resources[resource] < tradeQty {
			sellIdx++
			continue
		}

		buyer.Credits -= totalCost + buyerFee
		seller.Credits += totalCost - sellerFee

		seller.Resources[resource] -= tradeQty
		buyer.Resources[resource] += tradeQty

		seller.TotalTradeProfit += totalCost - sellerFee
		buyer.TotalTradeProfit -= totalCost + buyerFee

		buyOrder.FilledQty += tradeQty
		sellOrder.FilledQty += tradeQty

		if buyOrder.FilledQty >= buyOrder.Quantity {
			buyOrder.Status = models.OrderFilled
			delete(exchange.BuyOrders, buyOrder.ID)
			buyIdx++
		} else {
			buyOrder.Status = models.OrderPartial
		}

		if sellOrder.FilledQty >= sellOrder.Quantity {
			sellOrder.Status = models.OrderFilled
			delete(exchange.SellOrders, sellOrder.ID)
			sellIdx++
		} else {
			sellOrder.Status = models.OrderPartial
		}

		filledOrders = append(filledOrders, buyOrder, sellOrder)
	}

	return filledOrders
}

func getOrdersByResource(orderMap map[string]*models.Order, resource models.ResourceType) []*models.Order {
	var orders []*models.Order
	for _, order := range orderMap {
		if order.Resource == resource && order.Status == models.OrderPending {
			orders = append(orders, order)
		}
	}
	return orders
}

func UpdatePrices(exchange *models.Exchange) {
	resourceTypes := []models.ResourceType{
		models.IronOre,
		models.Titanium,
		models.Helium3,
		models.RareEarth,
		models.IceCrystal,
		models.Fuel,
	}

	for _, resource := range resourceTypes {
		buyVolume := getTotalVolume(exchange.BuyOrders, resource)
		sellVolume := getTotalVolume(exchange.SellOrders, resource)

		currentPrice := exchange.Prices[resource]

		var changeRatio float64
		if sellVolume > 0 {
			changeRatio = (buyVolume - sellVolume) / sellVolume
		} else if buyVolume > 0 {
			changeRatio = 0.1
		} else {
			changeRatio = 0
		}

		if changeRatio > MaxPriceChange {
			changeRatio = MaxPriceChange
		}
		if changeRatio < -MaxPriceChange {
			changeRatio = -MaxPriceChange
		}

		newPrice := currentPrice * (1 + changeRatio)

		if newPrice < 1.0 {
			newPrice = 1.0
		}

		exchange.Prices[resource] = newPrice
	}
}

func getTotalVolume(orderMap map[string]*models.Order, resource models.ResourceType) float64 {
	var total float64
	for _, order := range orderMap {
		if order.Resource == resource {
			remaining := order.Quantity - order.FilledQty
			if remaining > 0 {
				total += remaining
			}
		}
	}
	return total
}

func CalculateFee(exchange *models.Exchange, amount float64) float64 {
	return amount * exchange.FeeRate
}

func GetMarketPrice(exchange *models.Exchange, resource models.ResourceType) float64 {
	if price, exists := exchange.Prices[resource]; exists {
		return price
	}
	return 0
}

func GetBestBuyPrice(exchange *models.Exchange, resource models.ResourceType) float64 {
	var bestPrice float64
	first := true

	for _, order := range exchange.BuyOrders {
		if order.Resource == resource && order.Status == models.OrderPending {
			if first || order.Price > bestPrice {
				bestPrice = order.Price
				first = false
			}
		}
	}

	return bestPrice
}

func GetBestSellPrice(exchange *models.Exchange, resource models.ResourceType) float64 {
	var bestPrice float64
	first := true

	for _, order := range exchange.SellOrders {
		if order.Resource == resource && order.Status == models.OrderPending {
			if first || order.Price < bestPrice {
				bestPrice = order.Price
				first = false
			}
		}
	}

	return bestPrice
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func CalculateNetWorth(player *models.Player, exchanges []*models.Exchange) float64 {
	if len(exchanges) == 0 {
		return player.Credits
	}

	exchange := exchanges[0]

	totalValue := player.Credits

	for resource, amount := range player.Resources {
		if price, exists := exchange.Prices[resource]; exists {
			totalValue += amount * price
		}
	}

	return totalValue
}
