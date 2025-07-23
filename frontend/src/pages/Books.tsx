import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../utils/api';

interface Book {
  id: number;
  title: string;
  description?: string;
  cover_image_url?: string;
  genres: string[];
  author: {
    id: number;
    display_name: string;
  };
  chapters: Array<{
    id: number;
    title: string;
    chapter_number: number;
  }>;
}

const Books: React.FC = () => {
  const [books, setBooks] = useState<Book[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({
    genre: '',
    author: '',
    page: 1
  });
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 20,
    total: 0
  });

  useEffect(() => {
    fetchBooks();
  }, [filters]);

  const fetchBooks = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: filters.page.toString(),
        limit: '20'
      });
      
      if (filters.genre) params.append('genre', filters.genre);
      if (filters.author) params.append('author_id', filters.author);

      const response = await api.get(`/books?${params}`);
      setBooks(response.data.books);
      setPagination(response.data.pagination);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch books');
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1
    }));
  };

  const handlePageChange = (page: number) => {
    setFilters(prev => ({ ...prev, page }));
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="books-page">
      <div className="container">
        <div className="page-header">
          <h1>Browse Books</h1>
          <p>Discover amazing stories being written by talented authors</p>
        </div>

        {/* Filters */}
        <div className="filters">
          <div className="filter-group">
            <label htmlFor="genre" className="form-label">Genre</label>
            <select
              id="genre"
              value={filters.genre}
              onChange={(e) => handleFilterChange('genre', e.target.value)}
              className="form-input"
            >
              <option value="">All Genres</option>
              <option value="fantasy">Fantasy</option>
              <option value="scifi">Science Fiction</option>
              <option value="romance">Romance</option>
              <option value="mystery">Mystery</option>
              <option value="thriller">Thriller</option>
              <option value="historical">Historical</option>
              <option value="contemporary">Contemporary</option>
            </select>
          </div>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        {/* Books Grid */}
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
                <p className="author">by {book.author.display_name}</p>
                {book.description && (
                  <p className="description">{book.description}</p>
                )}
                <div className="book-meta">
                  <span className="chapters">
                    {book.chapters.length} chapter{book.chapters.length !== 1 ? 's' : ''}
                  </span>
                  {book.genres.length > 0 && (
                    <div className="genres">
                      {book.genres.slice(0, 2).map((genre, index) => (
                        <span key={index} className="genre-tag">{genre}</span>
                      ))}
                    </div>
                  )}
                </div>
                <Link to={`/books/${book.id}`} className="btn btn-primary">
                  Read Now
                </Link>
              </div>
            </div>
          ))}
        </div>

        {/* Pagination */}
        {pagination.total > pagination.limit && (
          <div className="pagination">
            {pagination.page > 1 && (
              <button
                onClick={() => handlePageChange(pagination.page - 1)}
                className="btn btn-outline"
              >
                Previous
              </button>
            )}
            <span className="page-info">
              Page {pagination.page} of {Math.ceil(pagination.total / pagination.limit)}
            </span>
            {pagination.page < Math.ceil(pagination.total / pagination.limit) && (
              <button
                onClick={() => handlePageChange(pagination.page + 1)}
                className="btn btn-outline"
              >
                Next
              </button>
            )}
          </div>
        )}

        {books.length === 0 && !loading && (
          <div className="empty-state">
            <p>No books found matching your criteria.</p>
            <button onClick={() => setFilters({ genre: '', author: '', page: 1 })} className="btn btn-primary">
              Clear Filters
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Books; 