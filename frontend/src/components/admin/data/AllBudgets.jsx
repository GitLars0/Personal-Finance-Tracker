import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import { formatDate, formatAmount } from '../../../utils/format';
import DeleteConfirmationModal from '../DeleteConfirmationModal';
import BudgetDetailsModal from './BudgetDetailsModal';

const AllBudgets = () => {
  const [budgets, setBudgets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [sortBy, setSortBy] = useState('amount_desc');
  
  // Delete modal state
  const [deleteModal, setDeleteModal] = useState({
    isOpen: false,
    budget: null,
    isDeleting: false
  });

  // Budget details modal state
  const [detailsModal, setDetailsModal] = useState({
    isOpen: false,
    budget: null
  });

  useEffect(() => {
    loadBudgets();
  }, []);

  const loadBudgets = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getAllBudgets();
      setBudgets(data.budgets);
    } catch (err) {
      setError('Failed to load budgets: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteBudget = async (budget) => {
    setDeleteModal({
      isOpen: true,
      budget: budget,
      isDeleting: false
    });
  };

  const confirmDeleteBudget = async () => {
    if (!deleteModal.budget) return;

    try {
      setDeleteModal(prev => ({ ...prev, isDeleting: true }));
      await adminApi.deleteBudget(deleteModal.budget.id);
      setBudgets(budgets.filter(b => b.id !== deleteModal.budget.id));
      setDeleteModal({ isOpen: false, budget: null, isDeleting: false });
    } catch (err) {
      setError('Failed to delete budget: ' + err.message);
      setDeleteModal(prev => ({ ...prev, isDeleting: false }));
    }
  };

  const handleViewBudget = (budget) => {
    setDetailsModal({
      isOpen: true,
      budget: budget
    });
  };

  const getBudgetStatus = (budget) => {
    const now = new Date();
    const startDate = new Date(budget.start_date);
    const endDate = new Date(budget.end_date);
    
    if (now < startDate) return 'upcoming';
    if (now > endDate) return 'expired';
    return 'active';
  };

  // Filter and sort budgets
  const filteredBudgets = budgets
    .filter(budget => {
      const matchesSearch = budget.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           budget.user_email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           budget.user_username.toLowerCase().includes(searchTerm.toLowerCase());
      
      const budgetStatus = getBudgetStatus(budget);
      const matchesStatus = statusFilter === 'all' || budgetStatus === statusFilter;
      
      return matchesSearch && matchesStatus;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'amount_desc':
          return b.amount - a.amount;
        case 'amount_asc':
          return a.amount - b.amount;
        case 'name':
          return a.name.localeCompare(b.name);
        case 'user':
          return a.user_username.localeCompare(b.user_username);
        case 'start_date':
          return new Date(b.start_date) - new Date(a.start_date);
        case 'end_date':
          return new Date(b.end_date) - new Date(a.end_date);
        default:
          return 0;
      }
    });

  if (loading) {
    return (
      <div className="admin-data-container">
        <div className="loading">Loading all budgets...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="admin-data-container">
        <div className="error">{error}</div>
        <button onClick={loadBudgets} className="retry-btn">Retry</button>
      </div>
    );
  }

  const totalBudgetAmount = budgets.reduce((sum, budget) => sum + budget.amount, 0);
  const activeBudgets = budgets.filter(b => getBudgetStatus(b) === 'active');

  return (
    <div className="admin-data-container">
      <div className="admin-data-header">
        <h2>All Budgets</h2>
        <p>Monitor and manage budget plans across all users</p>
      </div>

      {/* Controls */}
      <div className="admin-data-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search by budget name, user email, or username..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-container">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="status-filter"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="upcoming">Upcoming</option>
            <option value="expired">Expired</option>
          </select>
        </div>

        <div className="sort-container">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="sort-select"
          >
            <option value="amount_desc">Highest Amount</option>
            <option value="amount_asc">Lowest Amount</option>
            <option value="name">By Name</option>
            <option value="user">By User</option>
            <option value="start_date">By Start Date</option>
            <option value="end_date">By End Date</option>
          </select>
        </div>

        <button onClick={loadBudgets} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>

      {/* Statistics */}
      <div className="data-stats">
        <div className="stat-item">
          <span className="stat-value">{budgets.length}</span>
          <span className="stat-label">Total Budgets</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{formatAmount(totalBudgetAmount)}</span>
          <span className="stat-label">Total Budget Amount</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{activeBudgets.length}</span>
          <span className="stat-label">Active Budgets</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{filteredBudgets.length}</span>
          <span className="stat-label">Filtered Results</span>
        </div>
      </div>

      {/* Budgets Table */}
      <div className="admin-data-table">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Budget Name</th>
              <th>Amount</th>
              <th>Period</th>
              <th>Status</th>
              <th>Progress</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredBudgets.length === 0 ? (
              <tr>
                <td colSpan="9" className="no-data">
                  No budgets found matching your criteria.
                </td>
              </tr>
            ) : (
              filteredBudgets.map((budget) => {
                const status = getBudgetStatus(budget);
                const progressPercentage = budget.amount > 0 ? (budget.spent / budget.amount) * 100 : 0;
                
                return (
                  <tr key={budget.id}>
                    <td>#{budget.id}</td>
                    <td>
                      <div className="user-cell">
                        <div className="username">@{budget.user_username}</div>
                        <div className="email">{budget.user_email}</div>
                      </div>
                    </td>
                    <td className="budget-name-cell">
                      <strong>{budget.name}</strong>
                    </td>
                    <td className="amount-cell">
                      {formatAmount(budget.amount)}
                    </td>
                    <td className="period-cell">
                      <div className="date-range">
                        <div>{formatDate(budget.start_date)}</div>
                        <div>to {formatDate(budget.end_date)}</div>
                      </div>
                    </td>
                    <td>
                      <span className={`status-badge ${status}`}>
                        {status === 'active' && '🟢'}
                        {status === 'upcoming' && '🔵'}
                        {status === 'expired' && '🔴'}
                        {status}
                      </span>
                    </td>
                    <td className="progress-cell">
                      <div className="progress-container">
                        <div className="progress-bar">
                          <div 
                            className={`progress-fill ${progressPercentage > 100 ? 'over-budget' : ''}`}
                            style={{ width: `${Math.min(progressPercentage, 100)}%` }}
                          ></div>
                        </div>
                        <span className="progress-text">
                          {formatAmount(budget.spent || 0)} / {formatAmount(budget.amount)}
                        </span>
                        <span className={`progress-percentage ${progressPercentage > 100 ? 'over-budget' : ''}`}>
                          {progressPercentage.toFixed(1)}%
                        </span>
                      </div>
                    </td>
                    <td>{formatDate(budget.created_at)}</td>
                    <td>
                      <div className="action-buttons">
                        <button
                          onClick={() => handleViewBudget(budget)}
                          className="btn-view-small"
                          title="View budget details"
                        >
                          👁️
                        </button>
                        <button
                          onClick={() => handleDeleteBudget(budget)}
                          className="btn-danger-small"
                          title="Delete budget"
                        >
                          🗑️
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Delete Confirmation Modal */}
      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, budget: null, isDeleting: false })}
        onConfirm={confirmDeleteBudget}
        title="Delete Budget"
        message="Are you sure you want to delete this budget? This will also delete all associated budget items."
        itemName={deleteModal.budget?.name}
        isDeleting={deleteModal.isDeleting}
      />

      {/* Budget Details Modal */}
      {detailsModal.isOpen && detailsModal.budget && (
        <BudgetDetailsModal
          budget={detailsModal.budget}
          onClose={() => setDetailsModal({ isOpen: false, budget: null })}
        />
      )}
    </div>
  );
};

export default AllBudgets;