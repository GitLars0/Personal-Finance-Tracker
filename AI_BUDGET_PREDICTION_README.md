# AI Budget Prediction System

This document describes the new AI-powered budget prediction feature added to the Personal Finance Tracker.

## Overview

The AI system analyzes your historical budget and spending patterns to provide intelligent suggestions when creating new budgets. It uses machine learning to identify trends, seasonal patterns, and spending habits to predict appropriate budget amounts for each category.

## Architecture

The system consists of three main components:

1. **AI Service** (`ai-service/`): Python Flask application with ML algorithms
2. **Backend API** (`backend/controllers/ai_controller.go`): Go endpoints that interface with the AI service
3. **Frontend UI** (`frontend/src/components/budget/AiBudgetSuggestions.jsx`): React component that displays predictions

## Features

### AI Analysis
- **Historical Pattern Analysis**: Examines past budgets and actual spending
- **Trend Detection**: Identifies increasing, decreasing, or stable spending patterns
- **Seasonal Adjustments**: Applies month-specific adjustments (holidays, summer, etc.)
- **Confidence Scoring**: Provides reliability scores for each prediction
- **Smart Reasoning**: Explains why each amount was suggested

### User Experience
- **Intelligent Suggestions**: AI recommendations appear when creating new budgets
- **One-Click Apply**: Apply individual suggestions or all at once
- **Visual Confidence**: Color-coded confidence levels (high/medium/low)
- **Trend Indicators**: Visual icons showing spending trends (📈📉➡️)
- **Detailed Explanations**: Human-readable reasoning for each suggestion

## API Endpoints

### Backend Endpoints (Go)
- `GET /api/ai/budget-predictions` - Get AI-generated budget predictions
  - Query params: `target_month`, `target_year`, `historical_months`
- `GET /api/ai/spending-patterns` - Analyze spending patterns without predictions

### AI Service Endpoints (Python)
- `POST /predict-budget` - Generate budget predictions for a user
- `POST /analyze-patterns` - Analyze spending patterns
- `GET /health` - Health check endpoint

## Setup Instructions

### Prerequisites
- Docker and Docker Compose
- Existing Personal Finance Tracker with some historical budget data

### Installation

1. **Start the complete system:**
   ```bash
   docker-compose up -d
   ```

2. **Verify AI service is running:**
   ```bash
   curl http://localhost:5001/health
   ```

3. **Check backend connectivity:**
   ```bash
   curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/ai/budget-predictions
   ```

### Environment Variables

The following environment variables are configured in `docker-compose.yml`:

**Backend (app service):**
- `AI_SERVICE_HOST=ai-service`
- `AI_SERVICE_PORT=5001`

**AI Service:**
- `DB_HOST=db`
- `DB_USER=postgres`
- `DB_PASSWORD=password`
- `DB_NAME=finance_tracker`
- `DB_PORT=5432`

## Usage

### For End Users

1. **Create a Budget:**
   - Navigate to the Budget page
   - Click "Create Budget"
   - Select start and end dates

2. **Get AI Suggestions:**
   - Click "✨ Get AI Suggestions" button
   - The system will analyze your historical data
   - AI predictions will appear with confidence scores

3. **Apply Suggestions:**
   - Review each suggestion and its reasoning
   - Click "Use This Amount" for individual suggestions
   - Or click "Apply All" to use all suggestions at once

4. **Customize and Save:**
   - Modify suggested amounts as needed
   - Add additional categories manually
   - Save your budget

### AI Prediction Logic

The AI considers several factors:

1. **Historical Average**: Base amount from past budgets
2. **Trend Analysis**: Whether spending is increasing/decreasing
3. **Seasonal Factors**: Month-specific adjustments:
   - January: +10% (New Year expenses)
   - July-August: +15% (Summer activities)
   - November-December: +20-30% (Holiday season)
4. **Accuracy Score**: How consistent past budgets were with actual spending

### Confidence Levels

- **High (80%+)**: Green - Strong historical data, consistent patterns
- **Medium (60-79%)**: Orange - Some data available, moderate consistency
- **Low (<60%)**: Red - Limited data or high variance in past budgets

## Testing

### Manual Testing Steps

1. **Setup Test Data:**
   - Create at least 3 historical budgets with different months
   - Add some transactions in those periods
   - Ensure variety in categories and amounts

2. **Test AI Predictions:**
   ```bash
   # Test health endpoint
   curl http://localhost:5001/health
   
   # Test predictions (requires valid JWT token)
   curl -H "Authorization: Bearer YOUR_TOKEN" \
        "http://localhost:8080/api/ai/budget-predictions?target_month=4&target_year=2025"
   ```

3. **Frontend Testing:**
   - Navigate to Budget page
   - Click "Create Budget"
   - Set future dates
   - Click "Get AI Suggestions"
   - Verify suggestions appear with proper styling
   - Test "Apply" functionality

### Expected Behavior

- **With Historical Data**: Should show relevant predictions with confidence scores
- **No Historical Data**: Should show "No historical data available" message
- **Service Down**: Should show error message gracefully
- **Network Issues**: Should handle timeouts and display appropriate errors

## Development

### AI Service Development

```bash
cd ai-service
pip install -r requirements.txt
python app.py
```

### Adding New ML Features

1. **Extend Pattern Analysis**: Modify `analyze_spending_patterns()` in `ai-service/app.py`
2. **Add New Prediction Logic**: Update `predict_budget_amounts()` method
3. **Enhance Seasonal Factors**: Update `_get_seasonal_adjustment()` method

### Frontend Customization

- **Styling**: Modify `frontend/src/styles/AiBudgetSuggestions.css`
- **Behavior**: Update `frontend/src/components/budget/AiBudgetSuggestions.jsx`
- **Integration**: Enhance `frontend/src/components/budget/BudgetForm.jsx`

## Troubleshooting

### Common Issues

1. **AI Service Not Starting:**
   - Check Docker logs: `docker-compose logs ai-service`
   - Verify Python dependencies in requirements.txt
   - Ensure database connectivity

2. **No Predictions Returned:**
   - Verify user has historical budget data
   - Check database permissions for AI service
   - Review AI service logs for errors

3. **Frontend Not Showing Suggestions:**
   - Check browser console for JavaScript errors
   - Verify API endpoints are reachable
   - Check authentication token validity

4. **Database Connection Issues:**
   - Ensure PostgreSQL is running
   - Verify database credentials in docker-compose.yml
   - Check network connectivity between containers

### Logs and Debugging

```bash
# View AI service logs
docker-compose logs -f ai-service

# View backend logs
docker-compose logs -f app

# View database logs
docker-compose logs -f db

# Interactive debugging
docker-compose exec ai-service bash
docker-compose exec app bash
```

## Future Enhancements

### Potential Improvements

1. **Advanced ML Models**: 
   - Time series forecasting (ARIMA, LSTM)
   - Category-specific models
   - User behavior clustering

2. **Enhanced Features**:
   - Income prediction
   - Expense category recommendations
   - Budget optimization suggestions
   - Goal-based budgeting

3. **UI/UX Improvements**:
   - Interactive charts showing predictions
   - A/B testing for suggestion effectiveness
   - Personalization settings

4. **Performance Optimizations**:
   - Caching predictions
   - Batch processing
   - Model training pipelines

## Security Considerations

- All AI endpoints require authentication
- User data is isolated by user_id
- No sensitive data is logged
- Database connections use environment variables
- CORS is properly configured for frontend access

---

## Quick Start Guide

1. Ensure you have Docker running
2. Run `docker-compose up -d`
3. Create some historical budgets with transactions
4. Try creating a new budget and click "Get AI Suggestions"
5. Enjoy AI-powered budget recommendations! 🤖✨