import React, { useState, useEffect } from 'react';
import Modal from '../../Modal';
import { formatDate, formatAmount } from '../../../utils/format';
import adminApi from '../../../utils/adminApi';

const BudgetDetailsModal = ({ budget, onClose }) => {
  const [budgetDetails, setBudgetDetails] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!budget || !budget.id) {
      return;
    }

    const loadBudgetDetails = async () => {
      try {
        setLoading(true);
        const data = await adminApi.getBudgetDetails(budget.id);
        setBudgetDetails(data);
      } catch (err) {
        setError('Failed to load budget details: ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    loadBudgetDetails();
  }, [budget]);

  if (loading) {
    return (
      <Modal isOpen={true} onClose={onClose} title="Budget Details">
        <div className="loading">Loading budget details...</div>
      </Modal>
    );
  }

  if (error) {
    return (
      <Modal isOpen={true} onClose={onClose} title="Budget Details">
        <div className="error">{error}</div>
        <button onClick={onClose} className="btn-primary">Close</button>
      </Modal>
    );
  }

  const getProgressColor = (percentage) => {
    if (percentage > 100) return '#dc3545'; // Red - over budget
    if (percentage > 80) return '#ffc107';  // Yellow - approaching limit
    return '#28a745'; // Green - on track
  };

  const getStatusIcon = (percentage) => {
    if (percentage > 100) return '🔴';
    if (percentage > 80) return '🟡';
    return '🟢';
  };

  return (
    <Modal isOpen={true} onClose={onClose} title={`Budget Details - ${budget.name}`} maxWidth="800px">
      <div className="budget-details-modal">
        {/* Budget Summary */}
        <div className="budget-summary">
          <h3>Budget Overview</h3>
          <div className="summary-grid">
            <div className="summary-card">
              <div className="summary-value">{formatAmount(budgetDetails.summary.total_planned)}</div>
              <div className="summary-label">Total Planned</div>
            </div>
            <div className="summary-card">
              <div className="summary-value">{formatAmount(budgetDetails.summary.total_spent)}</div>
              <div className="summary-label">Total Spent</div>
            </div>
            <div className="summary-card">
              <div className="summary-value">{formatAmount(budgetDetails.summary.total_remaining)}</div>
              <div className="summary-label">Remaining</div>
            </div>
            <div className="summary-card">
              <div className="summary-value">{budgetDetails.summary.overall_progress.toFixed(1)}%</div>
              <div className="summary-label">Overall Progress</div>
            </div>
          </div>

          <div className="budget-period">
            <strong>Period:</strong> {formatDate(budget.start_date)} to {formatDate(budget.end_date)}
          </div>
        </div>

        {/* Category Breakdown */}
        <div className="categories-section">
          <h3>Category Breakdown</h3>
          <div className="categories-table">
            <table>
              <thead>
                <tr>
                  <th>Category</th>
                  <th>Planned</th>
                  <th>Spent</th>
                  <th>Remaining</th>
                  <th>Progress</th>
                  <th>Transactions</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {budgetDetails.categories.map((category) => {
                  const remaining = category.planned_amount - category.spent_amount;
                  return (
                    <tr key={category.id}>
                      <td className="category-name">{category.name}</td>
                      <td className="amount-cell">{formatAmount(category.planned_amount)}</td>
                      <td className="amount-cell">{formatAmount(category.spent_amount)}</td>
                      <td className={`amount-cell ${remaining < 0 ? 'negative' : 'positive'}`}>
                        {formatAmount(remaining)}
                      </td>
                      <td className="progress-cell">
                        <div className="progress-container">
                          <div className="progress-bar">
                            <div 
                              className="progress-fill"
                              style={{ 
                                width: `${Math.min(category.progress_percentage, 100)}%`,
                                backgroundColor: getProgressColor(category.progress_percentage)
                              }}
                            ></div>
                          </div>
                          <span className="progress-text">
                            {category.progress_percentage.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                      <td className="transaction-count">{category.transaction_count}</td>
                      <td className="status-cell">
                        <span className="status-indicator">
                          {getStatusIcon(category.progress_percentage)}
                          {category.progress_percentage > 100 ? 'Over Budget' : 
                           category.progress_percentage > 80 ? 'Near Limit' : 'On Track'}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="quick-stats">
          <div className="stat-item success">
            <span className="stat-icon">🟢</span>
            <span className="stat-number">{budgetDetails.summary.categories_on_track}</span>
            <span className="stat-label">On Track</span>
          </div>
          <div className="stat-item warning">
            <span className="stat-icon">🔴</span>
            <span className="stat-number">{budgetDetails.summary.categories_over_budget}</span>
            <span className="stat-label">Over Budget</span>
          </div>
        </div>

        <div className="modal-actions">
          <button onClick={onClose} className="btn-secondary">Close</button>
        </div>
      </div>

      <style jsx>{`
        .budget-details-modal {
          padding: 20px;
          max-height: 80vh;
          overflow-y: auto;
        }

        .budget-summary {
          margin-bottom: 30px;
          padding: 20px;
          background: #f8f9fa;
          border-radius: 8px;
        }

        .summary-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          gap: 15px;
          margin-bottom: 15px;
        }

        .summary-card {
          text-align: center;
          padding: 15px;
          background: white;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .summary-value {
          font-size: 24px;
          font-weight: bold;
          color: #333;
          margin-bottom: 5px;
        }

        .summary-label {
          font-size: 12px;
          color: #666;
          text-transform: uppercase;
        }

        .budget-period {
          text-align: center;
          color: #666;
          margin-top: 15px;
        }

        .categories-section h3 {
          margin-bottom: 15px;
          color: #333;
        }

        .categories-table {
          overflow-x: auto;
        }

        .categories-table table {
          width: 100%;
          border-collapse: collapse;
          background: white;
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .categories-table th,
        .categories-table td {
          padding: 12px;
          text-align: left;
          border-bottom: 1px solid #eee;
        }

        .categories-table th {
          background: #f8f9fa;
          font-weight: 600;
          color: #333;
        }

        .category-name {
          font-weight: 500;
        }

        .amount-cell {
          text-align: right;
          font-family: monospace;
        }

        .amount-cell.negative {
          color: #dc3545;
        }

        .amount-cell.positive {
          color: #28a745;
        }

        .progress-cell {
          min-width: 150px;
        }

        .progress-container {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .progress-bar {
          flex: 1;
          height: 20px;
          background: #e9ecef;
          border-radius: 10px;
          overflow: hidden;
        }

        .progress-fill {
          height: 100%;
          transition: width 0.3s ease;
        }

        .progress-text {
          font-size: 12px;
          min-width: 45px;
          text-align: right;
        }

        .transaction-count {
          text-align: center;
        }

        .status-cell {
          text-align: center;
        }

        .status-indicator {
          font-size: 12px;
          display: flex;
          align-items: center;
          gap: 5px;
        }

        .quick-stats {
          display: flex;
          justify-content: center;
          gap: 30px;
          margin: 30px 0;
        }

        .stat-item {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 10px 15px;
          border-radius: 8px;
        }

        .stat-item.success {
          background: #d4edda;
          border: 1px solid #c3e6cb;
        }

        .stat-item.warning {
          background: #f8d7da;
          border: 1px solid #f5c6cb;
        }

        .stat-number {
          font-weight: bold;
          font-size: 18px;
        }

        .stat-label {
          font-size: 12px;
          color: #666;
        }

        .modal-actions {
          display: flex;
          justify-content: center;
          margin-top: 30px;
        }

        .btn-secondary {
          padding: 10px 30px;
          border: 1px solid #ddd;
          background: white;
          color: #666;
          border-radius: 4px;
          cursor: pointer;
        }

        .btn-secondary:hover {
          background: #f8f9fa;
        }

        .loading, .error {
          text-align: center;
          padding: 40px;
          color: #666;
        }
      `}</style>
    </Modal>
  );
};

export default BudgetDetailsModal;