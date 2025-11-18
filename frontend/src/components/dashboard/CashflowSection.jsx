import React from 'react';

function CashflowSection({ cashflow, formatCurrency }) {
  if (!cashflow || !cashflow.periods?.length) {
    return null;
  }

  return (
    <section className="dashboard-section">
      <h2>Income vs Expenses</h2>
      <div className="cashflow-summary">
        <div className="cashflow-stat">
          <span>Total Income</span>
          <strong className="income">{formatCurrency(cashflow.summary.total_income_cents)}</strong>
        </div>
        <div className="cashflow-stat">
          <span>Total Expenses</span>
          <strong className="expense">{formatCurrency(cashflow.summary.total_expense_cents)}</strong>
        </div>
        <div className="cashflow-stat">
          <span>Net</span>
          <strong className={cashflow.summary.net_cents >= 0 ? 'income' : 'expense'}>
            {formatCurrency(cashflow.summary.net_cents)}
          </strong>
        </div>
      </div>
      <div className="cashflow-chart">
        {cashflow.periods.slice(-6).map((period, index) => {
          const maxAmount = Math.max(...cashflow.periods.map(p => 
            Math.max(p.income_cents, p.expense_cents)
          ));
          return (
            <div key={index} className="cashflow-period">
              <div className="period-label">
                {new Date(period.period).toLocaleDateString('en-US', { month: 'short' })}
              </div>
              <div className="period-bars">
                <div 
                  className="bar income-bar"
                  style={{ height: `${(period.income_cents / maxAmount) * 100}%` }}
                  title={`Income: ${formatCurrency(period.income_cents)}`}
                />
                <div 
                  className="bar expense-bar"
                  style={{ height: `${(period.expense_cents / maxAmount) * 100}%` }}
                  title={`Expenses: ${formatCurrency(period.expense_cents)}`}
                />
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}

export default CashflowSection;