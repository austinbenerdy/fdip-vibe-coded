import React from 'react';
import { Link } from 'react-router-dom';

const Home: React.FC = () => {
  return (
    <div className="home">
      {/* Hero Section */}
      <section className="hero">
        <div className="container">
          <div className="hero-content">
            <h1>Discover Amazing Stories in Progress</h1>
            <p>
              Connect with authors as they write their next masterpiece. 
              Read chapters as they're published and support your favorite writers.
            </p>
            <div className="hero-buttons">
              <Link to="/books" className="btn btn-primary">
                Browse Books
              </Link>
              <Link to="/register" className="btn btn-outline">
                Start Writing
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="how-it-works">
        <div className="container">
          <h2>How It Works</h2>
          <div className="grid grid-cols-3">
            <div className="card">
              <h3>1. Discover</h3>
              <p>Browse through thousands of books being written by talented authors.</p>
            </div>
            <div className="card">
              <h3>2. Read</h3>
              <p>Read chapters as they're published and follow your favorite stories.</p>
            </div>
            <div className="card">
              <h3>3. Support</h3>
              <p>Tip authors with tokens to show your appreciation and support their work.</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Home; 