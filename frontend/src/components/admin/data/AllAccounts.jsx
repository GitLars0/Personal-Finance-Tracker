import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import { formatDate, formatAmount } from '../../../utils/format';
import DeleteConfirmationModal from '../DeleteConfirmationModal';

const AllAccounts = () => {
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');
  const [sortBy, setSortBy] = useState('balance_desc');

  // Delete modal state
  const [deleteModal, setDeleteModal] = useState({
    isOpen: false,
    account: null,
    isDeleting: false
  });

  useEffect(() => {
    loadAccounts();
  }, []);

  const loadAccounts = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getAllAccounts();
      setAccounts(data.accounts);
    } catch (err) {
      setError('Failed to load accounts: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteAccount = async (account) => {
    setDeleteModal({
      isOpen: true,
      account: account,
      isDeleting: false
    });
  };

  const confirmDeleteAccount = async () => {
    if (!deleteModal.account) return;

    try {
      setDeleteModal(prev => ({ ...prev, isDeleting: true }));
      await adminApi.deleteAccount(deleteModal.account.id);
      setAccounts(accounts.filter(a => a.id !== deleteModal.account.id));
      setDeleteModal({ isOpen: false, account: null, isDeleting: false });
    } catch (err) {
      setError('Failed to delete account: ' + err.message);
      setDeleteModal(prev => ({ ...prev, isDeleting: false }));
    }
  };

  // Filter and sort accounts
  const filteredAccounts = accounts
    .filter(account => {
      const matchesSearch = account.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           account.user_email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           account.user_username.toLowerCase().includes(searchTerm.toLowerCase());
      
      const matchesType = typeFilter === 'all' || account.account_type === typeFilter;
      
      return matchesSearch && matchesType;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'balance_desc':
          return b.balance - a.balance;
        case 'balance_asc':
          return a.balance - b.balance;
        case 'name':
          return a.name.localeCompare(b.name);
        case 'user':
          return a.user_username.localeCompare(b.user_username);
        case 'type':
          return a.account_type.localeCompare(b.account_type);
        default:
          return 0;
      }
    });

  if (loading) {
    return (
      <div className="admin-data-container">
        <div className="loading">Loading all accounts...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="admin-data-container">
        <div className="error">{error}</div>
        <button onClick={loadAccounts} className="retry-btn">Retry</button>
      </div>
    );
  }

  const totalBalance = accounts.reduce((sum, account) => sum + account.balance, 0);

  return (
    <div className="admin-data-container">
      <div className="admin-data-header">
        <h2>All Accounts</h2>
        <p>Monitor and manage financial accounts across all users</p>
      </div>

      {/* Controls */}
      <div className="admin-data-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search by account name, user email, or username..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-container">
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="type-filter"
          >
            <option value="all">All Types</option>
            <option value="checking">Checking</option>
            <option value="savings">Savings</option>
            <option value="credit">Credit</option>
            <option value="investment">Investment</option>
            <option value="cash">Cash</option>
            <option value="other">Other</option>
          </select>
        </div>

        <div className="sort-container">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="sort-select"
          >
            <option value="balance_desc">Highest Balance</option>
            <option value="balance_asc">Lowest Balance</option>
            <option value="name">By Name</option>
            <option value="user">By User</option>
            <option value="type">By Type</option>
          </select>
        </div>

        <button onClick={loadAccounts} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>

      {/* Statistics */}
      <div className="data-stats">
        <div className="stat-item">
          <span className="stat-value">{accounts.length}</span>
          <span className="stat-label">Total Accounts</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{formatAmount(totalBalance)}</span>
          <span className="stat-label">Total Balance</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{accounts.filter(a => a.balance > 0).length}</span>
          <span className="stat-label">Positive Balance</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{filteredAccounts.length}</span>
          <span className="stat-label">Filtered Results</span>
        </div>
      </div>

      {/* Accounts Table */}
      <div className="admin-data-table">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Account Name</th>
              <th>Type</th>
              <th>Balance</th>
              <th>Currency</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredAccounts.length === 0 ? (
              <tr>
                <td colSpan="8" className="no-data">
                  No accounts found matching your criteria.
                </td>
              </tr>
            ) : (
              filteredAccounts.map((account) => (
                <tr key={account.id}>
                  <td>#{account.id}</td>
                  <td>
                    <div className="user-cell">
                      <div className="username">@{account.user_username}</div>
                      <div className="email">{account.user_email}</div>
                    </div>
                  </td>
                  <td className="account-name-cell">
                    <strong>{account.name}</strong>
                  </td>
                  <td>
                    <span className={`type-badge ${account.account_type}`}>
                      {account.account_type}
                    </span>
                  </td>
                  <td className={`balance-cell ${account.balance >= 0 ? 'positive' : 'negative'}`}>
                    {formatAmount(account.balance)}
                  </td>
                  <td>{account.currency || 'USD'}</td>
                  <td>{formatDate(account.created_at)}</td>
                  <td>
                    <button
                      onClick={() => handleDeleteAccount(account)}
                      className="btn-danger-small"
                      title="Delete account"
                    >
                      🗑️
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Delete Confirmation Modal */}
      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, account: null, isDeleting: false })}
        onConfirm={confirmDeleteAccount}
        title="Delete Account"
        message="Are you sure you want to delete this account? This will also delete all associated transactions."
        itemName={deleteModal.account?.name}
        isDeleting={deleteModal.isDeleting}
      />
    </div>
  );
};

export default AllAccounts;