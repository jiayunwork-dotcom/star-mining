import React from 'react';
import { GameProvider, useGameState } from './hooks/useGameState';
import LobbyPage from './pages/LobbyPage';
import GamePage from './pages/GamePage';
import { GAME_STATES } from './types/game';
import './App.css';

function AppContent() {
  const { state } = useGameState();

  const renderPage = () => {
    if (state.gameState === GAME_STATES.PLAYING || state.currentPage === 'game') {
      return <GamePage />;
    }
    return <LobbyPage />;
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>⭐ 星际矿业 Star Mining</h1>
        <div className="connection-status">
          <div className={`status-dot ${state.isConnected ? 'connected' : ''}`} />
          <span>{state.isConnected ? '已连接' : '未连接'}</span>
          {state.nickname && (
            <span style={{ marginLeft: '12px', color: '#a0a0d0' }}>
              {state.nickname}
            </span>
          )}
        </div>
      </header>
      <main className="main-content">
        {renderPage()}
      </main>
    </div>
  );
}

function App() {
  return (
    <GameProvider>
      <AppContent />
    </GameProvider>
  );
}

export default App;
