import { SaaSKitClient } from '../../../sdk/js/dist/index.js';

export const saaskitClient = new SaaSKitClient({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
  maxRetries: 3,
  retryDelay: 1000,
});
