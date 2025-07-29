import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../utils/api';

interface Chapter {
  id: number;
  title: string;
  chapter_number: number;
  is_published: boolean;
  is_private: boolean;
  word_count: number;
}

interface Book {
  id: number;
  title: string;
  description?: string;
  cover_image_url?: string;
  genres: string[];
  is_published: boolean;
  author: {
    id: number;
    display_name: string;
    bio?: string;
  };
  chapters: Chapter[];
}

const BookDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const [book, setBook] = useState<Book | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchBook = async () => {
      try {
        const response = await api.get(`/my-books/${id}`);
        setBook(response.data.book);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch book');
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchBook();
    }
  }, [id]);

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error || !book) {
    return (
      <div className="container">
        <div className="alert alert-error">
          {error || 'Book not found'}
        </div>
      </div>
    );
  }

  const publishedChapters = book.chapters.filter(chapter => 
    chapter.is_published && !chapter.is_private
  );

  const isAuthor = user && (user.id === book.author.id || user.role === 'admin');

  return (
    <div className="book-detail">
      <div className="container">
        {/* Book Header */}
        <div className="book-header">
          <div className="book-cover-section">
            {book.cover_image_url && (
              <img 
                src={book.cover_image_url} 
                alt={book.title}
                className="book-cover-large"
              />
            )}
          </div>
          <div className="book-info-section">
            <h1>{book.title}</h1>
            <p className="author">by {book.author.display_name}</p>
            
            {book.description && (
              <p className="description">{book.description}</p>
            )}

            <div className="book-meta">
              <div className="meta-item">
                <span className="label">Status:</span>
                <span className="value">
                  {book.is_published ? '📖 Published' : '📝 Draft'}
                </span>
              </div>
              <div className="meta-item">
                <span className="label">Chapters:</span>
                <span className="value">
                  {publishedChapters.length} published
                  {book.chapters.length > publishedChapters.length && 
                    ` (${book.chapters.length - publishedChapters.length} private)`
                  }
                </span>
              </div>
              {book.genres.length > 0 && (
                <div className="meta-item">
                  <span className="label">Genres:</span>
                  <div className="genres">
                    {book.genres.map((genre, index) => (
                      <span key={index} className="genre-tag">{genre}</span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {isAuthor && (
              <div className="author-actions">
                <Link to={`/my-books/${book.id}/edit`} className="btn btn-primary">
                  Edit Book
                </Link>
                <Link to={`/my-books/${book.id}/chapters/new`} className="btn btn-primary">
                  Add Chapter
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Author Info */}
        <div className="author-section">
          <h2>About the Author</h2>
          <div className="author-card">
            <h3>{book.author.display_name}</h3>
            {book.author.bio && <p>{book.author.bio}</p>}
            <Link to={`/authors/${book.author.id}`} className="btn btn-outline">
              View Author Profile
            </Link>
          </div>
        </div>

        {/* Chapters */}
        <div className="chapters-section">
          <h2>Chapters</h2>
          {publishedChapters.length === 0 ? (
            <div className="empty-state">
              <p>No published chapters yet.</p>
              {isAuthor && (
                <Link to={`/my-books/${book.id}/chapters/new`} className="btn btn-primary">
                  Add First Chapter
                </Link>
              )}
            </div>
          ) : (
            <div className="chapters-list">
              {publishedChapters.map((chapter) => (
                <div key={chapter.id} className="chapter-item">
                  <div className="chapter-info">
                    <h3>
                      Chapter {chapter.chapter_number}: {chapter.title}
                    </h3>
                    <p className="word-count">{chapter.word_count} words</p>
                  </div>
                  <div className="chapter-actions">
                    <Link to={`/chapters/${chapter.id}`} className="btn btn-primary">
                      Read Chapter
                    </Link>
                    {isAuthor && (
                      <Link to={`/chapters/${chapter.id}/edit`} className="btn btn-outline">
                        Edit
                      </Link>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default BookDetail; 