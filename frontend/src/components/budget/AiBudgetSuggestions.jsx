import React, { useState } from 'react';
import useAuthFetch from '../../hooks/useAuthFetch';
import '../../styles/AiBudgetSuggestions.css';

const AiBudgetSuggestions = ({ onApplySuggestion, targetMonth, targetYear, disabled = false }) => {
  const authFetch = useAuthFetch();
  const [predictions, setPredictions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const fetchPredictions = async () => {
    if (disabled) return;
    
    setLoading(true);
    setError(null);
    
    try {
      // Use current date if target date not provided
      const month = targetMonth || new Date().getMonth() + 1;
      const year = targetYear || new Date().getFullYear();
      
      const params = new URLSearchParams({
        target_month: month.toString(),
        target_year: year.toString(),
        historical_months: '12'
      });
      
      const response = await authFetch(`/api/ai/budget-predictions?${params}`);
      
      if (response && response.predictions) {
        setPredictions(response.predictions);
        if (response.predictions.length > 0) {
          setShowSuggestions(true);
        }
      } else {
        setError('No predictions available');
      }
    } catch (err) {
      console.error('AI prediction error:', err);
      setError(err.message || 'Failed to fetch AI budget suggestions');
    } finally {
      setLoading(false);
    }
  };

  const handleApplySuggestion = (prediction) => {
    if (onApplySuggestion) {
      onApplySuggestion({
        category_id: prediction.category_id,
        category_name: prediction.category_name,
        suggested_amount: prediction.predicted_amount_dollars
      });
    }
  };

  const handleApplyAllSuggestions = () => {
    if (onApplySuggestion && predictions.length > 0) {
      // Apply all suggestions at once by calling the callback for each prediction
      predictions.forEach(prediction => {
        onApplySuggestion({
          category_id: prediction.category_id,
          category_name: prediction.category_name,
          suggested_amount: prediction.predicted_amount_dollars
        });
      });
    }
  };

  const getConfidenceColor = (confidence) => {
    if (confidence >= 0.8) return 'high-confidence';
    if (confidence >= 0.6) return 'medium-confidence';
    return 'low-confidence';
  };

  const getTrendIcon = (trend) => {
    switch (trend) {
      case 'increasing': return '📈';
      case 'decreasing': return '📉';
      default: return '➡️';
    }
  };

  if (disabled) return null;

  return (
    <div className="ai-budget-suggestions">
      <div className="suggestions-header">
        <h4>🤖 AI Budget Suggestions</h4>
        <div className="suggestions-actions">
          {!showSuggestions && (
            <button 
              type="button"
              className="fetch-suggestions-btn"
              onClick={fetchPredictions}
              disabled={loading}
            >
              {loading ? '🔄 Analyzing...' : '✨ Get AI Suggestions'}
            </button>
          )}
          {showSuggestions && predictions.length > 0 && (
            <>
              <button 
                type="button"
                className="refresh-suggestions-btn"
                onClick={fetchPredictions}
                disabled={loading}
                title="Refresh suggestions"
              >
                🔄
              </button>
              <button 
                type="button"
                className="apply-all-btn"
                onClick={handleApplyAllSuggestions}
                title="Apply all suggestions"
              >
                Apply All
              </button>
              <button 
                type="button"
                className="hide-suggestions-btn"
                onClick={() => setShowSuggestions(false)}
                title="Hide suggestions"
              >
                ✕
              </button>
            </>
          )}
        </div>
      </div>

      {error && (
        <div className="ai-error">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {loading && (
        <div className="ai-loading">
          <div className="loading-spinner">🤖</div>
          <span>Analyzing your spending patterns...</span>
        </div>
      )}

      {showSuggestions && predictions.length === 0 && !loading && !error && (
        <div className="no-predictions">
          <span className="info-icon">ℹ️</span>
          <span>No historical data available for AI predictions. Create a few budgets first!</span>
        </div>
      )}

      {showSuggestions && predictions.length > 0 && (
        <div className="suggestions-list">
          <div className="suggestions-intro">
            <p>Based on your spending history, here are AI-generated budget suggestions:</p>
          </div>
          
          {predictions.map((prediction, index) => (
            <div key={prediction.category_id} className={`suggestion-item ${getConfidenceColor(prediction.confidence_score)}`}>
              <div className="suggestion-header">
                <div className="category-info">
                  <span className="category-name">{prediction.category_name}</span>
                  <span className="trend-indicator" title={`Trend: ${prediction.trend_direction}`}>
                    {getTrendIcon(prediction.trend_direction)}
                  </span>
                </div>
                <div className="suggestion-amount">
                  <span className="amount">${prediction.predicted_amount_dollars.toFixed(2)}</span>
                  <span className={`confidence ${getConfidenceColor(prediction.confidence_score)}`}>
                    {Math.round(prediction.confidence_score * 100)}% confidence
                  </span>
                </div>
              </div>
              
              <div className="suggestion-details">
                <div className="historical-comparison">
                  <div className="historical-averages">
                    <span className="historical-budget-avg">
                      Budget avg: ${prediction.historical_avg_dollars.toFixed(2)}
                    </span>
                    {prediction.historical_spending_avg_dollars !== undefined && (
                      <span className="historical-spending-avg">
                        Spending avg: ${prediction.historical_spending_avg_dollars.toFixed(2)}
                      </span>
                    )}
                  </div>
                  {Math.abs(prediction.predicted_amount_dollars - prediction.historical_avg_dollars) >= 0.01 && (
                    <span className="difference">
                      {(() => {
                        const diff = prediction.predicted_amount_dollars - prediction.historical_avg_dollars;
                        const sign = diff >= 0 ? '+' : '';
                        return `(${sign}$${diff.toFixed(2)} vs budget avg)`;
                      })()}
                    </span>
                  )}
                </div>
                
                <div className="reasoning">
                  <span className="reasoning-text">{prediction.reasoning}</span>
                </div>
              </div>
              
              <div className="suggestion-actions">
                <button 
                  type="button"
                  className="apply-suggestion-btn"
                  onClick={() => handleApplySuggestion(prediction)}
                  title="Apply this suggestion"
                >
                  Use This Amount
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AiBudgetSuggestions;