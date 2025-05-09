## Create Product
```bash
curl -X POST https://bug-free-bassoon-qprxx5rv5x5c4xww-8080.app.github.dev/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cloud9-espresso",
    "name": "Cloud 9 Espresso",
    "description": "Hit the high notes with our triple-threat blend spanning three legendary coffee regions, each separately roasted to full-city perfection. This headliner delivers berry brilliance backed by sweet orange-citrus acidity, then drops into smooth dark chocolate territory with brown sugar undertones before a crisp, clean finish. Rock-star performance as espresso or drip—either way, you'"'"'re floating on Cloud 9.",
    "image_url": "https://rockabillyroasting.com/wp-content/uploads/2021/06/cloud9espresso.jpg",
    "active": true,
    "stock_level": 100,
    "weight": 12,
    "origin": "Blend - Three Coffee Regions",
    "roast_level": "Full City",
    "flavor_notes": "Berry top notes, sweet orange-citrus acidity, smooth dark chocolate, brown sugar, crisp clean finish",
    "created_at": "2025-05-09T00:00:00Z",
    "updated_at": "2025-05-09T00:00:00Z"
  }'
  ```

  