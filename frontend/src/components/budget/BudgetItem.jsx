import React from 'react';
import { formatCurrency } from '../../utils/format';

export default function BudgetItem({ item }) {
  // The item should now have all necessary data from the unified API response
  const categoryName = item.category?.name || 'Unknown';
  const plannedCents = item.planned_cents || 0;
  const spentCents = item.spent_cents || 0;
  const progressPercent = item.progress_percent || 0;
  const status = item.status || 'under_budget';

  return (
    <div className="budget-item">
      <div className="item-header">
        <span className="category-name">{categoryName}</span>
        <span className="item-amount">{formatCurrency(spentCents)} / {formatCurrency(plannedCents)}</span>
      </div>
      <div className="progress-bar">
        <div 
          className={`progress-fill ${status}`} 
          style={{ width: `${Math.min(progressPercent, 100)}%` }} 
        />
      </div>
      <div className="progress-percentage">
        {progressPercent.toFixed(1)}%
      </div>
    </div>
  );
}
