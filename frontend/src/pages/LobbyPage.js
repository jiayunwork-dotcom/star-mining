import React, { useState, useEffect } from 'react';
import { useGameState } from '../hooks/useGameState';
import { MESSAGE_TYPES } from '../types/game';
import ws from '../utils/websocket';

function LobbyPage() {
  const { state, connect, createRoom, joinRoom, listRooms, startGame, leaveRoom } = useGameState();
  const [nickname, setNickname] = useState('');
  const [roomName, setRoomName] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [joinedRoom, setJoinedRoom] = useState(null);

  useEffect(() => {
    if (state.isConnected && state.rooms.length === 0) {
      listRooms();
    }
  }, [state.isConnected, state.rooms.length, listRooms]);

  useEffect(() => {
    const handleSystem = (data) => {
      if (!data) return;
      
      if (data.type === 'room_created' && data.success) {
        setJoinedRoom({ id: data.room_id, name: data.name, players: data.players || [] });
      } else if (data.type === 'room_joined' && data.success) {
        setJoinedRoom({ id: data.room_id, name: data.name, players: data.players || [] });
      } else if (data.type === 'rooms_list') {
        // 房间列表通过 state.rooms 处理
      }
    };

    const unlistenSystem = ws.on(MESSAGE_TYPES.SYSTEM, handleSystem);

    return () => {
      unlistenSystem();
    };
  }, []);

  const handleConnect = async () => {
    if (!nickname.trim()) return;
    setIsConnecting(true);
    const success = await connect(nickname.trim());
    setIsConnecting(false);
    if (success) {
      setTimeout(() => listRooms(), 500);
    }
  };

  const handleCreateRoom = () => {
    if (!roomName.trim()) return;
    createRoom(roomName.trim());
    setRoomName('');
  };

  const handleJoinRoom = (roomId) => {
    joinRoom(roomId);
  };

  const handleLeaveRoom = () => {
    leaveRoom();
    setJoinedRoom(null);
  };

  const handleStartGame = () => {
    startGame();
  };

  if (!state.isConnected && !isConnecting) {
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
        </div>
      </div>
    );
  }

  if (isConnecting) {
    return (
      <div className="lobby-page">
        <div className="lobby-container">
          <div className="loading">正在连接服务器...</div>
        </div>
      </div>
    );
  }

  if (joinedRoom || state.roomId) {
    const room = joinedRoom || { id: state.roomId, name: '游戏房间', players: state.players };
    return (
      <div className="lobby-page">
        <div className="lobby-container">
          <h2 className="lobby-title">{room.name}</h2>
          <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '16px' }}>
            房间号: {room.id}
          </p>

          <div className="players-in-room">
            <h3>玩家列表 ({state.players.length})</h3>
            {state.players.length === 0 ? (
              <div className="empty-state">暂无玩家</div>
            ) : (
              state.players.map((player) => (
                <div key={player.id || player.playerId} className="player-item">
                  <span className="player-name">{player.name || player.nickname}</span>
                  <span className="player-status">
                    {player.is_host ? '房主' : '玩家'}
                  </span>
                </div>
              ))
            )}
          </div>

          <div className="room-controls">
            <button className="btn btn-danger" onClick={handleLeaveRoom}>
              离开房间
            </button>
            <button
              className="btn btn-success"
              onClick={handleStartGame}
              disabled={state.players.length < 2}
            >
              开始游戏
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="lobby-page">
      <div className="lobby-container">
        <h2 className="lobby-title">游戏大厅</h2>
        <p style={{ textAlign: 'center', color: '#8080a0', marginBottom: '24px' }}>
          欢迎，{state.nickname}
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
                  <p>{room.players}/{room.max_players || 4} 名玩家</p>
                </div>
                <button
                  className="btn btn-success btn-small"
                  onClick={() => handleJoinRoom(room.id)}
                  disabled={room.players >= (room.max_players || 4)}
                >
                  加入
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
