import React, { useState } from 'react';
import useAuthFetch from '../../hooks/useAuthFetch';
import '../../styles/InlineAccount.css';

function InlineAccountForm({ onAccountCreated, onCancel }) {
  const authFetch = useAuthFetch();
  const [formData, setFormData] = useState({
    name: '',
    account_type: 'checking',
    initial_balance_cents: '',
    description: ''
  });
  const [error, setError] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const accountTypes = [
    { value: 'checking', label: 'Checking' },
    { value: 'savings', label: 'Savings' },
    { value: 'credit', label: 'Credit Card' },
    { value: 'cash', label: 'Cash' },
    { value: 'investment', label: 'Investment' },
    { value: 'other', label: 'Other' }
  ];

  const handleSubmit = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (!formData.name.trim()) {
      setError('Account name is required');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      console.log('InlineAccountForm: creating account with data:', formData);

      const newAccount = await authFetch('/api/accounts', {
        method: 'POST',
        body: JSON.stringify({
          name: formData.name.trim(),
          account_type: formData.account_type,
          initial_balance_cents: parseFloat(formData.initial_balance_cents || '0') * 100,
          description: formData.description.trim() || ''
        })
      });
      
      console.log('InlineAccountForm: created account:', newAccount);
      onAccountCreated(newAccount);
      
      // Reset form
      setFormData({
        name: '',
        account_type: 'checking',
        initial_balance_cents: '',
        description: ''
      });
    } catch (err) {
      console.error('InlineAccountForm: error creating account:', err);
      if (err && err.status === 409) {
        setError('An account with this name already exists. Please choose a different name.');
      } else if (err && err.message) {
        setError(err.message);
      } else {
        setError('Failed to create account');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = (e) => {
    e.preventDefault();
    e.stopPropagation();
    onCancel();
  };

  return (
    <div className="inline-account-form">
      <div className="form-header">
        <h4>Create New Account</h4>
      </div>
      
      {error && <div className="error-message">{error}</div>}
      
      <div className="form-grid">
        <div className="form-group">
          <label>Account Name *</label>
          <input
            type="text"
            placeholder="e.g. Main Checking, Savings Account"
            value={formData.name}
            onChange={(e) => setFormData({...formData, name: e.target.value})}
            disabled={isSubmitting}
            autoFocus
          />
        </div>

        <div className="form-group">
          <label>Account Type *</label>
          <select
            value={formData.account_type}
            onChange={(e) => setFormData({...formData, account_type: e.target.value})}
            disabled={isSubmitting}
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
            disabled={isSubmitting}
          />
          <small className="form-help">The starting balance for this account</small>
        </div>

        <div className="form-group">
          <label>Description</label>
          <input
            type="text"
            placeholder="Optional description"
            value={formData.description}
            onChange={(e) => setFormData({...formData, description: e.target.value})}
            disabled={isSubmitting}
          />
        </div>
      </div>

      <div className="form-actions">
        <button 
          type="button" 
          onClick={handleCancel}
          disabled={isSubmitting}
          className="cancel-btn"
        >
          Cancel
        </button>
        <button 
          type="button"
          onClick={handleSubmit}
          disabled={isSubmitting || !formData.name.trim()}
          className="submit-btn"
        >
          {isSubmitting ? 'Creating...' : 'Create Account'}
        </button>
      </div>
    </div>
  );
}

export default InlineAccountForm;