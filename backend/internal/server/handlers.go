package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"star-mining/internal/cache"
	"star-mining/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type APIHandler struct {
	roomManager    *RoomManager
	cache          *cache.RedisCache
	wsServer       *WebSocketServer
}

type CreateRoomRequest struct {
	RoomName   string `json:"room_name"`
	PlayerName string `json:"player_name"`
	PlayerID   string `json:"player_id,omitempty"`
}

type CreateRoomResponse struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type JoinRoomRequest struct {
	RoomID     string `json:"room_id"`
	PlayerName string `json:"player_name"`
	PlayerID   string `json:"player_id,omitempty"`
}

type JoinRoomResponse struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type LoginRequest struct {
	PlayerName string `json:"player_name"`
}

type LoginResponse struct {
	PlayerID string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type RoomListResponse struct {
	Rooms []RoomInfo `json:"rooms"`
}

type RoomInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
}

type PlayerActionRequest struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAPIHandler(roomManager *RoomManager, cache *cache.RedisCache, wsServer *WebSocketServer) *APIHandler {
	return &APIHandler{
		roomManager: roomManager,
		cache:       cache,
		wsServer:    wsServer,
	}
}

func (h *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RoomName == "" {
		sendError(w, http.StatusBadRequest, "room name is required")
		return
	}

	if req.PlayerName == "" {
		sendError(w, http.StatusBadRequest, "player name is required")
		return
	}

	playerID := req.PlayerID
	if playerID == "" {
		playerID = uuid.New().String()
	}

	room, err := h.roomManager.CreateRoom(req.RoomName, playerID, req.PlayerName)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.cache != nil {
		h.cache.SetJSON("room:"+room.ID, room, 24*time.Hour)
		h.cache.Set("player:"+playerID+":room", room.ID, 24*time.Hour)
	}

	resp := CreateRoomResponse{
		RoomID:   room.ID,
		PlayerID: playerID,
	}

	sendJSON(w, http.StatusCreated, resp)
}

func (h *APIHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerName == "" {
		sendError(w, http.StatusBadRequest, "player name is required")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	playerID := req.PlayerID
	if playerID == "" {
		playerID = uuid.New().String()
	}

	_, err := room.Join(playerID, req.PlayerName)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.cache != nil {
		h.cache.SetJSON("room:"+room.ID, room, 24*time.Hour)
		h.cache.Set("player:"+playerID+":room", room.ID, 24*time.Hour)
	}

	resp := JoinRoomResponse{
		RoomID:   room.ID,
		PlayerID: playerID,
	}

	sendJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := h.roomManager.ListRooms()

	roomInfos := make([]RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		if room.Status == RoomStatusWaiting {
			roomInfos = append(roomInfos, RoomInfo{
				ID:         room.ID,
				Name:       room.Name,
				Status:     room.Status,
				Players:    room.GetPlayerCount(),
				MaxPlayers: room.MaxPlayers,
			})
		}
	}

	resp := RoomListResponse{
		Rooms: roomInfos,
	}

	sendJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	gameState := room.GetGameState()
	sendJSON(w, http.StatusOK, gameState)
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerName == "" {
		sendError(w, http.StatusBadRequest, "player name is required")
		return
	}

	playerID := uuid.New().String()

	if h.cache != nil {
		h.cache.Set("player:"+playerID+":name", req.PlayerName, 24*time.Hour)
	}

	resp := LoginResponse{
		PlayerID:   playerID,
		PlayerName: req.PlayerName,
	}

	sendJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) PlayerReady(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.SetReady(playerID, true); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, gameState)
}

func (h *APIHandler) PlayerUnready(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.SetReady(playerID, false); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, gameState)
}

func (h *APIHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if !room.IsAllReady() {
		sendError(w, http.StatusBadRequest, "not all players are ready")
		return
	}

	if err := room.StartGame(); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	event := &EventData{
		Event:     "game_started",
		Timestamp: time.Now().Unix(),
	}
	eventMsg, _ := NewMessage(MsgTypeEvent, event)
	room.Broadcast(eventMsg)

	stateMsg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(stateMsg)

	sendJSON(w, http.StatusOK, gameState)
}

func (h *APIHandler) PlayerAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req PlayerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	action := &PlayerActionData{
		Action: req.Action,
		Params: req.Params,
	}

	msg, _ := NewMessageWithPlayer(MsgTypePlayerAction, roomID, playerID, action)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	h.wsServer.HandleConnection(w, r, playerID, roomID)
}

type BuildStationRequest struct {
	BodyID       string              `json:"body_id"`
	ResourceType models.ResourceType `json:"resource_type"`
}

type BuildRefineryRequest struct {
	BodyID         string              `json:"body_id"`
	InputResource  models.ResourceType `json:"input_resource"`
	OutputResource models.ResourceType `json:"output_resource"`
}

type PlaceOrderRequest struct {
	ExchangeID string              `json:"exchange_id"`
	Resource   models.ResourceType `json:"resource"`
	Quantity   float64             `json:"quantity"`
	Price      float64             `json:"price"`
}

type CancelOrderRequest struct {
	ExchangeID string `json:"exchange_id"`
	OrderID    string `json:"order_id"`
}

func (h *APIHandler) BuildStation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BuildStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BuildStation(playerID, req.BodyID, req.ResourceType); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) BuildRefinery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BuildRefineryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BuildRefinery(playerID, req.BodyID, req.InputResource, req.OutputResource); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) PlaceBuyOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.PlaceBuyOrder(playerID, req.ExchangeID, req.Resource, req.Quantity, req.Price); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) PlaceSellOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.PlaceSellOrder(playerID, req.ExchangeID, req.Resource, req.Quantity, req.Price); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.CancelOrder(playerID, req.ExchangeID, req.OrderID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) NextTurn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := room.ProcessTurn(); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()

	event := &EventData{
		Event:     "turn_ended",
		Timestamp: time.Now().Unix(),
		Params: map[string]interface{}{
			"turn": gameState.Turn,
		},
	}
	eventMsg, _ := NewMessage(MsgTypeEvent, event)
	room.Broadcast(eventMsg)

	stateMsg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(stateMsg)

	sendJSON(w, http.StatusOK, gameState)
}

func (h *APIHandler) GetPlayerState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	game := room.GetGame()
	if game == nil {
		sendError(w, http.StatusBadRequest, "game not started")
		return
	}

	player, exists := game.GetPlayer(playerID)
	if !exists {
		sendError(w, http.StatusNotFound, "player not found")
		return
	}

	sendJSON(w, http.StatusOK, player)
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, ErrorResponse{Error: message})
}

func validatePlayerInRoom(room *Room, playerID string) error {
	_, exists := room.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not in room")
	}
	return nil
}

type BuildShipyardRequest struct {
	BodyID string `json:"body_id"`
}

type BuildShipRequest struct {
	ShipyardID string            `json:"shipyard_id"`
	ShipType   models.ShipType `json:"ship_type"`
}

type CreateFleetRequest struct {
	Name    string   `json:"name"`
	ShipIDs []string `json:"ship_ids"`
}

type MoveFleetRequest struct {
	FleetID      string `json:"fleet_id"`
	TargetBodyID string `json:"target_body_id"`
}

type ResearchTechRequest struct {
	TechType models.TechType `json:"tech_type"`
}

type PlaceBidRequest struct {
	BodyID string  `json:"body_id"`
	Amount float64 `json:"amount"`
}

type BlockLaneRequest struct {
	LaneID string  `json:"lane_id"`
	Toll   float64 `json:"toll"`
}

type HirePiratesRequest struct {
	TargetPlayerID string `json:"target_player_id"`
}

type StockRequest struct {
	TargetPlayerID string `json:"target_player_id"`
	Shares         int    `json:"shares"`
}

type TakeoverRequest struct {
	TargetPlayerID string `json:"target_player_id"`
}

type CargoRequest struct {
	FleetID  string              `json:"fleet_id"`
	Resource models.ResourceType `json:"resource"`
	Amount   float64             `json:"amount"`
}

type UpgradeStationRequest struct {
	StationID string `json:"station_id"`
}

type UpgradeRefineryRequest struct {
	RefineryID string `json:"refinery_id"`
}

func (h *APIHandler) BuildShipyard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BuildShipyardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BuildShipyard(playerID, req.BodyID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) BuildShip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BuildShipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BuildShip(playerID, req.ShipyardID, req.ShipType); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) CreateFleet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req CreateFleetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	fleet, err := room.CreateFleet(playerID, req.Name, req.ShipIDs)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, fleet)
}

func (h *APIHandler) MoveFleet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req MoveFleetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.MoveFleet(playerID, req.FleetID, req.TargetBodyID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) ResearchTech(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req ResearchTechRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.ResearchTech(playerID, req.TechType); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.PlaceBid(playerID, req.BodyID, req.Amount); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) BlockLane(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BlockLaneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BlockLane(playerID, req.LaneID, req.Toll); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) HirePirates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req HirePiratesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.HirePirates(playerID, req.TargetPlayerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) BuyStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.BuyStock(playerID, req.TargetPlayerID, req.Shares); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) SellStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.SellStock(playerID, req.TargetPlayerID, req.Shares); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) ProposeTakeover(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req TakeoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.ProposeTakeover(playerID, req.TargetPlayerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) LoadCargo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req CargoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.LoadCargo(playerID, req.FleetID, req.Resource, req.Amount); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) UnloadCargo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req CargoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.UnloadCargo(playerID, req.FleetID, req.Resource, req.Amount); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) UpgradeStation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req UpgradeStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.UpgradeStation(playerID, req.StationID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) UpgradeRefinery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req UpgradeRefineryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.UpgradeRefinery(playerID, req.RefineryID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type RecruitSpyRequest struct{}

type AssignSpyMissionRequest struct {
	SpyID          string `json:"spy_id"`
	TargetPlayerID string `json:"target_player_id"`
	MissionType    string `json:"mission_type"`
	ThirdPartyID   string `json:"third_party_id,omitempty"`
}

type SetCounterSpyLevelRequest struct {
	Level string `json:"level"`
}

type SellIntelRequest struct {
	IntelID string  `json:"intel_id"`
	Price   float64 `json:"price"`
}

type BuyIntelRequest struct {
	ListingID string `json:"listing_id"`
}

type CancelIntelListingRequest struct {
	ListingID string `json:"listing_id"`
}

type ChooseSpySpecRequest struct {
	SpyID          string `json:"spy_id"`
	Specialization string `json:"specialization"`
}

func (h *APIHandler) RecruitSpy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	spy, err := room.RecruitSpy(playerID)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, spy)
}

func (h *APIHandler) AssignSpyMission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req AssignSpyMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	mission, err := room.AssignSpyMission(req.SpyID, playerID, req.TargetPlayerID, models.SpyMissionType(req.MissionType), req.ThirdPartyID)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, mission)
}

func (h *APIHandler) SetCounterSpyLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req SetCounterSpyLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.SetCounterSpyLevel(playerID, models.CounterSpyLevel(req.Level)); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) SellIntelOnMarket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req SellIntelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	listing, err := room.SellIntelOnMarket(playerID, req.IntelID, req.Price)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, listing)
}

func (h *APIHandler) BuyIntelFromMarket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req BuyIntelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	intel, err := room.BuyIntelFromMarket(playerID, req.ListingID)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, intel)
}

func (h *APIHandler) CancelIntelListing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req CancelIntelListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.CancelIntelListing(playerID, req.ListingID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *APIHandler) ChooseSpySpec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	playerID := vars["playerId"]

	var req ChooseSpySpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		sendError(w, http.StatusNotFound, "room not found")
		return
	}

	if err := validatePlayerInRoom(room, playerID); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := room.ChooseSpySpec(playerID, req.SpyID, models.SpySpecialization(req.Specialization)); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	gameState := room.GetGameState()
	msg, _ := NewMessage(MsgTypeGameState, gameState)
	room.Broadcast(msg)

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
