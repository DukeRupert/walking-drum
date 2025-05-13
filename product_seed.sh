#!/bin/bash

# Script to seed coffee product data
# Replace the API URL with your actual API endpoint
API_URL="https://bug-free-bassoon-qprxx5rv5x5c4xww-8080.app.github.dev/api/v1/products"

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
  "weight": 12,
  "origin": "Blend - Three Coffee Regions",
  "roast_level": "Full City",
  "flavor_notes": "Berry top notes, sweet orange-citrus acidity, smooth dark chocolate, brown sugar, crisp clean finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# 2-Stroke Blend
two_stroke=$(cat <<EOF
{
  "id": "2-stroke-blend",
  "name": "2-Stroke Blend",
  "description": "Rev up your morning with this high-octane blend. Sweet top notes race ahead before downshifting into toasted nuts and chocolate, finishing with a buttery smoothness that lingers like the purr of a vintage engine. Powerful yet refined rebel fuel that kickstarts your day with the perfect balance of intensity and smoothness. This blend delivers premium performance in any brewing method—from pour-over precision to the raw power of espresso.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134085/RockabillyRoasting/2-Stroke_frl4ox.jpg",
  "active": true,
  "stock_level": 85,
  "weight": 12,
  "origin": "Central & South America",
  "roast_level": "Medium",
  "flavor_notes": "Sweet top notes, toasted nuts, chocolate undertones, buttery smooth finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Chop Top
chop_top=$(cat <<EOF
{
  "id": "chop-top",
  "name": "Chop Top",
  "description": "Custom-built blend that cruises like a classic! This modified masterpiece combines Central American and East African beans for a ride worth remembering. Rummy fruit notes with raisin richness roll into a silky-smooth body, while subtle acidity gives just enough edge before cruising into a clean finish. Hot-rodded perfection in every cup.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134083/RockabillyRoasting/Chop_Top_zdszo5.jpg",
  "active": true,
  "stock_level": 75,
  "weight": 12,
  "origin": "Central America & East Africa",
  "roast_level": "Medium",
  "flavor_notes": "Rummy fruit notes, raisin richness, silky-smooth body, subtle acidity, clean finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Ethiopia
ethiopia=$(cat <<EOF
{
  "id": "ethiopia",
  "name": "Ethiopia",
  "description": "Rock n' roll goes back to its roots! Straight from coffee's birthplace comes this high-altitude rebel with a cause. Opens with a bright, attention-grabbing intro before mellowing into sweet lemon citrus notes. The encore? A clean, dry finish with subtle berry and floral undertones that'll have you applauding for more. The original coffee soundtrack, remixed to perfection.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134081/RockabillyRoasting/Ethiopia_p6y6td.jpg",
  "active": true,
  "stock_level": 90,
  "weight": 12,
  "origin": "Ethiopia",
  "roast_level": "Light-Medium",
  "flavor_notes": "Bright opening, sweet lemon citrus, clean dry finish, subtle berry and floral undertones",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Guatemala Tikal
guatemala_tikal=$(cat <<EOF
{
  "id": "guatemala-tikal",
  "name": "Guatemala Tikal",
  "description": "A greatest hits collection from Guatemala's premium growing regions! This chart-topper delivers dark chocolate richness that harmonizes perfectly with bright red apple notes. Every cup ends with a honey-sweet encore that'll leave you craving an instant replay. Consistently mind-blowing from the first sip to the final curtain—like your favorite vinyl played on perfect speakers.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134083/RockabillyRoasting/Guatemalan_fzq6gh.jpg",
  "active": true,
  "stock_level": 80,
  "weight": 12,
  "origin": "Guatemala",
  "roast_level": "Medium",
  "flavor_notes": "Dark chocolate richness, bright red apple notes, honey-sweet finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# White Coffee
white_coffee=$(cat <<EOF
{
  "id": "white-coffee",
  "name": "White Coffee",
  "description": "The rebel's secret weapon! Under-roasted for a nutty, electric flavor that hits different from your usual brew. Packing a caffeine punch that'll have you tuned up like a hot rod at a rally. Available only ground—just how this wild child wants it. Unleash your rockabilly spirit with every cup.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134081/RockabillyRoasting/White_Coffee_pkcyol.jpg",
  "active": true,
  "stock_level": 65,
  "weight": 12,
  "origin": "Specialty Blend",
  "roast_level": "Light (Under-roasted)",
  "flavor_notes": "Nutty, electric flavor, high caffeine content",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Cascadia Decaf
cascadia_decaf=$(cat <<EOF
{
  "id": "cascadia-decaf",
  "name": "Cascadia Decaf",
  "description": "No caffeine? No compromise! This fair-trade, organic rebel breaks all the rules with 100% chemical-free processing. Our tri-continental blend from Indonesia, Central and South America delivers a clean, sweet symphony that stays balanced whether you're brewing drip or pulling shots. Proof that going decaf doesn't mean turning down the volume.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134080/RockabillyRoasting/Cascadia_q2qqqr.jpg",
  "active": true,
  "stock_level": 70,
  "weight": 12,
  "origin": "Blend - Indonesia, Central & South America",
  "roast_level": "Medium",
  "flavor_notes": "Clean, sweet, balanced",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Bike Blend
bike_blend=$(cat <<EOF
{
  "id": "bike-blend",
  "name": "Bike Blend",
  "description": "Fuel your journey with our perfectly balanced Bike Blend. This expertly crafted combination of Central American and East African beans creates a versatile coffee made for life on the move. The fruit-forward opening notes give way to a vibrant yet moderate acidity, complemented by savory caramel midtones. The adventure concludes with a full-bodied baker's chocolate finish that leaves you ready for whatever lies ahead. Works brilliantly as both drip coffee for leisurely mornings or as espresso when you need that extra kick to hit the road.",
  "image_url": "https://res.cloudinary.com/rr-wholesale/image/upload/v1723134081/RockabillyRoasting/Bike_Blend_ad3t9v.jpg",
  "active": true,
  "stock_level": 85,
  "weight": 12,
  "origin": "Central America & East Africa",
  "roast_level": "Medium",
  "flavor_notes": "Fruit top notes, moderate acidity, savory caramel, full-bodied baker's chocolate finish",
  "created_at": "2025-05-12T00:00:00Z",
  "updated_at": "2025-05-12T00:00:00Z"
}
EOF
)

# Execute the creation of each product
create_product "$cloud9" "Cloud 9 Espresso"
create_product "$two_stroke" "2-Stroke Blend"
create_product "$chop_top" "Chop Top"
create_product "$ethiopia" "Ethiopia"
create_product "$guatemala_tikal" "Guatemala Tikal"
create_product "$white_coffee" "White Coffee"
create_product "$cascadia_decaf" "Cascadia Decaf"
create_product "$bike_blend" "Bike Blend"

echo "Coffee product seeding complete!"
