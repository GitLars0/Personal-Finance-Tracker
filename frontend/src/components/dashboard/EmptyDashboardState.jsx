import React from 'react';
import { Link } from 'react-router-dom';

function EmptyDashboardState() {
  return (
    <div className="empty-state">
      <div className="empty-state-content">
        <h2>🎯 Ready to start tracking your finances?</h2>
        <p>It looks like you haven't added any financial data yet. Let's get you started!</p>
        
        <div className="getting-started-cards">
          <div className="getting-started-card">
            <h3>🏦 Set Up Your Accounts</h3>
            <p>Add your bank accounts, credit cards, and cash to track all your money.</p>
            <Link to="/accounts" className="action-button primary">
              Manage Accounts
            </Link>
          </div>
          
          <div className="getting-started-card">
            <h3>🏷️ Create Categories</h3>
            <p>Organize your transactions with income and expense categories.</p>
            <Link to="/categories" className="action-button secondary">
              Set Up Categories
            </Link>
          </div>
          
          <div className="getting-started-card">
            <h3>📊 Add Your First Transaction</h3>
            <p>Start by recording your income and expenses to see your spending patterns.</p>
            <Link to="/transactions" className="action-button primary">
              Go to Transactions
            </Link>
          </div>
          
          <div className="getting-started-card">
            <h3>💰 Create a Budget</h3>
            <p>Set spending limits for different categories to stay on track with your goals.</p>
            <Link to="/budgets" className="action-button secondary">
              Create Budget
            </Link>
          </div>
        </div>
        
        <div className="quick-actions">
          <h3>Quick Actions</h3>
          <div className="quick-action-buttons">
            <button className="quick-action-btn" onClick={() => window.location.href = '/accounts'}>
              🏦 Add Account
            </button>
            <button className="quick-action-btn" onClick={() => window.location.href = '/categories'}>
              🏷️ Add Category
            </button>
            <button className="quick-action-btn" onClick={() => window.location.href = '/transactions'}>
              ➕ Add Transaction
            </button>
            <button className="quick-action-btn" onClick={() => window.location.href = '/budgets'}>
              🎯 Set Budget
            </button>
          </div>
        </div>
        
        <div className="help-text">
          <p><strong>💡 Getting Started:</strong> First set up your accounts and categories, then start adding transactions to see your dashboard come to life!</p>
        </div>
      </div>
    </div>
  );
}

export default EmptyDashboardState;