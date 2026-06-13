import React from 'react';
import { useGameState } from '../hooks/useGameState';

function EventLog() {
  const { state } = useGameState();
  const { eventLog } = state;

  const formatTime = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const getEventTypeClass = (eventType) => {
    return `event-${eventType || 'default'}`;
  };

  return (
    <div className="bottom-panel">
      <div className="bottom-panel-header">
        <h3>事件日志</h3>
        <span style={{ fontSize: '12px', color: '#8080a0' }}>
          {eventLog.length} 条记录
        </span>
      </div>
      <div className="bottom-panel-content">
        <div className="event-log">
          {eventLog.length === 0 ? (
            <div className="empty-state">暂无事件</div>
          ) : (
            eventLog.map((event, index) => (
              <div
                key={index}
                className={`event-item ${getEventTypeClass(event.type)}`}
              >
                <div className="event-time">
                  回合 {event.turn || '-'} · {formatTime(event.timestamp)}
                </div>
                <div className="event-message">
                  {event.message || event.description || JSON.stringify(event.data || {})}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

export default EventLog;
