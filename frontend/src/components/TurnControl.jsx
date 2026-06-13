import React, { useState } from 'react';
import { useGameState } from '../hooks/useGameState';

function TurnControl() {
  const { state, endTurn } = useGameState();
  const { turn, isConnected } = state;
  const [isEnding, setIsEnding] = useState(false);

  const handleEndTurn = async () => {
    if (isEnding) return;
    if (window.confirm('确定要结束当前回合吗？')) {
      setIsEnding(true);
      endTurn();
      setTimeout(() => setIsEnding(false), 1000);
    }
  };

  return (
    <div className="turn-control">
      <div className="turn-info">
        <div className="turn-number">第 {turn || 1} 回合</div>
        <div className="turn-label">回合</div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginLeft: '24px' }}>
        <div
          style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            background: isConnected ? '#51cf66' : '#ff6b6b',
          }}
        />
        <span style={{ fontSize: '12px', color: '#8080a0' }}>
          {isConnected ? '已连接' : '未连接'}
        </span>
      </div>

      <button
        className="btn btn-primary end-turn-btn"
        onClick={handleEndTurn}
        disabled={!isConnected || isEnding}
      >
        {isEnding ? '结算中...' : '结束回合'}
      </button>
    </div>
  );
}

export default TurnControl;
