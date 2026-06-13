import React, { createContext, useContext, useReducer, useEffect, useCallback } from 'react';
import ws from '../utils/websocket';
import { MESSAGE_TYPES, PLAYER_ACTIONS, GAME_STATES, GAME_PHASES } from '../types/game';

const GameContext = createContext(null);

const initialState = {
  gameState: GAME_STATES.LOBBY,
  currentPage: 'lobby',
  playerId: null,
  nickname: '',
  roomId: null,
  roomName: '',
  rooms: [],
  players: [],
  turn: 0,
  phase: GAME_PHASES.PLANNING,
  gameMap: {
    galaxies: [],
  },
  myPlayer: {
    id: '',
    name: '',
    company_name: '',
    credits: 5000,
    resources: {},
    stations: [],
    refineries: [],
    shipyards: [],
    fleets: [],
    ships: [],
    tech_tree: {},
    stocks: [],
    is_ai: false,
    is_bankrupt: false,
  },
  otherPlayers: [],
  exchanges: [],
  randomEvents: [],
  blockades: [],
  bids: [],
  gameOver: false,
  winnerId: '',
  maxTurns: 60,
  eventLog: [],
  selectedCelestial: null,
  isConnected: false,
  started: false,
};

function gameReducer(state, action) {
  switch (action.type) {
    case 'SET_CONNECTED':
      return { ...state, isConnected: action.payload };

    case 'SET_PLAYER':
      return {
        ...state,
        playerId: action.payload.playerId,
        nickname: action.payload.nickname,
      };

    case 'SET_ROOMS':
      return { ...state, rooms: action.payload };

    case 'SET_ROOM':
      return {
        ...state,
        roomId: action.payload.roomId,
        roomName: action.payload.roomName || state.roomName,
        players: action.payload.players || state.players,
      };

    case 'ADD_PLAYER':
      return {
        ...state,
        players: [...state.players, action.payload],
      };

    case 'REMOVE_PLAYER':
      return {
        ...state,
        players: state.players.filter((p) => p.id !== action.payload.id),
      };

    case 'SET_PAGE':
      return { ...state, currentPage: action.payload };

    case 'GAME_STARTED':
      return {
        ...state,
        gameState: GAME_STATES.PLAYING,
        currentPage: 'game',
        started: true,
        ...parseGameState(action.payload),
      };

    case 'UPDATE_GAME_STATE':
      return {
        ...state,
        ...parseGameState(action.payload),
      };

    case 'SELECT_CELESTIAL':
      return { ...state, selectedCelestial: action.payload };

    case 'ADD_EVENT':
      return {
        ...state,
        eventLog: [action.payload, ...state.eventLog].slice(0, 100),
      };

    case 'SET_EVENTS':
      return { ...state, eventLog: action.payload.slice(0, 100) };

    case 'RESET_GAME':
      return {
        ...initialState,
        playerId: state.playerId,
        nickname: state.nickname,
        isConnected: state.isConnected,
      };

    case 'LEAVE_ROOM':
      return {
        ...state,
        roomId: null,
        roomName: '',
        players: [],
        gameState: GAME_STATES.LOBBY,
        currentPage: 'lobby',
      };

    default:
      return state;
  }
}

function parseGameState(data) {
  if (!data) return {};

  const result = {};

  if (data.status !== undefined) {
    if (data.status === 'playing') {
      result.started = true;
      result.gameState = GAME_STATES.PLAYING;
    } else if (data.status === 'finished') {
      result.gameState = GAME_STATES.ENDED;
    }
  }

  const gameData = data.game_data || data;

  if (data.turn !== undefined) result.turn = data.turn;
  if (gameData.phase !== undefined) result.phase = gameData.phase;
  if (gameData.started !== undefined) result.started = gameData.started;
  if (gameData.game_over !== undefined) result.gameOver = gameData.game_over;
  if (gameData.winner_id !== undefined) result.winnerId = gameData.winner_id;
  if (gameData.max_turns !== undefined) result.maxTurns = gameData.max_turns;
  if (gameData.game_map !== undefined) result.gameMap = gameData.game_map;
  if (gameData.exchanges !== undefined) result.exchanges = gameData.exchanges;
  if (gameData.random_events !== undefined) result.randomEvents = gameData.random_events;
  if (gameData.blockades !== undefined) result.blockades = gameData.blockades;
  if (gameData.bids !== undefined) result.bids = gameData.bids;

  if (data.players && Array.isArray(data.players)) {
    result.players = data.players;
    const myId = ws.playerId;
    const me = data.players.find((p) => p.id === myId);
    const others = data.players.filter((p) => p.id !== myId);

    if (me) {
      result.myPlayer = {
        ...result.myPlayer,
        ...me,
        name: me.name || result.myPlayer.name,
      };
    }
    result.otherPlayers = others;
  }

  if (gameData.players && Array.isArray(gameData.players) && !data.players) {
    const myId = ws.playerId;
    const me = gameData.players.find((p) => p.id === myId);
    const others = gameData.players.filter((p) => p.id !== myId);

    if (me) {
      result.myPlayer = {
        ...result.myPlayer,
        ...me,
      };
    }
    result.otherPlayers = others;
    result.players = gameData.players;
  }

  return result;
}

function getAllCelestials(gameMap) {
  if (!gameMap || !gameMap.galaxies) return [];
  const celestials = [];
  for (const galaxy of gameMap.galaxies) {
    if (galaxy.celestials) {
      celestials.push(...galaxy.celestials);
    }
  }
  return celestials;
}

export function GameProvider({ children }) {
  const [state, dispatch] = useReducer(gameReducer, initialState);

  useEffect(() => {
    const handleOpen = () => dispatch({ type: 'SET_CONNECTED', payload: true });
    const handleClose = () => dispatch({ type: 'SET_CONNECTED', payload: false });

    const unlistenGameState = ws.on(MESSAGE_TYPES.GAME_STATE, (data) => {
      dispatch({ type: 'UPDATE_GAME_STATE', payload: data });

      if (data && data.status === 'playing' && state.currentPage !== 'game') {
        dispatch({ type: 'GAME_STARTED', payload: data });
      }
    });

    const unlistenEvent = ws.on(MESSAGE_TYPES.EVENT, (data) => {
      if (data && data.event === 'game_started') {
        dispatch({ type: 'ADD_EVENT', payload: { event: 'game_started', message: '游戏开始！' } });
      } else if (data && data.event === 'turn_ended') {
        dispatch({ type: 'ADD_EVENT', payload: { event: 'turn_ended', message: `回合 ${data.params?.turn} 结束` } });
      } else {
        dispatch({ type: 'ADD_EVENT', payload: data });
      }
    });

    const unlistenSystem = ws.on(MESSAGE_TYPES.SYSTEM, (data) => {
      if (!data) return;

      if (data.type === 'player_joined' && data.player) {
        dispatch({ type: 'ADD_PLAYER', payload: data.player });
      } else if (data.type === 'player_left' && data.player) {
        dispatch({ type: 'REMOVE_PLAYER', payload: data.player });
      } else if (data.type === 'game_started') {
        dispatch({ type: 'GAME_STARTED', payload: data.game_state });
      } else if (data.type === 'room_created') {
        dispatch({
          type: 'SET_ROOM',
          payload: {
            roomId: data.room_id,
            roomName: data.name,
            players: data.players || [],
          },
        });
      } else if (data.type === 'room_joined') {
        dispatch({
          type: 'SET_ROOM',
          payload: {
            roomId: data.room_id,
            roomName: data.name,
            players: data.players || [],
          },
        });
      }
    });

    const unlistenError = ws.on(MESSAGE_TYPES.ERROR, (data) => {
      console.error('[游戏错误]', data);
    });

    ws.on('open', handleOpen);
    ws.on('close', handleClose);

    if (ws.isConnected()) {
      dispatch({ type: 'SET_CONNECTED', payload: true });
    }

    return () => {
      ws.off('open', handleOpen);
      ws.off('close', handleClose);
      unlistenGameState();
      unlistenEvent();
      unlistenSystem();
      unlistenError();
    };
  }, [state.currentPage]);

  const connect = useCallback(async (nickname) => {
    try {
      const loginResult = await ws.login(nickname);
      dispatch({
        type: 'SET_PLAYER',
        payload: { playerId: loginResult.player_id, nickname: loginResult.player_name },
      });
      return true;
    } catch (e) {
      console.error('登录失败:', e);
      return false;
    }
  }, []);

  const disconnect = useCallback(() => {
    ws.disconnect();
    dispatch({ type: 'SET_CONNECTED', payload: false });
  }, []);

  const createRoom = useCallback(async (roomName) => {
    try {
      const result = await ws.createRoom(roomName);
      dispatch({
        type: 'SET_ROOM',
        payload: {
          roomId: result.room_id,
          roomName: result.room_name || roomName,
          players: result.players || [],
        },
      });
      await ws.connect();
      return true;
    } catch (e) {
      console.error('创建房间失败:', e);
      return false;
    }
  }, []);

  const joinRoom = useCallback(async (roomId) => {
    try {
      const result = await ws.joinRoom(roomId);
      dispatch({
        type: 'SET_ROOM',
        payload: {
          roomId: result.room_id,
          roomName: result.room_name || '游戏房间',
          players: result.players || [],
        },
      });
      await ws.connect();
      return true;
    } catch (e) {
      console.error('加入房间失败:', e);
      return false;
    }
  }, []);

  const leaveRoom = useCallback(() => {
    ws.disconnect();
    dispatch({ type: 'LEAVE_ROOM' });
  }, []);

  const listRooms = useCallback(async () => {
    try {
      const result = await ws.listRooms();
      dispatch({ type: 'SET_ROOMS', payload: result.rooms || [] });
      return true;
    } catch (e) {
      console.error('获取房间列表失败:', e);
      return false;
    }
  }, []);

  const startGame = useCallback(async () => {
    try {
      if (!state.roomId) {
        console.error('没有房间ID，无法开始游戏');
        return false;
      }
      await ws.startGame(state.roomId);
      return true;
    } catch (e) {
      console.error('开始游戏失败:', e);
      return false;
    }
  }, [state.roomId]);

  const endTurn = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.END_TURN, {});
  }, []);

  const buildShip = useCallback((shipType, celestialId) => {
    ws.sendPlayerAction('build_ship', {
      ship_type: shipType,
      shipyard_id: celestialId,
    });
  }, []);

  const moveFleet = useCallback((shipIds, targetCelestialId) => {
    ws.sendPlayerAction('move_fleet', {
      fleet_id: shipIds,
      target_body_id: targetCelestialId,
    });
  }, []);

  const researchTech = useCallback((techType) => {
    ws.sendPlayerAction('research_tech', {
      tech_type: techType,
    });
  }, []);

  const placeOrder = useCallback((orderType, resource, price, quantity) => {
    const action = orderType === 'buy' ? 'place_buy_order' : 'place_sell_order';
    ws.sendPlayerAction(action, {
      resource,
      price,
      quantity,
    });
  }, []);

  const cancelOrder = useCallback((orderId) => {
    ws.sendPlayerAction('cancel_order', {
      order_id: orderId,
    });
  }, []);

  const bidAuction = useCallback((auctionId, amount) => {
    ws.sendPlayerAction('place_bid', {
      body_id: auctionId,
      amount,
    });
  }, []);

  const tradeStock = useCallback((companyId, amount) => {
    const action = amount > 0 ? 'buy_stock' : 'sell_stock';
    ws.sendPlayerAction(action, {
      target_player_id: companyId,
      shares: Math.abs(amount),
    });
  }, []);

  const acquireCompany = useCallback((companyId, price) => {
    ws.sendPlayerAction('propose_takeover', {
      target_player_id: companyId,
    });
  }, []);

  const imposeEmbargo = useCallback((companyId) => {
    ws.sendPlayerAction('block_lane', {
      lane_id: companyId,
    });
  }, []);

  const ready = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.READY, {});
  }, []);

  const unready = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.UNREADY, {});
  }, []);

  const selectCelestial = useCallback((celestial) => {
    dispatch({ type: 'SELECT_CELESTIAL', payload: celestial });
  }, []);

  const setPage = useCallback((page) => {
    dispatch({ type: 'SET_PAGE', payload: page });
  }, []);

  const value = {
    state,
    connect,
    disconnect,
    createRoom,
    joinRoom,
    leaveRoom,
    listRooms,
    startGame,
    endTurn,
    buildShip,
    moveFleet,
    researchTech,
    placeOrder,
    cancelOrder,
    bidAuction,
    tradeStock,
    acquireCompany,
    imposeEmbargo,
    ready,
    unready,
    selectCelestial,
    setPage,
    getAllCelestials: () => getAllCelestials(state.gameMap),
  };

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}

export function useGameState() {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error('useGameState must be used within a GameProvider');
  }
  return context;
}

export default useGameState;
