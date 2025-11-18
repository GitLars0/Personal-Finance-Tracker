// Use relative URLs to match the proxy configuration

// Helper function to get auth headers
const getAuthHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };
};

// Helper function to handle API responses
const handleResponse = async (response) => {
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Request failed');
  }
  return response.json();
};

// User Management APIs
export const adminApi = {
  // Get all users
  getAllUsers: async () => {
    const response = await fetch(`/api/admin/users`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get specific user details
  getUserDetails: async (userId) => {
    const response = await fetch(`/api/admin/users/${userId}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Delete a user
  deleteUser: async (userId) => {
    const response = await fetch(`/api/admin/users/${userId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Update user role
  updateUserRole: async (userId, role) => {
    const response = await fetch(`/api/admin/users/${userId}/role`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({ role }),
    });
    return handleResponse(response);
  },

  // Data Oversight APIs
  
  // Get all transactions across all users
  getAllTransactions: async () => {
    const response = await fetch(`/api/admin/transactions`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get all accounts across all users
  getAllAccounts: async () => {
    const response = await fetch(`/api/admin/accounts`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get all categories across all users
  getAllCategories: async () => {
    const response = await fetch(`/api/admin/categories`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get all budgets across all users
  getAllBudgets: async () => {
    const response = await fetch(`/api/admin/budgets`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get specific budget details
  getBudgetDetails: async (budgetId) => {
    const response = await fetch(`/api/admin/budgets/${budgetId}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Delete transaction (admin override)
  deleteTransaction: async (transactionId) => {
    const response = await fetch(`/api/admin/transactions/${transactionId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Delete account (admin override)
  deleteAccount: async (accountId) => {
    const response = await fetch(`/api/admin/accounts/${accountId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Delete category (admin override)
  deleteCategory: async (categoryId) => {
    const response = await fetch(`/api/admin/categories/${categoryId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Delete budget (admin override)
  deleteBudget: async (budgetId) => {
    const response = await fetch(`/api/admin/budgets/${budgetId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // Get dashboard statistics
  getDashboardStats: async () => {
    const response = await fetch(`/api/admin/dashboard-stats`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
};

export default adminApi;