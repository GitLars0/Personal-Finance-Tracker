import React, { useState } from 'react';
import adminApi from '../../../utils/adminApi';
import Modal from '../../Modal';

const DeleteUserModal = ({ user, onClose, onUserDeleted }) => {
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState('');
  const [confirmText, setConfirmText] = useState('');

  const isAdmin = user.role === 'admin';
  const confirmationText = user.username;

  const handleDelete = async () => {
    if (confirmText !== confirmationText) {
      setError('Please type the username exactly to confirm deletion.');
      return;
    }

    try {
      setDeleting(true);
      setError('');
      await adminApi.deleteUser(user.id);
      onUserDeleted();
    } catch (err) {
      setError('Failed to delete user: ' + err.message);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <Modal isOpen={true} onClose={onClose} title="Delete User Account">
      <div className="delete-user-modal">
        <div className="warning-section">
          <div className="warning-icon">⚠️</div>
          <h3>Permanently Delete User Account</h3>
          <p>You are about to permanently delete the account for:</p>
          
          <div className="user-summary">
            <div className="user-info">
              <strong>{user.name || user.username}</strong>
              <span className="username">@{user.username}</span>
              <span className="email">{user.email}</span>
              <span className={`role-badge ${user.role}`}>
                {isAdmin ? '👑' : '👤'} {user.role}
              </span>
            </div>
          </div>
        </div>

        {isAdmin && (
          <div className="admin-warning">
            <h4>🔐 Administrator Account</h4>
            <p>This is an administrator account with elevated privileges. Deleting this account will:</p>
            <ul>
              <li>Remove all administrative access</li>
              <li>Cannot be undone</li>
              <li>May affect system administration if this is the only admin</li>
            </ul>
          </div>
        )}

        <div className="danger-section">
          <h4>⚡ This action cannot be undone!</h4>
          <p>Deleting this user will permanently remove:</p>
          <ul>
            <li>✗ User account and profile information</li>
            <li>✗ All financial accounts and balances</li>
            <li>✗ All transaction history</li>
            <li>✗ All budget plans and categories</li>
            <li>✗ All associated data and settings</li>
          </ul>
        </div>

        <div className="confirmation-section">
          <label htmlFor="confirm-input">
            Type <strong>{confirmationText}</strong> to confirm deletion:
          </label>
          <input
            id="confirm-input"
            type="text"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            placeholder={`Type "${confirmationText}" here`}
            className="confirm-input"
            disabled={deleting}
          />
        </div>

        {error && (
          <div className="error-message">
            {error}
          </div>
        )}

        <div className="modal-actions">
          <button
            onClick={onClose}
            disabled={deleting}
            className="btn-secondary"
          >
            Cancel
          </button>
          <button
            onClick={handleDelete}
            disabled={confirmText !== confirmationText || deleting}
            className="btn-danger"
          >
            {deleting ? 'Deleting...' : '🗑️ Delete User Permanently'}
          </button>
        </div>

        {deleting && (
          <div className="deletion-progress">
            <p>🔄 Deleting user account and all associated data...</p>
            <p className="deletion-warning">Please do not close this dialog.</p>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default DeleteUserModal;