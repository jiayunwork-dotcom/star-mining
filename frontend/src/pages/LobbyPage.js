import React, { useState, useEffect } from 'react';
import { useGameState } from '../hooks/useGameState';

function LobbyPage() {
  const { state, connect, createRoom, joinRoom, listRooms, startGame, leaveRoom, ready, unready } = useGameState();
  const [nickname, setNickname] = useState('');
  const [roomName, setRoomName] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);

  useEffect(() => {
    if (state.playerId && state.rooms.length === 0) {
      listRooms();
    }
  }, [state.playerId, state.rooms.length, listRooms]);

  const handleConnect = async () => {
    if (!nickname.trim()) return;
    setIsConnecting(true);
    const success = await connect(nickname.trim());
    setIsConnecting(false);
    if (success) {
      setTimeout(() => listRooms(), 300);
    }
  };

  const handleCreateRoom = async () => {
    if (!roomName.trim()) return;
    setIsConnecting(true);
    const success = await createRoom(roomName.trim());
    setIsConnecting(false);
    if (success) {
      setRoomName('');
    }
  };

  const handleJoinRoom = async (roomId) => {
    setIsConnecting(true);
    await joinRoom(roomId);
    setIsConnecting(false);
  };

  const handleLeaveRoom = () => {
    leaveRoom();
  };

  const handleStartGame = async () => {
    setIsConnecting(true);
    await startGame();
    setIsConnecting(false);
  };

  const myPlayer = state.players.find((p) => p.id === state.playerId);
  const isHost = state.players.length > 0 && state.players[0].id === state.playerId;
  const allReady = state.players.length >= 2 && state.players.every((p) => p.ready);

  if (!state.playerId && !isConnecting) {
    return (
      <div className="lobby-page">
        <div className="lobby-container">
          <h2 className="lobby-title">⭐ 星际矿业 ⭐</h2>
          <div className="nickname-input">
            <label>输入你的昵称</label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              placeholder="请输入昵称..."
              onKeyPress={(e) => e.key === 'Enter' && handleConnect()}
            />
          </div>
          <button
            className="btn btn-primary btn-block"
            onClick={handleConnect}
            disabled={!nickname.trim()}
          >
            连接服务器
          </button>
          <p style={{ textAlign: 'center', color: '#8080a0', marginTop: '16px', fontSize: '12px' }}>
            请先登录以进入游戏大厅
          </p>
        </div>
      </div>
    );
  }

  if (isConnecting) {
    return (
      <div className="lobby-page">
        <div className="lobby-container">
          <div className="loading">正在处理请求...</div>
        </div>
      </div>
    );
  }

  if (state.roomId) {
    return (
      <div className="lobby-page">
        <div className="lobby-container">
          <h2 className="lobby-title">{state.roomName || '游戏房间'}</h2>
          <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '8px' }}>
            房间号: {state.roomId}
          </p>
          <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '16px', fontSize: '12px' }}>
            连接状态: {state.isConnected ? '✅ 已连接' : '❌ 未连接'}
          </p>

          <div className="players-in-room">
            <h3>玩家列表 ({state.players.length})</h3>
            {state.players.length === 0 ? (
              <div className="empty-state">暂无玩家</div>
            ) : (
              state.players.map((player, index) => (
                <div key={player.id || index} className="player-item">
                  <span className="player-name" style={{ color: player.color || '#fff' }}>
                    {player.name || player.nickname}
                    {index === 0 && ' 👑'}
                  </span>
                  <span className="player-status">
                    {player.ready ? '✅ 已准备' : '⏳ 未准备'}
                  </span>
                </div>
              ))
            )}
          </div>

          <div className="room-controls">
            <button className="btn btn-danger" onClick={handleLeaveRoom}>
              离开房间
            </button>
            {!isHost && myPlayer && (
              myPlayer.ready ? (
                <button
                  className="btn btn-warning"
                  onClick={unready}
                >
                  取消准备
                </button>
              ) : (
                <button
                  className="btn btn-success"
                  onClick={ready}
                >
                  准备游戏
                </button>
              )
            )}
            {isHost && (
              <button
                className="btn btn-success"
                onClick={handleStartGame}
                disabled={!allReady || !state.isConnected}
              >
                开始游戏
              </button>
            )}
          </div>

          <div style={{ marginTop: '16px', fontSize: '12px', color: '#8080a0', textAlign: 'center' }}>
            {isHost
              ? `作为房主，需要 ${2 - state.players.length} 名更多玩家且全员准备后开始游戏`
              : '请点击"准备游戏"按钮，等待房主开始游戏'}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="lobby-page">
      <div className="lobby-container">
        <h2 className="lobby-title">游戏大厅</h2>
        <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '8px' }}>
          欢迎，{state.nickname}
        </p>
        <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '24px', fontSize: '12px' }}>
          玩家ID: {state.playerId?.substring(0, 8)}...
        </p>

        <div className="create-room-form">
          <input
            type="text"
            value={roomName}
            onChange={(e) => setRoomName(e.target.value)}
            placeholder="输入房间名..."
            onKeyPress={(e) => e.key === 'Enter' && handleCreateRoom()}
          />
          <button
            className="btn btn-primary"
            onClick={handleCreateRoom}
            disabled={!roomName.trim()}
          >
            创建房间
          </button>
        </div>

        <div className="room-list">
          <h3>房间列表</h3>
          {state.rooms.length === 0 ? (
            <div className="empty-state">暂无房间，创建一个吧</div>
          ) : (
            state.rooms.map((room) => (
              <div key={room.id} className="room-item">
                <div className="room-info">
                  <h4>{room.name}</h4>
                  <p>{room.players}/{room.max_players || 6} 名玩家</p>
                </div>
                <button
                  className="btn btn-success btn-small"
                  onClick={() => handleJoinRoom(room.id)}
                  disabled={room.players >= (room.max_players || 6) || room.status !== 'waiting'}
                >
                  {room.status === 'waiting' ? '加入' : '游戏中'}
                </button>
              </div>
            ))
          )}
        </div>

        <button className="btn btn-block" onClick={listRooms}>
          刷新列表
        </button>
      </div>
    </div>
  );
}

export default LobbyPage;
