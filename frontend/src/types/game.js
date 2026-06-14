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
  DECLARE_WAR: 'declare_war',
  SURRENDER_WAR: 'surrender_war',
  CREATE_SANCTION_PROPOSAL: 'create_sanction_proposal',
  SECOND_SANCTION_PROPOSAL: 'second_sanction_proposal',
  RECRUIT_SPY: 'recruit_spy',
  ASSIGN_SPY_MISSION: 'assign_spy_mission',
  SET_COUNTER_SPY_LEVEL: 'set_counter_spy_level',
  SELL_INTEL: 'sell_intel',
  BUY_INTEL: 'buy_intel',
  CANCEL_INTEL_LISTING: 'cancel_intel_listing',
  CHOOSE_SPY_SPEC: 'choose_spy_spec',
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

export const WAR_STATUS = {
  ACTIVE: 'active',
  SURRENDERED: 'surrendered',
};

export const SANCTION_PROPOSAL_STATUS = {
  PENDING: 'pending',
  ACTIVE: 'active',
  EXPIRED: 'expired',
  REJECTED: 'rejected',
};

export const SPY_LEVELS = {
  JUNIOR: 'junior',
  MIDDLE: 'middle',
  SENIOR: 'senior',
};

export const SPY_LEVEL_NAMES = {
  junior: '初级',
  middle: '中级',
  senior: '高级',
};

export const SPY_STATUS = {
  IDLE: 'idle',
  ON_MISSION: 'on_mission',
  CAPTURED: 'captured',
};

export const SPY_STATUS_NAMES = {
  idle: '待命',
  on_mission: '执行任务中',
  captured: '已被捕',
};

export const SPY_MISSION_TYPES = {
  STEAL_TECH: 'steal_tech',
  ECON_SABOTAGE: 'econ_sabotage',
  INTEL_GATHER: 'intel_gather',
  TURCOAT: 'turncoat',
  DIPLO_PRESSURE: 'diplo_pressure',
};

export const SPY_MISSION_NAMES = {
  steal_tech: '窃取科技',
  econ_sabotage: '经济破坏',
  intel_gather: '情报刺探',
  turncoat: '策反',
  diplo_pressure: '外交施压',
};

export const SPY_MISSION_DESCRIPTIONS = {
  steal_tech: '偷取目标一项科技(初级30%/中级50%/高级70%)',
  econ_sabotage: '销毁目标资金(初级5%/中级8%/高级12%), 成功率60%',
  intel_gather: '获取目标详细信息(初级50%/中级70%/高级90%)',
  turncoat: '策反目标间谍为双面间谍(仅高级, 成功率40%)',
  diplo_pressure: '降低目标与第三方关系值-20(中级45%/高级65%)',
};

export const COUNTER_SPY_LEVELS = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high',
};

export const COUNTER_SPY_NAMES = {
  low: '低(20%检测/100¢)',
  medium: '中(40%检测+识别/250¢)',
  high: '高(60%检测+识别+反制/500¢)',
};

export const COUNTER_SPY_COSTS = {
  low: 100,
  medium: 250,
  high: 500,
};

export const SPY_SPECIALIZATIONS = {
  INFILTRATION: 'infiltration',
  DESTRUCTION: 'destruction',
  SHADOW: 'shadow',
};

export const SPY_SPEC_NAMES = {
  infiltration: '渗透',
  destruction: '破坏',
  shadow: '影子',
};

export const SPY_SPEC_DESCRIPTIONS = {
  infiltration: '窃取科技/情报刺探+20%成功率，经济破坏/外交施压-15%，暴露增长减半',
  destruction: '经济破坏伤害翻倍，外交施压-35关系值，每次任务暴露+10',
  shadow: '闲置暴露-15/回合，被捕阈值提升至95，所有任务成功率-10%',
};

export const SPY_SPEC_COLORS = {
  infiltration: '#4488FF',
  destruction: '#FF4444',
  shadow: '#AA44FF',
};

export const INTEL_DURATION = 5;
