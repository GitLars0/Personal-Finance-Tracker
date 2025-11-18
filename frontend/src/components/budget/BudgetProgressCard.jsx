import React from 'react';
import { formatCurrency, formatDate } from '../../utils/format';

export default function BudgetProgressCard({ budgetData, onEdit, onDelete }) {
  if (!budgetData || !budgetData.budget) {
    return null;
  }

  const { budget, summary, categories } = budgetData;

  return (
    <div className="budget-card">
      <div className="budget-card-header">
        <h3>Budget ({formatDate(budget.period_start)} - {formatDate(budget.period_end)})</h3>
        <div className="budget-actions">
          <button onClick={() => onEdit(budgetData)} className="edit-btn">✏️</button>
          <button onClick={() => onDelete(budgetData)} className="delete-btn">🗑️</button>
        </div>
      </div>

      <div className="budget-summary">
        <div className="summary-item">Planned: <strong>{formatCurrency(summary.total_planned_cents)}</strong></div>
        <div className="summary-item">Spent: <strong className="spent">{formatCurrency(summary.total_spent_cents)}</strong></div>
        <div className="summary-item">Remaining: <strong className={`remaining ${summary.total_remaining_cents < 0 ? 'negative' : ''}`}>{formatCurrency(summary.total_remaining_cents)}</strong></div>
      </div>

      <div className="budget-items">
        {categories?.map((category) => (
          <div key={category.category_id} className="budget-item">
            <div className="item-header">
              <span className="category-name">{category.category_name}</span>
              <span className="item-amount">{formatCurrency(category.spent_cents)} / {formatCurrency(category.planned_cents)}</span>
            </div>
            <div className="progress-bar">
              <div 
                className={`progress-fill ${category.status}`} 
                style={{ width: `${Math.min(category.progress_percent || 0, 100)}%` }} 
              />
            </div>
            <div className="progress-percentage">
              {(category.progress_percent || 0).toFixed(1)}%
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}