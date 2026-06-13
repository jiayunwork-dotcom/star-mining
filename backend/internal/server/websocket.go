package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"star-mining/internal/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketConn struct {
	conn     *websocket.Conn
	send     chan *Message
	playerID string
	roomID   string
	mu       sync.Mutex
	closed   bool
}

type WebSocketServer struct {
	roomManager *RoomManager
	upgrader    websocket.Upgrader
}

func NewWebSocketServer(roomManager *RoomManager) *WebSocketServer {
	return &WebSocketServer{
		roomManager: roomManager,
		upgrader:    upgrader,
	}
}

func NewWebSocketConn(conn *websocket.Conn, playerID string, roomID string) *WebSocketConn {
	return &WebSocketConn{
		conn:     conn,
		send:     make(chan *Message, 256),
		playerID: playerID,
		roomID:   roomID,
		closed:   false,
	}
}

func (ws *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request, playerID string, roomID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade websocket: %v", err)
		return
	}

	wsConn := NewWebSocketConn(conn, playerID, roomID)

	room, exists := ws.roomManager.GetRoom(roomID)
	if !exists {
		log.Printf("room %s not found", roomID)
		conn.Close()
		return
	}

	player, exists := room.GetPlayer(playerID)
	if !exists {
		log.Printf("player %s not found in room %s", playerID, roomID)
		conn.Close()
		return
	}

	player.Conn = wsConn
	room.SetPlayerOnline(playerID, true)

	defer func() {
		room.SetPlayerOnline(playerID, false)
		wsConn.Close()
	}()

	go wsConn.writePump()
	wsConn.readPump(ws.roomManager)
}

func (c *WebSocketConn) readPump(roomManager *RoomManager) {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		c.handleMessage(messageData, roomManager)
	}
}

func (c *WebSocketConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := msg.ToBytes()
			if err != nil {
				log.Printf("failed to serialize message: %v", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *WebSocketConn) handleMessage(data []byte, roomManager *RoomManager) {
	msg, err := ParseMessage(data)
	if err != nil {
		log.Printf("failed to parse message: %v", err)
		return
	}

	room, exists := roomManager.GetRoom(c.roomID)
	if !exists {
		return
	}

	switch msg.Type {
	case MsgTypeHeartbeat:
		c.handleHeartbeat(room)
	case MsgTypePlayerAction:
		c.handlePlayerAction(msg, room)
	case MsgTypeChat:
		c.handleChat(msg, room)
	case MsgTypeGameState:
		c.handleGameStateRequest(room)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (c *WebSocketConn) handleHeartbeat(room *Room) {
	heartbeat := &HeartbeatData{
		Timestamp: time.Now().Unix(),
	}

	msg, err := NewMessage(MsgTypeHeartbeat, heartbeat)
	if err != nil {
		log.Printf("failed to create heartbeat message: %v", err)
		return
	}

	c.Send(msg)
}

func (c *WebSocketConn) handlePlayerAction(msg *Message, room *Room) {
	action, err := msg.ParsePlayerAction()
	if err != nil {
		c.sendError(400, fmt.Sprintf("failed to parse player action: %v", err))
		return
	}

	var execErr error

	switch action.Action {
	case ActionReady:
		execErr = room.SetReady(c.playerID, true)
	case ActionUnready:
		execErr = room.SetReady(c.playerID, false)
	case ActionStartGame:
		execErr = room.StartGame()
	case ActionNextTurn, ActionEndTurn:
		execErr = room.ProcessTurn()
		if execErr == nil {
			go room.BroadcastTurnReports()
			return
		}
	case ActionConfirmTurnReport:
		allConfirmed, confirmErr := room.ConfirmTurnReport(c.playerID)
		if confirmErr != nil {
			c.sendError(400, confirmErr.Error())
			return
		}
		ackMsg, _ := NewMessageWithPlayer(MsgTypeTurnReportAck, c.roomID, c.playerID, &TurnReportAckData{
			Turn:      room.Turn,
			PlayerID:  c.playerID,
			Confirmed: true,
		})
		room.Broadcast(ackMsg)
		if allConfirmed {
			c.broadcastGameState(room)
		}
		return
	case ActionBuildStation:
		execErr = c.handleBuildStation(action.Params, room)
	case ActionBuildRefinery:
		execErr = c.handleBuildRefinery(action.Params, room)
	case ActionBuildShipyard:
		execErr = c.handleBuildShipyard(action.Params, room)
	case ActionBuildShip:
		execErr = c.handleBuildShip(action.Params, room)
	case ActionCreateFleet:
		execErr = c.handleCreateFleet(action.Params, room)
	case ActionMoveFleet:
		execErr = c.handleMoveFleet(action.Params, room)
	case ActionResearchTech:
		execErr = c.handleResearchTech(action.Params, room)
	case ActionPlaceBuyOrder:
		execErr = c.handlePlaceBuyOrder(action.Params, room)
	case ActionPlaceSellOrder:
		execErr = c.handlePlaceSellOrder(action.Params, room)
	case ActionCancelOrder:
		execErr = c.handleCancelOrder(action.Params, room)
	case ActionPlaceBid:
		execErr = c.handlePlaceBid(action.Params, room)
	case ActionBlockLane:
		execErr = c.handleBlockLane(action.Params, room)
	case ActionHirePirates:
		execErr = c.handleHirePirates(action.Params, room)
	case ActionBuyStock:
		execErr = c.handleBuyStock(action.Params, room)
	case ActionSellStock:
		execErr = c.handleSellStock(action.Params, room)
	case ActionProposeTakeover:
		execErr = c.handleProposeTakeover(action.Params, room)
	case ActionLoadCargo:
		execErr = c.handleLoadCargo(action.Params, room)
	case ActionUnloadCargo:
		execErr = c.handleUnloadCargo(action.Params, room)
	case ActionUpgradeStation:
		execErr = c.handleUpgradeStation(action.Params, room)
	case ActionUpgradeRefinery:
		execErr = c.handleUpgradeRefinery(action.Params, room)
	case ActionCreateAlliance:
		execErr = c.handleCreateAlliance(action.Params, room)
	case ActionSendAllianceInvite:
		execErr = c.handleSendAllianceInvite(action.Params, room)
	case ActionAcceptAllianceInvite:
		execErr = c.handleAcceptAllianceInvite(action.Params, room)
	case ActionRejectAllianceInvite:
		execErr = c.handleRejectAllianceInvite(action.Params, room)
	case ActionLeaveAlliance:
		execErr = c.handleLeaveAlliance(room)
	case ActionKickAllianceMember:
		execErr = c.handleKickAllianceMember(action.Params, room)
	case ActionDisbandAlliance:
		execErr = c.handleDisbandAlliance(room)
	case ActionCreateTradeAgreement:
		execErr = c.handleCreateTradeAgreement(action.Params, room)
	case ActionRenewTradeAgreement:
		execErr = c.handleRenewTradeAgreement(action.Params, room)
	case ActionInitiateJointMilitary:
		execErr = c.handleInitiateJointMilitary(action.Params, room)
	case ActionJoinMilitaryAction:
		execErr = c.handleJoinMilitaryAction(action.Params, room)
	case ActionDeclineMilitaryAction:
		execErr = c.handleDeclineMilitaryAction(action.Params, room)
	case ActionTransferLeadership:
		execErr = c.handleTransferLeadership(action.Params, room)
	case ActionGetGameState:
		c.handleGameStateRequest(room)
		return
	default:
		c.sendError(400, fmt.Sprintf("unknown action: %s", action.Action))
		return
	}

	if execErr != nil {
		c.sendError(400, execErr.Error())
		return
	}

	c.broadcastGameState(room)
}

func (c *WebSocketConn) sendError(code int, message string) {
	errorData := &ErrorData{
		Code:    code,
		Message: message,
	}

	msg, err := NewMessage(MsgTypeError, errorData)
	if err != nil {
		log.Printf("failed to create error message: %v", err)
		return
	}

	c.Send(msg)
}

func (c *WebSocketConn) broadcastGameState(room *Room) {
	gameState := room.GetGameState()

	msg, err := NewMessage(MsgTypeGameState, gameState)
	if err != nil {
		log.Printf("failed to create game state message: %v", err)
		return
	}

	room.Broadcast(msg)
}

func getStringParam(params map[string]interface{}, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing parameter: %s", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return str, nil
}

func getFloatParam(params map[string]interface{}, key string) (float64, error) {
	val, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing parameter: %s", key)
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("parameter %s must be a number", key)
	}
}

func getIntParam(params map[string]interface{}, key string) (int, error) {
	val, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing parameter: %s", key)
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parameter %s must be an integer", key)
	}
}

func getStringSliceParam(params map[string]interface{}, key string) ([]string, error) {
	val, ok := params[key]
	if !ok {
		return nil, fmt.Errorf("missing parameter: %s", key)
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parameter %s must be an array", key)
	}
	result := make([]string, len(arr))
	for i, v := range arr {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("parameter %s must contain strings", key)
		}
		result[i] = str
	}
	return result, nil
}

func (c *WebSocketConn) handleBuildStation(params map[string]interface{}, room *Room) error {
	bodyID, err := getStringParam(params, "body_id")
	if err != nil {
		return err
	}

	resourceTypeStr, err := getStringParam(params, "resource_type")
	if err != nil {
		return err
	}

	return room.BuildStation(c.playerID, bodyID, models.ResourceType(resourceTypeStr))
}

func (c *WebSocketConn) handleBuildRefinery(params map[string]interface{}, room *Room) error {
	bodyID, err := getStringParam(params, "body_id")
	if err != nil {
		return err
	}

	inputResourceStr, err := getStringParam(params, "input_resource")
	if err != nil {
		return err
	}

	outputResourceStr, err := getStringParam(params, "output_resource")
	if err != nil {
		return err
	}

	return room.BuildRefinery(c.playerID, bodyID, models.ResourceType(inputResourceStr), models.ResourceType(outputResourceStr))
}

func (c *WebSocketConn) handleBuildShipyard(params map[string]interface{}, room *Room) error {
	bodyID, err := getStringParam(params, "body_id")
	if err != nil {
		return err
	}

	return room.BuildShipyard(c.playerID, bodyID)
}

func (c *WebSocketConn) handleBuildShip(params map[string]interface{}, room *Room) error {
	shipyardID, err := getStringParam(params, "shipyard_id")
	if err != nil {
		return err
	}

	shipTypeStr, err := getStringParam(params, "ship_type")
	if err != nil {
		return err
	}

	return room.BuildShip(c.playerID, shipyardID, models.ShipType(shipTypeStr))
}

func (c *WebSocketConn) handleCreateFleet(params map[string]interface{}, room *Room) error {
	name, err := getStringParam(params, "name")
	if err != nil {
		return err
	}

	shipIDs, err := getStringSliceParam(params, "ship_ids")
	if err != nil {
		return err
	}

	_, err = room.CreateFleet(c.playerID, name, shipIDs)
	return err
}

func (c *WebSocketConn) handleMoveFleet(params map[string]interface{}, room *Room) error {
	fleetID, err := getStringParam(params, "fleet_id")
	if err != nil {
		return err
	}

	targetBodyID, err := getStringParam(params, "target_body_id")
	if err != nil {
		return err
	}

	return room.MoveFleet(c.playerID, fleetID, targetBodyID)
}

func (c *WebSocketConn) handleResearchTech(params map[string]interface{}, room *Room) error {
	techTypeStr, err := getStringParam(params, "tech_type")
	if err != nil {
		return err
	}

	return room.ResearchTech(c.playerID, models.TechType(techTypeStr))
}

func (c *WebSocketConn) handlePlaceBuyOrder(params map[string]interface{}, room *Room) error {
	exchangeID, err := getStringParam(params, "exchange_id")
	if err != nil {
		return err
	}

	resourceStr, err := getStringParam(params, "resource")
	if err != nil {
		return err
	}

	quantity, err := getFloatParam(params, "quantity")
	if err != nil {
		return err
	}

	price, err := getFloatParam(params, "price")
	if err != nil {
		return err
	}

	return room.PlaceBuyOrder(c.playerID, exchangeID, models.ResourceType(resourceStr), quantity, price)
}

func (c *WebSocketConn) handlePlaceSellOrder(params map[string]interface{}, room *Room) error {
	exchangeID, err := getStringParam(params, "exchange_id")
	if err != nil {
		return err
	}

	resourceStr, err := getStringParam(params, "resource")
	if err != nil {
		return err
	}

	quantity, err := getFloatParam(params, "quantity")
	if err != nil {
		return err
	}

	price, err := getFloatParam(params, "price")
	if err != nil {
		return err
	}

	return room.PlaceSellOrder(c.playerID, exchangeID, models.ResourceType(resourceStr), quantity, price)
}

func (c *WebSocketConn) handleCancelOrder(params map[string]interface{}, room *Room) error {
	exchangeID, err := getStringParam(params, "exchange_id")
	if err != nil {
		return err
	}

	orderID, err := getStringParam(params, "order_id")
	if err != nil {
		return err
	}

	return room.CancelOrder(c.playerID, exchangeID, orderID)
}

func (c *WebSocketConn) handlePlaceBid(params map[string]interface{}, room *Room) error {
	bodyID, err := getStringParam(params, "body_id")
	if err != nil {
		return err
	}

	amount, err := getFloatParam(params, "amount")
	if err != nil {
		return err
	}

	return room.PlaceBid(c.playerID, bodyID, amount)
}

func (c *WebSocketConn) handleBlockLane(params map[string]interface{}, room *Room) error {
	laneID, err := getStringParam(params, "lane_id")
	if err != nil {
		return err
	}

	toll, err := getFloatParam(params, "toll")
	if err != nil {
		return err
	}

	return room.BlockLane(c.playerID, laneID, toll)
}

func (c *WebSocketConn) handleHirePirates(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}

	return room.HirePirates(c.playerID, targetPlayerID)
}

func (c *WebSocketConn) handleBuyStock(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}

	shares, err := getIntParam(params, "shares")
	if err != nil {
		return err
	}

	return room.BuyStock(c.playerID, targetPlayerID, shares)
}

func (c *WebSocketConn) handleSellStock(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}

	shares, err := getIntParam(params, "shares")
	if err != nil {
		return err
	}

	return room.SellStock(c.playerID, targetPlayerID, shares)
}

func (c *WebSocketConn) handleProposeTakeover(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}

	return room.ProposeTakeover(c.playerID, targetPlayerID)
}

func (c *WebSocketConn) handleLoadCargo(params map[string]interface{}, room *Room) error {
	fleetID, err := getStringParam(params, "fleet_id")
	if err != nil {
		return err
	}

	resourceStr, err := getStringParam(params, "resource")
	if err != nil {
		return err
	}

	amount, err := getFloatParam(params, "amount")
	if err != nil {
		return err
	}

	return room.LoadCargo(c.playerID, fleetID, models.ResourceType(resourceStr), amount)
}

func (c *WebSocketConn) handleUnloadCargo(params map[string]interface{}, room *Room) error {
	fleetID, err := getStringParam(params, "fleet_id")
	if err != nil {
		return err
	}

	resourceStr, err := getStringParam(params, "resource")
	if err != nil {
		return err
	}

	amount, err := getFloatParam(params, "amount")
	if err != nil {
		return err
	}

	return room.UnloadCargo(c.playerID, fleetID, models.ResourceType(resourceStr), amount)
}

func (c *WebSocketConn) handleUpgradeStation(params map[string]interface{}, room *Room) error {
	stationID, err := getStringParam(params, "station_id")
	if err != nil {
		return err
	}

	return room.UpgradeStation(c.playerID, stationID)
}

func (c *WebSocketConn) handleUpgradeRefinery(params map[string]interface{}, room *Room) error {
	refineryID, err := getStringParam(params, "refinery_id")
	if err != nil {
		return err
	}

	return room.UpgradeRefinery(c.playerID, refineryID)
}

func (c *WebSocketConn) handleCreateAlliance(params map[string]interface{}, room *Room) error {
	name, err := getStringParam(params, "name")
	if err != nil {
		return err
	}
	color, err := getStringParam(params, "color")
	if err != nil {
		return err
	}
	_, err2 := room.CreateAlliance(c.playerID, name, models.AllianceColor(color))
	return err2
}

func (c *WebSocketConn) handleSendAllianceInvite(params map[string]interface{}, room *Room) error {
	allianceID, err := getStringParam(params, "alliance_id")
	if err != nil {
		return err
	}
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}

	err = room.SendAllianceInvite(c.playerID, allianceID, targetPlayerID)
	if err != nil {
		return err
	}

	game := room.GetGame()
	if game != nil {
		alliance := game.FindPlayerAlliancePublic(c.playerID)
		if alliance != nil {
			inviteData := &AllianceInviteData{
				AllianceID:   allianceID,
				AllianceName: alliance.Name,
				InviterID:    c.playerID,
				TargetID:     targetPlayerID,
				ExpiryTurn:   game.GetGameState().Turn,
			}
			player, exists := room.GetPlayer(c.playerID)
			if exists {
				inviteData.InviterName = player.Name
			}
			inviteMsg, _ := NewMessage(MsgTypeAllianceInvite, inviteData)
			_ = room.SendToPlayer(targetPlayerID, inviteMsg)
		}
	}

	return nil
}

func (c *WebSocketConn) handleAcceptAllianceInvite(params map[string]interface{}, room *Room) error {
	allianceID, err := getStringParam(params, "alliance_id")
	if err != nil {
		return err
	}
	return room.AcceptAllianceInvite(c.playerID, allianceID)
}

func (c *WebSocketConn) handleRejectAllianceInvite(params map[string]interface{}, room *Room) error {
	allianceID, err := getStringParam(params, "alliance_id")
	if err != nil {
		return err
	}
	return room.RejectAllianceInvite(c.playerID, allianceID)
}

func (c *WebSocketConn) handleLeaveAlliance(room *Room) error {
	return room.LeaveAlliance(c.playerID)
}

func (c *WebSocketConn) handleKickAllianceMember(params map[string]interface{}, room *Room) error {
	targetID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}
	return room.KickAllianceMember(c.playerID, targetID)
}

func (c *WebSocketConn) handleDisbandAlliance(room *Room) error {
	return room.DisbandAlliance(c.playerID)
}

func (c *WebSocketConn) handleCreateTradeAgreement(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}
	_, err2 := room.CreateTradeAgreement(c.playerID, targetPlayerID)
	return err2
}

func (c *WebSocketConn) handleRenewTradeAgreement(params map[string]interface{}, room *Room) error {
	agreementID, err := getStringParam(params, "agreement_id")
	if err != nil {
		return err
	}
	return room.RenewTradeAgreement(c.playerID, agreementID)
}

func (c *WebSocketConn) handleInitiateJointMilitary(params map[string]interface{}, room *Room) error {
	targetPlayerID, err := getStringParam(params, "target_player_id")
	if err != nil {
		return err
	}
	targetBodyID, err := getStringParam(params, "target_body_id")
	if err != nil {
		return err
	}
	_, err2 := room.InitiateJointMilitaryAction(c.playerID, targetPlayerID, targetBodyID)
	return err2
}

func (c *WebSocketConn) handleJoinMilitaryAction(params map[string]interface{}, room *Room) error {
	actionID, err := getStringParam(params, "action_id")
	if err != nil {
		return err
	}
	fleetID, err := getStringParam(params, "fleet_id")
	if err != nil {
		return err
	}
	return room.JoinMilitaryAction(c.playerID, actionID, fleetID)
}

func (c *WebSocketConn) handleDeclineMilitaryAction(params map[string]interface{}, room *Room) error {
	actionID, err := getStringParam(params, "action_id")
	if err != nil {
		return err
	}
	return room.DeclineMilitaryAction(c.playerID, actionID)
}

func (c *WebSocketConn) handleTransferLeadership(params map[string]interface{}, room *Room) error {
	newLeaderID, err := getStringParam(params, "new_leader_id")
	if err != nil {
		return err
	}
	return room.TransferLeadership(c.playerID, newLeaderID)
}

func (c *WebSocketConn) handleChat(msg *Message, room *Room) {
	chat, err := msg.ParseChat()
	if err != nil {
		log.Printf("failed to parse chat: %v", err)
		return
	}

	chat.PlayerID = c.playerID
	chat.Timestamp = time.Now().Unix()

	player, exists := room.GetPlayer(c.playerID)
	if exists {
		chat.PlayerName = player.Name
	}

	chatMsg, err := NewMessage(MsgTypeChat, chat)
	if err != nil {
		log.Printf("failed to create chat message: %v", err)
		return
	}

	room.Broadcast(chatMsg)
}

func (c *WebSocketConn) handleGameStateRequest(room *Room) {
	gameState := room.GetGameState()

	msg, err := NewMessage(MsgTypeGameState, gameState)
	if err != nil {
		log.Printf("failed to create game state message: %v", err)
		return
	}

	c.Send(msg)
}

func (c *WebSocketConn) Send(msg *Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from send panic: %v", r)
		}
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.send <- msg:
	default:
		log.Printf("send channel full, dropping message for player %s", c.playerID)
	}
}

func (c *WebSocketConn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	close(c.send)
	c.conn.Close()
}

func (c *WebSocketConn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}
