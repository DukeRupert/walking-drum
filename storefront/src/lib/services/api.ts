import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ProductResponse, PriceResponse, CustomerResponse, Product, Price } from '$lib/types';

export async function getProducts(): Promise<ProductResponse> {
  const response = await fetch(`${PUBLIC_API_BASE_URL}/products`);
  if (!response.ok) {
    throw new Error('Failed to fetch products');
  }
  return await response.json();
}

export async function getProduct(id: string): Promise<Product> {
  const response = await fetch(`${PUBLIC_API_BASE_URL}/products/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch product with id ${id}`);
  }
  const data = await response.json();
  return data;
}

export async function getPrices(): Promise<PriceResponse> {
  const response = await fetch(`${PUBLIC_API_BASE_URL}/prices`);
  if (!response.ok) {
    throw new Error('Failed to fetch prices');
  }
  return await response.json();
}

export async function getPrice(id: string): Promise<Price> {
  const response = await fetch(`${PUBLIC_API_BASE_URL}/prices/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch price with id ${id}`);
  }
  const data = await response.json();
  return data;
}

export async function getCustomers(): Promise<CustomerResponse> {
  const response = await fetch(`${PUBLIC_API_BASE_URL}/customers`);
  if (!response.ok) {
    throw new Error('Failed to fetch customers');
  }
  const data = await response.json();
  return data;
}