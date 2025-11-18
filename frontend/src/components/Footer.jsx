import React from 'react';

function Footer() {
  return (
    <footer className="app-footer">
      <div className="footer-inner">
        <span>Personal Finance Tracker &copy; {new Date().getFullYear()}</span>
      </div>
    </footer>
  );
}

export default Footer;
