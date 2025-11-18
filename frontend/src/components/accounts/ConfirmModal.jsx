import React from 'react';

export default function ConfirmModal({ 
  isOpen, 
  onClose, 
  onConfirm, 
  title, 
  message, 
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  danger = false 
}) {
  if (!isOpen) return null;

  return (
    <div className="form-overlay" role="dialog" aria-modal="true" onClick={onClose}>
      <div className="modal" style={{ backgroundColor: '#ffffff', color: '#000', boxShadow: '0 8px 24px rgba(0,0,0,0.15)' }} onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{title}</h3>
          <button className="close-btn" onClick={onClose}>✖</button>
        </div>
        
        <div className="modal-body">
          <p style={{ marginBottom: '20px', fontSize: '16px', lineHeight: '1.5' }}>{message}</p>
          
          <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
            <button 
              type="button" 
              onClick={onClose} 
              className="btn-secondary"
              style={{ padding: '10px 20px', borderRadius: '4px', border: '1px solid #ddd', background: '#f5f5f5', cursor: 'pointer' }}
            >
              {cancelText}
            </button>
            <button 
              type="button" 
              onClick={() => {
                onConfirm();
                onClose();
              }} 
              className={danger ? 'btn-danger' : 'btn-primary'}
              style={{ 
                padding: '10px 20px', 
                borderRadius: '4px', 
                border: 'none', 
                background: danger ? '#dc3545' : '#007bff', 
                color: 'white', 
                cursor: 'pointer',
                fontWeight: '500'
              }}
            >
              {confirmText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
