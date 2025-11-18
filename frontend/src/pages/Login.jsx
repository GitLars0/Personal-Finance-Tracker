import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import "../styles/Login.css";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [info, setInfo] = useState("");
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    if (params.get("registered") === "true") {
      setInfo("Account created successfully! Please log in.");
    }
  }, [location.search]);

  const handleLogin = async () => {
    if (!username || !password) {
      setError("Please enter both username and password");
      return;
    }

    setLoading(true);
    setError("");
    setInfo("");

    try {
      const res = await fetch("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      
      const data = await res.json();
      
      if (res.ok) {
        // Store the JWT token
        localStorage.setItem("token", data.token);
        
        // Store user info
        if (data.user) {
          localStorage.setItem("user", JSON.stringify(data.user));
        }
        
        console.log("✅ Login successful:", data.user?.username);
        
        // Notify App component about authentication change
        window.dispatchEvent(new Event('authChange'));
        
        // Redirect to dashboard
        navigate("/dashboard");
      } else {
        setError(data.error || "Login failed");
      }
    } catch (err) {
      console.error("Login error:", err);
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleLogin();
    }
  };

  return (
    <div className="login-container">
      <div className="login-box">
        <h2 className="login-title">Login</h2>
        
        {error && <p className="login-error">{error}</p>}
        {info && <p className="info-message">{info}</p>}
        
        <div className="login-input-group">
          <input
            className="login-input"
            placeholder="Username or Email"
            value={username}
            onChange={e => setUsername(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={loading}
            autoComplete="username"
          />
        </div>
        
        <div className="login-input-group">
          <input
            className="login-input"
            placeholder="Password"
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={loading}
            autoComplete="current-password"
          />
        </div>
        
        <button 
          className="login-button"
          onClick={handleLogin}
          disabled={loading}
        >
          {loading ? "Logging in..." : "Login"}
        </button>

        <p className="login-demo-info">
          <strong>Demo Credentials:</strong><br/>
          Username: <code>demo</code><br/>
          Password: <code>demo123</code>
        </p>

        <p className="login-register-link">
          Don't have an account? <a href="/register">Register here</a>
        </p>
      </div>
    </div>
  );
}