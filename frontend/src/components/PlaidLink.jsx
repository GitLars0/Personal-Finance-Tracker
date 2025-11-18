import { usePlaidLink } from 'react-plaid-link';
import { useState, useEffect } from 'react';
import useAuthFetch from '../hooks/useAuthFetch';

function PlaidLink({ onSuccess, onExit }) {
  const [linkToken, setLinkToken] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const authFetch = useAuthFetch();

  useEffect(() => {
    // Get link token from backend
    const fetchLinkToken = async () => {
      try {
        setLoading(true);
        setError('');
        const data = await authFetch('/api/plaid/create_link_token', {
          method: 'POST',
        });
        setLinkToken(data.link_token);
      } catch (err) {
        setError(err.message || 'Failed to initialize Plaid');
        console.error('Failed to create link token:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchLinkToken();
  }, [authFetch]);

  const config = {
    token: linkToken,
    onSuccess: async (public_token, metadata) => {
      try {
        // Send public_token to backend to exchange for access_token
        const data = await authFetch('/api/plaid/exchange_public_token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            public_token,
            bank_name: metadata.institution?.name || 'Plaid Bank',
          }),
        });

        if (onSuccess) {
          onSuccess(data);
        }
      } catch (err) {
        console.error('Failed to exchange token:', err);
        setError(err.message || 'Failed to connect bank');
      }
    },
    onExit: (err, metadata) => {
      if (err) {
        console.error('Plaid Link error:', err);
        setError(err.display_message || 'Failed to connect bank');
      }
      if (onExit) {
        onExit(err, metadata);
      }
    },
  };

  const { open, ready } = usePlaidLink(config);

  if (error) {
    return (
      <div style={{ color: 'red', padding: '10px', fontSize: '14px' }}>
        ⚠️ {error}
      </div>
    );
  }

  if (loading || !linkToken) {
    return (
      <button disabled style={{ opacity: 0.6, cursor: 'not-allowed' }}>
        🔄 Loading Plaid...
      </button>
    );
  }

  return (
    <button 
      onClick={() => open()} 
      disabled={!ready}
      style={{
        padding: '12px 24px',
        backgroundColor: ready ? '#000' : '#ccc',
        color: 'white',
        border: 'none',
        borderRadius: '8px',
        fontSize: '16px',
        fontWeight: '600',
        cursor: ready ? 'pointer' : 'not-allowed',
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        transition: 'all 0.2s',
      }}
      onMouseOver={(e) => {
        if (ready) {
          e.target.style.backgroundColor = '#333';
          e.target.style.transform = 'translateY(-2px)';
        }
      }}
      onMouseOut={(e) => {
        if (ready) {
          e.target.style.backgroundColor = '#000';
          e.target.style.transform = 'translateY(0)';
        }
      }}
    >
      <span style={{ fontSize: '20px' }}>🏦</span>
      Connect Bank with Plaid
    </button>
  );
}

export default PlaidLink;
