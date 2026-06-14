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
  alliances: [],
  tradeAgreements: [],
  jointMilitaryActions: [],
  diplomacyRelations: [],
  playerCooldowns: [],
  allianceInvites: [],
  allianceWars: [],
  sanctionProposals: [],
  activeSanctions: [],
  spies: [],
  spyMissions: [],
  intelligences: [],
  intelMarketListings: [],
  counterSpySettings: [],
  gameOver: false,
  winnerId: '',
  maxTurns: 60,
  eventLog: [],
  selectedCelestial: null,
  isConnected: false,
  started: false,
  turnReport: null,
  showTurnReport: false,
  reportConfirmations: {},
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

    case 'SHOW_TURN_REPORT':
      return {
        ...state,
        turnReport: action.payload.report,
        showTurnReport: true,
        reportConfirmations: action.payload.confirmations || { [state.playerId]: false },
      };

    case 'HIDE_TURN_REPORT':
      return {
        ...state,
        showTurnReport: false,
      };

    case 'UPDATE_REPORT_CONFIRMATIONS':
      return {
        ...state,
        reportConfirmations: {
          ...state.reportConfirmations,
          ...action.payload,
        },
      };

    case 'MY_CONFIRM_REPORT':
      return {
        ...state,
        reportConfirmations: {
          ...state.reportConfirmations,
          [state.playerId]: true,
        },
      };

    case 'ADD_ALLIANCE_INVITE':
      return {
        ...state,
        allianceInvites: [...state.allianceInvites, action.payload],
      };

    case 'REMOVE_ALLIANCE_INVITE':
      return {
        ...state,
        allianceInvites: state.allianceInvites.filter(
          (inv) => inv.alliance_id !== action.payload.alliance_id
        ),
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
  if (gameData.alliances !== undefined) result.alliances = gameData.alliances;
  if (gameData.trade_agreements !== undefined) result.tradeAgreements = gameData.trade_agreements;
  if (gameData.joint_military_actions !== undefined) result.jointMilitaryActions = gameData.joint_military_actions;
  if (gameData.diplomacy_relations !== undefined) result.diplomacyRelations = gameData.diplomacy_relations;
  if (gameData.player_cooldowns !== undefined) result.playerCooldowns = gameData.player_cooldowns;
  if (gameData.alliance_wars !== undefined) result.allianceWars = gameData.alliance_wars;
  if (gameData.sanction_proposals !== undefined) result.sanctionProposals = gameData.sanction_proposals;
  if (gameData.active_sanctions !== undefined) result.activeSanctions = gameData.active_sanctions;
  if (gameData.spies !== undefined) result.spies = gameData.spies;
  if (gameData.spy_missions !== undefined) result.spyMissions = gameData.spy_missions;
  if (gameData.intelligences !== undefined) result.intelligences = gameData.intelligences;
  if (gameData.intel_market_listings !== undefined) result.intelMarketListings = gameData.intel_market_listings;
  if (gameData.counter_spy_settings !== undefined) result.counterSpySettings = gameData.counter_spy_settings;

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

    const unlistenTurnReport = ws.on(MESSAGE_TYPES.TURN_REPORT, (data) => {
      if (data && data.player_id === ws.playerId) {
        dispatch({
          type: 'SHOW_TURN_REPORT',
          payload: {
            report: data,
            confirmations: { [ws.playerId]: false },
          },
        });
      }
    });

    const unlistenTurnReportAck = ws.on(MESSAGE_TYPES.TURN_REPORT_ACK, (data) => {
      if (data && data.player_id) {
        dispatch({
          type: 'UPDATE_REPORT_CONFIRMATIONS',
          payload: { [data.player_id]: data.confirmed },
        });
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

    const unlistenAllianceInvite = ws.on(MESSAGE_TYPES.ALLIANCE_INVITE, (data) => {
      if (data && data.target_id === ws.playerId) {
        dispatch({ type: 'ADD_ALLIANCE_INVITE', payload: data });
      }
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
      unlistenTurnReport();
      unlistenTurnReportAck();
      unlistenEvent();
      unlistenSystem();
      unlistenError();
      unlistenAllianceInvite();
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

  const confirmTurnReport = useCallback(() => {
    if (!state.reportConfirmations || !state.reportConfirmations[state.playerId]) {
      ws.sendPlayerAction(PLAYER_ACTIONS.CONFIRM_TURN_REPORT, {});
      dispatch({ type: 'MY_CONFIRM_REPORT' });
    }
    dispatch({ type: 'HIDE_TURN_REPORT' });
  }, [state.reportConfirmations, state.playerId]);

  const createAlliance = useCallback((name, color) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.CREATE_ALLIANCE, { name, color });
  }, []);

  const sendAllianceInvite = useCallback((allianceId, targetPlayerId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.SEND_ALLIANCE_INVITE, {
      alliance_id: allianceId,
      target_player_id: targetPlayerId,
    });
  }, []);

  const acceptAllianceInvite = useCallback((allianceId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.ACCEPT_ALLIANCE_INVITE, { alliance_id: allianceId });
    dispatch({ type: 'REMOVE_ALLIANCE_INVITE', payload: { alliance_id: allianceId } });
  }, []);

  const rejectAllianceInvite = useCallback((allianceId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.REJECT_ALLIANCE_INVITE, { alliance_id: allianceId });
    dispatch({ type: 'REMOVE_ALLIANCE_INVITE', payload: { alliance_id: allianceId } });
  }, []);

  const leaveAlliance = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.LEAVE_ALLIANCE, {});
  }, []);

  const kickAllianceMember = useCallback((targetPlayerId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.KICK_ALLIANCE_MEMBER, { target_player_id: targetPlayerId });
  }, []);

  const disbandAlliance = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.DISBAND_ALLIANCE, {});
  }, []);

  const createTradeAgreement = useCallback((targetPlayerId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.CREATE_TRADE_AGREEMENT, { target_player_id: targetPlayerId });
  }, []);

  const renewTradeAgreement = useCallback((agreementId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.RENEW_TRADE_AGREEMENT, { agreement_id: agreementId });
  }, []);

  const initiateJointMilitary = useCallback((targetPlayerId, targetBodyId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.INITIATE_JOINT_MILITARY, {
      target_player_id: targetPlayerId,
      target_body_id: targetBodyId,
    });
  }, []);

  const joinMilitaryAction = useCallback((actionId, fleetId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.JOIN_MILITARY_ACTION, {
      action_id: actionId,
      fleet_id: fleetId,
    });
  }, []);

  const declineMilitaryAction = useCallback((actionId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.DECLINE_MILITARY_ACTION, { action_id: actionId });
  }, []);

  const transferLeadership = useCallback((newLeaderId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.TRANSFER_LEADERSHIP, { new_leader_id: newLeaderId });
  }, []);

  const declareWar = useCallback((targetAllianceId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.DECLARE_WAR, { target_alliance_id: targetAllianceId });
  }, []);

  const surrenderWar = useCallback((warId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.SURRENDER_WAR, { war_id: warId });
  }, []);

  const createSanctionProposal = useCallback((targetPlayerId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.CREATE_SANCTION_PROPOSAL, { target_player_id: targetPlayerId });
  }, []);

  const secondSanctionProposal = useCallback((proposalId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.SECOND_SANCTION_PROPOSAL, { proposal_id: proposalId });
  }, []);

  const recruitSpy = useCallback(() => {
    ws.sendPlayerAction(PLAYER_ACTIONS.RECRUIT_SPY, {});
  }, []);

  const assignSpyMission = useCallback((spyId, targetPlayerId, missionType, thirdPartyId) => {
    const params = {
      spy_id: spyId,
      target_player_id: targetPlayerId,
      mission_type: missionType,
    };
    if (thirdPartyId) {
      params.third_party_id = thirdPartyId;
    }
    ws.sendPlayerAction(PLAYER_ACTIONS.ASSIGN_SPY_MISSION, params);
  }, []);

  const setCounterSpyLevel = useCallback((level) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.SET_COUNTER_SPY_LEVEL, { level });
  }, []);

  const sellIntelOnMarket = useCallback((intelId, price) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.SELL_INTEL, { intel_id: intelId, price });
  }, []);

  const buyIntelFromMarket = useCallback((listingId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.BUY_INTEL, { listing_id: listingId });
  }, []);

  const cancelIntelListing = useCallback((listingId) => {
    ws.sendPlayerAction(PLAYER_ACTIONS.CANCEL_INTEL_LISTING, { listing_id: listingId });
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
    confirmTurnReport,
    createAlliance,
    sendAllianceInvite,
    acceptAllianceInvite,
    rejectAllianceInvite,
    leaveAlliance,
    kickAllianceMember,
    disbandAlliance,
    createTradeAgreement,
    renewTradeAgreement,
    initiateJointMilitary,
    joinMilitaryAction,
    declineMilitaryAction,
    transferLeadership,
    declareWar,
    surrenderWar,
    createSanctionProposal,
    secondSanctionProposal,
    recruitSpy,
    assignSpyMission,
    setCounterSpyLevel,
    sellIntelOnMarket,
    buyIntelFromMarket,
    cancelIntelListing,
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
