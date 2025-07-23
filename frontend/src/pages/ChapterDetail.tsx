import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../utils/api';
import ReactMarkdown from 'react-markdown';

interface Chapter {
  id: number;
  title: string;
  content: string;
  content_type: 'markdown' | 'html';
  chapter_number: number;
  word_count: number;
  image_url?: string;
  book: {
    id: number;
    title: string;
    author: {
      id: number;
      display_name: string;
    };
  };
}

const ChapterDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const [chapter, setChapter] = useState<Chapter | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [tipAmount, setTipAmount] = useState(10);
  const [tipping, setTipping] = useState(false);

  useEffect(() => {
    const fetchChapter = async () => {
      try {
        const response = await api.get(`/chapters/${id}`);
        setChapter(response.data.chapter);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch chapter');
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchChapter();
    }
  }, [id]);

  const handleTip = async () => {
    if (!user || !chapter) return;

    setTipping(true);
    try {
      await api.post('/tokens/tip', {
        chapter_id: chapter.id,
        amount: tipAmount
      });
      
      // Refresh user data to update token balance
      window.location.reload();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send tip');
    } finally {
      setTipping(false);
    }
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error || !chapter) {
    return (
      <div className="container">
        <div className="alert alert-error">
          {error || 'Chapter not found'}
        </div>
      </div>
    );
  }

  const isAuthor = user && (user.id === chapter.book.author.id || user.role === 'admin');

  return (
    <div className="chapter-detail">
      <div className="container">
        {/* Chapter Header */}
        <div className="chapter-header">
          <div className="breadcrumb">
            <Link to={`/books/${chapter.book.id}`}>{chapter.book.title}</Link>
            <span> / </span>
            <span>Chapter {chapter.chapter_number}</span>
          </div>
          
          <h1>Chapter {chapter.chapter_number}: {chapter.title}</h1>
          <p className="author">by {chapter.book.author.display_name}</p>
          
          <div className="chapter-meta">
            <span className="word-count">{chapter.word_count} words</span>
            {chapter.image_url && (
              <img 
                src={chapter.image_url} 
                alt={chapter.title}
                className="chapter-image"
              />
            )}
          </div>
        </div>

        {/* Chapter Content */}
        <div className="chapter-content">
          {chapter.content_type === 'markdown' ? (
            <ReactMarkdown>{chapter.content}</ReactMarkdown>
          ) : (
            <div dangerouslySetInnerHTML={{ __html: chapter.content }} />
          )}
        </div>

        {/* Tip Section */}
        {!isAuthor && user && (
          <div className="tip-section">
            <div className="card">
              <h3>Support the Author</h3>
              <p>If you enjoyed this chapter, consider tipping the author!</p>
              
              <div className="tip-options">
                <button
                  className={`tip-option ${tipAmount === 5 ? 'selected' : ''}`}
                  onClick={() => setTipAmount(5)}
                >
                  5 tokens
                </button>
                <button
                  className={`tip-option ${tipAmount === 10 ? 'selected' : ''}`}
                  onClick={() => setTipAmount(10)}
                >
                  10 tokens
                </button>
                <button
                  className={`tip-option ${tipAmount === 25 ? 'selected' : ''}`}
                  onClick={() => setTipAmount(25)}
                >
                  25 tokens
                </button>
                <button
                  className={`tip-option ${tipAmount === 50 ? 'selected' : ''}`}
                  onClick={() => setTipAmount(50)}
                >
                  50 tokens
                </button>
              </div>

              <div className="tip-custom">
                <label htmlFor="custom-tip" className="form-label">Custom amount:</label>
                <input
                  type="number"
                  id="custom-tip"
                  value={tipAmount}
                  onChange={(e) => setTipAmount(parseInt(e.target.value) || 0)}
                  className="form-input"
                  min="1"
                />
              </div>

              <button
                onClick={handleTip}
                disabled={tipping || tipAmount <= 0 || (user.token_balance || 0) < tipAmount}
                className="btn btn-primary"
              >
                {tipping ? 'Sending Tip...' : `Tip ${tipAmount} tokens`}
              </button>

              {user.token_balance !== undefined && (
                <p className="token-balance-info">
                  Your balance: {user.token_balance} tokens
                </p>
              )}
            </div>
          </div>
        )}

        {/* Navigation */}
        <div className="chapter-navigation">
          <Link to={`/books/${chapter.book.id}`} className="btn btn-outline">
            ← Back to Book
          </Link>
        </div>
      </div>
    </div>
  );
};

export default ChapterDetail; 