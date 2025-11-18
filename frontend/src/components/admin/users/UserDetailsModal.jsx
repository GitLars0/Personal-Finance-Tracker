import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import Modal from '../../Modal';
import { formatDate } from '../../../utils/format';

const UserDetailsModal = ({ user, onClose, onRoleUpdated }) => {
  const [userDetails, setUserDetails] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [updatingRole, setUpdatingRole] = useState(false);
  const [newRole, setNewRole] = useState(user.role);

  useEffect(() => {
    const loadUserDetails = async () => {
      try {
        setLoading(true);
        const data = await adminApi.getUserDetails(user.id);
        setUserDetails(data);
      } catch (err) {
        setError('Failed to load user details: ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    loadUserDetails();
  }, [user.id]);

  const retryLoadUserDetails = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await adminApi.getUserDetails(user.id);
      setUserDetails(data);
    } catch (err) {
      setError('Failed to load user details: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRoleUpdate = async () => {
    if (newRole === user.role) return;

    try {
      setUpdatingRole(true);
      await adminApi.updateUserRole(user.id, newRole);
      onRoleUpdated();
    } catch (err) {
      setError('Failed to update role: ' + err.message);
    } finally {
      setUpdatingRole(false);
    }
  };

  if (loading) {
    return (
      <Modal isOpen={true} onClose={onClose} title="User Details">
        <div className="loading">Loading user details...</div>
      </Modal>
    );
  }

  if (error) {
    return (
      <Modal isOpen={true} onClose={onClose} title="User Details">
        <div className="error">{error}</div>
        <button onClick={retryLoadUserDetails} className="btn-primary">Retry</button>
      </Modal>
    );
  }

  return (
    <Modal isOpen={true} onClose={onClose} title={`User Details - ${user.username}`}>
      <div className="user-details-modal">
        {/* Basic Info */}
        <div className="details-section">
          <h3>Basic Information</h3>
          <div className="details-grid">
            <div className="detail-item">
              <label>User ID:</label>
              <span>#{userDetails.user.id}</span>
            </div>
            <div className="detail-item">
              <label>Username:</label>
              <span>{userDetails.user.username}</span>
            </div>
            <div className="detail-item">
              <label>Email:</label>
              <span>{userDetails.user.email}</span>
            </div>
            <div className="detail-item">
              <label>Name:</label>
              <span>{userDetails.user.name || 'Not set'}</span>
            </div>
            <div className="detail-item">
              <label>Role:</label>
              <span className={`role-badge ${userDetails.user.role}`}>
                {userDetails.user.role === 'admin' ? '👑' : '👤'} {userDetails.user.role}
              </span>
            </div>
            <div className="detail-item">
              <label>Created:</label>
              <span>{formatDate(userDetails.user.created_at)}</span>
            </div>
            <div className="detail-item">
              <label>Last Updated:</label>
              <span>{formatDate(userDetails.user.updated_at)}</span>
            </div>
          </div>
        </div>

        {/* Statistics */}
        <div className="details-section">
          <h3>Account Statistics</h3>
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-number">{userDetails.statistics.accounts}</div>
              <div className="stat-label">Accounts</div>
            </div>
            <div className="stat-card">
              <div className="stat-number">{userDetails.statistics.transactions}</div>
              <div className="stat-label">Transactions</div>
            </div>
            <div className="stat-card">
              <div className="stat-number">{userDetails.statistics.categories}</div>
              <div className="stat-label">Categories</div>
            </div>
            <div className="stat-card">
              <div className="stat-number">{userDetails.statistics.budgets}</div>
              <div className="stat-label">Budgets</div>
            </div>
          </div>
        </div>

        {/* Role Management */}
        <div className="details-section">
          <h3>Role Management</h3>
          <div className="role-management">
            <div className="role-selector">
              <label htmlFor="role-select">Change Role:</label>
              <select
                id="role-select"
                value={newRole}
                onChange={(e) => setNewRole(e.target.value)}
                disabled={updatingRole}
              >
                <option value="user">👤 Regular User</option>
                <option value="admin">👑 Administrator</option>
              </select>
            </div>
            
            {newRole !== user.role && (
              <div className="role-change-warning">
                <p>⚠️ You are about to change this user's role from <strong>{user.role}</strong> to <strong>{newRole}</strong>.</p>
                {newRole === 'admin' && (
                  <p>🔐 This will grant administrative privileges to this user.</p>
                )}
                {newRole === 'user' && user.role === 'admin' && (
                  <p>📉 This will remove administrative privileges from this user.</p>
                )}
              </div>
            )}

            <button
              onClick={handleRoleUpdate}
              disabled={newRole === user.role || updatingRole}
              className="btn-primary"
            >
              {updatingRole ? 'Updating...' : 'Update Role'}
            </button>
          </div>
        </div>

        {/* Recent Activity (if available) */}
        {userDetails.recentTransactions && userDetails.recentTransactions.length > 0 && (
          <div className="details-section">
            <h3>Recent Activity</h3>
            <div className="recent-activity">
              {userDetails.recentTransactions.slice(0, 5).map(transaction => (
                <div key={transaction.id} className="activity-item">
                  <span className="activity-date">{formatDate(transaction.created_at)}</span>
                  <span className="activity-description">{transaction.description}</span>
                  <span className="activity-amount">${transaction.amount}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="modal-actions">
          <button onClick={onClose} className="btn-secondary">Close</button>
        </div>
      </div>
    </Modal>
  );
};

export default UserDetailsModal;