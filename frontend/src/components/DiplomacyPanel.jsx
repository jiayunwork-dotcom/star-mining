import React, { useState } from 'react';
import { useGameState } from '../hooks/useGameState';

function DiplomacyPanel() {
  const { state, bidAuction, tradeStock, acquireCompany, imposeEmbargo } = useGameState();
  const { otherPlayers, bids, myPlayer, blockades } = state;
  const [bidAmounts, setBidAmounts] = useState({});
  const [acquirePrices, setAcquirePrices] = useState({});
  const [stockAmounts, setStockAmounts] = useState({});

  const handleBid = (auctionId) => {
    const amount = parseFloat(bidAmounts[auctionId] || 0);
    if (amount <= 0) return;
    bidAuction(auctionId, amount);
    setBidAmounts({ ...bidAmounts, [auctionId]: '' });
  };

  const handleBuyStock = (playerId) => {
    const amount = parseInt(stockAmounts[playerId] || 0, 10);
    if (amount <= 0) return;
    tradeStock(playerId, amount);
    setStockAmounts({ ...stockAmounts, [playerId]: '' });
  };

  const handleSellStock = (playerId) => {
    const amount = parseInt(stockAmounts[playerId] || 0, 10);
    if (amount <= 0) return;
    tradeStock(playerId, -amount);
    setStockAmounts({ ...stockAmounts, [playerId]: '' });
  };

  const handleAcquire = (playerId) => {
    const price = parseFloat(acquirePrices[playerId] || 0);
    if (price <= 0) return;
    if (window.confirm(`确定要以 ¢${price} 的价格收购这家公司吗？`)) {
      acquireCompany(playerId, price);
      setAcquirePrices({ ...acquirePrices, [playerId]: '' });
    }
  };

  const handleEmbargo = (playerId, playerName) => {
    if (window.confirm(`确定要对 ${playerName} 实施贸易封锁吗？`)) {
      imposeEmbargo(playerId);
    }
  };

  const activeBids = (bids || []).filter((b) => b.status === 'active');

  return (
    <div className="panel">
      <h3 className="panel-title">外交中心</h3>

      <div className="diplomacy-section">
        <h4>矿区竞标 ({activeBids.length})</h4>
        {activeBids.length === 0 ? (
          <div className="empty-state">暂无进行中的拍卖</div>
        ) : (
          activeBids.map((bid) => (
            <div key={bid.id} className="auction-item">
              <div className="auction-title">{bid.name || '矿区拍卖'}</div>
              <div className="auction-info">
                最高出价: ¢{bid.highest_bid?.toLocaleString() || 0}
                {bid.highest_bidder && ` (${bid.highest_bidder})`}
              </div>
              <div className="auction-info">
                剩余回合: {bid.turns_left || '-'}
              </div>
              <div className="auction-bid-form">
                <input
                  type="number"
                  placeholder="出价"
                  value={bidAmounts[bid.id] || ''}
                  onChange={(e) =>
                    setBidAmounts({ ...bidAmounts, [bid.id]: e.target.value })
                  }
                />
                <button
                  className="btn btn-warning btn-small"
                  onClick={() => handleBid(bid.id)}
                >
                  出价
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      <div className="diplomacy-section">
        <h4>其他公司 ({otherPlayers?.length || 0})</h4>
        {!otherPlayers || otherPlayers.length === 0 ? (
          <div className="empty-state">暂无其他公司</div>
        ) : (
          <div className="company-list">
            {otherPlayers.map((player) => (
              <div key={player.id} className="company-item">
                <div className="company-item-header">
                  <span className="company-name">{player.company_name || player.name}</span>
                  <span className="company-stock-price">
                    ¢{player.stock_price?.toLocaleString() || 0}
                  </span>
                </div>
                <div style={{ fontSize: '11px', color: '#8080a0', marginBottom: '6px' }}>
                  市值: ¢{((player.stock_price || 0) * (player.shares || 1000)).toLocaleString()}
                </div>

                <div style={{ display: 'flex', gap: '4px', marginBottom: '8px' }}>
                  <input
                    type="number"
                    placeholder="股份数"
                    value={stockAmounts[player.id] || ''}
                    onChange={(e) =>
                      setStockAmounts({ ...stockAmounts, [player.id]: e.target.value })
                    }
                    style={{
                      flex: 1,
                      padding: '4px 6px',
                      background: 'rgba(0,0,0,0.4)',
                      border: '1px solid #4a4a7a',
                      borderRadius: '4px',
                      color: '#e0e0ff',
                      fontSize: '11px',
                      outline: 'none',
                    }}
                  />
                  <button
                    className="btn btn-success btn-small"
                    onClick={() => handleBuyStock(player.id)}
                  >
                    买
                  </button>
                  <button
                    className="btn btn-danger btn-small"
                    onClick={() => handleSellStock(player.id)}
                  >
                    卖
                  </button>
                </div>

                <div style={{ display: 'flex', gap: '4px', marginBottom: '8px' }}>
                  <input
                    type="number"
                    placeholder="收购价"
                    value={acquirePrices[player.id] || ''}
                    onChange={(e) =>
                      setAcquirePrices({ ...acquirePrices, [player.id]: e.target.value })
                    }
                    style={{
                      flex: 1,
                      padding: '4px 6px',
                      background: 'rgba(0,0,0,0.4)',
                      border: '1px solid #4a4a7a',
                      borderRadius: '4px',
                      color: '#e0e0ff',
                      fontSize: '11px',
                      outline: 'none',
                    }}
                  />
                  <button
                    className="btn btn-warning btn-small"
                    onClick={() => handleAcquire(player.id)}
                  >
                    收购
                  </button>
                </div>

                <button
                  className="btn btn-danger btn-small btn-block"
                  onClick={() => handleEmbargo(player.id, player.company_name || player.name)}
                >
                  贸易封锁
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {myPlayer?.stocks && myPlayer.stocks.length > 0 && (
        <div className="diplomacy-section">
          <h4>我的持股</h4>
          {myPlayer.stocks.map((stock) => (
            <div key={stock.player_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', fontSize: '12px' }}>
              <span style={{ color: '#a0a0d0' }}>{stock.company_name || stock.player_id}</span>
              <span style={{ color: '#e0e0ff' }}>{stock.shares} 股</span>
            </div>
          ))}
        </div>
      )}

      {blockades && blockades.length > 0 && (
        <div className="diplomacy-section">
          <h4>贸易封锁</h4>
          {blockades.map((blockade) => (
            <div key={blockade.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', fontSize: '12px' }}>
              <span style={{ color: '#ff6b6b' }}>{blockade.target_name || blockade.target_id}</span>
              <span style={{ color: '#8080a0' }}>{blockade.turns_left || '-'} 回合</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default DiplomacyPanel;
