"""
Advanced AI Budget Prediction Service with Machine Learning

This service uses sophisticated machine learning algorithms to analyze spending patterns
and predict optimal budget allocations:
- Linear Regression & Random Forest for trend prediction
- K-Means Clustering for spending behavior classification
- User Clustering for collaborative filtering recommendations
- Seasonal Decomposition for pattern recognition
- Feature Engineering with time series analysis
- Ensemble methods for robust predictions
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Dict, List, Any, Optional, Tuple
import numpy as np
import pandas as pd
from datetime import datetime, timedelta
import psycopg2
import os
import logging
from dataclasses import dataclass
import json

# Machine Learning imports
from sklearn.linear_model import LinearRegression, Ridge
from sklearn.ensemble import RandomForestRegressor
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler, MinMaxScaler
from sklearn.metrics import mean_absolute_error, r2_score
from sklearn.model_selection import train_test_split
from scipy import stats
from scipy.signal import find_peaks
import warnings
warnings.filterwarnings('ignore')

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Pydantic models for request/response validation
class BudgetPredictionRequest(BaseModel):
    user_id: int = Field(..., description="User ID for budget prediction")
    target_month: int = Field(default_factory=lambda: datetime.now().month, ge=1, le=12, description="Target month (1-12)")
    target_year: int = Field(default_factory=lambda: datetime.now().year, ge=2020, le=2030, description="Target year")
    historical_months: int = Field(default=18, ge=1, le=60, description="Number of historical months to analyze")

class PatternAnalysisRequest(BaseModel):
    user_id: int = Field(..., description="User ID for pattern analysis")
    historical_months: int = Field(default=18, ge=1, le=60, description="Number of historical months to analyze")

class FeatureImportance(BaseModel):
    feature_name: str
    importance: float

class BudgetPredictionResponse(BaseModel):
    category_id: int
    category_name: str
    predicted_amount_cents: int
    predicted_amount_dollars: float
    confidence_score: float
    historical_avg_cents: int  # Budget historical average
    historical_avg_dollars: float  # Budget historical average
    historical_spending_avg_cents: int = 0  # Spending historical average
    historical_spending_avg_dollars: float = 0.0  # Spending historical average
    trend_direction: str
    ml_model_used: str
    feature_importance: Dict[str, float]
    spending_cluster: str
    seasonal_pattern: str
    reasoning: str

class ModelInfo(BaseModel):
    algorithms_used: List[str]
    features_analyzed: List[str]
    confidence_factors: List[str]

class PredictionResult(BaseModel):
    predictions: List[BudgetPredictionResponse]
    target_month: int
    target_year: int
    user_id: int
    ml_enabled: bool
    model_info: ModelInfo
    message: str

class PatternAnalysisResult(BaseModel):
    analysis: Dict[str, Any]
    user_id: int
    ml_enabled: bool

class HealthResponse(BaseModel):
    status: str
    service: str

# Initialize FastAPI app
app = FastAPI(
    title="AI Budget Prediction Service",
    description="Advanced ML-based budget prediction service with feature engineering and ensemble methods",
    version="2.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure this properly for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@dataclass
class MLBudgetPrediction:
    category_id: int
    category_name: str
    predicted_amount_cents: int
    confidence_score: float
    historical_avg_cents: int  # Budget historical average
    historical_spending_avg_cents: int  # Spending historical average  
    trend_direction: str
    ml_model_used: str
    feature_importance: Dict[str, float]
    spending_cluster: str
    seasonal_pattern: str
    reasoning: str

class AdvancedBudgetPredictor:
    def __init__(self):
        self.db_connection = None
        self.scaler = StandardScaler()
        self.models = {}
        self.cluster_model = None
        
    def get_db_connection(self):
        """Get database connection using environment variables"""
        try:
            # Always create a fresh connection to avoid transaction issues
            connection = psycopg2.connect(
                host=os.getenv('DB_HOST', 'postgres'),
                database=os.getenv('DB_NAME', 'finance_tracker'),
                user=os.getenv('DB_USER', 'finance_user'),
                password=os.getenv('DB_PASSWORD', 'finance_password'),
                port=os.getenv('DB_PORT', '5432')
            )
            # Set autocommit to avoid transaction issues
            connection.autocommit = True
            return connection
        except Exception as e:
            logger.error(f"Database connection failed: {e}")
            raise
    
    def fetch_comprehensive_data(self, user_id: int, months_back: int = 18) -> Tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
        """Fetch comprehensive financial data for ML analysis"""
        
        # Enhanced budget query with more features
        budget_query = """
        SELECT 
            b.id as budget_id,
            b.period_start,
            b.period_end,
            bi.category_id,
            c.name as category_name,
            c.kind as category_kind,
            bi.planned_cents,
            EXTRACT(MONTH FROM b.period_start) as month,
            EXTRACT(YEAR FROM b.period_start) as year,
            EXTRACT(DOW FROM b.period_start) as day_of_week,
            EXTRACT(QUARTER FROM b.period_start) as quarter,
            DATE_PART('epoch', b.period_start) as timestamp_epoch
        FROM budgets b
        JOIN budget_items bi ON b.id = bi.budget_id
        JOIN categories c ON bi.category_id = c.id
        WHERE b.user_id = %s 
          AND b.period_start >= %s
        ORDER BY b.period_start DESC, c.name
        """
        
        # Enhanced spending query with transaction features
        spending_query = """
        SELECT 
            t.category_id,
            c.name as category_name,
            c.kind as category_kind,
            ABS(t.amount_cents) as spent_cents,
            t.txn_date,
            t.description,
            EXTRACT(MONTH FROM t.txn_date) as month,
            EXTRACT(YEAR FROM t.txn_date) as year,
            EXTRACT(DOW FROM t.txn_date) as day_of_week,
            EXTRACT(QUARTER FROM t.txn_date) as quarter,
            EXTRACT(DAY FROM t.txn_date) as day_of_month,
            DATE_PART('epoch', t.txn_date) as timestamp_epoch
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = %s 
          AND t.amount_cents < 0
          AND t.txn_date >= %s
        
        UNION ALL
        
        SELECT 
            ts.category_id,
            c.name as category_name,
            c.kind as category_kind,
            ABS(ts.amount_cents) as spent_cents,
            t.txn_date,
            t.description,
            EXTRACT(MONTH FROM t.txn_date) as month,
            EXTRACT(YEAR FROM t.txn_date) as year,
            EXTRACT(DOW FROM t.txn_date) as day_of_week,
            EXTRACT(QUARTER FROM t.txn_date) as quarter,
            EXTRACT(DAY FROM t.txn_date) as day_of_month,
            DATE_PART('epoch', t.txn_date) as timestamp_epoch
        FROM transaction_splits ts
        JOIN transactions t ON ts.parent_txn_id = t.id
        JOIN categories c ON ts.category_id = c.id
        WHERE t.user_id = %s 
          AND ts.amount_cents < 0
          AND t.txn_date >= %s
        """
        
        # Account data for context
        account_query = """
        SELECT 
            a.id as account_id,
            a.name as account_name,
            a.type as account_type,
            t.category_id,
            COUNT(*) as transaction_count,
            AVG(ABS(t.amount_cents)) as avg_transaction_amount
        FROM accounts a
        JOIN transactions t ON a.id = t.account_id
        JOIN categories c ON t.category_id = c.id
        WHERE a.user_id = %s 
          AND t.amount_cents < 0
          AND t.txn_date >= %s
        GROUP BY a.id, a.name, a.type, t.category_id
        """
        
        cutoff_date = datetime.now() - timedelta(days=months_back * 30)
        
        conn = None
        cursor = None
        
        try:
            conn = self.get_db_connection()
            cursor = conn.cursor()
            
            # Fetch budgets
            cursor.execute(budget_query, (user_id, cutoff_date))
            budget_columns = [desc[0] for desc in cursor.description]
            budget_data = cursor.fetchall()
            budgets_df = pd.DataFrame(budget_data, columns=budget_columns)
            
            # Convert numeric columns to proper types
            numeric_budget_cols = ['planned_cents', 'month', 'year', 'day_of_week', 'quarter', 'timestamp_epoch']
            for col in numeric_budget_cols:
                if col in budgets_df.columns:
                    budgets_df[col] = pd.to_numeric(budgets_df[col], errors='coerce')
            
            # Fetch spending
            cursor.execute(spending_query, (user_id, cutoff_date, user_id, cutoff_date))
            spending_columns = [desc[0] for desc in cursor.description]
            spending_data = cursor.fetchall()
            spending_df = pd.DataFrame(spending_data, columns=spending_columns)
            
            # Convert numeric columns to proper types
            numeric_spending_cols = ['spent_cents', 'month', 'year', 'day_of_week', 'quarter', 'day_of_month', 'timestamp_epoch']
            for col in numeric_spending_cols:
                if col in spending_df.columns:
                    spending_df[col] = pd.to_numeric(spending_df[col], errors='coerce')
            
            # Fetch account data
            cursor.execute(account_query, (user_id, cutoff_date))
            account_columns = [desc[0] for desc in cursor.description]
            account_data = cursor.fetchall()
            accounts_df = pd.DataFrame(account_data, columns=account_columns)
            
            # Convert numeric columns to proper types
            numeric_account_cols = ['transaction_count', 'avg_transaction_amount']
            for col in numeric_account_cols:
                if col in accounts_df.columns:
                    accounts_df[col] = pd.to_numeric(accounts_df[col], errors='coerce')
            
            return budgets_df, spending_df, accounts_df
            
        except Exception as e:
            logger.error(f"Database query failed: {e}")
            # Return empty DataFrames on error
            return pd.DataFrame(), pd.DataFrame(), pd.DataFrame()
        finally:
            # Always close cursor and connection
            if cursor:
                cursor.close()
            if conn:
                conn.close()
    
    def engineer_features(self, budgets_df: pd.DataFrame, spending_df: pd.DataFrame, accounts_df: pd.DataFrame) -> pd.DataFrame:
        """Engineer features for machine learning models"""
        
        if budgets_df.empty:
            return pd.DataFrame()
        
        features_list = []
        
        # Process each category
        for category_id in budgets_df['category_id'].unique():
            category_budgets = budgets_df[budgets_df['category_id'] == category_id].copy()
            category_spending = spending_df[spending_df['category_id'] == category_id].copy() if not spending_df.empty else pd.DataFrame()
            category_accounts = accounts_df[accounts_df['category_id'] == category_id].copy() if not accounts_df.empty else pd.DataFrame()
            
            # Sort by date for time series features
            category_budgets = category_budgets.sort_values('period_start')
            
            for idx, budget_row in category_budgets.iterrows():
                features = {}
                
                # Basic features
                features['category_id'] = category_id
                features['category_name'] = budget_row['category_name']
                features['target_amount'] = float(budget_row['planned_cents'])
                features['month'] = int(budget_row['month'])
                features['quarter'] = int(budget_row['quarter'])
                features['year'] = int(budget_row['year'])
                
                # Cyclical encoding for seasonality
                features['month_sin'] = np.sin(2 * np.pi * float(budget_row['month']) / 12)
                features['month_cos'] = np.cos(2 * np.pi * float(budget_row['month']) / 12)
                features['quarter_sin'] = np.sin(2 * np.pi * float(budget_row['quarter']) / 4)
                features['quarter_cos'] = np.cos(2 * np.pi * float(budget_row['quarter']) / 4)
                
                # Historical features (lookback)
                period_start = budget_row['period_start']
                historical_budgets = category_budgets[category_budgets['period_start'] < period_start]
                
                if len(historical_budgets) > 0:
                    # Budget-based historical stats (what was planned)
                    features['historical_mean'] = float(historical_budgets['planned_cents'].mean())
                    features['historical_std'] = float(historical_budgets['planned_cents'].std())
                    features['historical_min'] = float(historical_budgets['planned_cents'].min())
                    features['historical_max'] = float(historical_budgets['planned_cents'].max())
                    features['historical_trend'] = self._calculate_trend(historical_budgets['planned_cents'].values.astype(float))
                    features['historical_count'] = len(historical_budgets)

                    # Compute historical actual spending per past budget period (if any spending exists)
                    # For each historical budget period, sum spending that falls within that period
                    historical_spent_values = []
                    for _, past_row in historical_budgets.iterrows():
                        start = past_row['period_start']
                        end = past_row['period_end']
                        period_spent = category_spending[
                            (category_spending['txn_date'] >= start) &
                            (category_spending['txn_date'] <= end)
                        ]
                        if not period_spent.empty:
                            historical_spent_values.append(float(period_spent['spent_cents'].sum()))

                    if historical_spent_values:
                        features['historical_spent_mean'] = float(np.mean(historical_spent_values))
                        features['historical_spent_count'] = len(historical_spent_values)
                    else:
                        features['historical_spent_mean'] = 0.0
                        features['historical_spent_count'] = 0

                    # Combined historical mean: prefer actual spending mean when available
                    # Keep this for now but use budget mean as baseline for predictions
                    if features['historical_spent_count'] > 0:
                        features['historical_combined_mean'] = float(
                            0.7 * features['historical_spent_mean'] + 0.3 * features['historical_mean']
                        )
                    else:
                        features['historical_combined_mean'] = float(features['historical_mean'])
                    
                    # Recent trend (last 3 periods)
                    recent = historical_budgets.tail(3)
                    features['recent_mean'] = float(recent['planned_cents'].mean()) if len(recent) > 0 else features['historical_mean']
                    features['recent_trend'] = self._calculate_trend(recent['planned_cents'].values.astype(float)) if len(recent) > 1 else 0
                else:
                    # First budget - use defaults
                    features['historical_mean'] = float(budget_row['planned_cents'])
                    features['historical_std'] = 0.0
                    features['historical_min'] = float(budget_row['planned_cents'])
                    features['historical_max'] = float(budget_row['planned_cents'])
                    features['historical_trend'] = 0.0
                    features['historical_count'] = 0
                    features['recent_mean'] = float(budget_row['planned_cents'])
                    features['recent_trend'] = 0.0
                    # No spending history yet
                    features['historical_spent_mean'] = 0.0
                    features['historical_spent_count'] = 0
                    features['historical_combined_mean'] = float(budget_row['planned_cents'])
                
                # Spending pattern features
                if not category_spending.empty:
                    period_spending = category_spending[
                        (category_spending['txn_date'] >= budget_row['period_start']) &
                        (category_spending['txn_date'] <= budget_row['period_end'])
                    ]
                    
                    features['actual_spent'] = float(period_spending['spent_cents'].sum()) if not period_spending.empty else 0.0
                    features['spending_frequency'] = float(len(period_spending))
                    features['avg_transaction_size'] = float(period_spending['spent_cents'].mean()) if not period_spending.empty else 0.0
                    features['spending_volatility'] = float(period_spending['spent_cents'].std()) if len(period_spending) > 1 else 0.0
                    
                    # Calculate overall spending history for this category (not just this period)
                    total_category_spending = float(category_spending['spent_cents'].sum())
                    
                    # Budget vs actual accuracy - if this period has spending, calculate accuracy
                    if features['actual_spent'] > 0:
                        planned_amount = float(budget_row['planned_cents'])
                        features['budget_accuracy'] = 1.0 - abs(planned_amount - features['actual_spent']) / max(planned_amount, features['actual_spent'])
                    elif total_category_spending > 0:
                        # Category has spending history but not in this specific period
                        # Give it moderate accuracy (better than neutral but not perfect)
                        features['budget_accuracy'] = 0.7
                    else:
                        # No spending data for this category at all
                        features['budget_accuracy'] = 0.5  # Neutral
                else:
                    # No spending records for this category across the entire fetched window
                    features['actual_spent'] = 0.0
                    features['spending_frequency'] = 0.0
                    features['avg_transaction_size'] = 0.0
                    features['spending_volatility'] = 0.0
                    # Use neutral accuracy (0.5) so lack of data doesn't imply perfect matching
                    features['budget_accuracy'] = 0.5
                
                # Account diversity features
                if not category_accounts.empty:
                    features['account_diversity'] = float(len(category_accounts))
                    features['primary_account_usage'] = float(category_accounts['transaction_count'].max())
                else:
                    features['account_diversity'] = 1.0
                    features['primary_account_usage'] = 1.0
                
                features_list.append(features)
        
        return pd.DataFrame(features_list)
    
    def _calculate_trend(self, values: np.ndarray) -> float:
        """Calculate trend slope for a series of values"""
        if len(values) < 2:
            return 0.0
        x = np.arange(len(values))
        slope, _, _, _, _ = stats.linregress(x, values)
        return float(slope)
    
    def build_ml_models(self, features_df: pd.DataFrame) -> Dict[str, Any]:
        """Build and train multiple ML models for budget prediction"""
        
        if features_df.empty or len(features_df) < 5:
            return {}
        
        # Prepare features for ML
        feature_columns = [
            'month_sin', 'month_cos', 'quarter_sin', 'quarter_cos',
            'historical_mean', 'historical_spent_mean', 'historical_std', 'historical_trend', 'historical_count',
            'recent_mean', 'recent_trend', 'actual_spent', 'spending_frequency',
            'avg_transaction_size', 'spending_volatility', 'budget_accuracy',
            'account_diversity', 'primary_account_usage'
        ]
        
        # Fill any missing values
        for col in feature_columns:
            if col not in features_df.columns:
                if col == 'historical_spent_mean':
                    features_df[col] = 0.0  # Default for spending mean
                else:
                    features_df[col] = 0
        
        # Clean the data and handle NaN values
        X = features_df[feature_columns].copy()
        
        # Use smarter target: if we have actual spending data, use the higher of budget or actual spending
        # This trains the model to predict what someone should budget based on their actual behavior
        y = features_df.apply(lambda row: 
            max(row['target_amount'], row['actual_spent']) if row['actual_spent'] > 0 
            else row['target_amount'], axis=1)
        
        # Replace any remaining NaN values with appropriate defaults
        X = X.fillna({
            'month_sin': 0.0,
            'month_cos': 1.0,
            'quarter_sin': 0.0,
            'quarter_cos': 1.0,
            'historical_mean': 0.0,
            'historical_std': 0.0,
            'historical_trend': 0.0,
            'historical_count': 0.0,
            'recent_mean': 0.0,
            'recent_trend': 0.0,
            'actual_spent': 0.0,
            'spending_frequency': 0.0,
            'avg_transaction_size': 0.0,
            'spending_volatility': 0.0,
            'budget_accuracy': 1.0,
            'account_diversity': 1.0,
            'primary_account_usage': 1.0
        })
        
        # Ensure all values are finite
        X = X.replace([np.inf, -np.inf], 0.0)
        y = y.fillna(0.0).replace([np.inf, -np.inf], 0.0)
        
        # Verify no NaN values remain
        if X.isnull().any().any() or y.isnull().any():
            logger.warning("NaN values still present after cleaning, filling with zeros")
            X = X.fillna(0.0)
            y = y.fillna(0.0)
        
        # Scale features
        X_scaled = self.scaler.fit_transform(X)
        
        models = {}
        
        # Only train if we have enough data
        if len(X) >= 5:
            try:
                # Linear Regression with regularization
                models['linear'] = Ridge(alpha=1.0)
                models['linear'].fit(X_scaled, y)
                
                # Random Forest for non-linear patterns
                models['forest'] = RandomForestRegressor(n_estimators=50, random_state=42, max_depth=5)
                models['forest'].fit(X_scaled, y)
                
                # Store feature importance
                if hasattr(models['forest'], 'feature_importances_'):
                    feature_importance = dict(zip(feature_columns, models['forest'].feature_importances_))
                    models['feature_importance'] = feature_importance
                
                logger.info(f"Trained ML models on {len(X)} data points")
                
            except Exception as e:
                logger.warning(f"Failed to train ML models: {e}")
                
        return models
    
    def classify_spending_behavior(self, features_df: pd.DataFrame) -> Dict[int, str]:
        """Use K-Means clustering to classify spending behaviors"""
        
        if features_df.empty or len(features_df) < 3:
            return {}
        
        # Features for clustering
        cluster_features = ['budget_accuracy', 'spending_volatility', 'spending_frequency', 'historical_trend']
        available_features = [f for f in cluster_features if f in features_df.columns]
        
        if len(available_features) < 2:
            return {}
        
        try:
            cluster_data = features_df[available_features].fillna(0)
            
            # Normalize for clustering
            scaler = MinMaxScaler()
            cluster_data_scaled = scaler.fit_transform(cluster_data)
            
            # K-Means clustering (3 clusters: conservative, balanced, volatile)
            kmeans = KMeans(n_clusters=min(3, len(cluster_data)), random_state=42)
            clusters = kmeans.fit_predict(cluster_data_scaled)
            
            # Map clusters to behavior types
            cluster_mapping = {0: "Conservative", 1: "Balanced", 2: "Volatile"}
            
            category_clusters = {}
            for idx, row in features_df.iterrows():
                category_id = row['category_id']
                cluster_id = clusters[idx] if idx < len(clusters) else 0
                category_clusters[category_id] = cluster_mapping.get(cluster_id, "Balanced")
            
            self.cluster_model = kmeans
            return category_clusters
            
        except Exception as e:
            logger.warning(f"Clustering failed: {e}")
            return {}
    
    def detect_seasonal_patterns(self, features_df: pd.DataFrame) -> Dict[int, str]:
        """Detect seasonal spending patterns using statistical analysis"""
        
        patterns = {}
        
        for category_id in features_df['category_id'].unique():
            category_data = features_df[features_df['category_id'] == category_id]
            
            if len(category_data) < 4:  # Need at least 4 data points
                patterns[category_id] = "Insufficient data"
                continue
            
            # Analyze monthly patterns
            monthly_amounts = category_data.groupby('month')['target_amount'].mean()
            
            if len(monthly_amounts) >= 3:
                # Calculate coefficient of variation
                cv = monthly_amounts.std() / monthly_amounts.mean() if monthly_amounts.mean() > 0 else 0
                
                if cv > 0.3:
                    # Find peaks and troughs
                    amounts_array = monthly_amounts.reindex(range(1, 13), fill_value=monthly_amounts.mean()).values
                    peaks, _ = find_peaks(amounts_array, height=monthly_amounts.mean())
                    
                    if len(peaks) > 0:
                        peak_months = peaks + 1  # Convert to 1-based months
                        if any(month in [11, 12, 1] for month in peak_months):
                            patterns[category_id] = "Holiday seasonal"
                        elif any(month in [6, 7, 8] for month in peak_months):
                            patterns[category_id] = "Summer seasonal"
                        else:
                            patterns[category_id] = "Irregular seasonal"
                    else:
                        patterns[category_id] = "Stable"
                else:
                    patterns[category_id] = "Stable"
            else:
                patterns[category_id] = "Limited data"
        
        return patterns
    
    def create_user_profiles(self, historical_months: int = 18) -> pd.DataFrame:
        """Create user spending profiles for clustering"""
        
        try:
            # Get all users' spending data
            query = """
            WITH user_spending AS (
                SELECT 
                    u.id as user_id,
                    u.created_at as user_since,
                    COALESCE(cat.name, 'Uncategorized') as category,
                    cat.id as category_id,
                    SUM(ABS(t.amount_cents)) as total_spent,
                    COUNT(t.id) as transaction_count,
                    AVG(ABS(t.amount_cents)) as avg_transaction,
                    STDDEV(ABS(t.amount_cents)) as spending_volatility
                FROM users u
                LEFT JOIN transactions t ON u.id = t.user_id 
                LEFT JOIN categories cat ON t.category_id = cat.id
                WHERE t.txn_date >= CURRENT_DATE - INTERVAL '%s months'
                    AND t.amount_cents < 0  -- Only expenses
                GROUP BY u.id, u.created_at, cat.id, cat.name
            ),
            user_budgets AS (
                SELECT 
                    u.id as user_id,
                    COALESCE(cat.name, 'Uncategorized') as category,
                    cat.id as category_id,
                    AVG(bi.planned_cents) as avg_budget,
                    COUNT(bi.id) as budget_count
                FROM users u
                LEFT JOIN budgets b ON u.id = b.user_id
                LEFT JOIN budget_items bi ON b.id = bi.budget_id
                LEFT JOIN categories cat ON bi.category_id = cat.id
                WHERE b.period_start >= CURRENT_DATE - INTERVAL '%s months'
                GROUP BY u.id, cat.id, cat.name
            ),
            user_stats AS (
                SELECT 
                    u.id as user_id,
                    EXTRACT(DAYS FROM CURRENT_DATE - u.created_at) / 30.0 as account_age_months,
                    COUNT(DISTINCT acc.id) as account_count,
                    COUNT(DISTINCT t.category_id) as unique_categories,
                    SUM(CASE WHEN t.amount_cents < 0 THEN ABS(t.amount_cents) ELSE 0 END) as total_expenses,
                    SUM(CASE WHEN t.amount_cents > 0 THEN t.amount_cents ELSE 0 END) as total_income,
                    COUNT(DISTINCT DATE_TRUNC('month', t.txn_date)) as active_months
                FROM users u
                LEFT JOIN accounts acc ON u.id = acc.user_id
                LEFT JOIN transactions t ON u.id = t.user_id
                WHERE t.txn_date >= CURRENT_DATE - INTERVAL '%s months'
                GROUP BY u.id, u.created_at
            )
            SELECT 
                us.*,
                ub.avg_budget,
                ub.budget_count,
                ustat.account_age_months,
                ustat.account_count,
                ustat.unique_categories,
                ustat.total_expenses,
                ustat.total_income,
                ustat.active_months
            FROM user_spending us
            LEFT JOIN user_budgets ub ON us.user_id = ub.user_id AND us.category_id = ub.category_id
            LEFT JOIN user_stats ustat ON us.user_id = ustat.user_id
            WHERE us.total_spent > 0
            ORDER BY us.user_id, us.category
            """
            
            conn = psycopg2.connect(
                host=os.getenv('DB_HOST', 'postgres'),
                database=os.getenv('DB_NAME', 'finance_tracker'),
                user=os.getenv('DB_USER', 'finance_user'),
                password=os.getenv('DB_PASSWORD', 'finance_password'),
                port=os.getenv('DB_PORT', '5432')
            )
            df = pd.read_sql(query, conn, params=[historical_months, historical_months, historical_months])
            conn.close()
            
            return df
            
        except Exception as e:
            logger.error(f"Error creating user profiles: {e}")
            return pd.DataFrame()
    
    def cluster_users(self, user_profiles_df: pd.DataFrame, n_clusters: int = 5) -> Dict[int, Dict]:
        """Cluster users based on spending patterns and demographics"""
        
        if user_profiles_df.empty:
            return {}
        
        try:
            # Create user-level aggregated features
            user_features = user_profiles_df.groupby('user_id').agg({
                'total_spent': 'sum',
                'transaction_count': 'sum', 
                'avg_transaction': 'mean',
                'spending_volatility': 'mean',
                'avg_budget': 'mean',
                'budget_count': 'sum',
                'account_age_months': 'first',
                'account_count': 'first',
                'unique_categories': 'first',
                'total_expenses': 'first',
                'total_income': 'first',
                'active_months': 'first'
            }).fillna(0)
            
            # Calculate additional features
            user_features['expense_to_income_ratio'] = np.where(
                user_features['total_income'] > 0,
                user_features['total_expenses'] / user_features['total_income'],
                1.0
            )
            user_features['avg_monthly_spending'] = np.where(
                user_features['active_months'] > 0,
                user_features['total_expenses'] / user_features['active_months'],
                0
            )
            user_features['budget_adherence'] = np.where(
                user_features['avg_budget'] > 0,
                1 - abs(user_features['total_expenses'] - user_features['avg_budget']) / user_features['avg_budget'],
                0.5
            )
            
            # Select features for clustering
            clustering_features = [
                'avg_monthly_spending', 'expense_to_income_ratio', 'budget_adherence',
                'spending_volatility', 'unique_categories', 'account_age_months'
            ]
            
            feature_data = user_features[clustering_features].fillna(0)
            
            # Adaptive cluster count: use fewer clusters if we don't have enough users
            n_users = len(user_features)
            if n_users < 2:
                # Not enough users to cluster
                logger.info(f"Only {n_users} users available, skipping clustering")
                return {}
            
            # Adjust cluster count based on user count
            effective_clusters = min(n_clusters, max(2, n_users // 2))  # At least 2, at most n_users//2
            logger.info(f"Clustering {n_users} users into {effective_clusters} clusters")
            
            # Normalize features
            scaler = StandardScaler()
            normalized_features = scaler.fit_transform(feature_data)
            
            # Perform clustering
            kmeans = KMeans(n_clusters=effective_clusters, random_state=42, n_init=10)
            user_features['cluster'] = kmeans.fit_predict(normalized_features)
            
            # Create cluster profiles
            clusters = {}
            for cluster_id in range(effective_clusters):
                cluster_users = user_features[user_features['cluster'] == cluster_id]
                cluster_profile = {
                    'user_count': len(cluster_users),
                    'avg_monthly_spending': cluster_users['avg_monthly_spending'].mean(),
                    'avg_expense_ratio': cluster_users['expense_to_income_ratio'].mean(),
                    'avg_budget_adherence': cluster_users['budget_adherence'].mean(),
                    'avg_categories': cluster_users['unique_categories'].mean(),
                    'avg_age_months': cluster_users['account_age_months'].mean(),
                    'user_ids': cluster_users.index.tolist()
                }
                clusters[cluster_id] = cluster_profile
                
            # Return user cluster assignments with profiles
            user_clusters = {}
            for user_id in user_features.index:
                cluster_id = user_features.loc[user_id, 'cluster']
                user_clusters[user_id] = {
                    'cluster_id': cluster_id,
                    'cluster_profile': clusters[cluster_id],
                    'similarity_peers': [uid for uid in clusters[cluster_id]['user_ids'] if uid != user_id][:10]  # Top 10 similar users
                }
                
            return user_clusters
            
        except Exception as e:
            logger.error(f"Error clustering users: {e}")
            return {}
    
    def get_peer_recommendations(self, user_id: int, target_category_id: int, user_clusters: Dict, historical_months: int = 6) -> Dict:
        """Get budget recommendations based on similar users"""
        
        if user_id not in user_clusters:
            return {}
        
        try:
            similar_users = user_clusters[user_id]['similarity_peers']
            if not similar_users:
                return {}
            
            # Get successful budget patterns from similar users
            similar_users_str = ','.join(map(str, similar_users))
            
            query = """
            WITH peer_budgets AS (
                SELECT 
                    bi.category_id,
                    bi.planned_cents,
                    ABS(COALESCE(spent.total_spent, 0)) as actual_spent,
                    ABS(bi.planned_cents - COALESCE(spent.total_spent, 0)) as budget_variance
                FROM budget_items bi
                JOIN budgets b ON bi.budget_id = b.id
                LEFT JOIN (
                    SELECT 
                        t.user_id,
                        t.category_id,
                        DATE_TRUNC('month', t.txn_date) as month,
                        SUM(ABS(t.amount_cents)) as total_spent
                    FROM transactions t
                    WHERE t.amount_cents < 0
                        AND t.txn_date >= CURRENT_DATE - INTERVAL '%s months'
                    GROUP BY t.user_id, t.category_id, DATE_TRUNC('month', t.txn_date)
                ) spent ON bi.category_id = spent.category_id 
                    AND b.user_id = spent.user_id
                    AND DATE_TRUNC('month', b.period_start) = spent.month
                WHERE b.user_id IN (%s)
                    AND bi.category_id = %s
                    AND b.period_start >= CURRENT_DATE - INTERVAL '%s months'
            )
            SELECT 
                AVG(planned_cents) as avg_peer_budget,
                AVG(actual_spent) as avg_peer_spending,
                AVG(budget_variance) as avg_variance,
                COUNT(*) as peer_samples,
                STDDEV(planned_cents) as budget_std,
                MIN(planned_cents) as min_budget,
                MAX(planned_cents) as max_budget
            FROM peer_budgets
            WHERE planned_cents > 0
            """
            
            conn = psycopg2.connect(
                host=os.getenv('DB_HOST', 'postgres'),
                database=os.getenv('DB_NAME', 'finance_tracker'),
                user=os.getenv('DB_USER', 'finance_user'),
                password=os.getenv('DB_PASSWORD', 'finance_password'),
                port=os.getenv('DB_PORT', '5432')
            )
            cursor = conn.cursor()
            cursor.execute(query, [historical_months, similar_users_str, target_category_id, historical_months])
            result = cursor.fetchone()
            conn.close()
            
            if result and result[0]:
                return {
                    'avg_peer_budget': result[0],
                    'avg_peer_spending': result[1] or 0,
                    'avg_variance': result[2] or 0,
                    'peer_samples': result[3],
                    'budget_std': result[4] or 0,
                    'min_budget': result[5],
                    'max_budget': result[6],
                    'peer_count': len(similar_users)
                }
            
        except Exception as e:
            logger.error(f"Error getting peer recommendations: {e}")
        
        return {}
    
    def _generate_user_insights(self, user_cluster_info: Dict, cluster_profile: Dict) -> List[str]:
        """Generate personalized insights based on user's cluster"""
        
        insights = []
        
        if not cluster_profile:
            return ["Not enough data for personalized insights"]
        
        # Budget adherence insights
        budget_adherence = cluster_profile.get('avg_budget_adherence', 0)
        if budget_adherence > 0.8:
            insights.append("You're in a group of highly disciplined budgeters who stick close to their plans")
        elif budget_adherence > 0.6:
            insights.append("You're among users with good budget discipline, with room for improvement")
        else:
            insights.append("You're in a group that tends to go over budget - consider more realistic planning")
        
        # Spending level insights
        monthly_spending = cluster_profile.get('avg_monthly_spending', 0)
        if monthly_spending > 300000:  # > $3000
            insights.append("You're in a high-spending group - focus on identifying savings opportunities")
        elif monthly_spending > 150000:  # > $1500
            insights.append("You're in a moderate-spending group with balanced financial habits")
        else:
            insights.append("You're in a conservative-spending group - great for building savings")
        
        # Category diversity insights
        categories = cluster_profile.get('avg_categories', 0)
        if categories > 8:
            insights.append("Your group tracks spending across many categories - excellent for detailed budgeting")
        elif categories > 5:
            insights.append("Your group maintains good spending visibility across key categories")
        else:
            insights.append("Your group prefers simple budgeting with fewer categories")
        
        # Experience insights
        age_months = cluster_profile.get('avg_age_months', 0)
        if age_months > 12:
            insights.append("You're grouped with experienced users who have established spending patterns")
        else:
            insights.append("You're among newer users still developing their budgeting habits")
        
        return insights
    
    def predict_with_ml(self, user_id: int, target_month: int, target_year: int, historical_months: int = 18) -> List[MLBudgetPrediction]:
        """Generate ML-based budget predictions"""
        
        try:
            # Fetch comprehensive data
            budgets_df, spending_df, accounts_df = self.fetch_comprehensive_data(user_id, historical_months)
            
            if budgets_df.empty:
                return []
            
            # Engineer features
            features_df = self.engineer_features(budgets_df, spending_df, accounts_df)
            
            if features_df.empty:
                return []
            
            # Build ML models
            models = self.build_ml_models(features_df)
            
            # Classify spending behaviors
            spending_clusters = self.classify_spending_behavior(features_df)
            
            # Detect seasonal patterns
            seasonal_patterns = self.detect_seasonal_patterns(features_df)
            
            # Create user clusters for collaborative filtering
            user_profiles_df = self.create_user_profiles(historical_months)
            user_clusters = self.cluster_users(user_profiles_df)
            
            logger.info(f"User clustering: {len(user_clusters)} users clustered, user {user_id} in cluster {user_clusters.get(user_id, {}).get('cluster_id', 'Unknown')}")
            
            # Generate predictions for each category
            predictions = []
            
            category_ids = features_df['category_id'].unique()
            logger.info(f"Processing {len(category_ids)} categories: {list(category_ids)}")
            
            for category_id in category_ids:
                logger.info(f"Starting prediction for category {category_id}")
                category_data = features_df[features_df['category_id'] == category_id]
                latest_data = category_data.iloc[-1]  # Most recent data for this category
                
                # Prepare features for prediction
                prediction_features = self._prepare_prediction_features(
                    latest_data, target_month, target_year
                )
                
                # Make predictions using available models
                predictions_dict = {}
                
                if models and len(models) > 0:
                    feature_columns = [
                        'month_sin', 'month_cos', 'quarter_sin', 'quarter_cos',
                        'historical_mean', 'historical_spent_mean', 'historical_std', 'historical_trend', 'historical_count',
                        'recent_mean', 'recent_trend', 'actual_spent', 'spending_frequency',
                        'avg_transaction_size', 'spending_volatility', 'budget_accuracy',
                        'account_diversity', 'primary_account_usage'
                    ]
                    
                    X_pred = np.array([[prediction_features.get(col, 0) for col in feature_columns]])
                    
                    # Clean prediction features and handle NaN values
                    X_pred = np.nan_to_num(X_pred, nan=0.0, posinf=0.0, neginf=0.0)
                    
                    # Only scale if we have valid data
                    try:
                        X_pred_scaled = self.scaler.transform(X_pred)
                    except Exception as e:
                        logger.warning(f"Scaling failed: {e}, using unscaled features")
                        X_pred_scaled = X_pred
                    
                    # Get predictions from each model
                    for model_name, model in models.items():
                        if model_name not in ['feature_importance'] and hasattr(model, 'predict'):
                            try:
                                pred = model.predict(X_pred_scaled)[0]
                                # Ensure prediction is valid
                                if np.isfinite(pred) and pred >= 0:
                                    predictions_dict[model_name] = float(pred)
                                else:
                                    logger.warning(f"Model {model_name} produced invalid prediction: {pred}")
                            except Exception as e:
                                logger.warning(f"Model {model_name} prediction failed: {e}")
                
                # Initialize default values
                # Use budget historical mean as baseline for predictions (keep them separate)
                budget_historical_avg = float(latest_data.get('historical_mean', 0.0))
                spending_historical_avg = float(latest_data.get('historical_spent_mean', 0.0))
                historical_avg = budget_historical_avg  # Use budget mean as prediction baseline
                
                # Initialize prediction variables for this category (prevent carryover from previous categories)
                predicted_amount = None
                model_used = None
                confidence = None
                
                # Log both averages for transparency
                logger.info(f"  Budget historical avg: ${budget_historical_avg/100:.2f}")
                if latest_data.get('historical_spent_count', 0) > 0:
                    logger.info(f"  Spending historical avg: ${spending_historical_avg/100:.2f} (periods: {int(latest_data['historical_spent_count'])})")
                else:
                    logger.info(f"  No spending history available")
                logger.info(f"  Using budget avg as baseline: ${historical_avg/100:.2f}")
                
                ml_predicted_amount = historical_avg  # Default to historical
                model_agreement = 0.7
                max_reasonable_change = historical_avg * 0.5
                ml_change = 0.0
                
                # Ensemble prediction (average of available models)
                if predictions_dict:
                    ml_predicted_amount = np.mean(list(predictions_dict.values()))
                    
                    # Calculate model agreement for confidence
                    if len(predictions_dict) > 1:
                        pred_std = np.std(list(predictions_dict.values()))
                        pred_mean = np.mean(list(predictions_dict.values()))
                        model_agreement = max(0.3, 1.0 - (pred_std / pred_mean)) if pred_mean > 0 else 0.5
                    else:
                        model_agreement = 0.7
                    
                    # Check if ML prediction is reasonable (within 50% of historical)
                    ml_change = abs(ml_predicted_amount - historical_avg)
                
                # Check data quality to determine if we should suggest changes
                data_points = len(category_data)  # Total data points for this category, not just historical
                accuracy = latest_data.get('budget_accuracy', 0.5)
                
                # Debug logging to see what's happening
                logger.info(f"Category {latest_data['category_name']}: data_points={data_points}, accuracy={accuracy:.2f}, ml_change=${ml_change/100:.2f}")
                logger.info(f"  ML predicted: ${ml_predicted_amount/100:.2f}, Historical: ${historical_avg/100:.2f}")
                logger.info(f"  Budget avg: ${budget_historical_avg/100:.2f}, Spending avg: ${spending_historical_avg/100:.2f}")
                logger.info(f"  Model agreement: {model_agreement:.2f}, Max reasonable change: ${max_reasonable_change/100:.2f}")
                
                # Debug the accurate budgeter condition
                if spending_historical_avg > 0 and budget_historical_avg > 0:
                    ratio = spending_historical_avg / budget_historical_avg
                    ratio_diff = abs(ratio - 1.0)
                    logger.info(f"  Accurate budgeter check: spending/budget ratio = {ratio:.3f}, diff from 1.0 = {ratio_diff:.3f}")
                    logger.info(f"  Ratio diff < 0.02? {ratio_diff < 0.02}, Has predictions? {bool(predictions_dict)}")
                
                # Calculate base confidence based on data quality and spending history
                base_confidence = min(0.9, 0.3 + (data_points * 0.08))  # More data = higher base confidence
                
                # Apply accuracy bonus/penalty - but first check if we actually have spending data
                if accuracy > 0.7:  # Good accuracy (actual spending matches budgets well)
                    accuracy_factor = 1.2
                elif accuracy < 0.3:  # Poor accuracy (spending way off from budgets)
                    accuracy_factor = 0.8
                else:  # Neutral accuracy (0.3-0.7, including 0.5 for no spending data)
                    accuracy_factor = 1.0
                
                # Determine prediction strategy based on data quality
                if data_points < 3:
                    # Use pure historical average for insufficient data
                    predicted_amount = historical_avg
                    confidence = base_confidence * 0.6  # Lower confidence for insufficient data
                    model_used = f"Historical Average (only {int(data_points)} data points)"
                    logger.info(f"  -> Using historical (insufficient data points)")
                elif accuracy < 0.2:  # Very poor accuracy
                    # Use pure historical average for very poor accuracy
                    predicted_amount = historical_avg
                    confidence = base_confidence * accuracy_factor * 0.7
                    model_used = "Historical Average (very low accuracy)"
                    logger.info(f"  -> Using historical (very low accuracy)")
                elif predictions_dict and spending_historical_avg > 0 and abs(spending_historical_avg / budget_historical_avg - 1.0) < 0.02:
                    # PRIORITY CHECK: Perfect budgeters (spending within 2% of budget) - handle first!
                    predicted_amount = max(budget_historical_avg, spending_historical_avg * 1.02)  # Small 2% buffer for inflation
                    confidence = base_confidence * accuracy_factor * model_agreement * 0.92  # High confidence for accurate budgeters
                    model_used = f"Accurate Budgeter (spending ≈ budget)"
                    logger.info(f"  -> PRIORITY: Accurate budgeter detected, maintaining budget: ${predicted_amount/100:.2f}")
                elif data_points < 5 and user_id in user_clusters:
                    # PEER RECOMMENDATIONS: For users with limited data, use similar users' patterns
                    logger.info(f"  -> Checking peer recommendations for {latest_data['category_name']} (data_points={data_points})")
                    peer_data = self.get_peer_recommendations(user_id, category_id, user_clusters, historical_months=6)
                    
                    if peer_data and peer_data.get('peer_samples', 0) >= 3:
                        peer_budget = peer_data['avg_peer_budget']
                        peer_spending = peer_data.get('avg_peer_spending', peer_budget)
                        peer_confidence = min(0.8, 0.4 + (peer_data['peer_samples'] * 0.05))  # Confidence based on peer sample size
                        
                        # Blend peer recommendation with any available personal data
                        if historical_avg > 0:
                            # User has some history, blend with peer data
                            peer_weight = min(0.6, peer_data['peer_samples'] / 10)  # More peers = higher weight
                            predicted_amount = historical_avg * (1 - peer_weight) + peer_budget * peer_weight
                            model_used = f"Peer-Informed Prediction ({peer_data['peer_count']} similar users)"
                        else:
                            # New user, rely more heavily on peer data
                            predicted_amount = peer_budget
                            model_used = f"Peer-Based Recommendation ({peer_data['peer_count']} similar users)"
                        
                        confidence = peer_confidence
                        logger.info(f"  -> Using peer recommendations: ${predicted_amount/100:.2f} (peers: {peer_data['peer_count']}, samples: {peer_data['peer_samples']})")
                    else:
                        logger.info(f"  -> Peer recommendations insufficient: {peer_data}")
                        # Fall through to next condition - but don't leave predicted_amount unset!
                        predicted_amount = None  # Mark as needing to be set by subsequent logic
                        model_used = None
                elif predictions_dict and accuracy == 0.5:  # Special case for neutral accuracy (no spending data)
                    # Use ML blend but with moderate confidence for neutral accuracy
                    if ml_change > max_reasonable_change:
                        predicted_amount = historical_avg * 0.9 + ml_predicted_amount * 0.1
                        confidence = base_confidence * model_agreement * 0.7
                        model_used = "Conservative Historical (ML too extreme, neutral accuracy)"
                        logger.info(f"  -> Using mostly historical (extreme + neutral accuracy)")
                    else:
                        predicted_amount = historical_avg * 0.8 + ml_predicted_amount * 0.2
                        confidence = base_confidence * model_agreement * 0.85
                        model_used = f"Conservative Ensemble (neutral accuracy)"
                        logger.info(f"  -> Using ML blend with neutral accuracy")
                elif predictions_dict and accuracy > 0.5:  # Good accuracy - we have spending data that matches budgets
                    # When spending matches budgets well, be more aggressive with ML suggestions
                    if spending_historical_avg > 0 and spending_historical_avg > budget_historical_avg * 1.05:
                        # If actual spending is consistently higher than budgets, suggest closer to spending amount
                        suggested_base = max(budget_historical_avg, spending_historical_avg * 0.95)  # At least 95% of spending avg
                        predicted_amount = suggested_base * 0.7 + ml_predicted_amount * 0.3
                        model_used = "Spending-Informed Ensemble (spending > budget)"
                        logger.info(f"  -> Using spending-informed blend (high spender): ${predicted_amount/100:.2f}")
                    elif ml_change > max_reasonable_change:
                        # ML prediction is too extreme, use mostly historical average
                        predicted_amount = historical_avg * 0.95 + ml_predicted_amount * 0.05
                        model_used = f"Conservative Historical (ML too extreme)"
                        logger.info(f"  -> Using historical (ML too extreme)")
                    else:
                        # Good accuracy and reasonable ML change, trust it more
                        predicted_amount = historical_avg * 0.6 + ml_predicted_amount * 0.4
                        model_used = f"Confident Ensemble (good accuracy)"
                        logger.info(f"  -> Using confident ML blend: ${predicted_amount/100:.2f}")
                    
                    confidence = base_confidence * accuracy_factor * model_agreement * 0.85
                elif predictions_dict and ml_change < historical_avg * 0.02:  # Less than 2% change
                    # Very small change, just use historical average
                    predicted_amount = historical_avg
                    confidence = base_confidence * accuracy_factor * model_agreement * 0.95
                    model_used = "Historical Average (stable pattern)"
                    logger.info(f"  -> Using historical (stable pattern, change < 2%)")
                elif predictions_dict:
                    # ML prediction is reasonable, blend smartly based on spending vs budget history
                    
                    # If we have spending history, use it to inform our decision
                    if spending_historical_avg > 0:
                        spending_vs_budget_ratio = spending_historical_avg / budget_historical_avg
                        
                        if spending_vs_budget_ratio > 1.05:  # Consistently overspend by 5%+
                            # Suggest closer to spending average since they tend to overspend
                            predicted_amount = min(ml_predicted_amount, spending_historical_avg * 1.1)  # Cap at 110% of spending avg
                            model_used = f"Spending-Aware Ensemble (tend to overspend)"
                            logger.info(f"  -> Using spending-aware blend (overspender): ${predicted_amount/100:.2f}")
                        elif spending_vs_budget_ratio < 0.95:  # Consistently underspend by 5%+
                            # They're good at staying under budget, ML prediction is probably reasonable
                            predicted_amount = budget_historical_avg * 0.7 + ml_predicted_amount * 0.3
                            model_used = f"Conservative Ensemble (good at budgeting)"
                            logger.info(f"  -> Using conservative blend (good budgeter): ${predicted_amount/100:.2f}")
                        else:  # Spending is close to budget (95-105%) - note: very accurate (98-102%) handled above
                            # They're reasonable budgeters, trust a blend more heavily weighted toward spending
                            predicted_amount = spending_historical_avg * 0.8 + ml_predicted_amount * 0.2
                            model_used = f"Balanced Ensemble (reasonable budgeter)"
                            logger.info(f"  -> Using balanced blend (reasonable): ${predicted_amount/100:.2f}")
                    else:
                        # No spending history, use conservative ML blend with budget baseline
                        predicted_amount = budget_historical_avg * 0.85 + ml_predicted_amount * 0.15
                        model_used = f"Conservative Ensemble (no spending history)"
                        logger.info(f"  -> Using conservative ML blend: ${predicted_amount/100:.2f}")
                    
                    confidence = base_confidence * accuracy_factor * model_agreement * 0.9
                else:
                    # Fallback to historical average
                    predicted_amount = self._statistical_fallback_prediction(category_data, target_month)
                    model_used = "Historical Average"
                    confidence = base_confidence * 0.7
                    logger.info(f"  -> Using fallback historical average")
                
                # Safety check: ensure predicted_amount and model_used are always set
                if predicted_amount is None or model_used is None:
                    logger.warning(f"  -> SAFETY: predicted_amount or model_used not set, using fallback for {latest_data['category_name']}")
                    predicted_amount = historical_avg
                    model_used = "Fallback Historical Average"
                    confidence = base_confidence * 0.6
                
                # Ensure confidence is within reasonable bounds
                confidence = max(0.1, min(0.95, confidence))
                
                # Debug: Log the final prediction path taken
                logger.info(f"  -> FINAL: {latest_data['category_name']} prediction: ${predicted_amount/100:.2f}, model: {model_used}")
                
                # Get feature importance (works for both ML and fallback cases)
                feature_importance = models.get('feature_importance', {}) if models else {}
                
                # Create prediction object for ALL cases
                prediction = MLBudgetPrediction(
                    category_id=int(category_id),
                    category_name=str(latest_data['category_name']),
                    predicted_amount_cents=int(predicted_amount),
                    confidence_score=float(confidence),
                    historical_avg_cents=int(historical_avg),  # Budget historical average
                    historical_spending_avg_cents=int(spending_historical_avg),  # Spending historical average
                    trend_direction=self._classify_trend(latest_data['historical_trend']),
                    ml_model_used=model_used,
                    feature_importance=feature_importance,
                    spending_cluster=spending_clusters.get(category_id, "Unknown"),
                    seasonal_pattern=seasonal_patterns.get(category_id, "Unknown"),
                    reasoning=self._generate_ml_reasoning(
                        latest_data, predicted_amount, model_used, 
                        spending_clusters.get(category_id, "Unknown"),
                        seasonal_patterns.get(category_id, "Unknown"),
                        budget_historical_avg, spending_historical_avg  # Pass both budget and spending averages
                    )
                )
                
                predictions.append(prediction)
                logger.info(f"Added prediction for category {category_id} ({latest_data['category_name']}). Total predictions so far: {len(predictions)}")
            
            # Final logging (outside the for loop)
            logger.info(f"Completed prediction loop. Total predictions: {len(predictions)}")
            
            # Sort by confidence score
            predictions.sort(key=lambda x: x.confidence_score, reverse=True)
            
            return predictions
        
        except Exception as e:
            logger.error(f"ML prediction failed: {e}")
            return []
    
    def _prepare_prediction_features(self, latest_data: pd.Series, target_month: int, target_year: int) -> Dict:
        """Prepare features for prediction"""
        
        features = {}
        
        # Target month features
        features['month_sin'] = np.sin(2 * np.pi * target_month / 12)
        features['month_cos'] = np.cos(2 * np.pi * target_month / 12)
        target_quarter = (target_month - 1) // 3 + 1
        features['quarter_sin'] = np.sin(2 * np.pi * target_quarter / 4)
        features['quarter_cos'] = np.cos(2 * np.pi * target_quarter / 4)
        
        # Historical features (carry forward from latest data)
        for col in ['historical_mean', 'historical_std', 'historical_trend', 'historical_count',
                   'recent_mean', 'recent_trend', 'actual_spent', 'spending_frequency',
                   'avg_transaction_size', 'spending_volatility', 'budget_accuracy',
                   'account_diversity', 'primary_account_usage']:
            value = latest_data.get(col, 0)
            # Ensure value is finite and not NaN
            if pd.isna(value) or not np.isfinite(value):
                features[col] = 0.0
            else:
                features[col] = float(value)
        
        return features
    
    def _statistical_fallback_prediction(self, category_data: pd.DataFrame, target_month: int) -> float:
        """Fallback prediction using enhanced statistical methods"""
        
        if len(category_data) == 0:
            return 0.0
        
        # Use historical average without adjustments
        amounts = category_data['target_amount'].values
        return float(np.mean(amounts))
    
    def _classify_trend(self, trend_slope: float) -> str:
        """Classify trend direction"""
        if trend_slope > 100:  # More than $1 increase per period
            return "increasing"
        elif trend_slope < -100:  # More than $1 decrease per period
            return "decreasing"
        else:
            return "stable"
    
    def _generate_ml_reasoning(self, data: pd.Series, predicted_amount: float, model_used: str, 
                              cluster: str, seasonal_pattern: str, budget_historical_avg: float, spending_historical_avg: float) -> str:
        """Generate reasoning for ML prediction"""
        
        reasons = []
        
        # Historical context (main basis)
        budget_avg_dollars = budget_historical_avg / 100
        predicted_dollars = predicted_amount / 100
        
        # Different reasoning based on model used
        if "Accurate Budgeter" in model_used:
            reasons.append(f"Maintaining budget ${budget_avg_dollars:.2f} (accurate budgeter)")
            difference = predicted_dollars - budget_avg_dollars
            if abs(difference) > 0.50:  # More than 50 cents difference
                if difference > 0:
                    reasons.append(f"small +${difference:.2f} inflation buffer")
                else:
                    reasons.append(f"${difference:.2f} adjustment")
        elif "Peer-Based" in model_used or "Peer-Informed" in model_used:
            reasons.append(f"Based on similar users: ${predicted_dollars:.2f}")
            if budget_avg_dollars > 0:
                difference = predicted_dollars - budget_avg_dollars
                if abs(difference) > 0.50:
                    if difference > 0:
                        reasons.append(f"peers suggest +${difference:.2f} vs your history")
                    else:
                        reasons.append(f"peers suggest ${difference:.2f} vs your history")
        elif "Historical Average" in model_used:
            if "limited data" in model_used:
                reasons.append(f"Based on budget average ${budget_avg_dollars:.2f} (limited data)")
            elif "stable pattern" in model_used:
                reasons.append(f"Based on budget average ${budget_avg_dollars:.2f} (stable pattern)")
            else:
                reasons.append(f"Based on budget average ${budget_avg_dollars:.2f}")
        else:
            reasons.append(f"Conservative ML prediction")
            reasons.append(f"budget average ${budget_avg_dollars:.2f}")
            
            # Show the adjustment if significant
            difference = predicted_dollars - budget_avg_dollars
            if abs(difference) > 0.50:  # More than 50 cents difference
                if difference > 0:
                    reasons.append(f"ML suggests +${difference:.2f} adjustment")
                else:
                    reasons.append(f"ML suggests ${difference:.2f} adjustment")
        
        # Add spending history context if available
        if spending_historical_avg > 0:
            spending_avg_dollars = spending_historical_avg / 100
            reasons.append(f"spending avg ${spending_avg_dollars:.2f}")
        
        # Spending behavior - more intelligent pattern description
        if "Accurate Budgeter" in model_used:
            # For accurate budgeters, focus on their accuracy rather than volatility
            reasons.append("consistent spending pattern")
        elif cluster != "Unknown":
            reasons.append(f"{cluster.lower()} spending pattern")
        
        # Data quality indicators
        data_points = data.get('historical_count', 0)
        if data_points < 3:
            reasons.append("(limited historical data)")
        
        return "; ".join(reasons)

# Initialize predictor
predictor = AdvancedBudgetPredictor()

@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    return HealthResponse(status="healthy", service="ai-budget-predictor-ml")

@app.post("/predict-budget", response_model=PredictionResult)
async def predict_budget(request: BudgetPredictionRequest):
    """
    Generate ML-based budget predictions
    
    This endpoint analyzes historical spending patterns and budget data to generate
    intelligent budget predictions using machine learning algorithms.
    """
    try:
        logger.info(f"ML prediction for user {request.user_id}, target: {request.target_month}/{request.target_year}")
        
        # Generate ML predictions
        predictions = predictor.predict_with_ml(
            request.user_id, 
            request.target_month, 
            request.target_year, 
            request.historical_months
        )
        
        if not predictions:
            return PredictionResult(
                predictions=[],
                target_month=request.target_month,
                target_year=request.target_year,
                user_id=request.user_id,
                ml_enabled=True,
                model_info=ModelInfo(
                    algorithms_used=["Linear Regression", "Random Forest", "K-Means Clustering"],
                    features_analyzed=["Seasonal patterns", "Spending trends", "Historical accuracy", "Account diversity"],
                    confidence_factors=["Model agreement", "Data quality", "Historical consistency"]
                ),
                message="No historical data available for ML predictions"
            )
        
        # Convert to response format
        predictions_response = []
        for pred in predictions:
            predictions_response.append(BudgetPredictionResponse(
                category_id=int(pred.category_id),
                category_name=str(pred.category_name),
                predicted_amount_cents=int(pred.predicted_amount_cents),
                predicted_amount_dollars=float(pred.predicted_amount_cents / 100),
                confidence_score=float(pred.confidence_score),
                historical_avg_cents=int(pred.historical_avg_cents),  # Budget historical average
                historical_avg_dollars=float(pred.historical_avg_cents / 100),  # Budget historical average
                historical_spending_avg_cents=int(pred.historical_spending_avg_cents),  # Spending historical average
                historical_spending_avg_dollars=float(pred.historical_spending_avg_cents / 100),  # Spending historical average
                trend_direction=str(pred.trend_direction),
                ml_model_used=str(pred.ml_model_used),
                feature_importance=pred.feature_importance,
                spending_cluster=str(pred.spending_cluster),
                seasonal_pattern=str(pred.seasonal_pattern),
                reasoning=str(pred.reasoning)
            ))
        
        response = PredictionResult(
            predictions=predictions_response,
            target_month=int(request.target_month),
            target_year=int(request.target_year),
            user_id=int(request.user_id),
            ml_enabled=True,
            model_info=ModelInfo(
                algorithms_used=["Linear Regression", "Random Forest", "K-Means Clustering"],
                features_analyzed=["Seasonal patterns", "Spending trends", "Historical accuracy", "Account diversity"],
                confidence_factors=["Model agreement", "Data quality", "Historical consistency"]
            ),
            message=f"Generated {len(predictions)} ML-based budget predictions"
        )
        
        logger.info(f"Successfully generated {len(predictions)} ML predictions for user {request.user_id}")
        return response
        
    except Exception as e:
        logger.error(f"ML prediction error: {e}")
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")

@app.post("/analyze-patterns", response_model=PatternAnalysisResult)
async def analyze_patterns(request: PatternAnalysisRequest):
    """
    Advanced ML-based pattern analysis
    
    This endpoint performs comprehensive analysis of spending patterns,
    behavioral clustering, and seasonal trend detection.
    """
    try:
        # Fetch comprehensive data
        budgets_df, spending_df, accounts_df = predictor.fetch_comprehensive_data(
            request.user_id, 
            request.historical_months
        )
        
        if budgets_df.empty:
            return PatternAnalysisResult(
                analysis={
                    "patterns": {},
                    "message": "No historical data available",
                    "user_id": request.user_id,
                    "ml_enabled": True
                },
                user_id=request.user_id,
                ml_enabled=True
            )
        
        # Engineer features and run analysis
        features_df = predictor.engineer_features(budgets_df, spending_df, accounts_df)
        spending_clusters = predictor.classify_spending_behavior(features_df)
        seasonal_patterns = predictor.detect_seasonal_patterns(features_df)
        
        # Build summary
        analysis_results = {
            "spending_behavior_clusters": spending_clusters,
            "seasonal_patterns": seasonal_patterns,
            "data_quality": {
                "total_budget_records": len(budgets_df),
                "total_spending_records": len(spending_df),
                "categories_analyzed": len(features_df['category_id'].unique()) if not features_df.empty else 0,
                "time_span_months": request.historical_months
            },
            "ml_insights": {
                "clustering_algorithm": "K-Means",
                "seasonal_detection": "Statistical Peak Analysis",
                "feature_engineering": "Time series + Financial metrics"
            }
        }
        
        return PatternAnalysisResult(
            analysis=analysis_results,
            user_id=int(request.user_id),
            ml_enabled=True
        )
        
    except Exception as e:
        logger.error(f"Pattern analysis error: {e}")
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")

@app.post("/user-insights")
async def get_user_insights(request: BudgetPredictionRequest):
    """
    Get user clustering and peer comparison insights
    
    Analyzes user spending patterns and compares with similar users to provide
    personalized financial insights and peer-based recommendations.
    """
    try:
        logger.info(f"User insights request for user {request.user_id}")
        
        # Create user profiles and clusters
        user_profiles_df = predictor.create_user_profiles(request.historical_months)
        user_clusters = predictor.cluster_users(user_profiles_df)
        
        # Get user's cluster information
        user_cluster_info = user_clusters.get(request.user_id, {})
        
        if not user_cluster_info:
            return {
                "user_id": request.user_id,
                "cluster_id": None,
                "message": "Insufficient data for user clustering",
                "peer_recommendations": {}
            }
        
        # Get peer recommendations for user's main categories
        budgets_df, spending_df, accounts_df = predictor.fetch_comprehensive_data(request.user_id, request.historical_months)
        main_categories = budgets_df['category_id'].value_counts().head(5).index.tolist() if not budgets_df.empty else []
        
        peer_recommendations = {}
        for category_id in main_categories:
            peer_data = predictor.get_peer_recommendations(request.user_id, category_id, user_clusters)
            if peer_data:
                peer_recommendations[int(category_id)] = {
                    "avg_peer_budget": peer_data['avg_peer_budget'] / 100,  # Convert to dollars
                    "avg_peer_spending": peer_data.get('avg_peer_spending', 0) / 100,
                    "peer_count": peer_data['peer_count'],
                    "samples": peer_data['peer_samples']
                }
        
        cluster_profile = user_cluster_info.get('cluster_profile', {})
        
        return {
            "user_id": request.user_id,
            "cluster_id": user_cluster_info.get('cluster_id'),
            "cluster_profile": {
                "user_count": cluster_profile.get('user_count', 0),
                "avg_monthly_spending": round(cluster_profile.get('avg_monthly_spending', 0) / 100, 2),
                "avg_expense_ratio": round(cluster_profile.get('avg_expense_ratio', 0), 2),
                "avg_budget_adherence": round(cluster_profile.get('avg_budget_adherence', 0), 2),
                "avg_categories": round(cluster_profile.get('avg_categories', 0), 1),
                "avg_age_months": round(cluster_profile.get('avg_age_months', 0), 1)
            },
            "similar_users_count": len(user_cluster_info.get('similarity_peers', [])),
            "peer_recommendations": peer_recommendations,
            "insights": predictor._generate_user_insights(user_cluster_info, cluster_profile)
        }
        
    except Exception as e:
        logger.error(f"User insights error: {e}")
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")

if __name__ == '__main__':
    import uvicorn
    uvicorn.run(app, host='0.0.0.0', port=5001, log_level="info")