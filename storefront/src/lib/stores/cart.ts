// src/lib/stores/cart.ts
import { writable, get } from 'svelte/store';
import type { Price, Product } from '$lib/types';

export interface CartItem {
  priceId: string;
  price: Price;
  product: Product;
  quantity: number;
}

export interface Cart {
  items: CartItem[];
}

function createCartStore() {
  // Initialize the cart store
  const { subscribe, set, update } = writable<Cart>({ items: [] });

  return {
    subscribe,
    
    // Add an item to the cart
    addItem: (product: Product, price: Price, quantity = 1) => {
      update(cart => {
        // Check if item already exists in cart
        const existingItemIndex = cart.items.findIndex(item => item.priceId === price.id);
        
        if (existingItemIndex >= 0) {
          // Update quantity if item exists
          const items = [...cart.items];
          items[existingItemIndex] = {
            ...items[existingItemIndex],
            quantity: items[existingItemIndex].quantity + quantity
          };
          return { items };
        } else {
          // Add new item if it doesn't exist
          return {
            items: [
              ...cart.items,
              {
                priceId: price.id,
                price,
                product,
                quantity
              }
            ]
          };
        }
      });
    },
    
    // Remove an item from the cart
    removeItem: (priceId: string) => {
      update(cart => ({
        items: cart.items.filter(item => item.priceId !== priceId)
      }));
    },
    
    // Update item quantity
    updateQuantity: (priceId: string, quantity: number) => {
      update(cart => {
        const items = cart.items.map(item => {
          if (item.priceId === priceId) {
            return { ...item, quantity };
          }
          return item;
        });
        return { items };
      });
    },
    
    // Clear the cart
    clear: () => set({ items: [] }),
    
    // Get current cart
    getCart: () => get({ subscribe })
  };
}

export const cart = createCartStore();