import React from 'react';
import { useGameState } from '../hooks/useGameState';
import { RESOURCE_NAMES, SHIP_TYPE_NAMES } from '../types/game';

function CompanyPanel() {
  const { state } = useGameState();
  const { myPlayer } = state;

  if (!myPlayer) {
    return (
      <div className="panel">
        <h3 className="panel-title">公司信息</h3>
        <div className="empty-state">加载中...</div>
      </div>
    );
  }

  const resources = myPlayer.resources || {};
  const ships = myPlayer.ships || [];
  const techTree = myPlayer.tech_tree || {};

  const shipCountByType = ships.reduce((acc, ship) => {
    acc[ship.type] = (acc[ship.type] || 0) + 1;
    return acc;
  }, {});

  return (
    <div className="panel">
      <h3 className="panel-title">公司信息</h3>

      <div className="credits-display">
        ¢ {myPlayer.credits?.toLocaleString() || 0}
      </div>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '12px 0 8px' }}>
        资源库存
      </h4>
      <div className="resources-grid">
        {Object.entries(RESOURCE_NAMES).map(([key, name]) => {
          if (key === 'credits') return null;
          return (
            <div key={key} className="resource-item">
              <span className="resource-name">{name}</span>
              <span className="resource-amount">
                {resources[key]?.toLocaleString() || 0}
              </span>
            </div>
          );
        })}
      </div>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '16px 0 8px' }}>
        舰队 ({ships.length})
      </h4>
      <div className="resources-grid">
        {Object.entries(SHIP_TYPE_NAMES).map(([key, name]) => (
          <div key={key} className="resource-item">
            <span className="resource-name">{name}</span>
            <span className="resource-amount">{shipCountByType[key] || 0}</span>
          </div>
        ))}
      </div>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '16px 0 8px' }}>
        公司资产
      </h4>
      <div style={{ fontSize: '12px', color: '#8080a0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
          <span>公司名称</span>
          <span style={{ color: '#ffd43b' }}>{myPlayer.company_name || myPlayer.name}</span>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
          <span>空间站</span>
          <span>{(myPlayer.stations || []).length}</span>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
          <span>精炼厂</span>
          <span>{(myPlayer.refineries || []).length}</span>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <span>造船厂</span>
          <span>{(myPlayer.shipyards || []).length}</span>
        </div>
      </div>

      {Object.keys(techTree).length > 0 && (
        <>
          <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '16px 0 8px' }}>
            科技等级
          </h4>
          <div style={{ fontSize: '12px', color: '#8080a0' }}>
            {Object.entries(techTree).map(([key, level]) => (
              <div key={key} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                <span></span>
                <span style={{ color: '#cc5de8' }}>Lv.{level}</span>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

export default CompanyPanel;
