import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const Header: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <header className="header">
      <div className="container">
        <div className="header-content">
          <Link to="/" className="logo">
            First Draft In Progress
          </Link>
          
          <nav className="nav">
            <Link to="/books" className="nav-link">
              Browse Books
            </Link>
            <Link to="/authors" className="nav-link">
              Authors
            </Link>
            
            {user ? (
              <>
                {user.role === 'author' || user.role === 'admin' ? (
                  <Link to="/dashboard" className="nav-link">
                    Dashboard
                  </Link>
                ) : null}
                <Link to="/profile" className="nav-link">
                  Profile
                </Link>
                <Link to="/tokens" className="nav-link">
                  Tokens
                </Link>
                <Link to="/buy-tokens" className="nav-link">
                  Buy Tokens
                </Link>
                <div className="user-menu">
                  <span className="user-name">{user.display_name}</span>
                  <span className="token-balance">💰 {user.token_balance || 0} tokens</span>
                  <button onClick={handleLogout} className="btn btn-outline">
                    Logout
                  </button>
                </div>
              </>
            ) : (
              <>
                <Link to="/login" className="btn btn-outline">
                  Login
                </Link>
                <Link to="/register" className="btn btn-primary">
                  Register
                </Link>
              </>
            )}
          </nav>
        </div>
      </div>
    </header>
  );
};

export default Header; 