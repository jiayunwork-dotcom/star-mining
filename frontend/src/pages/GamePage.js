import React, { useState } from 'react';
import GalaxyMap from '../components/GalaxyMap.jsx';
import CompanyPanel from '../components/CompanyPanel.jsx';
import TradePanel from '../components/TradePanel.jsx';
import FleetPanel from '../components/FleetPanel.jsx';
import TechPanel from '../components/TechPanel.jsx';
import DiplomacyPanel from '../components/DiplomacyPanel.jsx';
import EventLog from '../components/EventLog.jsx';
import TurnControl from '../components/TurnControl.jsx';
import TurnReportModal from '../components/TurnReportModal.jsx';
import { useGameState } from '../hooks/useGameState';

const sidebarTabs = [
  { id: 'company', name: '公司', component: CompanyPanel },
  { id: 'fleet', name: '舰队', component: FleetPanel },
  { id: 'trade', name: '贸易', component: TradePanel },
  { id: 'tech', name: '科技', component: TechPanel },
  { id: 'diplomacy', name: '外交', component: DiplomacyPanel },
];

function GamePage() {
  const { state, leaveRoom, confirmTurnReport } = useGameState();
  const [activeTab, setActiveTab] = useState('company');

  const ActivePanel = sidebarTabs.find((t) => t.id === activeTab)?.component || CompanyPanel;

  const handleBackToLobby = () => {
    if (window.confirm('确定要返回大厅吗？当前游戏进度将丢失。')) {
      leaveRoom();
    }
  };

  return (
    <div className="game-page">
      <div className="game-sidebar">
        <div className="sidebar-tabs">
          {sidebarTabs.map((tab) => (
            <div
              key={tab.id}
              className={`sidebar-tab ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.name}
            </div>
          ))}
        </div>
        <div className="sidebar-content">
          <ActivePanel />
        </div>
        <div style={{ padding: '12px', borderTop: '1px solid #3a3a6a' }}>
          <button
            className="btn btn-danger btn-small btn-block"
            onClick={handleBackToLobby}
          >
            返回大厅
          </button>
        </div>
      </div>

      <div className="game-main">
        <TurnControl />
        <GalaxyMap />
        <div className="bottom-panels">
          <EventLog />
        </div>
      </div>

      {state.showTurnReport && state.turnReport && (
        <TurnReportModal
          report={state.turnReport}
          onConfirm={confirmTurnReport}
          confirmations={state.reportConfirmations}
          players={state.players}
        />
      )}
    </div>
  );
}

export default GamePage;
