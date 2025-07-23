import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../utils/api';

interface BookFormData {
  title: string;
  description: string;
  genres: string[];
  cover_image_url: string;
  is_published: boolean;
}

const BookForm: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState<BookFormData>({
    title: '',
    description: '',
    genres: [],
    cover_image_url: '',
    is_published: false
  });

  const availableGenres = [
    'fantasy', 'scifi', 'romance', 'mystery', 'thriller', 
    'historical', 'contemporary', 'horror', 'adventure', 'literary'
  ];

  useEffect(() => {
    if (id) {
      fetchBook();
    }
  }, [id]);

  const fetchBook = async () => {
    try {
      setLoading(true);
      const response = await api.get(`/books/${id}`);
      const book = response.data.book;
      setFormData({
        title: book.title,
        description: book.description || '',
        genres: book.genres || [],
        cover_image_url: book.cover_image_url || '',
        is_published: book.is_published
      });
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch book');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleGenreChange = (genre: string) => {
    setFormData(prev => ({
      ...prev,
      genres: prev.genres.includes(genre)
        ? prev.genres.filter(g => g !== genre)
        : [...prev.genres, genre]
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (id) {
        await api.put(`/books/${id}`, formData);
      } else {
        await api.post('/books', formData);
      }
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save book');
    } finally {
      setLoading(false);
    }
  };

  if (loading && id) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="book-form-page">
      <div className="container">
        <div className="page-header">
          <h1>{id ? 'Edit Book' : 'Create New Book'}</h1>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        <div className="card">
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="title" className="form-label">Title *</label>
              <input
                type="text"
                id="title"
                name="title"
                value={formData.title}
                onChange={handleChange}
                className="form-input"
                required
                maxLength={200}
              />
            </div>

            <div className="form-group">
              <label htmlFor="description" className="form-label">Description</label>
              <textarea
                id="description"
                name="description"
                value={formData.description}
                onChange={handleChange}
                className="form-input form-textarea"
                rows={4}
                maxLength={1000}
                placeholder="Tell readers what your book is about..."
              />
            </div>

            <div className="form-group">
              <label className="form-label">Genres</label>
              <div className="genres-grid">
                {availableGenres.map((genre) => (
                  <label key={genre} className="genre-checkbox">
                    <input
                      type="checkbox"
                      checked={formData.genres.includes(genre)}
                      onChange={() => handleGenreChange(genre)}
                    />
                    <span className="genre-label">{genre}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="cover_image_url" className="form-label">Cover Image URL</label>
              <input
                type="url"
                id="cover_image_url"
                name="cover_image_url"
                value={formData.cover_image_url}
                onChange={handleChange}
                className="form-input"
                placeholder="https://example.com/cover.jpg"
              />
              {formData.cover_image_url && (
                <img 
                  src={formData.cover_image_url} 
                  alt="Cover preview"
                  className="cover-preview"
                  onError={(e) => {
                    e.currentTarget.style.display = 'none';
                  }}
                />
              )}
            </div>

            <div className="form-group">
              <label className="form-label">
                <input
                  type="checkbox"
                  checked={formData.is_published}
                  onChange={(e) => setFormData(prev => ({ ...prev, is_published: e.target.checked }))}
                />
                Publish this book (make it visible to readers)
              </label>
            </div>

            <div className="form-actions">
              <button 
                type="submit" 
                className="btn btn-primary"
                disabled={loading}
              >
                {loading ? 'Saving...' : (id ? 'Update Book' : 'Create Book')}
              </button>
              <button 
                type="button" 
                onClick={() => navigate('/dashboard')}
                className="btn btn-outline"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default BookForm; 