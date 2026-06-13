import React, { useState, useEffect, useRef } from 'react';
import { useGameState } from '../hooks/useGameState';
import { ALLIANCE_COLORS, DIPLOMACY_STATUS } from '../types/game';

function AllianceInvitePopup({ invite, onAccept, onReject }) {
  const [timeLeft, setTimeLeft] = useState(30);
  const timerRef = useRef(null);

  useEffect(() => {
    timerRef.current = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timerRef.current);
          onReject(invite.alliance_id);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(timerRef.current);
  }, [invite.alliance_id, onReject]);

  const radius = 40;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - timeLeft / 30);

  return (
    <div className="alliance-invite-popup">
      <div className="invite-timer-ring">
        <svg width="100" height="100" viewBox="0 0 100 100">
          <circle
            cx="50" cy="50" r={radius}
            fill="none"
            stroke="rgba(255,255,255,0.1)"
            strokeWidth="4"
          />
          <circle
            cx="50" cy="50" r={radius}
            fill="none"
            stroke={timeLeft > 10 ? '#4488FF' : '#FF4444'}
            strokeWidth="4"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            transform="rotate(-90 50 50)"
            style={{ transition: 'stroke-dashoffset 1s linear' }}
          />
        </svg>
        <div className="invite-timer-text">{timeLeft}</div>
      </div>
      <div className="invite-info">
        <div className="invite-title">联盟邀请</div>
        <div className="invite-detail">
          <strong>{invite.inviter_name || invite.inviter_id}</strong> 邀请你加入联盟
        </div>
        <div className="invite-detail">
          「<span style={{ color: '#4488FF' }}>{invite.alliance_name}</span>」
        </div>
      </div>
      <div className="invite-actions">
        <button className="btn btn-success btn-small" onClick={() => { clearInterval(timerRef.current); onAccept(invite.alliance_id); }}>
          接受
        </button>
        <button className="btn btn-danger btn-small" onClick={() => { clearInterval(timerRef.current); onReject(invite.alliance_id); }}>
          拒绝
        </button>
      </div>
    </div>
  );
}

function DiplomacyPanel() {
  const {
    state,
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
  } = useGameState();

  const {
    otherPlayers,
    myPlayer,
    alliances,
    tradeAgreements,
    jointMilitaryActions,
    diplomacyRelations,
    playerCooldowns,
    allianceInvites,
  } = state;

  const [subTab, setSubTab] = useState('alliance');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newAllianceName, setNewAllianceName] = useState('');
  const [newAllianceColor, setNewAllianceColor] = useState('');
  const [inviteTarget, setInviteTarget] = useState('');
  const [tradeTarget, setTradeTarget] = useState('');
  const [militaryTarget, setMilitaryTarget] = useState('');
  const [militaryTargetBody, setMilitaryTargetBody] = useState('');
  const [selectedActionFleet, setSelectedActionFleet] = useState({});

  const myAlliance = (alliances || []).find(
    (a) => a.status === 'active' && (a.member_ids || []).includes(myPlayer?.id)
  );
  const isLeader = myAlliance && myAlliance.leader_id === myPlayer?.id;

  const myCooldown = (playerCooldowns || []).find((c) => c.player_id === myPlayer?.id);
  const isInCooldown = myCooldown && myCooldown.cooldown_turns > 0;

  const usedColors = (alliances || [])
    .filter((a) => a.status === 'active')
    .map((a) => a.color);
  const availableColors = ALLIANCE_COLORS.filter((c) => !usedColors.includes(c.id));

  const getRelationValue = (otherId) => {
    const rel = (diplomacyRelations || []).find(
      (r) =>
        (r.player1_id === myPlayer?.id && r.player2_id === otherId) ||
        (r.player2_id === myPlayer?.id && r.player1_id === otherId)
    );
    return rel ? rel.value : 50;
  };

  const getDiplomacyStatus = (value) => {
    if (value < 20) return DIPLOMACY_STATUS.HOSTILE;
    if (value > 70) return DIPLOMACY_STATUS.FRIENDLY;
    return DIPLOMACY_STATUS.NEUTRAL;
  };

  const getRelationColor = (value) => {
    if (value <= 20) return '#FF4444';
    if (value <= 40) return '#FF8844';
    if (value <= 60) return '#FFCC44';
    if (value <= 80) return '#88FF44';
    return '#44FF44';
  };

  const handleCreateAlliance = () => {
    if (!newAllianceName.trim() || !newAllianceColor) return;
    createAlliance(newAllianceName.trim(), newAllianceColor);
    setNewAllianceName('');
    setNewAllianceColor('');
    setShowCreateForm(false);
  };

  const handleSendInvite = () => {
    if (!inviteTarget || !myAlliance) return;
    sendAllianceInvite(myAlliance.id, inviteTarget);
    setInviteTarget('');
  };

  const handleCreateTrade = () => {
    if (!tradeTarget) return;
    createTradeAgreement(tradeTarget);
    setTradeTarget('');
  };

  const handleInitiateMilitary = () => {
    if (!militaryTarget || !militaryTargetBody) return;
    initiateJointMilitary(militaryTarget, militaryTargetBody);
    setMilitaryTarget('');
    setMilitaryTargetBody('');
  };

  const handleJoinAction = (actionId) => {
    const fleetId = selectedActionFleet[actionId];
    if (!fleetId) return;
    joinMilitaryAction(actionId, fleetId);
    setSelectedActionFleet({ ...selectedActionFleet, [actionId]: '' });
  };

  const getPlayerName = (id) => {
    if (id === myPlayer?.id) return myPlayer.company_name || myPlayer.name || '我';
    const other = (otherPlayers || []).find((p) => p.id === id);
    return other ? (other.company_name || other.name || id) : id;
  };

  const isPlayerOnline = (id) => {
    return (state.players || []).some((p) => p.id === id && p.connected !== false);
  };

  const subTabs = [
    { key: 'alliance', label: '联盟' },
    { key: 'relations', label: '外交关系' },
    { key: 'trade', label: '贸易协定' },
    { key: 'military', label: '军事行动' },
  ];

  return (
    <div className="panel diplomacy-panel">
      {allianceInvites && allianceInvites.length > 0 && (
        <div className="alliance-invites-overlay">
          {allianceInvites.map((inv) => (
            <AllianceInvitePopup
              key={inv.alliance_id}
              invite={inv}
              onAccept={acceptAllianceInvite}
              onReject={rejectAllianceInvite}
            />
          ))}
        </div>
      )}

      <h3 className="panel-title">外交中心</h3>

      <div className="diplomacy-subtabs">
        {subTabs.map((tab) => (
          <button
            key={tab.key}
            className={`diplomacy-subtab ${subTab === tab.key ? 'active' : ''}`}
            onClick={() => setSubTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {subTab === 'alliance' && (
        <div className="diplomacy-section">
          {myAlliance ? (
            <div className="alliance-info">
              <div className="alliance-header">
                <div
                  className="alliance-color-badge"
                  style={{ backgroundColor: myAlliance.color }}
                />
                <span className="alliance-name">{myAlliance.name}</span>
                {isLeader && <span className="alliance-leader-badge">盟主</span>}
              </div>

              <div className="alliance-members">
                <h5>成员 ({(myAlliance.member_ids || []).length}/3)</h5>
                {(myAlliance.member_ids || []).map((mid) => (
                  <div key={mid} className="alliance-member-row">
                    <span
                      className={`member-online ${isPlayerOnline(mid) ? 'online' : 'offline'}`}
                    />
                    <span className="member-name">{getPlayerName(mid)}</span>
                    {mid === myAlliance.leader_id && (
                      <span className="member-role">盟主</span>
                    )}
                    {isLeader && mid !== myAlliance.leader_id && mid !== myPlayer?.id && (
                      <button
                        className="btn btn-danger btn-tiny"
                        onClick={() => kickAllianceMember(mid)}
                      >
                        踢出
                      </button>
                    )}
                    {mid === myPlayer?.id && !isLeader && (
                      <button
                        className="btn btn-warning btn-tiny"
                        onClick={() => {
                          if (window.confirm('确定退出联盟？退出后有3回合冷却期。')) {
                            leaveAlliance();
                          }
                        }}
                      >
                        退出
                      </button>
                    )}
                  </div>
                ))}
              </div>

              {isLeader && (myAlliance.member_ids || []).length < 3 && (
                <div className="invite-form">
                  <h5>邀请成员</h5>
                  <div className="invite-form-row">
                    <select
                      value={inviteTarget}
                      onChange={(e) => setInviteTarget(e.target.value)}
                      className="diplomacy-select"
                    >
                      <option value="">选择玩家</option>
                      {(otherPlayers || [])
                        .filter((p) => !(myAlliance.member_ids || []).includes(p.id))
                        .map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.company_name || p.name}
                          </option>
                        ))}
                    </select>
                    <button
                      className="btn btn-primary btn-small"
                      onClick={handleSendInvite}
                      disabled={!inviteTarget}
                    >
                      邀请
                    </button>
                  </div>
                </div>
              )}

              {isLeader && (
                <div className="alliance-leader-actions">
                  <button
                    className="btn btn-danger btn-small"
                    onClick={() => {
                      if (window.confirm('确定解散联盟？此操作不可撤销。')) {
                        disbandAlliance();
                      }
                    }}
                  >
                    解散联盟
                  </button>
                </div>
              )}

              {isLeader && (myAlliance.member_ids || []).length > 1 && (
                <div className="transfer-leadership-section">
                  <h5>转让盟主</h5>
                  <div className="invite-form-row">
                    <select
                      value=""
                      onChange={(e) => {
                        if (e.target.value && window.confirm(`确定将盟主转让给 ${getPlayerName(e.target.value)}？`)) {
                          transferLeadership(e.target.value);
                        }
                      }}
                      className="diplomacy-select"
                    >
                      <option value="">选择成员</option>
                      {(myAlliance.member_ids || [])
                        .filter((mid) => mid !== myAlliance.leader_id)
                        .map((mid) => (
                          <option key={mid} value={mid}>
                            {getPlayerName(mid)}
                          </option>
                        ))}
                    </select>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="no-alliance">
              {isInCooldown ? (
                <div className="cooldown-notice">
                  <span className="cooldown-icon">⏳</span>
                  联盟冷却中，剩余 {myCooldown.cooldown_turns} 回合
                </div>
              ) : (
                <>
                  {!showCreateForm ? (
                    <button
                      className="btn btn-primary"
                      onClick={() => setShowCreateForm(true)}
                    >
                      创建联盟
                    </button>
                  ) : (
                    <div className="create-alliance-form">
                      <h5>创建新联盟</h5>
                      <input
                        type="text"
                        placeholder="联盟名称（创建后不可修改）"
                        value={newAllianceName}
                        onChange={(e) => setNewAllianceName(e.target.value)}
                        className="diplomacy-input"
                        maxLength={20}
                      />
                      <div className="color-picker">
                        <label>选择标识色：</label>
                        <div className="color-options">
                          {availableColors.map((c) => (
                            <button
                              key={c.id}
                              className={`color-option ${newAllianceColor === c.id ? 'selected' : ''}`}
                              style={{ backgroundColor: c.id }}
                              onClick={() => setNewAllianceColor(c.id)}
                              title={c.name}
                            />
                          ))}
                        </div>
                      </div>
                      <div className="form-actions">
                        <button
                          className="btn btn-success btn-small"
                          onClick={handleCreateAlliance}
                          disabled={!newAllianceName.trim() || !newAllianceColor}
                        >
                          确认创建
                        </button>
                        <button
                          className="btn btn-secondary btn-small"
                          onClick={() => {
                            setShowCreateForm(false);
                            setNewAllianceName('');
                            setNewAllianceColor('');
                          }}
                        >
                          取消
                        </button>
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      )}

      {subTab === 'relations' && (
        <div className="diplomacy-section">
          <h4>外交关系</h4>
          {(otherPlayers || []).length === 0 ? (
            <div className="empty-state">暂无其他玩家</div>
          ) : (
            <div className="relations-list">
              {(otherPlayers || []).map((player) => {
                const value = getRelationValue(player.id);
                const status = getDiplomacyStatus(value);
                const color = getRelationColor(value);
                const isAlly = myAlliance && (myAlliance.member_ids || []).includes(player.id);
                return (
                  <div key={player.id} className="relation-row">
                    <div className="relation-player">
                      <span className={`relation-icon ${status}`}>
                        {status === DIPLOMACY_STATUS.HOSTILE && '🔴'}
                        {status === DIPLOMACY_STATUS.FRIENDLY && '🟢'}
                        {status === DIPLOMACY_STATUS.NEUTRAL && '⚪'}
                      </span>
                      <span className="relation-name">
                        {player.company_name || player.name}
                      </span>
                      {isAlly && <span className="ally-badge">盟友</span>}
                    </div>
                    <div className="relation-bar-container">
                      <div className="relation-bar">
                        <div
                          className="relation-bar-fill"
                          style={{
                            width: `${value}%`,
                            background: `linear-gradient(90deg, #FF4444, ${color})`,
                          }}
                        />
                      </div>
                      <span className="relation-value" style={{ color }}>
                        {value}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {subTab === 'trade' && (
        <div className="diplomacy-section">
          <h4>贸易协定</h4>

          {myAlliance && (
            <div className="create-trade-form">
              <div className="invite-form-row">
                <select
                  value={tradeTarget}
                  onChange={(e) => setTradeTarget(e.target.value)}
                  className="diplomacy-select"
                >
                  <option value="">选择盟友</option>
                  {(myAlliance.member_ids || [])
                    .filter((mid) => mid !== myPlayer?.id)
                    .map((mid) => (
                      <option key={mid} value={mid}>
                        {getPlayerName(mid)}
                      </option>
                    ))}
                </select>
                <button
                  className="btn btn-success btn-small"
                  onClick={handleCreateTrade}
                  disabled={!tradeTarget}
                >
                  签订协定
                </button>
              </div>
            </div>
          )}

          {!myAlliance && (
            <div className="empty-state" style={{ color: '#FF8844' }}>
              需要加入联盟才能签订贸易协定
            </div>
          )}

          {(tradeAgreements || []).length === 0 ? (
            <div className="empty-state">暂无贸易协定</div>
          ) : (
            <div className="trade-agreements-list">
              {(tradeAgreements || []).map((ta) => {
                const otherId =
                  ta.player_id1 === myPlayer?.id ? ta.player_id2 : ta.player_id1;
                const isActive = ta.status === 'active';
                const turnsLeft = isActive ? ta.expiry_turn - state.turn : 0;
                return (
                  <div key={ta.id} className={`trade-agreement-item ${isActive ? 'active' : 'expired'}`}>
                    <div className="ta-header">
                      <span className="ta-player">{getPlayerName(otherId)}</span>
                      <span className={`ta-status ${isActive ? 'active' : 'expired'}`}>
                        {isActive ? '生效中' : '已过期'}
                      </span>
                    </div>
                    <div className="ta-detail">
                      {isActive ? (
                        <>剩余 {turnsLeft} 回合 | 手续费减半</>
                      ) : (
                        <>第 {ta.expiry_turn} 回合过期</>
                      )}
                    </div>
                    {isActive && turnsLeft <= 3 && (
                      <button
                        className="btn btn-warning btn-tiny"
                        onClick={() => renewTradeAgreement(ta.id)}
                      >
                        续约
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {subTab === 'military' && (
        <div className="diplomacy-section">
          <h4>联合军事行动</h4>

          {isLeader && (
            <div className="initiate-military-form">
              <h5>发起联合军事行动</h5>
              <div className="military-form-row">
                <select
                  value={militaryTarget}
                  onChange={(e) => setMilitaryTarget(e.target.value)}
                  className="diplomacy-select"
                >
                  <option value="">选择目标玩家</option>
                  {(otherPlayers || []).map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.company_name || p.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="military-form-row">
                <select
                  value={militaryTargetBody}
                  onChange={(e) => setMilitaryTargetBody(e.target.value)}
                  className="diplomacy-select"
                >
                  <option value="">选择目标星球</option>
                  {(state.gameMap?.galaxies || []).flatMap((g) =>
                    (g.celestial_bodies || []).map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name || c.id}
                      </option>
                    ))
                  )}
                </select>
              </div>
              <button
                className="btn btn-danger btn-small"
                onClick={handleInitiateMilitary}
                disabled={!militaryTarget || !militaryTargetBody}
              >
                发起行动
              </button>
            </div>
          )}

          {(jointMilitaryActions || []).length === 0 ? (
            <div className="empty-state">暂无军事行动</div>
          ) : (
            <div className="military-actions-list">
              {(jointMilitaryActions || []).map((action) => {
                const isRecruiting = action.status === 'recruiting';
                const isInProgress = action.status === 'in_progress';
                const myParticipation = (action.participants || []).find(
                  (p) => p.player_id === myPlayer?.id
                );
                const myStationaryFleets = (myPlayer?.fleets || []).filter(
                  (f) => !f.is_moving
                );

                return (
                  <div key={action.id} className={`military-action-item ${action.status}`}>
                    <div className="ma-header">
                      <span className="ma-target">
                        目标: {getPlayerName(action.target_player_id)} / {action.target_body_id}
                      </span>
                      <span className={`ma-status ${action.status}`}>
                        {action.status === 'recruiting' && '招募中'}
                        {action.status === 'in_progress' && '进行中'}
                        {action.status === 'completed' && '已完成'}
                        {action.status === 'cancelled' && '已取消'}
                      </span>
                    </div>

                    {isRecruiting && (
                      <div className="ma-deadline">
                        截止: 第 {action.deadline_turn} 回合
                      </div>
                    )}

                    {isInProgress && (
                      <div className="ma-progress">
                        到达: 第 {action.arrival_turn} 回合
                        {action.total_attack > 0 && (
                          <span className="ma-attack">
                            总攻击力: {Math.round(action.total_attack)} (含1.2x协同加成)
                          </span>
                        )}
                      </div>
                    )}

                    <div className="ma-participants">
                      <span className="ma-label">参战方:</span>
                      {(action.participants || []).map((p) => (
                        <span
                          key={p.player_id}
                          className={`ma-participant ${p.joined ? 'joined' : 'pending'}`}
                        >
                          {getPlayerName(p.player_id)}
                          {p.joined ? '✓' : '⏳'}
                        </span>
                      ))}
                    </div>

                    {isRecruiting && myAlliance &&
                      (myAlliance.member_ids || []).includes(myPlayer?.id) &&
                      (!myParticipation || !myParticipation.joined) && (
                        <div className="ma-join-form">
                          <select
                            value={selectedActionFleet[action.id] || ''}
                            onChange={(e) =>
                              setSelectedActionFleet({
                                ...selectedActionFleet,
                                [action.id]: e.target.value,
                              })
                            }
                            className="diplomacy-select"
                          >
                            <option value="">选择舰队</option>
                            {myStationaryFleets.map((f) => (
                              <option key={f.id} value={f.id}>
                                舰队 {f.id} ({(f.ships || []).length} 艘)
                              </option>
                            ))}
                          </select>
                          <button
                            className="btn btn-success btn-tiny"
                            onClick={() => handleJoinAction(action.id)}
                            disabled={!selectedActionFleet[action.id]}
                          >
                            参战
                          </button>
                          <button
                            className="btn btn-secondary btn-tiny"
                            onClick={() => declineMilitaryAction(action.id)}
                          >
                            放弃
                          </button>
                        </div>
                      )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default DiplomacyPanel;
