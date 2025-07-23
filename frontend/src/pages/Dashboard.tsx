import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../utils/api';

interface Book {
  id: number;
  title: string;
  description?: string;
  cover_image_url?: string;
  is_published: boolean;
  chapters: Array<{
    id: number;
    title: string;
    chapter_number: number;
    is_published: boolean;
    is_private: boolean;
  }>;
}

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [books, setBooks] = useState<Book[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchBooks = async () => {
      try {
        const response = await api.get('/books');
        setBooks(response.data.books);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch books');
      } finally {
        setLoading(false);
      }
    };

    fetchBooks();
  }, []);

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container">
        <div className="alert alert-error">{error}</div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="container">
        <div className="dashboard-header">
          <h1>Author Dashboard</h1>
          <Link to="/books/new" className="btn btn-primary">
            Create New Book
          </Link>
        </div>

        {/* Stats */}
        <div className="stats-grid">
          <div className="card">
            <h3>Total Books</h3>
            <p className="stat-number">{books.length}</p>
          </div>
          <div className="card">
            <h3>Published Books</h3>
            <p className="stat-number">{books.filter(b => b.is_published).length}</p>
          </div>
          <div className="card">
            <h3>Total Chapters</h3>
            <p className="stat-number">
              {books.reduce((total, book) => total + book.chapters.length, 0)}
            </p>
          </div>
          <div className="card">
            <h3>Token Balance</h3>
            <p className="stat-number">💰 {user?.token_balance || 0}</p>
          </div>
        </div>

        {/* Books */}
        <div className="books-section">
          <h2>Your Books</h2>
          {books.length === 0 ? (
            <div className="empty-state">
              <p>You haven't created any books yet.</p>
              <Link to="/books/new" className="btn btn-primary">
                Create Your First Book
              </Link>
            </div>
          ) : (
            <div className="books-grid">
              {books.map((book) => (
                <div key={book.id} className="card book-card">
                  {book.cover_image_url && (
                    <img 
                      src={book.cover_image_url} 
                      alt={book.title}
                      className="book-cover"
                    />
                  )}
                  <div className="book-info">
                    <h3>{book.title}</h3>
                    {book.description && (
                      <p className="description">{book.description}</p>
                    )}
                    <div className="book-stats">
                      <span className="status">
                        {book.is_published ? '📖 Published' : '📝 Draft'}
                      </span>
                      <span className="chapters">
                        {book.chapters.length} chapter{book.chapters.length !== 1 ? 's' : ''}
                      </span>
                    </div>
                    <div className="book-actions">
                      <Link to={`/books/${book.id}`} className="btn btn-outline">
                        View
                      </Link>
                      <Link to={`/books/${book.id}/edit`} className="btn btn-primary">
                        Edit
                      </Link>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Quick Actions */}
        <div className="quick-actions">
          <h2>Quick Actions</h2>
          <div className="actions-grid">
            <Link to="/books/new" className="card action-card">
              <h3>📚 Create New Book</h3>
              <p>Start writing a new story</p>
            </Link>
            <Link to="/profile" className="card action-card">
              <h3>👤 Edit Profile</h3>
              <p>Update your author profile</p>
            </Link>
            <Link to="/tokens" className="card action-card">
              <h3>💰 Manage Tokens</h3>
              <p>View earnings and cash out</p>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard; 