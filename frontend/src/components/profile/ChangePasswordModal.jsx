import React from 'react';
import Modal from '../Modal';

function ChangePasswordModal({ 
  isOpen, 
  onClose, 
  passwordFormData, 
  setPasswordFormData, 
  onSubmit, 
  error 
}) {
  if (!isOpen) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit();
  };

  const handleClose = () => {
    setPasswordFormData({
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
    onClose();
  };

  return (
    <Modal 
      title="Change Password" 
      onClose={handleClose}
    >
      <form onSubmit={handleSubmit} className="profile-form">
        <div className="form-group">
          <label htmlFor="currentPassword">Current Password</label>
          <input
            id="currentPassword"
            type="password"
            value={passwordFormData.currentPassword}
            onChange={(e) => setPasswordFormData(prev => ({ ...prev, currentPassword: e.target.value }))}
            placeholder="Enter current password"
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="newPassword">New Password</label>
          <input
            id="newPassword"
            type="password"
            value={passwordFormData.newPassword}
            onChange={(e) => setPasswordFormData(prev => ({ ...prev, newPassword: e.target.value }))}
            placeholder="Enter new password (min 6 characters)"
            required
            minLength="6"
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="confirmPassword">Confirm New Password</label>
          <input
            id="confirmPassword"
            type="password"
            value={passwordFormData.confirmPassword}
            onChange={(e) => setPasswordFormData(prev => ({ ...prev, confirmPassword: e.target.value }))}
            placeholder="Confirm new password"
            required
            minLength="6"
          />
        </div>
        
        {error && <div className="error-message">{error}</div>}
        
        <div className="form-actions">
          <button type="button" onClick={handleClose}>
            Cancel
          </button>
          <button type="submit" className="btn-primary">
            Change Password
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default ChangePasswordModal;