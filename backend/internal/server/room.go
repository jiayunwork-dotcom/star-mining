package server

import (
	"errors"
	"sync"
	"time"

	"star-mining/internal/game"
	"star-mining/internal/models"

	"github.com/google/uuid"
)

const (
	RoomStatusWaiting  = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

const (
	MaxPlayersPerRoom = 6
	MinPlayersToStart = 2
)

type Player struct {
	ID     string
	Name   string
	Ready  bool
	Color  string
	Score  int
	Online bool
	Conn   *WebSocketConn
}

type Room struct {
	ID                   string
	Name                 string
	Status               string
	Players              map[string]*Player
	MaxPlayers           int
	CreatorID            string
	CreatedAt            time.Time
	Turn                 int
	mu                   sync.RWMutex
	game                 *game.GameInstance
	reportConfirmations  map[string]bool
	waitingForReports    bool
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func NewRoom(name string, creatorID string, creatorName string) *Room {
	room := &Room{
		ID:                  generateRoomID(),
		Name:                name,
		Status:              RoomStatusWaiting,
		Players:             make(map[string]*Player),
		MaxPlayers:        MaxPlayersPerRoom,
		CreatorID:         creatorID,
		CreatedAt:         time.Now(),
		Turn:                0,
		reportConfirmations: make(map[string]bool),
		waitingForReports: false,
	}

	player := &Player{
		ID:     creatorID,
		Name:   creatorName,
		Ready:  false,
		Color:  getPlayerColor(0),
		Score:  0,
		Online: true,
	}

	room.Players[creatorID] = player

	return room
}

func generateRoomID() string {
	return uuid.New().String()[:8]
}

func getPlayerColor(index int) string {
	colors := []string{
		"#FF6B6B",
		"#4ECDC4",
		"#45B7D1",
		"#96CEB4",
		"#FFEAA7",
		"#DDA0DD",
	}
	return colors[index%len(colors)]
}

func (rm *RoomManager) CreateRoom(name string, creatorID string, creatorName string) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room := NewRoom(name, creatorID, creatorName)
	rm.rooms[room.ID] = room

	return room, nil
}

func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	return room, exists
}

func (rm *RoomManager) ListRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.rooms, roomID)
}

func (r *Room) Join(playerID string, playerName string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomStatusWaiting {
		return nil, errors.New("room is not in waiting status")
	}

	if len(r.Players) >= r.MaxPlayers {
		return nil, errors.New("room is full")
	}

	if _, exists := r.Players[playerID]; exists {
		return r.Players[playerID], nil
	}

	player := &Player{
		ID:     playerID,
		Name:   playerName,
		Ready:  false,
		Color:  getPlayerColor(len(r.Players)),
		Score:  0,
		Online: true,
	}

	r.Players[playerID] = player

	return player, nil
}

func (r *Room) Leave(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Players[playerID]; !exists {
		return errors.New("player not in room")
	}

	delete(r.Players, playerID)

	if len(r.Players) == 0 {
		r.Status = RoomStatusFinished
	}

	return nil
}

func (r *Room) SetReady(playerID string, ready bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, exists := r.Players[playerID]
	if !exists {
		return errors.New("player not in room")
	}

	player.Ready = ready

	return nil
}

func (r *Room) IsAllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.Players) < MinPlayersToStart {
		return false
	}

	for _, player := range r.Players {
		if !player.Ready {
			return false
		}
	}

	return true
}

func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomStatusWaiting {
		return errors.New("room is not in waiting status")
	}

	if len(r.Players) < MinPlayersToStart {
		return errors.New("not enough players")
	}

	gameInstance := game.NewGameInstance(r.ID, time.Now().UnixNano())

	playerIDs := make([]string, 0, len(r.Players))
	playerNames := make(map[string]string)
	for id, player := range r.Players {
		playerIDs = append(playerIDs, id)
		playerNames[id] = player.Name
	}

	if err := gameInstance.Initialize(playerIDs, playerNames); err != nil {
		return err
	}

	if err := gameInstance.Start(); err != nil {
		return err
	}

	r.game = gameInstance
	r.Status = RoomStatusPlaying
	r.Turn = 1

	return nil
}

func (r *Room) EndGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomStatusPlaying {
		return errors.New("room is not in playing status")
	}

	if r.game != nil {
		r.game.EndGame()
	}

	r.Status = RoomStatusFinished

	return nil
}

func (r *Room) GetPlayer(playerID string) (*Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	player, exists := r.Players[playerID]
	return player, exists
}

func (r *Room) GetPlayersList() []PlayerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]PlayerInfo, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, PlayerInfo{
			ID:     player.ID,
			Name:   player.Name,
			Ready:  player.Ready,
			Color:  player.Color,
			Score:  player.Score,
			Online: player.Online,
		})
	}

	return players
}

func (r *Room) GetPlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Players)
}

func (r *Room) SetPlayerOnline(playerID string, online bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if player, exists := r.Players[playerID]; exists {
		player.Online = online
	}
}

func (r *Room) Broadcast(msg *Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, player := range r.Players {
		if player.Online && player.Conn != nil {
			player.Conn.Send(msg)
		}
	}
}

func (r *Room) SendToPlayer(playerID string, msg *Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	player, exists := r.Players[playerID]
	if !exists {
		return errors.New("player not found")
	}

	if player.Conn != nil {
		player.Conn.Send(msg)
	}

	return nil
}

func (r *Room) GetGameState() *GameStateData {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gameState := &GameStateData{
		RoomID:  r.ID,
		Status:  r.Status,
		Players: r.GetPlayersList(),
		Turn:    r.Turn,
	}

	if r.game != nil {
		gameState.GameData = r.game.GetGameState()
		if r.game.GetGameState() != nil {
			gameState.Turn = r.game.GetGameState().Turn
		}
	}

	return gameState
}

func (r *Room) GetGame() *game.GameInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.game
}

func (r *Room) ProcessTurn() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	if r.waitingForReports {
		return errors.New("waiting for turn report confirmations")
	}

	if err := r.game.ProcessTurn(); err != nil {
		return err
	}

	r.Turn = r.game.GetGameState().Turn

	if r.game.IsGameOver() {
		r.Status = RoomStatusFinished
	}

	r.reportConfirmations = make(map[string]bool)
	r.waitingForReports = true

	return nil
}

func (r *Room) BroadcastTurnReports() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.game == nil || !r.game.IsReportReady() {
		return
	}

	for playerID := range r.Players {
		report := r.game.GetTurnReport(playerID)
		if report == nil {
			continue
		}

		msg, err := NewMessageWithPlayer(MsgTypeTurnReport, r.ID, playerID, report)
		if err != nil {
			continue
		}

		_ = r.SendToPlayer(playerID, msg)
	}
}

func (r *Room) ConfirmTurnReport(playerID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.waitingForReports {
		return false, errors.New("not waiting for report confirmations")
	}

	r.reportConfirmations[playerID] = true

	allConfirmed := true
	for pid := range r.Players {
		if !r.reportConfirmations[pid] {
			allConfirmed = false
			break
		}
	}

	if allConfirmed {
		r.waitingForReports = false
	}

	return allConfirmed, nil
}

func (r *Room) IsWaitingForReports() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.waitingForReports
}

func (r *Room) GetReportConfirmations() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool)
	for k, v := range r.reportConfirmations {
		result[k] = v
	}
	return result
}

func (r *Room) BuildStation(playerID string, bodyID string, resourceType models.ResourceType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BuildStation(playerID, bodyID, resourceType)
}

func (r *Room) BuildRefinery(playerID string, bodyID string, inputResource, outputResource models.ResourceType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BuildRefinery(playerID, bodyID, inputResource, outputResource)
}

func (r *Room) PlaceBuyOrder(playerID string, exchangeID string, resource models.ResourceType, quantity float64, price float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.PlaceBuyOrder(playerID, exchangeID, resource, quantity, price)
}

func (r *Room) PlaceSellOrder(playerID string, exchangeID string, resource models.ResourceType, quantity float64, price float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.PlaceSellOrder(playerID, exchangeID, resource, quantity, price)
}

func (r *Room) CancelOrder(playerID string, exchangeID string, orderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.CancelOrder(playerID, exchangeID, orderID)
}

func (r *Room) BuildShipyard(playerID string, bodyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BuildShipyard(playerID, bodyID)
}

func (r *Room) BuildShip(playerID string, shipyardID string, shipType models.ShipType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BuildShip(playerID, shipyardID, shipType)
}

func (r *Room) CreateFleet(playerID string, name string, shipIDs []string) (*models.Fleet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return nil, errors.New("game not started")
	}

	return r.game.CreateFleet(playerID, name, shipIDs)
}

func (r *Room) MoveFleet(playerID string, fleetID string, targetBodyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.MoveFleet(playerID, fleetID, targetBodyID)
}

func (r *Room) ResearchTech(playerID string, techType models.TechType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.ResearchTech(playerID, techType)
}

func (r *Room) PlaceBid(playerID string, bodyID string, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.PlaceBid(playerID, bodyID, amount)
}

func (r *Room) BlockLane(playerID string, laneID string, toll float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BlockLane(playerID, laneID, toll)
}

func (r *Room) HirePirates(playerID string, targetPlayerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.HirePirates(playerID, targetPlayerID)
}

func (r *Room) BuyStock(playerID string, targetPlayerID string, shares int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.BuyStock(playerID, targetPlayerID, shares)
}

func (r *Room) SellStock(playerID string, targetPlayerID string, shares int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.SellStock(playerID, targetPlayerID, shares)
}

func (r *Room) ProposeTakeover(playerID string, targetPlayerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.ProposeTakeover(playerID, targetPlayerID)
}

func (r *Room) LoadCargo(playerID string, fleetID string, resource models.ResourceType, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.LoadCargo(playerID, fleetID, resource, amount)
}

func (r *Room) UnloadCargo(playerID string, fleetID string, resource models.ResourceType, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.UnloadCargo(playerID, fleetID, resource, amount)
}

func (r *Room) UpgradeStation(playerID string, stationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.UpgradeStation(playerID, stationID)
}

func (r *Room) UpgradeRefinery(playerID string, refineryID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.game == nil {
		return errors.New("game not started")
	}

	return r.game.UpgradeRefinery(playerID, refineryID)
}
