import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import '../styles/Sidebar.css';

function Sidebar({ isAuthenticated }) {
  const [isOpen, setIsOpen] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  
  // Update user data when authentication changes
  useEffect(() => {
    if (isAuthenticated) {
      const userData = JSON.parse(localStorage.getItem('user') || '{}');
      setIsAdmin(userData.role === 'admin');
      
      // Debug logging
      console.log('Sidebar - User data:', userData);
      console.log('Sidebar - Is admin:', userData.role === 'admin');
      console.log('Sidebar - User role:', userData.role);
    } else {
      setIsAdmin(false);
    }
  }, [isAuthenticated]);
  
  // Listen for auth changes
  useEffect(() => {
    const handleAuthChange = () => {
      if (isAuthenticated) {
        const userData = JSON.parse(localStorage.getItem('user') || '{}');
        setIsAdmin(userData.role === 'admin');
      }
    };

    window.addEventListener('authChange', handleAuthChange);
    return () => window.removeEventListener('authChange', handleAuthChange);
  }, [isAuthenticated]);

  const toggleSidebar = () => {
    setIsOpen(!isOpen);
  };

  const closeSidebar = () => {
    setIsOpen(false);
  };

  if (!isAuthenticated) {
    return null; // Don't show sidebar if not authenticated
  }

  return (
    <>
      {/* Sidebar Toggle Button */}
      <button 
        className={`sidebar-toggle ${isOpen ? 'active' : ''}`}
        onClick={toggleSidebar}
        aria-label="Toggle sidebar"
      >
        <span className="hamburger-line"></span>
        <span className="hamburger-line"></span>
        <span className="hamburger-line"></span>
      </button>

      {/* Sidebar Overlay */}
      {isOpen && <div className="sidebar-overlay" onClick={closeSidebar}></div>}

      {/* Sidebar */}
      <div className={`sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-header">
          <h3>Quick Access</h3>
          <button className="sidebar-close" onClick={closeSidebar}>
            ×
          </button>
        </div>

        <div className="sidebar-content">
          {/* Admin Section - only show for admin users */}
          {isAdmin && (
            <div className="sidebar-section admin-section">
              <h4>🔐 Administration</h4>
              <ul className="sidebar-links">
                <li>
                  <Link 
                    to="/admin" 
                    className="sidebar-link admin-link"
                    onClick={closeSidebar}
                  >
                    <span className="sidebar-icon">👑</span>
                    <span>Admin Dashboard</span>
                  </Link>
                </li>
                <li>
                  <Link 
                    to="/admin/users" 
                    className="sidebar-link admin-link"
                    onClick={closeSidebar}
                  >
                    <span className="sidebar-icon">👥</span>
                    <span>User Management</span>
                  </Link>
                </li>
                <li>
                  <Link 
                    to="/admin/transactions" 
                    className="sidebar-link admin-link"
                    onClick={closeSidebar}
                  >
                    <span className="sidebar-icon">💳</span>
                    <span>All Transactions</span>
                  </Link>
                </li>
                <li>
                  <Link 
                    to="/admin/accounts" 
                    className="sidebar-link admin-link"
                    onClick={closeSidebar}
                  >
                    <span className="sidebar-icon">🏦</span>
                    <span>All Accounts</span>
                  </Link>
                </li>
              </ul>
            </div>
          )}

          <div className="sidebar-section">
            <h4>Management</h4>
            <ul className="sidebar-links">
              <li>
                <Link 
                  to="/accounts" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">🏦</span>
                  <span>My Accounts</span>
                </Link>
              </li>
              <li>
                <Link 
                  to="/categories" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">📂</span>
                  <span>Categories</span>
                </Link>
              </li>
              <li>
                <Link 
                  to="/banks" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">🔗</span>
                  <span>Bank Connections</span>
                </Link>
              </li>
            </ul>
          </div>

          <div className="sidebar-section">
            <h4>Quick Actions</h4>
            <ul className="sidebar-links">
              <li>
                <Link 
                  to="/transactions" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">💰</span>
                  <span>Add Transaction</span>
                </Link>
              </li>
              <li>
                <Link 
                  to="/budgets" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">📊</span>
                  <span>Manage Budgets</span>
                </Link>
              </li>
            </ul>
          </div>

          <div className="sidebar-section">
            <h4>Overview</h4>
            <ul className="sidebar-links">
              <li>
                <Link 
                  to="/dashboard" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">📈</span>
                  <span>Dashboard</span>
                </Link>
              </li>
            </ul>
          </div>

          <div className="sidebar-section">
            <h4>Account</h4>
            <ul className="sidebar-links">
              <li>
                <Link 
                  to="/profile" 
                  className="sidebar-link"
                  onClick={closeSidebar}
                >
                  <span className="sidebar-icon">👤</span>
                  <span>Profile Settings</span>
                </Link>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </>
  );
}

export default Sidebar;