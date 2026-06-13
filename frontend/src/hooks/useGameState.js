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

  if (data.turn !== undefined) result.turn = data.turn;
  if (data.phase !== undefined) result.phase = data.phase;
  if (data.started !== undefined) result.started = data.started;
  if (data.game_over !== undefined) result.gameOver = data.game_over;
  if (data.winner_id !== undefined) result.winnerId = data.winner_id;
  if (data.max_turns !== undefined) result.maxTurns = data.max_turns;
  if (data.game_map !== undefined) result.gameMap = data.game_map;
  if (data.exchanges !== undefined) result.exchanges = data.exchanges;
  if (data.random_events !== undefined) result.randomEvents = data.random_events;
  if (data.blockades !== undefined) result.blockades = data.blockades;
  if (data.bids !== undefined) result.bids = data.bids;

  if (data.players && Array.isArray(data.players)) {
    const myId = ws.playerId;
    const me = data.players.find((p) => p.id === myId);
    const others = data.players.filter((p) => p.id !== myId);
    
    if (me) result.myPlayer = me;
    result.otherPlayers = others;
    result.players = data.players;
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
    });

    const unlistenEvent = ws.on(MESSAGE_TYPES.EVENT, (data) => {
      dispatch({ type: 'ADD_EVENT', payload: data });
    });

    const unlistenSystem = ws.on(MESSAGE_TYPES.SYSTEM, (data) => {
      if (data && data.type === 'player_joined') {
        dispatch({ type: 'ADD_PLAYER', payload: data.player });
      } else if (data && data.type === 'player_left') {
        dispatch({ type: 'REMOVE_PLAYER', payload: data.player });
      } else if (data && data.type === 'game_started') {
        dispatch({ type: 'GAME_STARTED', payload: data.game_state });
      } else if (data && data.type === 'room_created') {
        dispatch({ type: 'SET_ROOM', payload: { roomId: data.room_id, players: data.players } });
      } else if (data && data.type === 'room_joined') {
        dispatch({ type: 'SET_ROOM', payload: { roomId: data.room_id, players: data.players } });
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
  }, []);

  const connect = useCallback(async (nickname) => {
    try {
      const result = await ws.connect(nickname);
      if (result && result.success) {
        dispatch({
          type: 'SET_PLAYER',
          payload: { playerId: ws.playerId, nickname },
        });
      }
      return true;
    } catch (e) {
      console.error('连接失败:', e);
      return false;
    }
  }, []);

  const disconnect = useCallback(() => {
    ws.disconnect();
    dispatch({ type: 'SET_CONNECTED', payload: false });
  }, []);

  const createRoom = useCallback((roomName) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.CREATE_ROOM, { name: roomName });
  }, []);

  const joinRoom = useCallback((roomId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.JOIN_ROOM, { room_id: roomId });
  }, []);

  const leaveRoom = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.LEAVE_ROOM, {});
    dispatch({ type: 'LEAVE_ROOM' });
  }, []);

  const listRooms = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.LIST_ROOMS, {});
  }, []);

  const startGame = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.START_GAME, {});
  }, []);

  const endTurn = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.END_TURN, {});
  }, []);

  const buildShip = useCallback((shipType, celestialId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.BUILD, {
      type: 'ship',
      ship_type: shipType,
      celestial_id: celestialId,
    });
  }, []);

  const moveFleet = useCallback((shipIds, targetCelestialId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.MOVE, {
      ship_ids: shipIds,
      target_celestial_id: targetCelestialId,
    });
  }, []);

  const researchTech = useCallback((techType) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.RESEARCH, {
      tech_type: techType,
    });
  }, []);

  const placeOrder = useCallback((orderType, resource, price, quantity) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      order_type: orderType,
      resource,
      price,
      quantity,
    });
  }, []);

  const cancelOrder = useCallback((orderId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      action: 'cancel',
      order_id: orderId,
    });
  }, []);

  const bidAuction = useCallback((auctionId, amount) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      action: 'bid',
      auction_id: auctionId,
      amount,
    });
  }, []);

  const tradeStock = useCallback((companyId, amount) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      action: 'stock',
      company_id: companyId,
      amount,
    });
  }, []);

  const acquireCompany = useCallback((companyId, price) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      action: 'acquire',
      company_id: companyId,
      price,
    });
  }, []);

  const imposeEmbargo = useCallback((companyId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRADE, {
      action: 'embargo',
      target_id: companyId,
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
