import React from 'react';

function AccountCard({ 
  account, 
  formatCurrency, 
  getAccountTypeLabel, 
  handleEdit, 
  handleDelete 
}) {
  return (
    <div className="account-card">
      <div className="account-card-header">
        <div className="account-info">
          <h3>{account.name}</h3>
          <span className="account-type">
            {getAccountTypeLabel(account.account_type)}
          </span>
        </div>
        <div className="account-actions">
          <button 
            onClick={() => handleEdit(account)}
            className="edit-btn"
            title="Edit account"
          >
            ✏️
          </button>
          <button 
            onClick={() => handleDelete(account.id)}
            className="delete-btn"
            title="Delete account"
          >
            🗑️
          </button>
        </div>
      </div>

      <div className="account-balance">
        <span className="balance-label">Balance:</span>
        <span className={`balance-amount ${(account.current_balance_cents || account.initial_balance_cents || 0) >= 0 ? 'positive' : 'negative'}`}>
          {formatCurrency(account.current_balance_cents || account.initial_balance_cents || 0)}
        </span>
      </div>

      {account.description && (
        <div className="account-description">
          {account.description}
        </div>
      )}

      <div className="account-stats">
        <div className="stat-item">
          <span>Created:</span>
          <span>{new Date(account.created_at).toLocaleDateString()}</span>
        </div>
        {account.initial_balance_cents !== (account.current_balance_cents || 0) && (
          <div className="stat-item">
            <span>Initial Balance:</span>
            <span>{formatCurrency(account.initial_balance_cents || 0)}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default AccountCard;