import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import api from '../utils/api';

interface UserProfile {
  id: number;
  username: string;
  email: string;
  display_name: string;
  bio?: string;
  role: string;
  token_balance: number;
  created_at: string;
}

const Profile: React.FC = () => {
  const { user, updateProfile } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [promoting, setPromoting] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [formData, setFormData] = useState({
    display_name: '',
    bio: '',
    email: ''
  });

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await api.get('/profile');
        setProfile(response.data.user);
        setFormData({
          display_name: response.data.user.display_name,
          bio: response.data.user.bio || '',
          email: response.data.user.email
        });
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch profile');
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      const response = await api.put('/profile', formData);
      setProfile(response.data.user);
      updateProfile(response.data.user);
      setEditing(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update profile');
    }
  };

  const handlePromoteToAuthor = async () => {
    setPromoting(true);
    setError('');
    setSuccessMessage('');

    try {
      const response = await api.post('/profile/promote');
      
      // Update the token if a new one is provided
      if (response.data.token) {
        localStorage.setItem('token', response.data.token);
        // Update the auth context with new token
        window.location.reload(); // Simple way to refresh the auth context
      }
      
      setProfile(response.data.user);
      updateProfile(response.data.user);
      setSuccessMessage(response.data.message);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to promote to author');
    } finally {
      setPromoting(false);
    }
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error && !profile) {
    return (
      <div className="container">
        <div className="alert alert-error">{error}</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="container">
        <div className="alert alert-error">Profile not found</div>
      </div>
    );
  }

  return (
    <div className="profile-page">
      <div className="container">
        <div className="profile-header">
          <h1>Profile</h1>
          {!editing && (
            <button onClick={() => setEditing(true)} className="btn btn-primary">
              Edit Profile
            </button>
          )}
        </div>

        {error && <div className="alert alert-error">{error}</div>}
        {successMessage && <div className="alert alert-success">{successMessage}</div>}

        <div className="profile-content">
          <div className="profile-section">
            <h2>Account Information</h2>
            <div className="profile-info">
              <div className="info-item">
                <label>Username:</label>
                <span>{profile.username}</span>
              </div>
              <div className="info-item">
                <label>Role:</label>
                <span className="role-badge">{profile.role}</span>
              </div>
              <div className="info-item">
                <label>Member Since:</label>
                <span>{new Date(profile.created_at).toLocaleDateString()}</span>
              </div>
              <div className="info-item">
                <label>Token Balance:</label>
                <span className="token-balance">💰 {profile.token_balance} tokens</span>
              </div>
            </div>

            {/* Self-Promotion Section for Readers */}
            {profile.role === 'reader' && (
              <div className="promotion-section">
                <div className="card">
                  <h3>🚀 Become an Author</h3>
                  <p>
                    Ready to share your stories with the world? Promote yourself to author status 
                    and start creating books and chapters that readers can enjoy and support with tips.
                  </p>
                  <div className="author-benefits">
                    <h4>Author Benefits:</h4>
                    <ul>
                      <li>✨ Create and publish your own books</li>
                      <li>📝 Write chapters and manage your content</li>
                      <li>💰 Earn tokens from reader tips</li>
                      <li>📊 Access author dashboard and analytics</li>
                      <li>👥 Build a following of readers</li>
                    </ul>
                  </div>
                  <button
                    onClick={handlePromoteToAuthor}
                    disabled={promoting}
                    className="btn btn-primary btn-large"
                  >
                    {promoting ? 'Promoting...' : '🚀 Promote to Author'}
                  </button>
                </div>
              </div>
            )}
          </div>

          {editing ? (
            <div className="profile-section">
              <h2>Edit Profile</h2>
              <form onSubmit={handleSubmit}>
                <div className="form-group">
                  <label htmlFor="display_name" className="form-label">Display Name</label>
                  <input
                    type="text"
                    id="display_name"
                    name="display_name"
                    value={formData.display_name}
                    onChange={handleChange}
                    className="form-input"
                    required
                    maxLength={100}
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="email" className="form-label">Email</label>
                  <input
                    type="email"
                    id="email"
                    name="email"
                    value={formData.email}
                    onChange={handleChange}
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="bio" className="form-label">Bio</label>
                  <textarea
                    id="bio"
                    name="bio"
                    value={formData.bio}
                    onChange={handleChange}
                    className="form-input form-textarea"
                    rows={4}
                    maxLength={500}
                  />
                </div>

                <div className="form-actions">
                  <button type="submit" className="btn btn-primary">
                    Save Changes
                  </button>
                  <button 
                    type="button" 
                    onClick={() => setEditing(false)}
                    className="btn btn-outline"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          ) : (
            <div className="profile-section">
              <h2>Personal Information</h2>
              <div className="profile-info">
                <div className="info-item">
                  <label>Display Name:</label>
                  <span>{profile.display_name}</span>
                </div>
                <div className="info-item">
                  <label>Email:</label>
                  <span>{profile.email}</span>
                </div>
                {profile.bio && (
                  <div className="info-item">
                    <label>Bio:</label>
                    <p className="bio-text">{profile.bio}</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Quick Actions */}
          <div className="profile-section">
            <h2>Quick Actions</h2>
            <div className="actions-grid">
              <a href="/tokens" className="card action-card">
                <h3>💰 Manage Tokens</h3>
                <p>Buy tokens or cash out earnings</p>
              </a>
              {(profile.role === 'author' || profile.role === 'admin') && (
                <a href="/dashboard" className="card action-card">
                  <h3>📚 Author Dashboard</h3>
                  <p>Manage your books and chapters</p>
                </a>
              )}
              <a href="/books" className="card action-card">
                <h3>📖 Browse Books</h3>
                <p>Discover new stories to read</p>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile; 