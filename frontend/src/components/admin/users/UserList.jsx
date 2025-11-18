import React, { useState, useEffect } from 'react';
import adminApi from '../../../utils/adminApi';
import UserCard from './UserCard';
import UserDetailsModal from './UserDetailsModal';
import DeleteUserModal from './DeleteUserModal';

const UserList = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedUser, setSelectedUser] = useState(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [userToDelete, setUserToDelete] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getAllUsers();
      setUsers(data.users);
    } catch (err) {
      setError('Failed to load users: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = (user) => {
    setSelectedUser(user);
    setShowDetailsModal(true);
  };

  const handleDeleteUser = (user) => {
    setUserToDelete(user);
    setShowDeleteModal(true);
  };

  const handleUserDeleted = () => {
    loadUsers(); // Refresh the list
    setShowDeleteModal(false);
    setUserToDelete(null);
  };

  const handleRoleUpdated = () => {
    loadUsers(); // Refresh the list
    setShowDetailsModal(false);
    setSelectedUser(null);
  };

  // Filter users based on search term and role
  const filteredUsers = users.filter(user => {
    const matchesSearch = user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         user.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (user.name && user.name.toLowerCase().includes(searchTerm.toLowerCase()));
    
    const matchesRole = roleFilter === 'all' || user.role === roleFilter;
    
    return matchesSearch && matchesRole;
  });

  if (loading) {
    return (
      <div className="user-list-container">
        <div className="loading">Loading users...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="user-list-container">
        <div className="error">{error}</div>
        <button onClick={loadUsers} className="retry-btn">Retry</button>
      </div>
    );
  }

  return (
    <div className="user-list-container">
      <div className="user-list-header">
        <h2>User Management</h2>
        <p>Manage user accounts and permissions</p>
      </div>

      {/* Filters and Search */}
      <div className="user-list-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search users by name, username, or email..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>
        
        <div className="filter-container">
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="role-filter"
          >
            <option value="all">All Roles</option>
            <option value="user">Users</option>
            <option value="admin">Admins</option>
          </select>
        </div>

        <button onClick={loadUsers} className="refresh-btn">
          🔄 Refresh
        </button>
      </div>

      {/* User Stats */}
      <div className="user-stats">
        <div className="stat-item">
          <span className="stat-value">{users.length}</span>
          <span className="stat-label">Total Users</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{users.filter(u => u.role === 'admin').length}</span>
          <span className="stat-label">Admins</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{users.filter(u => u.role === 'user').length}</span>
          <span className="stat-label">Regular Users</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{filteredUsers.length}</span>
          <span className="stat-label">Filtered Results</span>
        </div>
      </div>

      {/* User Cards Grid */}
      <div className="users-grid">
        {filteredUsers.length === 0 ? (
          <div className="no-users">
            <p>No users found matching your criteria.</p>
          </div>
        ) : (
          filteredUsers.map((user) => (
            <UserCard
              key={user.id}
              user={user}
              onViewDetails={handleViewDetails}
              onDeleteUser={handleDeleteUser}
            />
          ))
        )}
      </div>

      {/* Modals */}
      {showDetailsModal && selectedUser && (
        <UserDetailsModal
          user={selectedUser}
          onClose={() => setShowDetailsModal(false)}
          onRoleUpdated={handleRoleUpdated}
        />
      )}

      {showDeleteModal && userToDelete && (
        <DeleteUserModal
          user={userToDelete}
          onClose={() => setShowDeleteModal(false)}
          onUserDeleted={handleUserDeleted}
        />
      )}
    </div>
  );
};

export default UserList;