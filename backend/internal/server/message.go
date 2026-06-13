package server

import (
	"encoding/json"
	"fmt"
)

const (
	MsgTypeGameState    = "game_state"
	MsgTypePlayerAction = "player_action"
	MsgTypeChat         = "chat"
	MsgTypeEvent        = "event"
	MsgTypeHeartbeat    = "heartbeat"
	MsgTypeSystem       = "system"
	MsgTypeError        = "error"
)

const (
	ActionReady           = "ready"
	ActionUnready         = "unready"
	ActionStartGame       = "start_game"
	ActionNextTurn        = "next_turn"
	ActionEndTurn         = "end_turn"
	ActionBuildStation    = "build_station"
	ActionBuildRefinery   = "build_refinery"
	ActionBuildShipyard   = "build_shipyard"
	ActionBuildShip       = "build_ship"
	ActionCreateFleet     = "create_fleet"
	ActionMoveFleet       = "move_fleet"
	ActionResearchTech    = "research_tech"
	ActionPlaceBuyOrder   = "place_buy_order"
	ActionPlaceSellOrder  = "place_sell_order"
	ActionCancelOrder     = "cancel_order"
	ActionPlaceBid        = "place_bid"
	ActionBlockLane       = "block_lane"
	ActionHirePirates     = "hire_pirates"
	ActionBuyStock        = "buy_stock"
	ActionSellStock       = "sell_stock"
	ActionProposeTakeover = "propose_takeover"
	ActionLoadCargo       = "load_cargo"
	ActionUnloadCargo     = "unload_cargo"
	ActionUpgradeStation  = "upgrade_station"
	ActionUpgradeRefinery = "upgrade_refinery"
	ActionGetGameState    = "get_game_state"
)

type Message struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id,omitempty"`
	PlayerID string         `json:"player_id,omitempty"`
	Data    json.RawMessage `json:"data"`
}

type GameStateData struct {
	RoomID   string        `json:"room_id"`
	Status   string        `json:"status"`
	Players  []PlayerInfo  `json:"players"`
	GameData interface{}   `json:"game_data,omitempty"`
	Turn     int           `json:"turn,omitempty"`
}

type PlayerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	Color    string `json:"color,omitempty"`
	Score    int    `json:"score,omitempty"`
	Online   bool   `json:"online"`
}

type PlayerActionData struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

type ChatData struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

type EventData struct {
	Event     string                 `json:"event"`
	PlayerID  string                 `json:"player_id,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

type HeartbeatData struct {
	Timestamp int64 `json:"timestamp"`
}

type SystemData struct {
	Message string `json:"message"`
	Level   string `json:"level"`
}

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewMessage(msgType string, data interface{}) (*Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message data: %w", err)
	}

	return &Message{
		Type: msgType,
		Data: dataBytes,
	}, nil
}

func NewMessageWithPlayer(msgType string, roomID string, playerID string, data interface{}) (*Message, error) {
	msg, err := NewMessage(msgType, data)
	if err != nil {
		return nil, err
	}
	msg.RoomID = roomID
	msg.PlayerID = playerID
	return msg, nil
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}

func (m *Message) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) ParseGameState() (*GameStateData, error) {
	var data GameStateData
	if err := json.Unmarshal(m.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse game state data: %w", err)
	}
	return &data, nil
}

func (m *Message) ParsePlayerAction() (*PlayerActionData, error) {
	var data PlayerActionData
	if err := json.Unmarshal(m.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse player action data: %w", err)
	}
	return &data, nil
}

func (m *Message) ParseChat() (*ChatData, error) {
	var data ChatData
	if err := json.Unmarshal(m.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse chat data: %w", err)
	}
	return &data, nil
}

func (m *Message) ParseEvent() (*EventData, error) {
	var data EventData
	if err := json.Unmarshal(m.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse event data: %w", err)
	}
	return &data, nil
}
