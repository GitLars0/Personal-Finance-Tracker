import React from 'react';

function BudgetProgressSection({ budgetProgress, formatCurrency, formatDate }) {
  if (!budgetProgress) {
    return null;
  }

  return (
    <section className="dashboard-section">
      <h2>Budget Progress</h2>
      <div className="budget-overview">
        <div className="budget-summary">
          <div className="summary-item">
            <span>Planned:</span>
            <strong>{formatCurrency(budgetProgress.summary.total_planned_cents)}</strong>
          </div>
          <div className="summary-item">
            <span>Spent:</span>
            <strong className="spent">{formatCurrency(budgetProgress.summary.total_spent_cents)}</strong>
          </div>
          <div className="summary-item">
            <span>Remaining:</span>
            <strong className="remaining">{formatCurrency(budgetProgress.summary.total_remaining_cents)}</strong>
          </div>
        </div>
        <div className="budget-period">
          <small>
            {formatDate(budgetProgress.budget.period_start)} - {formatDate(budgetProgress.budget.period_end)}
            ({budgetProgress.budget.days_remaining} days remaining)
          </small>
        </div>
      </div>

      <div className="budget-categories">
        {budgetProgress.categories?.map((category) => (
          <div key={category.category_id} className="budget-item">
            <div className="budget-item-header">
              <span className="category-name">{category.category_name}</span>
              <span className={`status-badge ${category.status}`}>
                {category.status.replace('_', ' ')}
              </span>
            </div>
            <div className="progress-bar">
              <div 
                className={`progress-fill ${category.status}`}
                style={{ width: `${Math.min(category.progress_percent || 0, 100)}%` }}
              />
            </div>
            <div className="budget-item-details">
              <span>{formatCurrency(category.spent_cents)} / {formatCurrency(category.planned_cents)}</span>
              <span>{(category.progress_percent || 0).toFixed(1)}%</span>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

export default BudgetProgressSection;