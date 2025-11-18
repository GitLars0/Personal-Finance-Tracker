import React from 'react';
import { formatDate } from '../../../utils/format';

const UserCard = ({ user, onViewDetails, onDeleteUser }) => {
  const isAdmin = user.role === 'admin';
  
  return (
    <div className={`user-card ${isAdmin ? 'admin-user' : 'regular-user'}`}>
      <div className="user-card-header">
        <div className="user-avatar">
          {user.name ? user.name.charAt(0).toUpperCase() : user.username.charAt(0).toUpperCase()}
        </div>
        <div className="user-info">
          <h3>{user.name || user.username}</h3>
          <p className="username">@{user.username}</p>
          <p className="email">{user.email}</p>
        </div>
        <div className={`role-badge ${user.role}`}>
          {user.role === 'admin' ? '👑' : '👤'} {user.role}
        </div>
      </div>

      <div className="user-card-details">
        <div className="detail-row">
          <span className="label">User ID:</span>
          <span className="value">#{user.id}</span>
        </div>
        <div className="detail-row">
          <span className="label">Created:</span>
          <span className="value">{formatDate(user.created_at)}</span>
        </div>
        <div className="detail-row">
          <span className="label">Last Updated:</span>
          <span className="value">{formatDate(user.updated_at)}</span>
        </div>
      </div>

      <div className="user-card-actions">
        <button
          onClick={() => onViewDetails(user)}
          className="btn-secondary"
        >
          📊 View Details
        </button>
        {!isAdmin && (
          <button
            onClick={() => onDeleteUser(user)}
            className="btn-danger"
          >
            🗑️ Delete
          </button>
        )}
      </div>
    </div>
  );
};

export default UserCard;