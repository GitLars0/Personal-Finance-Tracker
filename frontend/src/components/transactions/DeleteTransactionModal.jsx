import React from 'react';
import Modal from '../Modal';

function DeleteTransactionModal({ transaction, onConfirm, onCancel, formatCurrency, formatDate }) {
  if (!transaction) return null;

  return (
    <Modal title="Delete Transaction" onClose={onCancel}>
      <div className="delete-modal-content">
        <div className="warning-icon">⚠️</div>
        <p>Are you sure you want to delete this transaction?</p>
        
        <div className="transaction-details">
          <div className="detail-row">
            <strong>Amount:</strong> 
            <span className={transaction.amount_cents < 0 ? 'expense' : 'income'}>
              {formatCurrency(Math.abs(transaction.amount_cents))}
            </span>
          </div>
          <div className="detail-row">
            <strong>Description:</strong> 
            <span>{transaction.description || 'No description'}</span>
          </div>
          <div className="detail-row">
            <strong>Date:</strong> 
            <span>{formatDate(transaction.txn_date)}</span>
          </div>
          {transaction.category?.name && (
            <div className="detail-row">
              <strong>Category:</strong> 
              <span>{transaction.category.name}</span>
            </div>
          )}
          {transaction.account?.name && (
            <div className="detail-row">
              <strong>Account:</strong> 
              <span>{transaction.account.name}</span>
            </div>
          )}
        </div>

        <div className="modal-actions">
          <button type="button" className="cancel-btn" onClick={onCancel}>
            Cancel
          </button>
          <button type="button" className="delete-btn" onClick={onConfirm}>
            Delete Transaction
          </button>
        </div>
      </div>
    </Modal>
  );
}

export default DeleteTransactionModal;