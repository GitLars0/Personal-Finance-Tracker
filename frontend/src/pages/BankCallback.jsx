import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthFetch from '../hooks/useAuthFetch';

export default function BankCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const authFetch = useAuthFetch();
  const [status, setStatus] = useState('processing');
  const [message, setMessage] = useState('Processing bank connection...');

  useEffect(() => {
    const processCallback = async () => {
      try {
        const consentId = searchParams.get('consent_id');
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        const error = searchParams.get('error');

        if (error) {
          setStatus('error');
          setMessage('Bank connection was cancelled or failed. Please try again.');
          return;
        }

        let response;

        if (code && state) {
          // OAuth callback (Sparebank 1)
          response = await authFetch('/api/sparebank1/callback', {
            method: 'POST',
            body: JSON.stringify({
              code: code,
              state: state
            })
          });
        } else if (consentId) {
          // PSD2 callback (Sparebanken Norge, Bulder Bank)
          response = await authFetch('/api/banks/callback', {
            method: 'POST',
            body: JSON.stringify({
              consent_id: consentId,
              status: 'success'
            })
          });
        } else {
          setStatus('error');
          setMessage('Invalid callback - missing required parameters.');
          return;
        }

        if (response.status === 'connected' || response.status === 'valid') {
          setStatus('success');
          setMessage('Bank connection successful! Your transactions will be synced automatically.');
          
          // Redirect to bank connections page after 3 seconds
          setTimeout(() => {
            navigate('/banks');
          }, 3000);
        } else {
          setStatus('error');
          setMessage('Bank connection failed. Please try connecting again.');
        }
      } catch (err) {
        setStatus('error');
        setMessage('Failed to process bank connection: ' + err.message);
      }
    };

    processCallback();
  }, [searchParams, authFetch, navigate]);

  const getStatusIcon = () => {
    switch (status) {
      case 'processing':
        return (
          <div className="spinner">
            <div className="bounce1"></div>
            <div className="bounce2"></div>
            <div className="bounce3"></div>
          </div>
        );
      case 'success':
        return <div className="success-icon">✓</div>;
      case 'error':
        return <div className="error-icon">✗</div>;
      default:
        return null;
    }
  };

  return (
    <div className="bank-callback">
      <div className="callback-container">
        <div className={`status-card ${status}`}>
          {getStatusIcon()}
          <h2>{status === 'processing' ? 'Processing...' : 
               status === 'success' ? 'Connection Successful!' : 
               'Connection Failed'}</h2>
          <p>{message}</p>
          
          {status === 'success' && (
            <p className="redirect-info">
              Redirecting to bank connections page in 3 seconds...
            </p>
          )}
          
          {status === 'error' && (
            <div className="error-actions">
              <button 
                className="btn-primary"
                onClick={() => navigate('/banks')}
              >
                Try Again
              </button>
              <button 
                className="btn-secondary"
                onClick={() => navigate('/dashboard')}
              >
                Go to Dashboard
              </button>
            </div>
          )}
        </div>
      </div>
      
      <style jsx>{`
        .bank-callback {
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          padding: 2rem;
        }
        
        .callback-container {
          max-width: 500px;
          width: 100%;
        }
        
        .status-card {
          background: white;
          border-radius: 1rem;
          padding: 3rem 2rem;
          text-align: center;
          box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
        }
        
        .status-card h2 {
          margin: 1.5rem 0 1rem 0;
          font-size: 1.5rem;
          font-weight: 600;
        }
        
        .status-card.processing h2 {
          color: #1e40af;
        }
        
        .status-card.success h2 {
          color: #059669;
        }
        
        .status-card.error h2 {
          color: #dc2626;
        }
        
        .status-card p {
          color: #64748b;
          font-size: 1.1rem;
          margin: 0 0 1.5rem 0;
          line-height: 1.6;
        }
        
        .spinner {
          display: flex;
          justify-content: center;
          gap: 0.5rem;
        }
        
        .spinner > div {
          width: 12px;
          height: 12px;
          background-color: #1e40af;
          border-radius: 100%;
          display: inline-block;
          animation: sk-bouncedelay 1.4s infinite ease-in-out both;
        }
        
        .spinner .bounce1 {
          animation-delay: -0.32s;
        }
        
        .spinner .bounce2 {
          animation-delay: -0.16s;
        }
        
        @keyframes sk-bouncedelay {
          0%, 80%, 100% { 
            transform: scale(0);
          } 40% { 
            transform: scale(1.0);
          }
        }
        
        .success-icon {
          width: 60px;
          height: 60px;
          background: #059669;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-size: 2rem;
          font-weight: bold;
          margin: 0 auto;
        }
        
        .error-icon {
          width: 60px;
          height: 60px;
          background: #dc2626;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          font-size: 2rem;
          font-weight: bold;
          margin: 0 auto;
        }
        
        .redirect-info {
          font-size: 0.9rem !important;
          color: #059669 !important;
          font-weight: 500;
        }
        
        .error-actions {
          display: flex;
          gap: 1rem;
          justify-content: center;
          margin-top: 1rem;
        }
        
        .btn-primary {
          background: #1e40af;
          color: white;
          border: none;
          padding: 0.75rem 1.5rem;
          border-radius: 0.5rem;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s;
          text-decoration: none;
        }
        
        .btn-primary:hover {
          background: #1e3a8a;
        }
        
        .btn-secondary {
          background: #f1f5f9;
          color: #475569;
          border: 1px solid #cbd5e1;
          padding: 0.75rem 1.5rem;
          border-radius: 0.5rem;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s;
        }
        
        .btn-secondary:hover {
          background: #e2e8f0;
        }
        
        @media (max-width: 768px) {
          .bank-callback {
            padding: 1rem;
          }
          
          .status-card {
            padding: 2rem 1.5rem;
          }
          
          .error-actions {
            flex-direction: column;
          }
        }
      `}</style>
    </div>
  );
}