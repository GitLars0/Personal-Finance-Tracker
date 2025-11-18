import React from 'react';

function DashboardHeader({ onRefresh }) {
  return (
    <header className="dashboard-header">
      <h1>Financial Dashboard</h1>
      <button className="refresh-btn" onClick={onRefresh}>
        🔄 Refresh
      </button>
    </header>
  );
}

export default DashboardHeader;