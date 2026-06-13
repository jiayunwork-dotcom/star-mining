import React, { useState, useRef, useEffect } from 'react';
import { useGameState } from '../hooks/useGameState';
import { CELESTIAL_TYPES, RESOURCE_NAMES, CELESTIAL_TYPE_NAMES } from '../types/game';

const celestialColors = {
  star: '#ffd43b',
  planet: '#6dd5ed',
  asteroid_belt: '#a0a0d0',
  gas_giant: '#cc5de8',
  terrestrial: '#51cf66',
};

const celestialSizes = {
  star: 20,
  planet: 14,
  asteroid_belt: 8,
  gas_giant: 16,
  terrestrial: 12,
};

function GalaxyMap() {
  const { state, selectCelestial, getAllCelestials } = useGameState();
  const { selectedCelestial, myPlayer, gameMap } = state;
  const svgRef = useRef(null);
  const [viewBox, setViewBox] = useState({ x: 0, y: 0, w: 800, h: 600 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

  const celestials = getAllCelestials();
  const lanes = gameMap?.lanes || [];

  useEffect(() => {
    if (celestials && celestials.length > 0) {
      const xs = celestials.map((c) => c.x);
      const ys = celestials.map((c) => c.y);
      const minX = Math.min(...xs) - 100;
      const maxX = Math.max(...xs) + 100;
      const minY = Math.min(...ys) - 100;
      const maxY = Math.max(...ys) + 100;
      setViewBox({ x: minX, y: minY, w: maxX - minX, h: maxY - minY });
    }
  }, [celestials]);

  const handleMouseDown = (e) => {
    if (e.target.tagName === 'svg') {
      setIsDragging(true);
      setDragStart({ x: e.clientX, y: e.clientY });
    }
  };

  const handleMouseMove = (e) => {
    if (!isDragging) return;
    const dx = (e.clientX - dragStart.x) * (viewBox.w / svgRef.current.clientWidth);
    const dy = (e.clientY - dragStart.y) * (viewBox.h / svgRef.current.clientHeight);
    setViewBox((prev) => ({
      ...prev,
      x: prev.x - dx,
      y: prev.y - dy,
    }));
    setDragStart({ x: e.clientX, y: e.clientY });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleWheel = (e) => {
    e.preventDefault();
    const scale = e.deltaY > 0 ? 1.1 : 0.9;
    const rect = svgRef.current.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const worldX = viewBox.x + (mouseX / rect.width) * viewBox.w;
    const worldY = viewBox.y + (mouseY / rect.height) * viewBox.h;
    const newW = viewBox.w * scale;
    const newH = viewBox.h * scale;
    const newX = worldX - (mouseX / rect.width) * newW;
    const newY = worldY - (mouseY / rect.height) * newH;
    setViewBox({ x: newX, y: newY, w: newW, h: newH });
  };

  const handleCelestialClick = (celestial) => {
    selectCelestial(celestial);
  };

  const handleCloseDetail = () => {
    selectCelestial(null);
  };

  const shipsAtCelestial = (celestialId) => {
    if (!myPlayer || !myPlayer.ships) return 0;
    return myPlayer.ships.filter((s) => s.celestial_id === celestialId).length;
  };

  const renderStars = () => {
    const stars = [];
    for (let i = 0; i < 100; i++) {
      const x = Math.random() * 2000 - 500;
      const y = Math.random() * 1500 - 300;
      const r = Math.random() * 1.5 + 0.5;
      const opacity = Math.random() * 0.5 + 0.2;
      stars.push(
        <circle key={i} cx={x} cy={y} r={r} fill="white" opacity={opacity} />
      );
    }
    return stars;
  };

  return (
    <div className="galaxy-map-container">
      <svg
        ref={svgRef}
        className="galaxy-map-svg"
        viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.w} ${viewBox.h}`}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onWheel={handleWheel}
      >
        {renderStars()}

        {lanes && lanes.map((lane, index) => {
          const from = celestials.find((c) => c.id === lane.from);
          const to = celestials.find((c) => c.id === lane.to);
          if (!from || !to) return null;
          return (
            <line
              key={`lane-${index}`}
              className="lane-line"
              x1={from.x}
              y1={from.y}
              x2={to.x}
              y2={to.y}
            />
          );
        })}

        {celestials && celestials.map((celestial) => {
          const size = celestialSizes[celestial.type] || 10;
          const color = celestialColors[celestial.type] || '#a0a0d0';
          const isSelected = selectedCelestial && selectedCelestial.id === celestial.id;
          const shipCount = shipsAtCelestial(celestial.id);

          return (
            <g
              key={celestial.id}
              className={`celestial-node ${isSelected ? 'selected' : ''}`}
              onClick={() => handleCelestialClick(celestial)}
              transform={`translate(${celestial.x}, ${celestial.y})`}
            >
              {celestial.type === CELESTIAL_TYPES.STAR && (
                <>
                  <circle r={size + 10} fill={color} opacity="0.2" />
                  <circle r={size + 5} fill={color} opacity="0.4" />
                </>
              )}

              {celestial.type === CELESTIAL_TYPES.PLANET && (
                <defs>
                  <radialGradient id={`grad-${celestial.id}`}>
                    <stop offset="0%" stopColor={color} stopOpacity="1" />
                    <stop offset="100%" stopColor={color} stopOpacity="0.5" />
                  </radialGradient>
                </defs>
              )}

              <circle
                r={size}
                fill={celestial.type === 'planet' ? `url(#grad-${celestial.id})` : color}
                stroke={isSelected ? '#6dd5ed' : 'none'}
                strokeWidth={isSelected ? 3 : 0}
              />

              {shipCount > 0 && (
                <g transform={`translate(${size + 5}, ${-size - 5})`}>
                  <circle r="8" fill="#51cf66" />
                  <text
                    textAnchor="middle"
                    dy="3"
                    fill="white"
                    fontSize="10"
                    fontWeight="bold"
                  >
                    {shipCount}
                  </text>
                </g>
              )}

              <text
                y={size + 16}
                textAnchor="middle"
                fill="#a0a0d0"
                fontSize="11"
              >
                {celestial.name}
              </text>
            </g>
          );
        })}
      </svg>

      {selectedCelestial && (
        <div className="celestial-detail">
          <button className="close-detail-btn" onClick={handleCloseDetail}>
            ×
          </button>
          <h3>{selectedCelestial.name}</h3>
          <p className="celestial-type">
            {CELESTIAL_TYPE_NAMES[selectedCelestial.type] || selectedCelestial.type}
          </p>

          <div className="celestial-stats">
            {selectedCelestial.resources && (
              <>
                {Object.entries(selectedCelestial.resources).map(([key, value]) => (
                  <div key={key} className="celestial-stat">
                    <span className="celestial-stat-label">
                      {RESOURCE_NAMES[key] || key}
                    </span>
                    <span className="celestial-stat-value">{value}</span>
                  </div>
                ))}
              </>
            )}

            {selectedCelestial.owner && (
              <div className="celestial-stat">
                <span className="celestial-stat-label">所有者</span>
                <span className="celestial-stat-value">{selectedCelestial.owner}</span>
              </div>
            )}

            {selectedCelestial.mining_rate !== undefined && (
              <div className="celestial-stat">
                <span className="celestial-stat-label">采矿效率</span>
                <span className="celestial-stat-value">{selectedCelestial.mining_rate}x</span>
              </div>
            )}
          </div>

          {shipsAtCelestial(selectedCelestial.id) > 0 && (
            <div style={{ marginTop: '12px', paddingTop: '12px', borderTop: '1px solid #3a3a6a' }}>
              <p style={{ fontSize: '13px', color: '#a0a0d0' }}>
                驻扎飞船: {shipsAtCelestial(selectedCelestial.id)} 艘
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default GalaxyMap;
