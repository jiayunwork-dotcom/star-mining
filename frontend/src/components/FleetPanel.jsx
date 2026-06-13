import React, { useState } from 'react';
import { useGameState } from '../hooks/useGameState';
import { SHIP_TYPE_NAMES, SHIP_COSTS, RESOURCE_NAMES } from '../types/game';

const shipTypes = Object.keys(SHIP_TYPE_NAMES);

function FleetPanel() {
  const { state, buildShip, moveFleet, getAllCelestials } = useGameState();
  const { myPlayer, selectedCelestial } = state;
  const [selectedShipType, setSelectedShipType] = useState('mining');
  const [selectedShips, setSelectedShips] = useState([]);
  const [targetCelestial, setTargetCelestial] = useState('');

  const ships = myPlayer?.ships || [];
  const celestials = getAllCelestials();

  const handleBuildShip = () => {
    if (!selectedCelestial) {
      alert('请先在地图上选择一个天体作为建造位置');
      return;
    }
    buildShip(selectedShipType, selectedCelestial.id);
  };

  const handleToggleShip = (shipId) => {
    setSelectedShips((prev) =>
      prev.includes(shipId)
        ? prev.filter((id) => id !== shipId)
        : [...prev, shipId]
    );
  };

  const handleMoveFleet = () => {
    if (selectedShips.length === 0) {
      alert('请先选择要移动的飞船');
      return;
    }
    if (!targetCelestial) {
      alert('请选择目标天体');
      return;
    }
    moveFleet(selectedShips, targetCelestial);
    setSelectedShips([]);
    setTargetCelestial('');
  };

  const getCelestialName = (celestialId) => {
    const celestial = celestials.find((c) => c.id === celestialId);
    return celestial ? celestial.name : '未知';
  };

  const cost = SHIP_COSTS[selectedShipType] || {};

  return (
    <div className="panel">
      <h3 className="panel-title">舰队管理</h3>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '0 0 8px' }}>
        我的舰队 ({ships.length})
      </h4>

      {ships.length === 0 ? (
        <div className="empty-state">暂无飞船</div>
      ) : (
        <div className="fleet-list" style={{ maxHeight: '180px', overflowY: 'auto' }}>
          {ships.map((ship) => (
            <div
              key={ship.id}
              className="fleet-item"
              style={{
                borderColor: selectedShips.includes(ship.id) ? '#6dd5ed' : undefined,
                cursor: 'pointer',
              }}
              onClick={() => handleToggleShip(ship.id)}
            >
              <div className="fleet-item-header">
                <span className="fleet-ship-name">{ship.name || SHIP_TYPE_NAMES[ship.type]}</span>
                <span className="fleet-ship-type">{SHIP_TYPE_NAMES[ship.type]}</span>
              </div>
              <div className="fleet-ship-location">
                位置: {getCelestialName(ship.celestial_id)}
              </div>
              {ship.status && (
                <div style={{ fontSize: '11px', color: '#ffd43b', marginTop: '4px' }}>
                  {ship.status}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {selectedShips.length > 0 && (
        <div style={{ marginTop: '12px', padding: '12px', background: 'rgba(0,0,0,0.3)', borderRadius: '8px' }}>
          <div style={{ fontSize: '12px', color: '#a0a0d0', marginBottom: '8px' }}>
            已选择 {selectedShips.length} 艘飞船
          </div>
          <select
            value={targetCelestial}
            onChange={(e) => setTargetCelestial(e.target.value)}
            style={{
              width: '100%',
              padding: '8px',
              background: 'rgba(0,0,0,0.4)',
              border: '1px solid #4a4a7a',
              borderRadius: '6px',
              color: '#e0e0ff',
              marginBottom: '8px',
              outline: 'none',
            }}
          >
            <option value="">选择目标天体...</option>
            {celestials.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          <button
            className="btn btn-primary btn-small btn-block"
            onClick={handleMoveFleet}
          >
            移动舰队
          </button>
        </div>
      )}

      <div className="build-ship-form">
        <h4>建造飞船</h4>

        <div className="ship-type-grid">
          {shipTypes.map((type) => (
            <div
              key={type}
              className={`ship-type-option ${selectedShipType === type ? 'selected' : ''}`}
              onClick={() => setSelectedShipType(type)}
            >
              <div className="ship-type-option-name">
                {SHIP_TYPE_NAMES[type]}
              </div>
              <div className="ship-type-option-cost">
                ¢{SHIP_COSTS[type].credits}
              </div>
            </div>
          ))}
        </div>

        <div style={{ fontSize: '11px', color: '#8080a0', marginBottom: '8px' }}>
          消耗: 
          {Object.entries(cost).map(([key, value], index) => (
            <span key={key}>
              {index > 0 && '，'}
              {key === 'credits' ? '信用币' : RESOURCE_NAMES[key] || key}: {value}
            </span>
          ))}
        </div>

        {selectedCelestial ? (
          <div style={{ fontSize: '11px', color: '#a0a0d0', marginBottom: '8px' }}>
            建造位置: {selectedCelestial.name}
          </div>
        ) : (
          <div style={{ fontSize: '11px', color: '#ff6b6b', marginBottom: '8px' }}>
            请在地图上选择建造位置
          </div>
        )}

        <button
          className="btn btn-success btn-small btn-block"
          onClick={handleBuildShip}
          disabled={!selectedCelestial}
        >
          建造飞船
        </button>
      </div>
    </div>
  );
}

export default FleetPanel;
