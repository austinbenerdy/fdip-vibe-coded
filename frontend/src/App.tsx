import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import Header from './components/Header';
import PrivateRoute from './components/PrivateRoute';

// Pages
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Books from './pages/Books';
import BookDetail from './pages/BookDetail';
import ChapterDetail from './pages/ChapterDetail';
import BookForm from './pages/BookForm';
import Profile from './pages/Profile';
import Tokens from './pages/Tokens';
import BuyTokens from './pages/BuyTokens';

function App() {
  return (
    <AuthProvider>
      <div className="App">
        <Header />
        <main>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/books" element={<Books />} />
            <Route path="/books/new" element={<BookForm />} />
            <Route path="/books/:id" element={<BookDetail />} />
            <Route path="/books/:id/edit" element={<BookForm />} />
            <Route path="/my-books/:id" element={<BookDetail />} />
            <Route path="/my-books/:id/edit" element={<BookForm />} />
            <Route path="/chapters/:id" element={<ChapterDetail />} />
            <Route path="/buy-tokens" element={<BuyTokens />} />
            
            {/* Protected Routes */}
            <Route 
              path="/dashboard" 
              element={
                <PrivateRoute>
                  <Dashboard />
                </PrivateRoute>
              } 
            />
            <Route 
              path="/profile" 
              element={
                <PrivateRoute>
                  <Profile />
                </PrivateRoute>
              } 
            />
            <Route 
              path="/tokens" 
              element={
                <PrivateRoute>
                  <Tokens />
                </PrivateRoute>
              } 
            />
          </Routes>
        </main>
      </div>
    </AuthProvider>
  );
}

export default App; 