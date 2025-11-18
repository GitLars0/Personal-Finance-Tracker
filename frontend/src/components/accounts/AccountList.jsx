import React from 'react';
import AccountCard from './AccountCard';

function AccountList({ 
  accounts, 
  formatCurrency, 
  getAccountTypeLabel, 
  handleEdit, 
  handleDelete,
  resetForm,
  setShowForm 
}) {
  if (accounts.length === 0) {
    return (
      <div className="empty-state">
        <h3>No accounts found</h3>
        <p>Add your first account to start tracking transactions.</p>
        <button 
          className="action-button"
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
        >
          Add Your First Account
        </button>
        
        <div className="getting-started">
          <h4>Getting Started</h4>
          <p>Accounts represent where your money is stored or owed. Common examples:</p>
          <ul>
            <li><strong>Checking Account:</strong> Your main bank account for daily expenses</li>
            <li><strong>Savings Account:</strong> Long-term savings and emergency funds</li>
            <li><strong>Credit Card:</strong> Credit cards for tracking debt and payments</li>
            <li><strong>Cash:</strong> Physical cash you carry</li>
          </ul>
        </div>
      </div>
    );
  }

  return (
    <div className="accounts-grid">
      {accounts.map(account => (
        <AccountCard
          key={account.id}
          account={account}
          formatCurrency={formatCurrency}
          getAccountTypeLabel={getAccountTypeLabel}
          handleEdit={handleEdit}
          handleDelete={handleDelete}
        />
      ))}
    </div>
  );
}

export default AccountList;