import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import { formatDate } from '../../../utils/format';
import DeleteConfirmationModal from '../DeleteConfirmationModal';

const AllCategories = () => {
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');
  const [sortBy, setSortBy] = useState('name');

  // Delete modal state
  const [deleteModal, setDeleteModal] = useState({
    isOpen: false,
    category: null,
    isDeleting: false
  });

  useEffect(() => {
    loadCategories();
  }, []);

  const loadCategories = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getAllCategories();
      setCategories(data.categories);
    } catch (err) {
      setError('Failed to load categories: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteCategory = async (category) => {
    setDeleteModal({
      isOpen: true,
      category: category,
      isDeleting: false
    });
  };

  const confirmDeleteCategory = async () => {
    if (!deleteModal.category) return;

    try {
      setDeleteModal(prev => ({ ...prev, isDeleting: true }));
      await adminApi.deleteCategory(deleteModal.category.id);
      setCategories(categories.filter(c => c.id !== deleteModal.category.id));
      setDeleteModal({ isOpen: false, category: null, isDeleting: false });
    } catch (err) {
      setError('Failed to delete category: ' + err.message);
      setDeleteModal(prev => ({ ...prev, isDeleting: false }));
    }
  };

  // Filter and sort categories
  const filteredCategories = categories
    .filter(category => {
      const matchesSearch = category.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           category.user_email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           category.user_username.toLowerCase().includes(searchTerm.toLowerCase());
      
      const matchesType = typeFilter === 'all' || category.type === typeFilter;
      
      return matchesSearch && matchesType;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'user':
          return a.user_username.localeCompare(b.user_username);
        case 'type':
          return a.type.localeCompare(b.type);
        case 'date':
          return new Date(b.created_at) - new Date(a.created_at);
        default:
          return 0;
      }
    });

  if (loading) {
    return (
      <div className="admin-data-container">
        <div className="loading">Loading all categories...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="admin-data-container">
        <div className="error">{error}</div>
        <button onClick={loadCategories} className="retry-btn">Retry</button>
      </div>
    );
  }

  return (
    <div className="admin-data-container">
      <div className="admin-data-header">
        <h2>All Categories</h2>
        <p>Monitor and manage transaction categories across all users</p>
      </div>

      {/* Controls */}
      <div className="admin-data-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search by category name, user email, or username..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-container">
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="type-filter"
          >
            <option value="all">All Types</option>
            <option value="income">Income</option>
            <option value="expense">Expense</option>
          </select>
        </div>

        <div className="sort-container">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="sort-select"
          >
            <option value="name">By Name</option>
            <option value="user">By User</option>
            <option value="type">By Type</option>
            <option value="date">By Date Created</option>
          </select>
        </div>

        <button onClick={loadCategories} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>

      {/* Statistics */}
      <div className="data-stats">
        <div className="stat-item">
          <span className="stat-value">{categories.length}</span>
          <span className="stat-label">Total Categories</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{categories.filter(c => c.type === 'income').length}</span>
          <span className="stat-label">Income Categories</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{categories.filter(c => c.type === 'expense').length}</span>
          <span className="stat-label">Expense Categories</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{filteredCategories.length}</span>
          <span className="stat-label">Filtered Results</span>
        </div>
      </div>

      {/* Categories Table */}
      <div className="admin-data-table">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>User</th>
              <th>Category Name</th>
              <th>Type</th>
              <th>Parent Category</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredCategories.length === 0 ? (
              <tr>
                <td colSpan="6" className="no-data">
                  No categories found matching your criteria.
                </td>
              </tr>
            ) : (
              filteredCategories.map((category) => (
                <tr key={category.id}>
                  <td>#{category.id}</td>
                  <td>
                    <div className="user-cell">
                      <div className="username">@{category.user_username}</div>
                      <div className="email">{category.user_email}</div>
                    </div>
                  </td>
                  <td className="category-name-cell">
                    <div className="category-info">
                      <span className="category-name">{category.name}</span>
                      {category.icon && <span className="category-icon">{category.icon}</span>}
                    </div>
                  </td>
                  <td>
                    <span className={`type-badge ${category.type}`}>
                      {category.type}
                    </span>
                  </td>
                  <td>
                    {category.parent_name ? (
                      <span className="parent-category">{category.parent_name}</span>
                    ) : (
                      <span className="no-parent">Top Level</span>
                    )}
                  </td>
                  <td>{formatDate(category.created_at)}</td>
                  <td>
                    <button
                      onClick={() => handleDeleteCategory(category)}
                      className="btn-danger-small"
                      title="Delete category"
                    >
                      🗑️
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Delete Confirmation Modal */}
      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, category: null, isDeleting: false })}
        onConfirm={confirmDeleteCategory}
        title="Delete Category"
        message="Are you sure you want to delete this category? This may affect associated transactions."
        itemName={deleteModal.category?.name}
        isDeleting={deleteModal.isDeleting}
      />
    </div>
  );
};

export default AllCategories;