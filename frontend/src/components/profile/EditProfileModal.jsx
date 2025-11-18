import React from 'react';
import Modal from '../Modal';

function EditProfileModal({ 
  isOpen, 
  onClose, 
  editFormData, 
  setEditFormData, 
  onSubmit, 
  error 
}) {
  if (!isOpen) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit();
  };

  return (
    <Modal 
      title="Edit Profile" 
      onClose={onClose}
    >
      <form onSubmit={handleSubmit} className="profile-form">
        <div className="form-group">
          <label htmlFor="name">Display Name</label>
          <input
            id="name"
            type="text"
            value={editFormData.name}
            onChange={(e) => setEditFormData(prev => ({ ...prev, name: e.target.value }))}
            placeholder="Enter your display name"
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="username">Username</label>
          <input
            id="username"
            type="text"
            value={editFormData.username}
            onChange={(e) => setEditFormData(prev => ({ ...prev, username: e.target.value }))}
            placeholder="Enter your username"
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={editFormData.email}
            onChange={(e) => setEditFormData(prev => ({ ...prev, email: e.target.value }))}
            placeholder="Enter your email"
            required
          />
        </div>
        
        {error && <div className="error-message">{error}</div>}
        
        <div className="form-actions">
          <button type="button" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="btn-primary">
            Save Changes
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default EditProfileModal;