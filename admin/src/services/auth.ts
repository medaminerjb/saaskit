import { saaskitClient } from './saaskit';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
  name?: string;
}

export const authService = {
  async login(credentials: LoginCredentials) {
    return await saaskitClient.auth.login(credentials);
  },

  async register(credentials: RegisterCredentials) {
    return await saaskitClient.auth.register(credentials);
  },

  async logout(accessToken: string, refreshToken: string) {
    await saaskitClient.auth.logout(accessToken, { refresh_token: refreshToken });
  },

  async refresh(refreshToken: string) {
    return await saaskitClient.auth.refresh({ refresh_token: refreshToken });
  },
};

