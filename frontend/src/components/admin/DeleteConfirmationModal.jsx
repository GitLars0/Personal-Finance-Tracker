import React from 'react';
import Modal from '../Modal';

const DeleteConfirmationModal = ({ 
  isOpen, 
  onClose, 
  onConfirm, 
  title, 
  message, 
  itemName,
  isDeleting = false 
}) => {
  if (!isOpen) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title}>
      <div className="delete-confirmation-modal">
        <div className="delete-message">
          <div className="warning-icon">⚠️</div>
          <p>{message}</p>
          {itemName && (
            <div className="item-highlight">
              <strong>"{itemName}"</strong>
            </div>
          )}
          <p className="warning-text">
            This action cannot be undone.
          </p>
        </div>

        <div className="modal-actions">
          <button 
            onClick={onClose} 
            className="btn-secondary"
            disabled={isDeleting}
          >
            Cancel
          </button>
          <button 
            onClick={onConfirm} 
            className="btn-danger"
            disabled={isDeleting}
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </div>

      <style jsx>{`
        .delete-confirmation-modal {
          padding: 20px;
          text-align: center;
        }

        .delete-message {
          margin-bottom: 30px;
        }

        .warning-icon {
          font-size: 48px;
          margin-bottom: 15px;
        }

        .delete-message p {
          margin: 10px 0;
          color: #666;
          line-height: 1.5;
        }

        .item-highlight {
          margin: 15px 0;
          padding: 10px;
          background-color: #fff3cd;
          border: 1px solid #ffeaa7;
          border-radius: 4px;
        }

        .item-highlight strong {
          color: #856404;
        }

        .warning-text {
          color: #dc3545 !important;
          font-weight: 500;
        }

        .modal-actions {
          display: flex;
          gap: 10px;
          justify-content: center;
        }

        .btn-secondary {
          padding: 10px 20px;
          border: 1px solid #ddd;
          background: white;
          color: #666;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
        }

        .btn-secondary:hover:not(:disabled) {
          background: #f8f9fa;
        }

        .btn-danger {
          padding: 10px 20px;
          border: 1px solid #dc3545;
          background: #dc3545;
          color: white;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
        }

        .btn-danger:hover:not(:disabled) {
          background: #c82333;
        }

        .btn-secondary:disabled,
        .btn-danger:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }
      `}</style>
    </Modal>
  );
};

export default DeleteConfirmationModal;