import React from 'react';

function TransactionFilters({ 
  filters, 
  setFilters, 
  categories, 
  accounts 
}) {
  return (
    <div className="filters-section">
      <div className="filters-row">
        <input
          type="text"
          placeholder="Search descriptions..."
          value={filters.search}
          onChange={(e) => setFilters({...filters, search: e.target.value})}
          className="search-input"
        />
        
        <select
          value={filters.category_id}
          onChange={(e) => setFilters({...filters, category_id: e.target.value})}
        >
          <option value="">All Categories</option>
          {categories.map(category => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>

        <select
          value={filters.account_id}
          onChange={(e) => setFilters({...filters, account_id: e.target.value})}
        >
          <option value="">All Accounts</option>
          {accounts.map(account => (
            <option key={account.id} value={account.id}>
              {account.name}
            </option>
          ))}
        </select>

        <input
          type="date"
          placeholder="From"
          value={filters.from}
          onChange={(e) => setFilters({...filters, from: e.target.value})}
        />

        <input
          type="date"
          placeholder="To"
          value={filters.to}
          onChange={(e) => setFilters({...filters, to: e.target.value})}
        />

        <button 
          onClick={() => setFilters({search: '', category_id: '', account_id: '', from: '', to: ''})}
          className="clear-filters-btn"
        >
          Clear Filters
        </button>
      </div>
    </div>
  );
}

export default TransactionFilters;