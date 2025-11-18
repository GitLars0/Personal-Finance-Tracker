import React, { useState, useEffect, useCallback } from 'react';
import '../styles/Profile.css';
import useAuthFetch from '../hooks/useAuthFetch';
import { ProfileCard, EditProfileModal, ChangePasswordModal, DeleteAccountModal } from '../components/profile';

function Profile() {
  const authFetch = useAuthFetch();
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showEditForm, setShowEditForm] = useState(false);
  const [showPasswordForm, setShowPasswordForm] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState('');
  
  const [editFormData, setEditFormData] = useState({
    name: '',
    email: '',
    username: ''
  });
  
  const [passwordFormData, setPasswordFormData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  });

  const fetchUserProfile = useCallback(async () => {
    try {
      setLoading(true);
      const userData = await authFetch('/api/user/profile');
      setUser(userData);
      setEditFormData({
        name: userData.name || '',
        email: userData.email || '',
        username: userData.username || ''
      });
      setError(null);
    } catch (err) {
      setError(err.message || 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  }, [authFetch]);

  useEffect(() => {
    fetchUserProfile();
  }, [fetchUserProfile]);

  const handleEditSubmit = async () => {
    try {
      const updatedUser = await authFetch('/api/user/profile', {
        method: 'PUT',
        body: JSON.stringify(editFormData)
      });
      setUser(updatedUser);
      setShowEditForm(false);
      setError(null);
      
      // Update localStorage user data
      const currentUser = JSON.parse(localStorage.getItem('user') || '{}');
      localStorage.setItem('user', JSON.stringify({ ...currentUser, ...updatedUser }));
    } catch (err) {
      setError(err.message || 'Failed to update profile');
    }
  };

  const handlePasswordSubmit = async () => {
    if (passwordFormData.newPassword !== passwordFormData.confirmPassword) {
      setError('New passwords do not match');
      return;
    }
    
    if (passwordFormData.newPassword.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }
    
    try {
      await authFetch('/api/user/change-password', {
        method: 'PUT',
        body: JSON.stringify({
          current_password: passwordFormData.currentPassword,
          new_password: passwordFormData.newPassword
        })
      });
      
      // Close the password modal and show an inline success message
      setShowPasswordForm(false);
      setPasswordFormData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      });
      setError(null);
      setPasswordSuccess('Password changed successfully.');
    } catch (err) {
      setError(err.message || 'Failed to change password');
    }
  };

  const handleDeleteAccount = async () => {
    try {
      await authFetch('/api/user/account', {
        method: 'DELETE'
      });
      
      // Clear local storage and redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.dispatchEvent(new Event('authChange'));
      window.location.href = '/';
    } catch (err) {
      setError(err.message || 'Failed to delete account');
      setShowDeleteConfirm(false);
    }
  };

  const handleEditProfileClick = () => {
    setError(null);
    setShowEditForm(true);
  };

  const handleChangePasswordClick = () => {
    setPasswordSuccess('');
    setError(null);
    setShowPasswordForm(true);
  };

  const handleDeleteAccountClick = () => {
    setError(null);
    setShowDeleteConfirm(true);
  };

  const handleCloseEditForm = () => {
    setShowEditForm(false);
    setError(null);
  };

  const handleClosePasswordForm = () => {
    setShowPasswordForm(false);
    setError(null);
  };

  const handleCloseDeleteConfirm = () => {
    setShowDeleteConfirm(false);
    setError(null);
  };

  if (loading) {
    return (
      <div className="profile-container">
        <div className="loading">Loading profile...</div>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="profile-container">
        <div className="error-message">{error}</div>
        <button onClick={fetchUserProfile}>Retry</button>
      </div>
    );
  }

  return (
    <div className="profile-container">
      <header className="profile-header">
        <h1>Profile Settings</h1>
      </header>

      <div className="profile-content">
        <ProfileCard 
          user={user}
          passwordSuccess={passwordSuccess}
          onEditProfile={handleEditProfileClick}
          onChangePassword={handleChangePasswordClick}
          onDeleteAccount={handleDeleteAccountClick}
        />
      </div>

      <EditProfileModal 
        isOpen={showEditForm}
        onClose={handleCloseEditForm}
        editFormData={editFormData}
        setEditFormData={setEditFormData}
        onSubmit={handleEditSubmit}
        error={error}
      />

      <ChangePasswordModal 
        isOpen={showPasswordForm}
        onClose={handleClosePasswordForm}
        passwordFormData={passwordFormData}
        setPasswordFormData={setPasswordFormData}
        onSubmit={handlePasswordSubmit}
        error={error}
      />

      <DeleteAccountModal 
        isOpen={showDeleteConfirm}
        onClose={handleCloseDeleteConfirm}
        onConfirm={handleDeleteAccount}
      />
    </div>
  );
}

export default Profile;