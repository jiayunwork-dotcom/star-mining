import React from 'react';
import { useGameState } from '../hooks/useGameState';
import { TECH_NAMES, TECH_DESCRIPTIONS, TECH_TYPES } from '../types/game';

const techTypes = Object.values(TECH_TYPES);
const MAX_TECH_LEVEL = 10;

const techCosts = [
  { credits: 500 },
  { credits: 1000 },
  { credits: 2000 },
  { credits: 4000 },
  { credits: 8000 },
  { credits: 16000 },
  { credits: 32000 },
  { credits: 64000 },
  { credits: 128000 },
  { credits: 256000 },
];

function TechPanel() {
  const { state, researchTech } = useGameState();
  const { myPlayer } = state;

  const techTree = myPlayer?.tech_tree || {};

  const handleResearch = (techType) => {
    const currentLevel = techTree[techType] || 0;
    if (currentLevel >= MAX_TECH_LEVEL) return;
    const cost = techCosts[currentLevel];
    if (!cost) return;
    if (myPlayer.credits < cost.credits) {
      alert('信用币不足');
      return;
    }
    researchTech(techType);
  };

  const getLevelStatus = (techType, level) => {
    const currentLevel = techTree[techType] || 0;
    if (level < currentLevel) return 'researched';
    if (level === currentLevel) return 'current';
    return 'locked';
  };

  return (
    <div className="panel">
      <h3 className="panel-title">科技树</h3>

      <div className="tech-tree">
        {techTypes.map((techType) => {
          const currentLevel = techTree[techType] || 0;
          const nextCost = currentLevel < MAX_TECH_LEVEL ? techCosts[currentLevel] : null;
          const canResearch = currentLevel < MAX_TECH_LEVEL && myPlayer.credits >= (nextCost?.credits || 0);

          return (
            <div key={techType} className="tech-route">
              <div className="tech-route-header">
                <span className="tech-route-name">
                  {TECH_NAMES[techType]}
                </span>
                <span className="tech-route-level">
                  Lv.{currentLevel}/{MAX_TECH_LEVEL}
                </span>
              </div>

              <p style={{ fontSize: '11px', color: '#8080a0', marginBottom: '8px' }}>
                {TECH_DESCRIPTIONS[techType]}
              </p>

              <div className="tech-levels">
                {Array.from({ length: MAX_TECH_LEVEL }).map((_, i) => {
                  const status = getLevelStatus(techType, i);
                  return (
                    <div
                      key={i}
                      className={`tech-level-dot ${status === 'researched' ? 'researched' : ''}`}
                      title={`等级 ${i + 1}`}
                    />
                  );
                })}
              </div>

              {currentLevel < MAX_TECH_LEVEL ? (
                <>
                  <div style={{ fontSize: '11px', color: '#8080a0', marginBottom: '8px' }}>
                    下一级: ¢{nextCost?.credits?.toLocaleString() || 0}
                  </div>
                  <button
                    className="btn btn-small research-btn"
                    onClick={() => handleResearch(techType)}
                    disabled={!canResearch}
                  >
                    研发 Lv.{currentLevel + 1}
                  </button>
                </>
              ) : (
                <div style={{ textAlign: 'center', fontSize: '12px', color: '#51cf66' }}>
                  已达最高等级
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default TechPanel;
