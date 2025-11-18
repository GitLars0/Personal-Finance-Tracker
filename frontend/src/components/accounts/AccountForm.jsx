import React from 'react';

function AccountForm({ 
  showForm, 
  editingAccount, 
  formData, 
  setFormData, 
  handleFormSubmit, 
  resetForm, 
  setShowForm, 
  setEditingAccount,
  accountTypes 
}) {
  if (!showForm) return null;

  return (
    <div className="form-overlay">
      <div className="account-form">
        <h2>{editingAccount ? 'Edit Account' : 'Add New Account'}</h2>
        
        <form onSubmit={handleFormSubmit}>
          <div className="form-group">
            <label>Account Name *</label>
            <input
              type="text"
              placeholder="e.g., Main Checking, Credit Card, Savings"
              value={formData.name}
              onChange={(e) => setFormData({...formData, name: e.target.value})}
              required
            />
          </div>

          <div className="form-group">
            <label>Account Type *</label>
            <select
              value={formData.account_type}
              onChange={(e) => setFormData({...formData, account_type: e.target.value})}
              required
            >
              {accountTypes.map(type => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Initial Balance (USD)</label>
            <input
              type="number"
              step="0.01"
              placeholder="0.00"
              value={formData.initial_balance_cents}
              onChange={(e) => setFormData({...formData, initial_balance_cents: e.target.value})}
            />
            <small className="form-help">
              The starting balance for this account. Leave empty for $0.00
            </small>
          </div>

          <div className="form-group">
            <label>Description</label>
            <textarea
              placeholder="Optional description for this account"
              value={formData.description}
              onChange={(e) => setFormData({...formData, description: e.target.value})}
              rows={3}
            />
          </div>

          <div className="form-actions">
            <button type="button" onClick={() => {
              setShowForm(false);
              setEditingAccount(null);
              resetForm();
            }}>
              Cancel
            </button>
            <button type="submit" className="primary">
              {editingAccount ? 'Update Account' : 'Add Account'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default AccountForm;