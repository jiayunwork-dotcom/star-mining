import React, { useState } from 'react';
import { useGameState } from '../hooks/useGameState';
import {
  SPY_LEVEL_NAMES,
  SPY_STATUS_NAMES,
  SPY_MISSION_NAMES,
  SPY_MISSION_DESCRIPTIONS,
  COUNTER_SPY_NAMES,
  COUNTER_SPY_LEVELS,
  SPY_SPEC_NAMES,
  SPY_SPEC_DESCRIPTIONS,
  SPY_SPEC_COLORS,
  SPY_SPECIALIZATIONS,
  INTEL_DURATION,
} from '../types/game';

function IntelPanel() {
  const {
    state,
    recruitSpy,
    assignSpyMission,
    setCounterSpyLevel,
    sellIntelOnMarket,
    buyIntelFromMarket,
    cancelIntelListing,
    chooseSpySpec,
  } = useGameState();

  const {
    otherPlayers,
    myPlayer,
    spies,
    spyMissions,
    intelligences,
    intelMarketListings,
    counterSpySettings,
    alliances,
    activeSanctions,
  } = state;

  const [subTab, setSubTab] = useState('spies');
  const [selectedSpy, setSelectedSpy] = useState('');
  const [missionTarget, setMissionTarget] = useState('');
  const [missionType, setMissionType] = useState('');
  const [thirdParty, setThirdParty] = useState('');
  const [sellIntelId, setSellIntelId] = useState('');
  const [sellPrice, setSellPrice] = useState('');
  const [specSelectionSpyId, setSpecSelectionSpyId] = useState('');

  const mySpies = (spies || []).filter((s) => s.player_id === myPlayer?.id);
  const myIntel = (intelligences || []).filter((i) => i.owner_player_id === myPlayer?.id && !i.expired);
  const myCounterSpy = (counterSpySettings || []).find((s) => s.player_id === myPlayer?.id);
  const currentCounterLevel = myCounterSpy ? myCounterSpy.level : 'low';
  const activeListings = (intelMarketListings || []).filter((l) => l.active);
  const myListings = activeListings.filter((l) => l.seller_id === myPlayer?.id);
  const otherListings = activeListings.filter((l) => l.seller_id !== myPlayer?.id);

  const isSanctioned = (activeSanctions || []).some((s) => s.target_id === myPlayer?.id);

  const getSpecModifierLabel = (spec, missionType) => {
    if (!spec) return '';
    const bonusMissions = {
      infiltration: ['steal_tech', 'intel_gather'],
      destruction: ['econ_sabotage', 'diplo_pressure'],
      shadow: [],
    };
    const penaltyMissions = {
      infiltration: ['econ_sabotage', 'diplo_pressure'],
      destruction: [],
      shadow: ['steal_tech', 'econ_sabotage', 'intel_gather', 'diplo_pressure', 'turncoat'],
    };
    if (bonusMissions[spec]?.includes(missionType)) return ' ▲';
    if (penaltyMissions[spec]?.includes(missionType)) return ' ▼';
    return '';
  };

  const getPlayerName = (id) => {
    if (id === myPlayer?.id) return myPlayer?.company_name || myPlayer?.name || '我';
    const other = (otherPlayers || []).find((p) => p.id === id);
    return other ? (other.company_name || other.name || id) : id;
  };

  const handleRecruit = () => {
    if (mySpies.length >= 5) return;
    recruitSpy();
  };

  const handleAssignMission = () => {
    if (!selectedSpy || !missionTarget || !missionType) return;
    assignSpyMission(selectedSpy, missionTarget, missionType, thirdParty || undefined);
    setSelectedSpy('');
    setMissionTarget('');
    setMissionType('');
    setThirdParty('');
  };

  const handleSetCounterSpy = (level) => {
    setCounterSpyLevel(level);
  };

  const handleSellIntel = () => {
    if (!sellIntelId || !sellPrice || parseFloat(sellPrice) <= 0) return;
    sellIntelOnMarket(sellIntelId, parseFloat(sellPrice));
    setSellIntelId('');
    setSellPrice('');
  };

  const handleBuyIntel = (listingId) => {
    buyIntelFromMarket(listingId);
  };

  const handleCancelListing = (listingId) => {
    cancelIntelListing(listingId);
  };

  const getAvailableMissions = (spyLevel, spySpec) => {
    const missions = [
      { type: 'steal_tech', name: '窃取科技', minLevel: 'junior' },
      { type: 'econ_sabotage', name: '经济破坏', minLevel: 'junior' },
      { type: 'intel_gather', name: '情报刺探', minLevel: 'junior' },
      { type: 'turncoat', name: '策反', minLevel: 'senior' },
      { type: 'diplo_pressure', name: '外交施压', minLevel: 'middle' },
    ];
    const levelOrder = { junior: 0, middle: 1, senior: 2 };
    return missions.filter((m) => {
      if (levelOrder[spyLevel] < levelOrder[m.minLevel]) return false;
      if (m.type === 'turncoat' && spySpec === 'destruction') return false;
      return true;
    });
  };

  const idleSpies = mySpies.filter((s) => s.status === 'idle');
  const onMissionSpies = mySpies.filter((s) => s.status === 'on_mission');

  const subTabs = [
    { key: 'spies', label: '间谍管理' },
    { key: 'intel', label: '情报库' },
    { key: 'counter', label: '反间谍' },
    { key: 'market', label: '情报市场' },
  ];

  return (
    <div className="panel intel-panel">
      <h3 className="panel-title">情报中心</h3>

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

      {subTab === 'spies' && (
        <div className="diplomacy-section">
          <div className="spy-recruit-section" style={{ marginBottom: '12px', padding: '10px', border: '1px solid #3a3a6a', borderRadius: '6px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>间谍数量: <strong>{mySpies.length}/5</strong></span>
              <button
                className="btn btn-primary btn-small"
                onClick={handleRecruit}
                disabled={mySpies.length >= 5 || myPlayer?.credits < 500}
              >
                招募间谍 (500¢)
              </button>
            </div>
            <div style={{ fontSize: '12px', color: '#888', marginTop: '4px' }}>
              维护费: 初级50¢/中级100¢/高级200¢ 每回合
            </div>
          </div>

          {mySpies.length === 0 ? (
            <div className="empty-state">暂无间谍</div>
          ) : (
            <div className="spy-list">
              {mySpies.map((spy) => (
                <div key={spy.id} className="spy-item" style={{ padding: '8px', marginBottom: '6px', border: `1px solid ${spy.specialization ? SPY_SPEC_COLORS[spy.specialization] : '#3a3a6a'}`, borderRadius: '4px', background: spy.status === 'on_mission' ? 'rgba(68,136,255,0.1)' : 'rgba(255,255,255,0.03)', boxShadow: spy.specialization ? `0 0 6px ${SPY_SPEC_COLORS[spy.specialization]}40` : 'none' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                      <strong>{spy.specialization ? `${SPY_LEVEL_NAMES[spy.level]}·${SPY_SPEC_NAMES[spy.specialization]}` : SPY_LEVEL_NAMES[spy.level] || spy.level}</strong>
                      {spy.is_double_agent && (
                        <span style={{ color: '#FF4444', marginLeft: '6px', fontSize: '11px' }}>双面间谍</span>
                      )}
                      {!spy.specialization && spy.spec_eligible && (
                        <span style={{ color: '#FFCC44', marginLeft: '6px', fontSize: '11px', cursor: 'pointer' }} onClick={() => setSpecSelectionSpyId(spy.id)}>
                          可选专精 ▶
                        </span>
                      )}
                      {spy.specialization && (
                        <span style={{ color: SPY_SPEC_COLORS[spy.specialization], marginLeft: '6px', fontSize: '11px' }}>
                          {SPY_SPEC_NAMES[spy.specialization]}专精
                        </span>
                      )}
                    </div>
                    <span className={`status-badge ${spy.status === 'idle' ? 'status-success' : spy.status === 'on_mission' ? 'status-info' : 'status-danger'}`}>
                      {SPY_STATUS_NAMES[spy.status] || spy.status}
                    </span>
                  </div>
                  <div style={{ fontSize: '12px', color: '#AAA', marginTop: '4px' }}>
                    <span>暴露值: </span>
                    <span style={{ color: spy.exposure >= 80 ? '#FF4444' : spy.exposure >= 50 ? '#FFCC44' : '#88FF44' }}>
                      {spy.exposure}/100
                    </span>
                    <span style={{ marginLeft: '12px' }}>完成任务: {spy.completed_missions}次</span>
                  </div>
                  <div style={{ marginTop: '4px', height: '4px', background: '#333', borderRadius: '2px', overflow: 'hidden' }}>
                    <div style={{ width: `${spy.exposure}%`, height: '100%', background: spy.exposure >= 80 ? '#FF4444' : spy.exposure >= 50 ? '#FFCC44' : '#88FF44', borderRadius: '2px' }} />
                  </div>
                </div>
              ))}
            </div>
          )}

          {specSelectionSpyId && (
            <div style={{ marginTop: '12px', padding: '10px', border: '1px solid #FFCC44', borderRadius: '6px', background: 'rgba(255,204,68,0.05)' }}>
              <h5 style={{ marginBottom: '8px', color: '#FFCC44' }}>选择专精路线 (不可更改)</h5>
              {Object.values(SPY_SPECIALIZATIONS).map((spec) => {
                const isDestructionSanctioned = spec === 'destruction' && isSanctioned;
                return (
                  <div key={spec} style={{ padding: '8px', marginBottom: '6px', border: `1px solid ${SPY_SPEC_COLORS[spec]}60`, borderRadius: '4px', background: `${SPY_SPEC_COLORS[spec]}10`, opacity: isDestructionSanctioned ? 0.5 : 1 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <strong style={{ color: SPY_SPEC_COLORS[spec] }}>{SPY_SPEC_NAMES[spec]}专精</strong>
                      <button
                        className="btn btn-primary btn-tiny"
                        onClick={() => { chooseSpySpec(specSelectionSpyId, spec); setSpecSelectionSpyId(''); }}
                        disabled={isDestructionSanctioned}
                      >
                        选择
                      </button>
                    </div>
                    <div style={{ fontSize: '12px', color: '#AAA', marginTop: '4px' }}>
                      {SPY_SPEC_DESCRIPTIONS[spec]}
                    </div>
                    {isDestructionSanctioned && (
                      <div style={{ fontSize: '11px', color: '#FF4444', marginTop: '4px' }}>被制裁状态下不可选择</div>
                    )}
                  </div>
                );
              })}
              <button className="btn btn-secondary btn-tiny" onClick={() => setSpecSelectionSpyId('')} style={{ marginTop: '4px' }}>
                取消
              </button>
            </div>
          )}

          {idleSpies.length > 0 && (
            <div className="spy-mission-form" style={{ marginTop: '12px', padding: '10px', border: '1px solid #4488FF', borderRadius: '6px' }}>
              <h5 style={{ marginBottom: '8px' }}>派遣任务</h5>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <select value={selectedSpy} onChange={(e) => { setSelectedSpy(e.target.value); setMissionType(''); }} className="diplomacy-select">
                  <option value="">选择间谍</option>
                  {idleSpies.map((s) => (
                    <option key={s.id} value={s.id}>
                      {SPY_LEVEL_NAMES[s.level]}{s.specialization ? `·${SPY_SPEC_NAMES[s.specialization]}` : ''} - 暴露值{s.exposure} ({s.id.slice(0, 12)})
                    </option>
                  ))}
                </select>
                <select value={missionTarget} onChange={(e) => setMissionTarget(e.target.value)} className="diplomacy-select">
                  <option value="">选择目标玩家</option>
                  {(otherPlayers || []).map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.company_name || p.name}
                    </option>
                  ))}
                </select>
                {selectedSpy && (
                  <select value={missionType} onChange={(e) => setMissionType(e.target.value)} className="diplomacy-select">
                    <option value="">选择任务</option>
                    {getAvailableMissions(mySpies.find((s) => s.id === selectedSpy)?.level || 'junior', mySpies.find((s) => s.id === selectedSpy)?.specialization).map((m) => (
                      <option key={m.type} value={m.type}>
                        {m.name} - {SPY_MISSION_DESCRIPTIONS[m.type]}{getSpecModifierLabel(mySpies.find((s) => s.id === selectedSpy)?.specialization, m.type)}
                      </option>
                    ))}
                  </select>
                )}
                {missionType === 'diplo_pressure' && (
                  <select value={thirdParty} onChange={(e) => setThirdParty(e.target.value)} className="diplomacy-select">
                    <option value="">选择第三方玩家</option>
                    {(otherPlayers || []).filter((p) => p.id !== missionTarget).map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.company_name || p.name}
                      </option>
                    ))}
                  </select>
                )}
                <button
                  className="btn btn-danger btn-small"
                  onClick={handleAssignMission}
                  disabled={!selectedSpy || !missionTarget || !missionType || (missionType === 'diplo_pressure' && !thirdParty)}
                >
                  执行任务
                </button>
              </div>
            </div>
          )}

          {onMissionSpies.length > 0 && (
            <div style={{ marginTop: '12px' }}>
              <h5>任务中的间谍</h5>
              {spyMissions && spyMissions.filter((m) => !m.resolved && m.owner_player_id === myPlayer?.id).map((m) => (
                <div key={m.id} style={{ padding: '6px 8px', marginBottom: '4px', background: 'rgba(68,136,255,0.08)', borderRadius: '4px', fontSize: '12px' }}>
                  {SPY_MISSION_NAMES[m.mission_type] || m.mission_type} → {getPlayerName(m.target_player_id)}
                  {m.third_party_id && ` (第三方: ${getPlayerName(m.third_party_id)})`}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {subTab === 'intel' && (
        <div className="diplomacy-section">
          <h4>已获取情报</h4>
          {myIntel.length === 0 ? (
            <div className="empty-state">暂无情报</div>
          ) : (
            <div className="intel-list">
              {myIntel.map((intel) => {
                const turnsLeft = intel.expiry_turn - state.turn;
                const totalDuration = INTEL_DURATION;
                const ratio = totalDuration > 0 ? turnsLeft / totalDuration : 0;
                const barColor = ratio > 0.6 ? '#44FF44' : ratio > 0.3 ? '#FFCC44' : '#FF4444';
                const shouldFlash = ratio <= 0.3 && ratio > 0;
                return (
                  <div key={intel.id} className="intel-item" style={{ padding: '8px', marginBottom: '6px', border: '1px solid #3a3a6a', borderRadius: '4px', background: turnsLeft <= 1 ? 'rgba(255,68,68,0.1)' : 'rgba(255,255,255,0.03)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <strong>{intel.summary || intel.intel_type}</strong>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <div style={{ flex: 1, height: '6px', background: '#333', borderRadius: '3px', overflow: 'hidden' }}>
                          <div style={{
                            width: `${Math.max(0, ratio * 100)}%`,
                            height: '100%',
                            background: barColor,
                            borderRadius: '3px',
                            transition: 'width 0.3s, background 0.3s',
                            animation: shouldFlash ? 'intelFlash 1s ease-in-out infinite' : 'none',
                          }} />
                        </div>
                        <span style={{ fontSize: '11px', color: barColor, whiteSpace: 'nowrap' }}>
                          {turnsLeft}/{totalDuration}回合
                        </span>
                      </div>
                    </div>
                    <div style={{ fontSize: '12px', color: '#AAA', marginTop: '4px' }}>
                      来源: {getPlayerName(intel.source_player_id)}
                    </div>
                    <div style={{ fontSize: '12px', color: '#CCC', marginTop: '4px' }}>
                      {intel.content}
                    </div>
                    {intel.target_techs && intel.target_techs.length > 0 && (
                      <div style={{ fontSize: '11px', color: '#888', marginTop: '2px' }}>
                        科技: {intel.target_techs.join(', ')}
                      </div>
                    )}
                    {intel.target_alliances && intel.target_alliances.length > 0 && (
                      <div style={{ fontSize: '11px', color: '#888', marginTop: '2px' }}>
                        联盟: {intel.target_alliances.join(', ')}
                      </div>
                    )}
                    {!isSanctioned && (
                      <button
                        className="btn btn-secondary btn-tiny"
                        onClick={() => { setSellIntelId(intel.id); setSellPrice('250'); }}
                        style={{ marginTop: '4px' }}
                      >
                        上架到市场
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {sellIntelId && (
            <div style={{ marginTop: '8px', padding: '8px', border: '1px solid #FFCC44', borderRadius: '4px' }}>
              <h5>出售情报</h5>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <input
                  type="number"
                  value={sellPrice}
                  onChange={(e) => setSellPrice(e.target.value)}
                  placeholder="价格"
                  className="diplomacy-input"
                  style={{ width: '80px' }}
                  min="1"
                />
                <span style={{ fontSize: '12px', color: '#888' }}>¢</span>
                <button className="btn btn-success btn-tiny" onClick={handleSellIntel} disabled={!sellPrice || parseFloat(sellPrice) <= 0}>
                  确认上架
                </button>
                <button className="btn btn-secondary btn-tiny" onClick={() => { setSellIntelId(''); setSellPrice(''); }}>
                  取消
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {subTab === 'counter' && (
        <div className="diplomacy-section">
          <h4>反间谍设置</h4>
          <div style={{ marginBottom: '12px' }}>
            <div style={{ fontSize: '13px', color: '#CCC', marginBottom: '8px' }}>
              当前等级: <strong style={{ color: '#4488FF' }}>{COUNTER_SPY_NAMES[currentCounterLevel]}</strong>
            </div>
            <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
              {Object.values(COUNTER_SPY_LEVELS).map((level) => (
                <button
                  key={level}
                  className={`btn ${currentCounterLevel === level ? 'btn-primary' : 'btn-secondary'} btn-small`}
                  onClick={() => handleSetCounterSpy(level)}
                >
                  {COUNTER_SPY_NAMES[level]}
                </button>
              ))}
            </div>
          </div>
          <div style={{ fontSize: '12px', color: '#888' }}>
            <div>低级: 20%概率检测到间谍活动(仅通知)</div>
            <div>中级: 40%概率检测并识别来源</div>
            <div>高级: 60%概率检测+识别+反制(间谍暴露值+30)</div>
            <div style={{ marginTop: '4px', color: '#FF8844' }}>战争期间反间谍效率自动降低一档</div>
          </div>
        </div>
      )}

      {subTab === 'market' && (
        <div className="diplomacy-section">
          <h4>情报市场</h4>

          {otherListings.length > 0 && (
            <div style={{ marginBottom: '12px' }}>
              <h5>在售情报</h5>
              {otherListings.map((l) => {
                const intel = (intelligences || []).find((i) => i.id === l.intel_id);
                const listingTotal = INTEL_DURATION;
                const listingRemaining = l.remaining_turns;
                const listingRatio = listingTotal > 0 ? listingRemaining / listingTotal : 0;
                const listingBarColor = listingRatio > 0.6 ? '#44FF44' : listingRatio > 0.3 ? '#FFCC44' : '#FF4444';
                const listingShouldFlash = listingRatio <= 0.3 && listingRatio > 0;
                return (
                  <div key={l.id} style={{ padding: '8px', marginBottom: '6px', border: '1px solid #3a3a6a', borderRadius: '4px', background: 'rgba(255,255,255,0.03)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontSize: '12px', color: '#CCC' }}>
                        卖家: {getPlayerName(l.seller_id)}
                      </span>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginTop: '4px' }}>
                        <div style={{ flex: 1, height: '4px', background: '#333', borderRadius: '2px', overflow: 'hidden' }}>
                          <div style={{
                            width: `${Math.max(0, listingRatio * 100)}%`,
                            height: '100%',
                            background: listingBarColor,
                            borderRadius: '2px',
                            animation: listingShouldFlash ? 'intelFlash 1s ease-in-out infinite' : 'none',
                          }} />
                        </div>
                        <span style={{ fontSize: '11px', color: listingBarColor, whiteSpace: 'nowrap' }}>
                          {listingRemaining}/{listingTotal}回合
                        </span>
                      </div>
                    </div>
                    <div style={{ fontSize: '13px', marginTop: '4px' }}>
                      {intel ? intel.summary || intel.content : '情报详情'}
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '6px' }}>
                      <span style={{ color: '#FFCC44', fontWeight: 'bold' }}>{Math.round(l.price)}¢</span>
                      <button className="btn btn-success btn-tiny" onClick={() => handleBuyIntel(l.id)}>
                        购买
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {myListings.length > 0 && (
            <div style={{ marginBottom: '12px' }}>
              <h5>我的挂单</h5>
              {myListings.map((l) => {
                const myListingTotal = INTEL_DURATION;
                const myListingRemaining = l.remaining_turns;
                const myListingRatio = myListingTotal > 0 ? myListingRemaining / myListingTotal : 0;
                const myListingBarColor = myListingRatio > 0.6 ? '#44FF44' : myListingRatio > 0.3 ? '#FFCC44' : '#FF4444';
                const myListingShouldFlash = myListingRatio <= 0.3 && myListingRatio > 0;
                return (
                  <div key={l.id} style={{ padding: '6px 8px', marginBottom: '4px', border: '1px solid #FFCC44', borderRadius: '4px', fontSize: '12px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span>价格: {Math.round(l.price)}¢</span>
                      <button className="btn btn-danger btn-tiny" onClick={() => handleCancelListing(l.id)}>
                        下架
                      </button>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginTop: '4px' }}>
                      <div style={{ flex: 1, height: '4px', background: '#333', borderRadius: '2px', overflow: 'hidden' }}>
                        <div style={{
                          width: `${Math.max(0, myListingRatio * 100)}%`,
                          height: '100%',
                          background: myListingBarColor,
                          borderRadius: '2px',
                          animation: myListingShouldFlash ? 'intelFlash 1s ease-in-out infinite' : 'none',
                        }} />
                      </div>
                      <span style={{ fontSize: '11px', color: myListingBarColor, whiteSpace: 'nowrap' }}>
                        {myListingRemaining}/{myListingTotal}回合
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {otherListings.length === 0 && myListings.length === 0 && (
            <div className="empty-state">情报市场暂无在售情报</div>
          )}

          {isSanctioned && (
            <div style={{ color: '#FF4444', fontSize: '12px', marginTop: '8px' }}>
              你已被制裁，无法在情报市场挂单
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default IntelPanel;
