import React from 'react';

function ProfileCard({ user, passwordSuccess, onEditProfile, onChangePassword, onDeleteAccount }) {
  return (
    <div className="profile-card">
      {passwordSuccess && (
        <div className="success-message">
          {passwordSuccess}
        </div>
      )}
      
      <div className="profile-info">
        <div className="avatar">
          <span className="avatar-initial">
            {(user?.name || user?.username || 'U').charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="user-details">
          <h2>{user?.name || 'No name set'}</h2>
          <p className="username">@{user?.username}</p>
          <p className="email">{user?.email}</p>
          <p className="member-since">
            Member since {new Date(user?.created_at).toLocaleDateString()}
          </p>
        </div>
      </div>

      <div className="profile-actions">
        <button 
          className="btn btn-primary"
          onClick={onEditProfile}
        >
          Edit Profile
        </button>
        <button 
          className="btn btn-secondary"
          onClick={onChangePassword}
        >
          Change Password
        </button>
        <button 
          className="btn btn-danger"
          onClick={onDeleteAccount}
        >
          Delete Account
        </button>
      </div>
    </div>
  );
}

export default ProfileCard;