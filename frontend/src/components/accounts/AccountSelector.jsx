import React, { useState } from 'react';
import InlineAccountForm from './InlineAccountForm';

function AccountSelector({ 
  value, 
  onChange, 
  accounts, 
  onAccountCreated, 
  required = false, 
  label = "Account"
}) {
  const [showCreateForm, setShowCreateForm] = useState(false);

  const handleAccountCreated = (newAccount) => {
    console.log('AccountSelector: handling new account:', newAccount);
    onAccountCreated(newAccount);
    onChange(newAccount.id.toString()); // Auto-select the new account
    setShowCreateForm(false);
  };

  if (showCreateForm) {
    console.log('AccountSelector: showing create form');
    return (
      <div className="form-group">
        <label>{label} {required && '*'}</label>
        <InlineAccountForm 
          onAccountCreated={handleAccountCreated}
          onCancel={() => setShowCreateForm(false)}
        />
      </div>
    );
  }

  return (
    <div className="form-group">
      <label>{label} {required && '*'}</label>
      <div className="account-selector-wrapper">
        <select
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={required}
        >
          <option value="">Select Account</option>
          {accounts.map(account => (
            <option key={account.id} value={account.id}>
              {account.name} ({account.account_type})
            </option>
          ))}
        </select>
        
        <button 
          type="button" 
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            console.log('+ New Account button clicked!');
            setShowCreateForm(true);
          }}
          className="create-account-btn"
          title="Create new account"
        >
          + New
        </button>
      </div>
      
      {accounts.length === 0 && (
        <small className="form-help">
          No accounts found. <button type="button" onClick={() => setShowCreateForm(true)} className="link-btn">Create your first account</button>
        </small>
      )}
    </div>
  );
}

export default AccountSelector;