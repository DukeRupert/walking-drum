# Coffee Subscription System Architecture

## Overview

This system will manage recurring coffee subscriptions for a coffee roaster, allowing customers to subscribe to regular coffee deliveries on a weekly, bi-weekly, or monthly basis.

## Core Entities

### Products

- Coffee varieties offered by the roaster
- Each product has attributes like name, description, origin, roast level, etc.
- Products are synchronized with Stripe product catalog

### Prices

- Different subscription options for each product
- Weekly, bi-weekly, or monthly delivery frequencies
- Price points for different quantities (e.g., 250g, 500g, 1kg)
- Each price is linked to a Stripe price object

### Customers

- Customer information (name, email, shipping address)
- Payment methods (stored securely via Stripe)
- Each customer is synced with a Stripe customer object

### Subscriptions

- Links customers to specific products at specific prices
- Tracks status (active, paused, canceled)
- Manages delivery schedules
- Each subscription is backed by a Stripe subscription

## System Components

### API Layer

- RESTful endpoints for managing products, prices, customers, and subscriptions
- Authentication and authorization for customer and admin access
- Echo framework for HTTP routing and middleware

### Service Layer

- Business logic for subscription management
- Integration with Stripe API
- Event handling for subscription lifecycle events

### Data Layer

- PostgreSQL database for persistence
- Repository pattern for data access
- Database migrations for schema management

### Integration Layer

- Stripe webhook handler for asynchronous events
- Email notification service
- Potentially shipping/delivery integration

## Database Schema

### Products Table

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    stripe_product_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    origin VARCHAR(255),
    roast_level VARCHAR(50),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Prices Table

```sql
CREATE TABLE prices (
    id SERIAL PRIMARY KEY,
    stripe_price_id VARCHAR(255) UNIQUE NOT NULL,
    product_id INTEGER REFERENCES products(id),
    nickname VARCHAR(255),
    unit_amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    recurring_interval VARCHAR(10) NOT NULL, -- 'week', 'month'
    recurring_interval_count INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Customers Table

```sql
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    stripe_customer_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Customer Addresses Table

```sql
CREATE TABLE customer_addresses (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    line1 VARCHAR(255) NOT NULL,
    line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255),
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(2) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Subscriptions Table

```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
    customer_id INTEGER REFERENCES customers(id),
    price_id INTEGER REFERENCES prices(id),
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Key Workflows

### New Subscription Flow

1. Customer selects coffee product(s)
2. Customer chooses subscription frequency (weekly, bi-weekly, monthly)
3. Customer provides shipping and billing information
4. System creates Stripe customer if new
5. System creates Stripe subscription
6. Customer completes payment setup
7. System activates subscription
8. Confirmation email sent to customer

### Subscription Management Flow

1. Customer logs into account
2. Customer can view active subscriptions
3. Customer can modify subscription (change frequency, pause, cancel)
4. System updates Stripe subscription accordingly
5. Confirmation email sent for changes

### Subscription Renewal Flow

1. Stripe attempts automatic payment before renewal date
2. Stripe webhook notifies system of payment status
3. If successful, system schedules next delivery
4. If failed, system notifies customer and tries again

### Admin Management Flow

1. Admin can view and manage all products, prices, customers, and subscriptions
2. Admin can create or update products and prices
3. Admin can view subscription analytics
4. Admin can manually adjust subscriptions as needed
