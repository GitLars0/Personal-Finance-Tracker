import React, { useEffect, useState, useCallback } from 'react';
import '../styles/Categories.css';
import { CategoryForm, CategoryList, DefaultCategoriesSetup } from '../components/categories';
import DeleteCategoryModal from '../components/categories/DeleteCategoryModal';

function Categories() {
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingCategory, setEditingCategory] = useState(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [categoryToDelete, setCategoryToDelete] = useState(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    kind: 'expense',
    parent_id: '',
    description: ''
  });

  // Filter state
  const [filter, setFilter] = useState('all'); // 'all', 'income', 'expense'

  const token = localStorage.getItem('token');

  const fetchCategories = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/categories', {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok && response.status !== 404) {
        throw new Error('Failed to fetch categories');
      }

      const categories = response.ok ? await response.json() : [];
      setCategories(categories);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) {
      setError('Not authenticated');
      setLoading(false);
      return;
    }
    fetchCategories();
  }, [token, fetchCategories]);

  const handleFormSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.name || !formData.kind) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      const submitData = {
        name: formData.name,
        kind: formData.kind,
        parent_id: formData.parent_id ? parseInt(formData.parent_id) : null,
        description: formData.description || null
      };

      const url = editingCategory 
        ? `/api/categories/${editingCategory.id}`
        : '/api/categories';
      
      const method = editingCategory ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers,
        body: JSON.stringify(submitData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save category');
      }

      // Success - refresh data and close form
      await fetchCategories();
      resetForm();
      setShowForm(false);
      setEditingCategory(null);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleEdit = (category) => {
    setEditingCategory(category);
    setFormData({
      name: category.name || '',
      kind: category.kind || 'expense',
      parent_id: category.parent_id?.toString() || '',
      description: category.description || ''
    });
    setShowForm(true);
  };

  const handleDelete = (category) => {
    setCategoryToDelete(category);
    setShowDeleteModal(true);
  };

  const confirmDelete = async (categoryId) => {
    try {
      await fetchCategories();
      setShowDeleteModal(false);
      setCategoryToDelete(null);
    } catch (err) {
      setError(err.message);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      kind: 'expense',
      parent_id: '',
      description: ''
    });
    setEditingCategory(null);
    setError(null);
  };

  const createDefaultCategories = async () => {
    const defaultCategories = [
      // Income categories
      { name: 'Salary', kind: 'income', description: 'Regular employment income' },
      { name: 'Freelance', kind: 'income', description: 'Freelance and contract work' },
      { name: 'Investment Returns', kind: 'income', description: 'Dividends, interest, capital gains' },
      { name: 'Other Income', kind: 'income', description: 'Miscellaneous income' },
      
      // Expense categories
      { name: 'Housing', kind: 'expense', description: 'Rent, mortgage, utilities' },
      { name: 'Food & Dining', kind: 'expense', description: 'Groceries, restaurants, takeout' },
      { name: 'Transportation', kind: 'expense', description: 'Gas, public transit, car maintenance' },
      { name: 'Healthcare', kind: 'expense', description: 'Medical bills, insurance, pharmacy' },
      { name: 'Entertainment', kind: 'expense', description: 'Movies, games, subscriptions' },
      { name: 'Shopping', kind: 'expense', description: 'Clothing, electronics, general shopping' },
      { name: 'Bills & Utilities', kind: 'expense', description: 'Phone, internet, electricity' },
      { name: 'Education', kind: 'expense', description: 'Books, courses, tuition' },
      { name: 'Personal Care', kind: 'expense', description: 'Haircuts, cosmetics, gym' },
      { name: 'Travel', kind: 'expense', description: 'Vacations, business trips' },
      { name: 'Other Expenses', kind: 'expense', description: 'Miscellaneous expenses' }
    ];

    try {
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      for (const category of defaultCategories) {
        await fetch('/api/categories', {
          method: 'POST',
          headers,
          body: JSON.stringify(category)
        });
      }

      await fetchCategories();
    } catch (err) {
      setError('Failed to create default categories: ' + err.message);
    }
  };

  // Filter categories
  const filteredCategories = categories.filter(category => {
    if (filter === 'all') return true;
    return category.kind === filter;
  });

  // Group categories by parent
  const groupedCategories = filteredCategories.reduce((groups, category) => {
    const key = category.parent_id || 'root';
    if (!groups[key]) groups[key] = [];
    groups[key].push(category);
    return groups;
  }, {});

  if (loading) {
    return (
      <div className="categories-container">
        <div className="loading">Loading categories...</div>
      </div>
    );
  }

  if (error && !showForm) {
    return (
      <div className="categories-container">
        <div className="error-message">{error}</div>
        <button onClick={fetchCategories}>Retry</button>
      </div>
    );
  }

  return (
    <div className="categories-container">
      <header className="categories-header">
        <h1>Categories</h1>
        <div className="header-actions">
          <button 
            className="add-category-btn"
            onClick={() => {
              resetForm();
              setShowForm(true);
            }}
          >
            ➕ Add Category
          </button>
        </div>
      </header>

      {error && <div className="error-message">{error}</div>}

      <CategoryForm 
        showForm={showForm}
        editingCategory={editingCategory}
        formData={formData}
        setFormData={setFormData}
        handleFormSubmit={handleFormSubmit}
        resetForm={resetForm}
        setShowForm={setShowForm}
        setEditingCategory={setEditingCategory}
        categories={categories}
      />

      {categories.length === 0 ? (
        <DefaultCategoriesSetup 
          createDefaultCategories={createDefaultCategories}
          resetForm={resetForm}
          setShowForm={setShowForm}
        />
      ) : (
        <>
          {/* Filter Tabs */}
          <div className="filter-tabs">
            <button 
              className={filter === 'all' ? 'active' : ''}
              onClick={() => setFilter('all')}
            >
              All Categories ({categories.length})
            </button>
            <button 
              className={filter === 'income' ? 'active' : ''}
              onClick={() => setFilter('income')}
            >
              Income ({categories.filter(c => c.kind === 'income').length})
            </button>
            <button 
              className={filter === 'expense' ? 'active' : ''}
              onClick={() => setFilter('expense')}
            >
              Expenses ({categories.filter(c => c.kind === 'expense').length})
            </button>
          </div>

          <div className="categories-list">
            <CategoryList 
              filteredCategories={filteredCategories}
              filter={filter}
              resetForm={resetForm}
              setFormData={setFormData}
              setShowForm={setShowForm}
              groupedCategories={groupedCategories}
              handleEdit={handleEdit}
              handleDelete={handleDelete}
            />
          </div>

          {/* Quick Actions */}
          {categories.length > 0 && categories.length < 10 && (
            <div className="quick-actions">
              <h3>Need more categories?</h3>
              <button 
                className="action-button secondary"
                onClick={createDefaultCategories}
              >
                Add Default Categories
              </button>
            </div>
          )}
        </>
      )}

      <DeleteCategoryModal
        isOpen={showDeleteModal}
        onClose={() => {
          setShowDeleteModal(false);
          setCategoryToDelete(null);
        }}
        category={categoryToDelete}
        onConfirm={confirmDelete}
        token={token}
      />
    </div>
  );
}

export default Categories;