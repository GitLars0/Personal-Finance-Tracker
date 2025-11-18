import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import adminApi from '../utils/adminApi';
import '../styles/Admin.css';
import '../styles/AdminComponents.css';

const Admin = () => {
  const [stats, setStats] = useState({
    totalUsers: 0,
    totalTransactions: 0,
    totalAccounts: 0,
    totalCategories: 0,
    totalBudgets: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadDashboardStats();
  }, []);

  const loadDashboardStats = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getDashboardStats();
      setStats(data);
    } catch (err) {
      setError('Failed to load dashboard statistics: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="admin-container">
        <div className="loading">Loading admin dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="admin-container">
        <div className="error">{error}</div>
      </div>
    );
  }

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h1>Admin Dashboard</h1>
        <p>System overview and management tools</p>
      </div>

      {/* Statistics Cards */}
      <div className="admin-stats-grid">
        <div className="stat-card">
          <div className="stat-icon">👥</div>
          <div className="stat-info">
            <h3>{stats.totalUsers}</h3>
            <p>Total Users</p>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">💳</div>
          <div className="stat-info">
            <h3>{stats.totalTransactions}</h3>
            <p>Total Transactions</p>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">🏦</div>
          <div className="stat-info">
            <h3>{stats.totalAccounts}</h3>
            <p>Total Accounts</p>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">📂</div>
          <div className="stat-info">
            <h3>{stats.totalCategories}</h3>
            <p>Total Categories</p>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">📊</div>
          <div className="stat-info">
            <h3>{stats.totalBudgets}</h3>
            <p>Total Budgets</p>
          </div>
        </div>
      </div>

      {/* Management Sections */}
      <div className="admin-sections">
        <h2>Management Tools</h2>
        <div className="admin-nav-grid">
          <Link to="/admin/users" className="admin-nav-card">
            <div className="nav-icon">👥</div>
            <h3>User Management</h3>
            <p>View, edit, and manage user accounts</p>
          </Link>

          <Link to="/admin/transactions" className="admin-nav-card">
            <div className="nav-icon">💳</div>
            <h3>All Transactions</h3>
            <p>Monitor and manage all user transactions</p>
          </Link>

          <Link to="/admin/accounts" className="admin-nav-card">
            <div className="nav-icon">🏦</div>
            <h3>All Accounts</h3>
            <p>Oversee all financial accounts</p>
          </Link>

          <Link to="/admin/categories" className="admin-nav-card">
            <div className="nav-icon">📂</div>
            <h3>All Categories</h3>
            <p>Manage category structures</p>
          </Link>

          <Link to="/admin/budgets" className="admin-nav-card">
            <div className="nav-icon">📊</div>
            <h3>All Budgets</h3>
            <p>Review user budget plans</p>
          </Link>

          <div className="admin-nav-card system-card">
            <div className="nav-icon">⚙️</div>
            <h3>System Tools</h3>
            <p>Database management and system utilities</p>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="admin-quick-actions">
        <h2>Quick Actions</h2>
        <div className="quick-action-buttons">
          <button 
            className="action-btn refresh-btn"
            onClick={loadDashboardStats}
          >
            🔄 Refresh Stats
          </button>
          <Link to="/admin/users" className="action-btn primary-btn">
            👤 Manage Users
          </Link>
          <Link to="/admin/transactions" className="action-btn secondary-btn">
            📊 View All Data
          </Link>
        </div>
      </div>
    </div>
  );
};

export default Admin;