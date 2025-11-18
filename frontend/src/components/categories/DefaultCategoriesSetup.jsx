import React from 'react';

function DefaultCategoriesSetup({ createDefaultCategories, resetForm, setShowForm }) {
  return (
    <div className="empty-state">
      <h3>No categories found</h3>
      <p>Add categories to organize your transactions and budgets.</p>
      
      <div className="empty-actions">
        <button 
          className="action-button primary"
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
        >
          Create Custom Category
        </button>
        <button 
          className="action-button secondary"
          onClick={createDefaultCategories}
        >
          Create Default Categories
        </button>
      </div>
      
      <div className="getting-started">
        <h4>Getting Started</h4>
        <p>Categories help you organize and track your financial transactions. We can set up common categories for you, or you can create your own.</p>
        
        <div className="category-examples">
          <div className="example-group">
            <h5>💰 Income Categories</h5>
            <ul>
              <li>Salary</li>
              <li>Freelance</li>
              <li>Investment Returns</li>
              <li>Other Income</li>
            </ul>
          </div>
          
          <div className="example-group">
            <h5>💸 Expense Categories</h5>
            <ul>
              <li>Housing & Utilities</li>
              <li>Food & Dining</li>
              <li>Transportation</li>
              <li>Entertainment</li>
              <li>Healthcare</li>
              <li>Shopping</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DefaultCategoriesSetup;