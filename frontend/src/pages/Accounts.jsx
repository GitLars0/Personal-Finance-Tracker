import React, { useEffect, useState, useCallback } from 'react';
import '../styles/Accounts.css';
import { AccountForm, AccountList, AccountSummary } from '../components/accounts';
import ConfirmModal from '../components/accounts/ConfirmModal';


function Accounts() {
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingAccount, setEditingAccount] = useState(null);
  const [deleteConfirm, setDeleteConfirm] = useState({ isOpen: false, accountId: null, accountName: '' });

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    account_type: 'checking',
    initial_balance_cents: '',
    description: ''
  });

  const token = localStorage.getItem('token');

  const accountTypes = [
    { value: 'checking', label: 'Checking Account' },
    { value: 'savings', label: 'Savings Account' },
    { value: 'credit', label: 'Credit Card' },
    { value: 'investment', label: 'Investment Account' },
    { value: 'cash', label: 'Cash' },
    { value: 'other', label: 'Other' }
  ];

  const fetchAccounts = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/accounts', {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok && response.status !== 404) {
        throw new Error('Failed to fetch accounts');
      }

      const accounts = response.ok ? await response.json() : [];
      setAccounts(accounts);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) {
      setError('Not authenticated');
      setLoading(false);
      return;
    }
    fetchAccounts();
  }, [token, fetchAccounts]);

  const handleFormSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.name || !formData.account_type) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      const submitData = {
        name: formData.name,
        account_type: formData.account_type,
        initial_balance_cents: formData.initial_balance_cents 
          ? Math.round(parseFloat(formData.initial_balance_cents) * 100) 
          : 0,
        description: formData.description || null
      };

      const url = editingAccount 
        ? `/api/accounts/${editingAccount.id}`
        : '/api/accounts';
      
      const method = editingAccount ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers,
        body: JSON.stringify(submitData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save account');
      }

      // Success - refresh data and close form
      await fetchAccounts();
      resetForm();
      setShowForm(false);
      setEditingAccount(null);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleEdit = (account) => {
    setEditingAccount(account);
    setFormData({
      name: account.name || '',
      account_type: account.account_type || 'checking',
      initial_balance_cents: account.initial_balance_cents 
        ? (account.initial_balance_cents / 100).toString() 
        : '',
      description: account.description || ''
    });
    setShowForm(true);
  };

  const handleDelete = async (id) => {
    const account = accounts.find(a => a.id === id);
    if (!account) return;
    
    // Show confirmation modal
    setDeleteConfirm({
      isOpen: true,
      accountId: id,
      accountName: account.name
    });
  };

  const confirmDelete = async () => {
    const { accountId } = deleteConfirm;
    
    try {
      const response = await fetch(`/api/accounts/${accountId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete account');
      }

      await fetchAccounts();
    } catch (err) {
      setError(err.message);
    } finally {
      setDeleteConfirm({ isOpen: false, accountId: null, accountName: '' });
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      account_type: 'checking',
      initial_balance_cents: '',
      description: ''
    });
    setEditingAccount(null);
    setError(null);
  };

  const formatCurrency = (cents) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(cents / 100);
  };

  const getAccountTypeLabel = (type) => {
    const accountType = accountTypes.find(t => t.value === type);
    return accountType ? accountType.label : type;
  };

  if (loading) {
    return (
      <div className="accounts-container">
        <div className="loading">Loading accounts...</div>
      </div>
    );
  }

  if (error && !showForm) {
    return (
      <div className="accounts-container">
        <div className="error-message">{error}</div>
        <button onClick={fetchAccounts}>Retry</button>
      </div>
    );
  }

  return (
    <div className="accounts-container">
      <header className="accounts-header">
        <h1>Accounts</h1>
        <button 
          className="add-account-btn"
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
        >
          ➕ Add Account
        </button>
      </header>

      {error && <div className="error-message">{error}</div>}

      <AccountForm 
        showForm={showForm}
        editingAccount={editingAccount}
        formData={formData}
        setFormData={setFormData}
        handleFormSubmit={handleFormSubmit}
        resetForm={resetForm}
        setShowForm={setShowForm}
        setEditingAccount={setEditingAccount}
        accountTypes={accountTypes}
      />

      <AccountList 
        accounts={accounts}
        formatCurrency={formatCurrency}
        getAccountTypeLabel={getAccountTypeLabel}
        handleEdit={handleEdit}
        handleDelete={handleDelete}
        resetForm={resetForm}
        setShowForm={setShowForm}
      />

      <AccountSummary 
        accounts={accounts}
        formatCurrency={formatCurrency}
      />

      <ConfirmModal
        isOpen={deleteConfirm.isOpen}
        onClose={() => setDeleteConfirm({ isOpen: false, accountId: null, accountName: '' })}
        onConfirm={confirmDelete}
        title="Delete Account"
        message={`Are you sure you want to delete "${deleteConfirm.accountName}"? This action cannot be undone and will fail if the account has existing transactions.`}
        confirmText="Delete Account"
        cancelText="Cancel"
        danger={true}
      />
    </div>
  );
}

export default Accounts;