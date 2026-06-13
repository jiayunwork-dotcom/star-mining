import React, { useEffect, useCallback } from 'react';
import { RESOURCE_NAMES } from '../types/game';

function formatNumber(num) {
  if (num === undefined || num === null || isNaN(num)) return '0';
  return Math.round(num * 100) / 100;
}

function getColorClass(num) {
  if (num > 0) return 'text-positive';
  if (num < 0) return 'text-negative';
  return '';
}

function formatPercent(num) {
  const sign = num > 0 ? '+' : '';
  return `${sign}${formatNumber(num)}%`;
}

function TurnReportModal({ report, onConfirm, confirmations, players }) {
  const handleKeyDown = useCallback((e) => {
    if (e.key === 'Escape') {
      onConfirm();
    }
  }, [onConfirm]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
    };
  }, [handleKeyDown]);

  if (!report) return null;

  const playerMap = {};
  (players || []).forEach(p => {
    playerMap[p.id] = p;
  });

  const {
    resource_changes = [],
    finance = {},
    fleet_activity = {},
    market_changes = [],
    random_events = [],
    rankings = [],
  } = report;

  const confirmedCount = Object.keys(confirmations || {}).filter(k => confirmations[k]).length;
  const totalPlayers = Object.keys(playerMap).length;

  return (
    <div className="modal-overlay" onClick={onConfirm}>
      <div className="turn-report-modal" onClick={e => e.stopPropagation()}>
        <div className="turn-report-header">
          <div>
            <h2 className="turn-report-title">回合 {report.turn} 结算报告</h2>
            <p className="turn-report-subtitle">{report.player_name} 的回合详情</p>
          </div>
          <button className="modal-close-btn" onClick={onConfirm} aria-label="关闭">×</button>
        </div>

        <div className="turn-report-content">
          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">📦</span> 资源变动
            </h3>
            <div className="table-container">
              <table className="report-table">
                <thead>
                  <tr>
                    <th>矿物</th>
                    <th className="num-col">产出</th>
                    <th className="num-col">消耗</th>
                    <th className="num-col">交易</th>
                    <th className="num-col">净变化</th>
                  </tr>
                </thead>
                <tbody>
                  {resource_changes.map(rc => (
                    <tr key={rc.resource_type}>
                      <td>{RESOURCE_NAMES[rc.resource_type] || rc.resource_name}</td>
                      <td className={`num-col ${getColorClass(rc.produced)}`}>
                        {rc.produced > 0 ? `+${formatNumber(rc.produced)}` : formatNumber(rc.produced)}
                      </td>
                      <td className={`num-col ${getColorClass(-rc.consumed)}`}>
                        {rc.consumed > 0 ? `-${formatNumber(rc.consumed)}` : '0'}
                      </td>
                      <td className={`num-col ${getColorClass(rc.traded)}`}>
                        {rc.traded > 0 ? `+${formatNumber(rc.traded)}` : formatNumber(rc.traded)}
                      </td>
                      <td className={`num-col ${getColorClass(rc.net_change)} bold`}>
                        {rc.net_change > 0 ? `+${formatNumber(rc.net_change)}` : formatNumber(rc.net_change)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>

          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">💰</span> 财务收支
            </h3>
            <div className="finance-grid">
              <div className="finance-col">
                <h4 className="finance-col-title income">收入明细</h4>
                <div className="finance-list">
                  {(finance.income_items || []).map((item, idx) => (
                    <div key={idx} className="finance-item">
                      <span className="finance-label">{item.label}</span>
                      <span className="finance-amount text-positive">+{formatNumber(item.amount)}</span>
                    </div>
                  ))}
                  <div className="finance-item total-row">
                    <span className="finance-label">总收入</span>
                    <span className="finance-amount bold text-positive">+{formatNumber(finance.total_income)}</span>
                  </div>
                </div>
              </div>
              <div className="finance-col">
                <h4 className="finance-col-title expense">支出明细</h4>
                <div className="finance-list">
                  {(finance.expense_items || []).map((item, idx) => (
                    <div key={idx} className="finance-item">
                      <span className="finance-label">{item.label}</span>
                      <span className="finance-amount text-negative">-{formatNumber(item.amount)}</span>
                    </div>
                  ))}
                  <div className="finance-item total-row">
                    <span className="finance-label">总支出</span>
                    <span className="finance-amount bold text-negative">-{formatNumber(finance.total_expense)}</span>
                  </div>
                </div>
              </div>
            </div>
            <div className="finance-summary">
              <div className="summary-item">
                <span className="summary-label">本回合净收入</span>
                <span className={`summary-value bold ${getColorClass(finance.net_income)}`}>
                  {finance.net_income > 0 ? `+${formatNumber(finance.net_income)}` : formatNumber(finance.net_income)}
                </span>
              </div>
              <div className="summary-item highlight">
                <span className="summary-label">当前余额</span>
                <span className="summary-value credits-value">{formatNumber(finance.current_balance)} ¢</span>
              </div>
            </div>
          </section>

          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">🚀</span> 舰队动态
            </h3>

            <div className="fleet-subsection">
              <h4 className="subsection-title">移动情况</h4>
              {(fleet_activity.movements && fleet_activity.movements.length > 0) ? (
                <div className="event-card-list">
                  {fleet_activity.movements.map((mv, idx) => (
                    <div key={idx} className={`event-card ${mv.arrived ? 'arrived' : 'moving'}`}>
                      <div className="event-card-title">
                        <strong>{mv.fleet_name}</strong>
                        {mv.arrived ? (
                          <span className="status-badge status-success">已到达</span>
                        ) : (
                          <span className="status-badge status-info">航行中</span>
                        )}
                      </div>
                      <div className="event-card-body">
                        <span className="route-text">
                          {mv.from_body_name || mv.from_body_id} → {mv.to_body_name || mv.to_body_id}
                        </span>
                        {!mv.arrived && (
                          <span className="remain-turns">剩余 {mv.turns_remaining} 回合</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-notice">本回合无舰队移动</div>
              )}
            </div>

            <div className="fleet-subsection">
              <h4 className="subsection-title">战斗记录</h4>
              {(fleet_activity.combats && fleet_activity.combats.length > 0) ? (
                <div className="event-card-list">
                  {fleet_activity.combats.map((cb, idx) => (
                    <div key={idx} className="event-card combat-card">
                      <div className="event-card-title">
                        <strong>战斗</strong>
                        <span className={`status-badge ${
                          cb.winner === 'attacker' ? 'status-success' :
                          cb.winner === 'defender' ? 'status-danger' : 'status-warning'
                        }`}>
                          {cb.winner === 'attacker' ? '攻击方胜' :
                           cb.winner === 'defender' ? '防御方胜' : '平局'}
                        </span>
                      </div>
                      <div className="combat-line">
                        <span className="combat-side attacker">{cb.attacker_name}</span>
                        <span className="combat-vs">VS</span>
                        <span className="combat-side defender">{cb.defender_name}</span>
                      </div>
                      <div className="combat-stats">
                        <div>攻击方损失: <span className="text-negative">{cb.attacker_losses}</span> 艘舰船</div>
                        <div>防御方损失: <span className="text-negative">{cb.defender_losses}</span> 艘舰船</div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-notice">本回合无战斗</div>
              )}
            </div>

            <div className="fleet-subsection">
              <h4 className="subsection-title">海盗袭击</h4>
              {(fleet_activity.pirate_attacks && fleet_activity.pirate_attacks.length > 0) ? (
                <div className="event-card-list">
                  {fleet_activity.pirate_attacks.map((pa, idx) => (
                    <div key={idx} className={`event-card pirate-card ${pa.defended ? 'defended' : 'attacked'}`}>
                      <div className="event-card-title">
                        <strong>⚔️ 海盗袭击 - {pa.fleet_name}</strong>
                        <span className={`status-badge ${pa.defended ? 'status-success' : 'status-danger'}`}>
                          {pa.defended ? '击退' : '受损'}
                        </span>
                      </div>
                      <div className="event-card-body">
                        <span>位置: {pa.location_name}</span>
                        <span>损失舰船: <span className="text-negative">{pa.player_losses}</span></span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-notice">本回合无海盗袭击</div>
              )}
            </div>
          </section>

          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">📈</span> 市场行情
            </h3>
            <div className="market-grid">
              {market_changes.map(mc => (
                <div key={mc.resource_type} className="market-card">
                  <div className="market-name">{RESOURCE_NAMES[mc.resource_type] || mc.resource_name}</div>
                  <div className="market-price-row">
                    <span className="price-old">{formatNumber(mc.old_price)}¢</span>
                    <span className="price-arrow">→</span>
                    <span className="price-new">{formatNumber(mc.new_price)}¢</span>
                  </div>
                  <div className={`market-change ${getColorClass(mc.change_percent)}`}>
                    {mc.change_percent > 0 ? '🔺' : mc.change_percent < 0 ? '🔻' : '➡️'}
                    {formatPercent(mc.change_percent)}
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">🎲</span> 随机事件
            </h3>
            {random_events && random_events.length > 0 ? (
              <div className="event-card-list">
                {random_events.map(ev => (
                  <div key={ev.event_id}
                       className={`event-card random-event ${ev.affects_me ? 'affects-me' : ''} ${ev.is_global ? 'global' : ''}`}>
                    <div className="event-card-title">
                      <strong>{ev.name}</strong>
                      {ev.is_global ? (
                        <span className="status-badge status-warning tag-global">全局事件</span>
                      ) : (
                        ev.affects_me && <span className="status-badge status-danger tag-affects">影响我方</span>
                      )}
                    </div>
                    <div className="event-card-body">
                      {ev.description}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="calm-notice">
                🌤️ 本回合风平浪静
              </div>
            )}
          </section>

          {report.diplomacy && report.diplomacy.changes && report.diplomacy.changes.length > 0 && (
            <section className="report-section">
              <h3 className="report-section-title">
                <span className="section-icon">🤝</span> 外交动态
              </h3>
              <div className="diplomacy-report-list">
                {report.diplomacy.changes.map((dc, idx) => {
                  const playerName = playerMap[dc.player_id]
                    ? (playerMap[dc.player_id].company_name || playerMap[dc.player_id].name)
                    : dc.player_id;
                  const isPositive = dc.change > 0;
                  const changeStr = isPositive ? `+${dc.change}` : `${dc.change}`;
                  return (
                    <div key={idx} className="diplomacy-report-item">
                      <span className="diplomacy-report-player">{playerName}</span>
                      <span className="diplomacy-report-reason">{dc.reason}</span>
                      <span className={`diplomacy-report-change ${isPositive ? 'text-positive' : 'text-negative'}`}>
                        {changeStr}
                      </span>
                      <span className="diplomacy-report-values">
                        {dc.old_value} → {dc.new_value}
                      </span>
                    </div>
                  );
                })}
              </div>
            </section>
          )}

          <section className="report-section">
            <h3 className="report-section-title">
              <span className="section-icon">🏆</span> 排名变动
            </h3>
            <div className="table-container">
              <table className="report-table ranking-table">
                <thead>
                  <tr>
                    <th>排名</th>
                    <th>玩家</th>
                    <th>公司</th>
                    <th className="num-col">评分</th>
                    <th className="num-col">变动</th>
                  </tr>
                </thead>
                <tbody>
                  {rankings.map(r => (
                    <tr key={r.player_id} className={r.is_me ? 'my-row' : ''}>
                      <td>
                        <span className={`rank-badge rank-${r.rank}`}>#{r.rank}</span>
                      </td>
                      <td className={r.is_me ? 'bold' : ''}>
                        {r.is_me && <span className="me-tag">我</span>}
                        {r.player_name}
                        {(r.is_bankrupt || r.is_defeated) && (
                          <span className="status-badge status-danger ml-8">出局</span>
                        )}
                      </td>
                      <td>{r.company_name}</td>
                      <td className="num-col">{formatNumber(r.score)}</td>
                      <td className="num-col">
                        {r.rank_change > 0 ? (
                          <span className="text-positive">🔺 +{r.rank_change}</span>
                        ) : r.rank_change < 0 ? (
                          <span className="text-negative">🔻 {r.rank_change}</span>
                        ) : (
                          <span className="text-neutral">—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <div className="turn-report-footer">
          <div className="confirmations-info">
            已确认: {confirmedCount} / {totalPlayers} 人
          </div>
          <button className="btn btn-primary confirm-turn-btn" onClick={onConfirm}>
            确认并进入下一回合 (ESC)
          </button>
        </div>
      </div>
    </div>
  );
}

export default TurnReportModal;
