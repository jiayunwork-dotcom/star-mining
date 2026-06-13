export const MESSAGE_TYPES = {
  GAME_STATE: 'game_state',
  PLAYER_ACTION: 'player_action',
  CHAT: 'chat',
  EVENT: 'event',
  HEARTBEAT: 'heartbeat',
  SYSTEM: 'system',
  ERROR: 'error',
  TURN_REPORT: 'turn_report',
  TURN_REPORT_ACK: 'turn_report_ack',
  ALLIANCE_INVITE: 'alliance_invite',
};

export const PLAYER_ACTIONS = {
  BUILD: 'build',
  TRADE: 'trade',
  MOVE: 'move',
  UPGRADE: 'upgrade',
  RESEARCH: 'research',
  READY: 'ready',
  UNREADY: 'unready',
  END_TURN: 'end_turn',
  CREATE_ROOM: 'create_room',
  JOIN_ROOM: 'join_room',
  LEAVE_ROOM: 'leave_room',
  LIST_ROOMS: 'list_rooms',
  START_GAME: 'start_game',
  CONFIRM_TURN_REPORT: 'confirm_turn_report',
  CREATE_ALLIANCE: 'create_alliance',
  SEND_ALLIANCE_INVITE: 'send_alliance_invite',
  ACCEPT_ALLIANCE_INVITE: 'accept_alliance_invite',
  REJECT_ALLIANCE_INVITE: 'reject_alliance_invite',
  LEAVE_ALLIANCE: 'leave_alliance',
  KICK_ALLIANCE_MEMBER: 'kick_alliance_member',
  DISBAND_ALLIANCE: 'disband_alliance',
  CREATE_TRADE_AGREEMENT: 'create_trade_agreement',
  RENEW_TRADE_AGREEMENT: 'renew_trade_agreement',
  INITIATE_JOINT_MILITARY: 'initiate_joint_military',
  JOIN_MILITARY_ACTION: 'join_military_action',
  DECLINE_MILITARY_ACTION: 'decline_military_action',
  TRANSFER_LEADERSHIP: 'transfer_leadership',
};

export const CELESTIAL_TYPES = {
  STAR: 'star',
  PLANET: 'planet',
  ASTEROID_BELT: 'asteroid_belt',
  GAS_GIANT: 'gas_giant',
  TERRESTRIAL: 'terrestrial',
};

export const SHIP_TYPES = {
  CARGO: 'cargo',
  FRIGATE: 'frigate',
  MINING: 'mining',
};

export const RESOURCE_TYPES = {
  IRON_ORE: 'iron_ore',
  TITANIUM: 'titanium',
  HELIUM_3: 'helium_3',
  RARE_EARTH: 'rare_earth',
  ICE_CRYSTAL: 'ice_crystal',
  CREDITS: 'credits',
  FUEL: 'fuel',
};

export const RESOURCE_NAMES = {
  iron_ore: '铁矿',
  titanium: '钛矿',
  helium_3: '氦-3',
  rare_earth: '稀土',
  ice_crystal: '冰晶',
  credits: '信用币',
  fuel: '燃料',
};

export const TECH_TYPES = {
  MINING_EFFICIENCY: 'mining_efficiency',
  REFINING_TECH: 'refining_tech',
  ENGINE_IMPROVEMENT: 'engine_improvement',
  WEAPON_UPGRADE: 'weapon_upgrade',
};

export const TECH_NAMES = {
  mining_efficiency: '采矿效率',
  refining_tech: '精炼技术',
  engine_improvement: '引擎改进',
  weapon_upgrade: '武器升级',
};

export const TECH_DESCRIPTIONS = {
  mining_efficiency: '提高采矿效率，增加每回合资源产出',
  refining_tech: '提升精炼技术，提高资源加工效率',
  engine_improvement: '加快飞船移动速度，缩短航程',
  weapon_upgrade: '增强武器系统，提高战斗能力',
};

export const SHIP_TYPE_NAMES = {
  cargo: '货运舰',
  frigate: '护卫舰',
  mining: '采矿船',
};

export const SHIP_COSTS = {
  cargo: { credits: 800, iron_ore: 150, fuel: 50 },
  frigate: { credits: 2000, iron_ore: 200, titanium: 100, fuel: 80 },
  mining: { credits: 1500, iron_ore: 100, titanium: 50, fuel: 60 },
};

export const GAME_PHASES = {
  PLANNING: 'planning',
  EXECUTION: 'execution',
  RESOLUTION: 'resolution',
};

export const GAME_STATES = {
  LOBBY: 'lobby',
  PLAYING: 'playing',
  ENDED: 'ended',
};

export const ORDER_TYPES = {
  BUY: 'buy',
  SELL: 'sell',
};

export const CELESTIAL_TYPE_NAMES = {
  star: '恒星',
  planet: '行星',
  asteroid_belt: '小行星带',
  gas_giant: '气态巨星',
  terrestrial: '类地行星',
};

export const ALLIANCE_COLORS = [
  { id: '#FF4444', name: '红色' },
  { id: '#4488FF', name: '蓝色' },
  { id: '#44FF44', name: '绿色' },
  { id: '#FFFF44', name: '黄色' },
  { id: '#FF44FF', name: '紫色' },
  { id: '#44FFFF', name: '青色' },
];

export const DIPLOMACY_STATUS = {
  HOSTILE: 'hostile',
  NEUTRAL: 'neutral',
  FRIENDLY: 'friendly',
};

export const MILITARY_ACTION_STATUS = {
  RECRUITING: 'recruiting',
  IN_PROGRESS: 'in_progress',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
};
