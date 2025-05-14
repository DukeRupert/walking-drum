#!/bin/bash

# Script to seed coffee product data
# Replace the API URL with your actual API endpoint
API_URL="https://refactored-giggle-4w4rrp4q5r925j7w-8080.app.github.dev/api/v1/products"

# Helper function to make API calls
create_product() {
  echo "Creating product: $2"
  curl -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "$1"
  echo -e "\n"
}

# Cloud 9 Espresso (from your example)
echo "Seeding coffee products to $API_URL..."
cloud9=$(cat <<EOF
{
  "id": "cloud9-espresso",
  "name": "Cloud 9 Espresso",
  "description": "Hit the high notes with our triple-threat blend spanning three legendary coffee regions, each separately roasted to full-city perfection. This headliner delivers berry brilliance backed by sweet orange-citrus acidity, then drops into smooth dark chocolate territory with brown sugar undertones before a crisp, clean finish. Rock-star performance as espresso or drip—either way, you're floating on Cloud 9.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134080/RockabillyRoasting/Cloud_9_a3mgso.jpg",
  "active": true,
  "stock_level": 100,
  "options": {
    "weight": ["12oz", "3lb", "5lb"],
    "grind": ["Whole Bean", "Drip Ground"]
  },
  "allow_subscription": true,
  "origin": "Blend - Three Coffee Regions",
  "roast_level": "medium",
  "flavor_notes": "Berry top notes, sweet orange-citrus acidity, smooth dark chocolate, brown sugar, crisp clean finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Execute the creation of each product
create_product "$cloud9" "Cloud 9 Espresso"

echo "Coffee product seeding complete!"