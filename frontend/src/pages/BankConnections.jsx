import React, { useState, useEffect, useCallback } from 'react';
import useAuthFetch from '../hooks/useAuthFetch';
import PlaidLink from '../components/PlaidLink';
import ConfirmModal from '../components/accounts/ConfirmModal';
import '../styles/BankConnections.css';

export default function BankConnections() {
  const authFetch = useAuthFetch();
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [disconnectConfirm, setDisconnectConfirm] = useState({ isOpen: false, connectionId: null, bankName: '' });
  const [successModal, setSuccessModal] = useState({ isOpen: false, message: '' });

  const loadConnections = useCallback(async () => {
    try {
      setLoading(true);
      const data = await authFetch('/api/banks/connections');
      setConnections(data.connections || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [authFetch]);

  useEffect(() => {
    loadConnections();
  }, [loadConnections]);

  const handleDisconnect = async (connectionId) => {
    const connection = connections.find(c => c.id === connectionId);
    if (!connection) return;
    
    // Show confirmation modal
    setDisconnectConfirm({
      isOpen: true,
      connectionId: connectionId,
      bankName: connection.bank_name
    });
  };

  const confirmDisconnect = async () => {
    const { connectionId } = disconnectConfirm;
    
    try {
      await authFetch(`/api/banks/connections/${connectionId}`, {
        method: 'DELETE'
      });
      
      setSuccessModal({
        isOpen: true,
        message: `Successfully disconnected from ${disconnectConfirm.bankName}!`
      });
      
      loadConnections(); // Refresh the list
    } catch (err) {
      setError(err.message);
    } finally {
      setDisconnectConfirm({ isOpen: false, connectionId: null, bankName: '' });
    }
  };

  const handleSync = async (connectionId) => {
    try {
      setError(null);
      
      const data = await authFetch(`/api/plaid/sync/${connectionId}`, {
        method: 'POST'
      });
      
      if (data.success || data.transactions_synced !== undefined) {
        setSuccessModal({
          isOpen: true,
          message: `✅ Sync completed! ${data.transactions_synced || 0} transactions imported.`
        });
        loadConnections(); // Refresh the list
      } else {
        throw new Error('Sync failed');
      }
    } catch (err) {
      setError(`Sync failed: ${err.message}`);
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusBadge = (status) => {
    const statusMap = {
      pending: { class: 'status-pending', text: 'Pending Authentication' },
      connected: { class: 'status-connected', text: 'Connected' },
      expired: { class: 'status-expired', text: 'Expired' },
      failed: { class: 'status-failed', text: 'Failed' }
    };
    
    const statusInfo = statusMap[status] || { class: 'status-unknown', text: status };
    return <span className={`status-badge ${statusInfo.class}`}>{statusInfo.text}</span>;
  };

  if (loading) {
    return <div className="loading">Loading bank connections...</div>;
  }

  return (
    <div className="bank-connections">
      <div className="page-header">
        <h1>Bank Connections</h1>
        <p>Connect your bank accounts via Plaid to automatically import transactions</p>
        <PlaidLink 
          onSuccess={(data) => {
            console.log('Plaid bank connected!', data);
            setError('');
            setSuccessModal({
              isOpen: true,
              message: '🎉 Bank successfully connected! Your accounts are now linked.'
            });
            loadConnections(); // Refresh connections list
          }}
          onExit={(err) => {
            if (err) {
              setError('Plaid connection cancelled or failed');
            }
          }}
        />
      </div>

      {error && (
        <div className="alert alert-error">
          {error}
        </div>
      )}

      <div className="connections-list">
        {connections.length === 0 ? (
          <div className="empty-state">
            <h3>No Bank Connections</h3>
            <p>Connect your bank account via Plaid to start automatically importing transactions.</p>
            <div style={{ display: 'flex', gap: '10px', justifyContent: 'center', marginTop: '20px' }}>
              <PlaidLink 
                onSuccess={(data) => {
                  console.log('Plaid bank connected!', data);
                  setError('');
                  setSuccessModal({
                    isOpen: true,
                    message: '🎉 Bank successfully connected! Your accounts are now linked.'
                  });
                  loadConnections();
                }}
                onExit={(err) => {
                  if (err) {
                    setError('Plaid connection cancelled or failed');
                  }
                }}
              />
            </div>
          </div>
        ) : (
          connections.map(connection => (
            <div key={connection.id} className="connection-card">
              <div className="connection-header">
                <div className="bank-info">
                  <h3>{connection.bank_name}</h3>
                  {getStatusBadge(connection.status)}
                </div>
                <div style={{ display: 'flex', gap: '10px' }}>
                  <button 
                    className="btn-primary btn-sm"
                    onClick={() => handleSync(connection.id)}
                    title="Sync transactions now"
                  >
                    🔄 Sync Now
                  </button>
                  <button 
                    className="btn-danger btn-sm"
                    onClick={() => handleDisconnect(connection.id)}
                  >
                    Disconnect
                  </button>
                </div>
              </div>
              
              <div className="connection-details">
                <div className="detail-row">
                  <span className="label">Connected:</span>
                  <span>{formatDate(connection.created_at)}</span>
                </div>
                
                {connection.consent_valid_until && (
                  <div className="detail-row">
                    <span className="label">Consent Valid Until:</span>
                    <span>{formatDate(connection.consent_valid_until)}</span>
                  </div>
                )}
                
                {connection.last_sync_at && (
                  <div className="detail-row">
                    <span className="label">Last Sync:</span>
                    <span>{formatDate(connection.last_sync_at)}</span>
                  </div>
                )}
                
                <div className="detail-row">
                  <span className="label">Sync Frequency:</span>
                  <span>{connection.frequency_per_day} times per day</span>
                </div>
                
                <div className="detail-row">
                  <span className="label">Total Syncs:</span>
                  <span>{connection.sync_count}</span>
                </div>
              </div>

              {connection.linked_accounts && connection.linked_accounts.length > 0 && (
                <div className="linked-accounts">
                  <h4>Linked Accounts ({connection.linked_accounts.length})</h4>
                  {connection.linked_accounts.map(account => (
                    <div key={account.id} className="account-item">
                      <span className="account-name">{account.account_name}</span>
                      <span className="account-iban">{account.iban}</span>
                      <span className="account-currency">{account.currency}</span>
                      {account.last_transaction_sync && (
                        <span className="sync-time">
                          Last sync: {formatDate(account.last_transaction_sync)}
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* Disconnect Confirmation Modal */}
      <ConfirmModal
        isOpen={disconnectConfirm.isOpen}
        onClose={() => setDisconnectConfirm({ isOpen: false, connectionId: null, bankName: '' })}
        onConfirm={confirmDisconnect}
        title="Disconnect Bank"
        message={`Are you sure you want to disconnect from "${disconnectConfirm.bankName}"? This will stop automatic transaction syncing.`}
        confirmText="Disconnect"
        cancelText="Cancel"
        danger={true}
      />

      {/* Success Modal */}
      <ConfirmModal
        isOpen={successModal.isOpen}
        onClose={() => setSuccessModal({ isOpen: false, message: '' })}
        onConfirm={() => setSuccessModal({ isOpen: false, message: '' })}
        title="Success"
        message={successModal.message}
        confirmText="OK"
        cancelText=""
        danger={false}
      />
    </div>
  );
}