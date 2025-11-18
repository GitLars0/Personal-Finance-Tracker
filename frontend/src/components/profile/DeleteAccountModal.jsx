import React from 'react';
import Modal from '../Modal';

function DeleteAccountModal({ isOpen, onClose, onConfirm }) {
  if (!isOpen) return null;

  return (
    <Modal 
      title="Delete Account" 
      onClose={onClose}
    >
      <div className="delete-confirmation">
        <div className="warning-icon">⚠️</div>
        <h3>Are you sure you want to delete your account?</h3>
        <p>
          This action cannot be undone. This will permanently delete your account,
          including all your transactions, budgets, categories, and accounts.
        </p>
        <div className="form-actions">
          <button 
            type="button" 
            onClick={onClose}
          >
            Cancel
          </button>
          <button 
            type="button" 
            className="btn-danger"
            onClick={onConfirm}
          >
            Yes, Delete My Account
          </button>
        </div>
      </div>
    </Modal>
  );
}

export default DeleteAccountModal;