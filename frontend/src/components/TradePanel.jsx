import React, { useState, useMemo } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import { useGameState } from '../hooks/useGameState';
import { RESOURCE_NAMES, ORDER_TYPES } from '../types/game';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const resourceList = Object.keys(RESOURCE_NAMES).filter((r) => r !== 'credits');

function TradePanel() {
  const { state, placeOrder, cancelOrder } = useGameState();
  const { exchanges } = state;
  const [selectedResource, setSelectedResource] = useState('iron_ore');
  const [orderType, setOrderType] = useState(ORDER_TYPES.BUY);
  const [price, setPrice] = useState('');
  const [quantity, setQuantity] = useState('');

  const exchange = exchanges && exchanges.length > 0 ? exchanges[0] : null;
  const prices = exchange?.prices || {};
  const orders = exchange?.orders || [];

  const chartData = useMemo(() => {
    const priceHistory = exchange?.price_history || {};
    const history = priceHistory[selectedResource] || [];
    const labels = history.map((_, i) => `T-${history.length - i - 1}`);
    
    return {
      labels,
      datasets: [
        {
          label: RESOURCE_NAMES[selectedResource] + ' 价格',
          data: history,
          borderColor: '#6dd5ed',
          backgroundColor: 'rgba(109, 213, 237, 0.1)',
          fill: true,
          tension: 0.4,
          pointRadius: 2,
          pointHoverRadius: 4,
        },
      ],
    };
  }, [selectedResource, exchange]);

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      x: {
        display: false,
      },
      y: {
        grid: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
        ticks: {
          color: '#8080a0',
        },
      },
    },
  };

  const filteredOrders = orders.filter((o) => o.resource === selectedResource);
  const buyOrders = filteredOrders.filter((o) => o.type === ORDER_TYPES.BUY).sort((a, b) => b.price - a.price);
  const sellOrders = filteredOrders.filter((o) => o.type === ORDER_TYPES.SELL).sort((a, b) => a.price - b.price);

  const handlePlaceOrder = () => {
    const p = parseFloat(price);
    const q = parseInt(quantity, 10);
    if (isNaN(p) || isNaN(q) || p <= 0 || q <= 0) return;
    placeOrder(orderType, selectedResource, p, q);
    setPrice('');
    setQuantity('');
  };

  const handleCancelOrder = (orderId) => {
    cancelOrder(orderId);
  };

  return (
    <div className="panel">
      <h3 className="panel-title">贸易所</h3>

      <div className="resource-selector">
        {resourceList.map((res) => (
          <button
            key={res}
            className={`resource-selector-btn ${selectedResource === res ? 'active' : ''}`}
            onClick={() => setSelectedResource(res)}
          >
            {RESOURCE_NAMES[res]}
          </button>
        ))}
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
        <span style={{ fontSize: '12px', color: '#8080a0' }}>当前价格</span>
        <span style={{ fontSize: '18px', fontWeight: 'bold', color: '#ffd43b' }}>
          ¢ {prices[selectedResource]?.toLocaleString() || 0}
        </span>
      </div>

      <div className="price-chart-container">
        <Line data={chartData} options={chartOptions} />
      </div>

      <div className="market-tabs">
        <div
          className={`market-tab ${orderType === ORDER_TYPES.BUY ? 'active' : ''}`}
          onClick={() => setOrderType(ORDER_TYPES.BUY)}
        >
          买入
        </div>
        <div
          className={`market-tab ${orderType === ORDER_TYPES.SELL ? 'active' : ''}`}
          onClick={() => setOrderType(ORDER_TYPES.SELL)}
        >
          卖出
        </div>
      </div>

      <div className="order-form">
        <input
          type="number"
          placeholder="价格"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
        />
        <input
          type="number"
          placeholder="数量"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
        />
        <button
          className={`btn btn-small ${orderType === ORDER_TYPES.BUY ? 'btn-success' : 'btn-danger'}`}
          onClick={handlePlaceOrder}
        >
          挂单
        </button>
      </div>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '12px 0 8px' }}>
        买单 (最高 {buyOrders.length > 0 ? buyOrders[0].price : '-'})
      </h4>
      <div className="order-book">
        {buyOrders.length === 0 ? (
          <div className="empty-state" style={{ padding: '10px' }}>暂无买单</div>
        ) : (
          buyOrders.slice(0, 5).map((order) => (
            <div key={order.id} className="order-item buy">
            <span>¢{order.price}</span>
            <span>{order.quantity}</span>
            {order.is_mine && (
              <button
                className="btn btn-danger btn-small"
                onClick={() => handleCancelOrder(order.id)}
              >
                撤
              </button>
            )}
          </div>
          ))
        )}
      </div>

      <h4 style={{ fontSize: '13px', color: '#a0a0d0', margin: '12px 0 8px' }}>
        卖单 (最低 {sellOrders.length > 0 ? sellOrders[0].price : '-'})
      </h4>
      <div className="order-book">
        {sellOrders.length === 0 ? (
          <div className="empty-state" style={{ padding: '10px' }}>暂无卖单</div>
        ) : (
          sellOrders.slice(0, 5).map((order) => (
            <div key={order.id} className="order-item sell">
            <span>¢{order.price}</span>
            <span>{order.quantity}</span>
            {order.is_mine && (
              <button
                className="btn btn-danger btn-small"
                onClick={() => handleCancelOrder(order.id)}
              >
                撤
              </button>
            )}
          </div>
          ))
        )}
      </div>
    </div>
  );
}

export default TradePanel;
