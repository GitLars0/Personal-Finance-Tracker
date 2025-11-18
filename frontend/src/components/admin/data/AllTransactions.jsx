import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import { formatDate, formatAmount } from '../../../utils/format';
import DeleteConfirmationModal from '../DeleteConfirmationModal';

const AllTransactions = () => {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');
  const [sortBy, setSortBy] = useState('date_desc');

  // Delete modal state
  const [deleteModal, setDeleteModal] = useState({
    isOpen: false,
    transaction: null,
    isDeleting: false
  });

  useEffect(() => {
    loadTransactions();
  }, []);

  const loadTransactions = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getAllTransactions();
      setTransactions(data.transactions);
    } catch (err) {
      setError('Failed to load transactions: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteTransaction = async (transaction) => {
    setDeleteModal({
      isOpen: true,
      transaction: transaction,
      isDeleting: false
    });
  };

  const confirmDeleteTransaction = async () => {
    if (!deleteModal.transaction) return;

    try {
      setDeleteModal(prev => ({ ...prev, isDeleting: true }));
      await adminApi.deleteTransaction(deleteModal.transaction.id);
      setTransactions(transactions.filter(t => t.id !== deleteModal.transaction.id));
      setDeleteModal({ isOpen: false, transaction: null, isDeleting: false });
    } catch (err) {
      setError('Failed to delete transaction: ' + err.message);
      setDeleteModal(prev => ({ ...prev, isDeleting: false }));
    }
  };

  // Filter and sort transactions
  const filteredTransactions = transactions
    .filter(transaction => {
      const matchesSearch = transaction.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           transaction.user_email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           transaction.user_username.toLowerCase().includes(searchTerm.toLowerCase());
      
      const matchesType = typeFilter === 'all' || transaction.type === typeFilter;
      
      return matchesSearch && matchesType;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'date_desc':
          return new Date(b.created_at) - new Date(a.created_at);
        case 'date_asc':
          return new Date(a.created_at) - new Date(b.created_at);
        case 'amount_desc':
          return b.amount - a.amount;
        case 'amount_asc':
          return a.amount - b.amount;
        case 'user':
          return a.user_username.localeCompare(b.user_username);
        default:
          return 0;
      }
    });

  if (loading) {
    return (
      <div className="admin-data-container">
        <div className="loading">Loading all transactions...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="admin-data-container">
        <div className="error">{error}</div>
        <button onClick={loadTransactions} className="retry-btn">Retry</button>
      </div>
    );
  }

  return (
    <div className="admin-data-container">
      <div className="admin-data-header">
        <h2>All Transactions</h2>
        <p>Monitor and manage transactions across all users</p>
      </div>

      {/* Controls */}
      <div className="admin-data-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search by description, user email, or username..."
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
            <option value="income">Income</option>
            <option value="expense">Expense</option>
            <option value="transfer">Transfer</option>
          </select>
        </div>

        <div className="sort-container">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="sort-select"
          >
            <option value="date_desc">Newest First</option>
            <option value="date_asc">Oldest First</option>
            <option value="amount_desc">Highest Amount</option>
            <option value="amount_asc">Lowest Amount</option>
            <option value="user">By User</option>
          </select>
        </div>

        <button onClick={loadTransactions} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>

      {/* Statistics */}
      <div className="data-stats">
        <div className="stat-item">
          <span className="stat-value">{transactions.length}</span>
          <span className="stat-label">Total Transactions</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{transactions.filter(t => t.type === 'income').length}</span>
          <span className="stat-label">Income</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{transactions.filter(t => t.type === 'expense').length}</span>
          <span className="stat-label">Expense</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{filteredTransactions.length}</span>
          <span className="stat-label">Filtered Results</span>
        </div>
      </div>

      {/* Transactions Table */}
      <div className="admin-data-table">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Description</th>
              <th>Type</th>
              <th>Amount</th>
              <th>Account</th>
              <th>Category</th>
              <th>Date</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredTransactions.length === 0 ? (
              <tr>
                <td colSpan="9" className="no-data">
                  No transactions found matching your criteria.
                </td>
              </tr>
            ) : (
              filteredTransactions.map((transaction) => (
                <tr key={transaction.id}>
                  <td>#{transaction.id}</td>
                  <td>
                    <div className="user-cell">
                      <div className="username">@{transaction.user_username}</div>
                      <div className="email">{transaction.user_email}</div>
                    </div>
                  </td>
                  <td className="description-cell">{transaction.description}</td>
                  <td>
                    <span className={`type-badge ${transaction.type}`}>
                      {transaction.type}
                    </span>
                  </td>
                  <td className={`amount-cell ${transaction.type}`}>
                    {formatAmount(transaction.amount)}
                  </td>
                  <td>{transaction.account_name}</td>
                  <td>{transaction.category_name || 'Uncategorized'}</td>
                  <td>{formatDate(transaction.created_at)}</td>
                  <td>
                    <button
                      onClick={() => handleDeleteTransaction(transaction)}
                      className="btn-danger-small"
                      title="Delete transaction"
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
        onClose={() => setDeleteModal({ isOpen: false, transaction: null, isDeleting: false })}
        onConfirm={confirmDeleteTransaction}
        title="Delete Transaction"
        message="Are you sure you want to delete this transaction?"
        itemName={deleteModal.transaction?.description}
        isDeleting={deleteModal.isDeleting}
      />
    </div>
  );
};

export default AllTransactions;