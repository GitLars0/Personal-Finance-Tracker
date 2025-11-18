import React from 'react';

function Modal({ children, onClose, title }) {
  return (
    <div className="form-overlay" role="dialog" aria-modal="true">
  <div className="modal" style={{ backgroundColor: '#ffffff', color: '#000', boxShadow: '0 8px 24px rgba(0,0,0,0.15)' }}>
        <div className="modal-header">
          <h3>{title}</h3>
          <button className="close-btn" onClick={onClose}>✖</button>
        </div>
        <div className="modal-body">{children}</div>
      </div>
    </div>
  );
}

export default Modal;
