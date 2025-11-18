import React from 'react';

function SpendingSummarySection({ spendSummary, formatCurrency }) {
  if (!spendSummary || !spendSummary.categories?.length) {
    return null;
  }

  return (
    <section className="dashboard-section">
      <h2>Spending by Category</h2>
      <div className="spend-summary">
        <div className="donut-chart-placeholder">
          {spendSummary.categories.slice(0, 5).map((cat) => (
            <div key={cat.category_id} className="category-spend-item">
              <div className="category-info">
                <span className="category-name">{cat.category_name}</span>
                <span className="category-amount">{formatCurrency(cat.total_cents)}</span>
              </div>
              <div className="category-bar-container">
                <div 
                  className="category-bar"
                  style={{ width: `${cat.percentage}%` }}
                />
              </div>
              <span className="category-percentage">{cat.percentage.toFixed(1)}%</span>
            </div>
          ))}
        </div>
        <div className="total-spending">
          <h3>Total Spending</h3>
          <p className="amount">{formatCurrency(spendSummary.total_spent_cents)}</p>
          <small>{spendSummary.period.from} - {spendSummary.period.to}</small>
        </div>
      </div>
    </section>
  );
}

export default SpendingSummarySection;