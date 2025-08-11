import axios, { AxiosInstance, AxiosResponse } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.api.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // 认证相关
  async login(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/auth/login', data);
    return response.data;
  }

  async register(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/auth/register', data);
    return response.data;
  }

  // 地址管理
  async getAddresses(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/addresses');
    return response.data;
  }

  async generateAddress(chainType: string): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/addresses/generate', { chain_type: chainType });
    return response.data;
  }

  async bindAddress(addressId: number, userId: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/addresses/bind', { 
      address_id: addressId, 
      user_id: userId 
    });
    return response.data;
  }

  // 余额管理
  async getBalances(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/balances');
    return response.data;
  }

  async getBalance(currency: string, chain: string): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/balances/${currency}/${chain}`);
    return response.data;
  }

  // 提币管理
  async createWithdrawal(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/withdrawals', data);
    return response.data;
  }

  async getWithdrawals(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/withdrawals');
    return response.data;
  }

  async getWithdrawal(id: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/withdrawals/${id}`);
    return response.data;
  }

  // 充币记录
  async getDeposits(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/deposits');
    return response.data;
  }

  async getDeposit(id: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/deposits/${id}`);
    return response.data;
  }

  // 交易记录
  async getTransactions(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/transactions');
    return response.data;
  }

  // 币种配置
  async getCurrencies(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/currencies');
    return response.data;
  }
}

export const apiService = new ApiService();
export default apiService; 