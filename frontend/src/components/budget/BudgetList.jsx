import React from 'react';
import BudgetItem from './BudgetItem';
import { formatCurrency, formatDate } from '../../utils/format';

export default function BudgetList({ budget, onEdit, onDelete }) {
  // The budget should now have a summary object with spending calculations
  const summary = budget.summary || {
    total_planned_cents: 0,
    total_spent_cents: 0,
    total_remaining_cents: 0
  };

  // Use the items array directly from the budget
  const budgetItems = budget.items || [];

  return (
    <div className="budget-card">
      <div className="budget-card-header">
        <h3>Budget ({formatDate(budget.period_start)} - {formatDate(budget.period_end)})</h3>
        <div className="budget-actions">
          <button onClick={() => onEdit(budget)} className="edit-btn">✏️</button>
          <button onClick={() => onDelete(budget)} className="delete-btn">🗑️</button>
        </div>
      </div>

      <div className="budget-summary">
        <div className="summary-item">Planned: <strong>{formatCurrency(summary.total_planned_cents)}</strong></div>
        <div className="summary-item">Spent: <strong className="spent">{formatCurrency(summary.total_spent_cents)}</strong></div>
        <div className="summary-item">Remaining: <strong className={`remaining ${summary.total_remaining_cents < 0 ? 'negative' : ''}`}>{formatCurrency(summary.total_remaining_cents)}</strong></div>
      </div>

      <div className="budget-items">
        {budgetItems.map((item, idx) => (
          <BudgetItem 
            key={item.id || item.category_id || idx} 
            item={item} 
          />
        ))}
      </div>
    </div>
  );
}
