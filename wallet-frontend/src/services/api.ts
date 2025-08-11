import axios, { AxiosInstance, AxiosResponse } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081/api/v1';

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
        // 确保 headers 存在
        if (!config.headers) {
          config.headers = {} as any;
        }
        if (token) {
          (config.headers as any).Authorization = `Bearer ${token}`;
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

  // 统一解包 { data: ... }
  private unwrap<T>(resp: AxiosResponse<any, any>): T {
    return (resp.data && resp.data.data !== undefined) ? resp.data.data as T : resp.data as T;
  }

  // 认证相关
  async login(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/auth/login', data);
    return this.unwrap<any>(response);
  }

  async register(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/auth/register', data);
    return this.unwrap<any>(response);
  }

  // 地址管理
  async getAddresses(): Promise<any[]> {
    const response: AxiosResponse<any> = await this.api.get('/addresses');
    return this.unwrap<any[]>(response);
  }

  async generateAddress(chainType: string): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/addresses/generate', { chain_type: chainType });
    return this.unwrap<any>(response);
  }

  async bindAddress(addressId: number, userId: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/addresses/bind', { 
      address_id: addressId, 
      user_id: userId 
    });
    return this.unwrap<any>(response);
  }

  // 余额管理
  async getBalances(): Promise<any[]> {
    const response: AxiosResponse<any> = await this.api.get('/balances');
    return this.unwrap<any[]>(response);
  }

  async getBalance(currency: string, chain: string): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/balances/${currency}/${chain}`);
    return this.unwrap<any>(response);
  }

  // 提币管理
  async createWithdrawal(data: any): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/withdraws', data);
    return this.unwrap<any>(response);
  }

  async getWithdrawals(): Promise<any[]> {
    const response: AxiosResponse<any> = await this.api.get('/withdraws');
    return this.unwrap<any[]>(response);
  }

  async getWithdrawal(id: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/withdraws/${id}`);
    return this.unwrap<any>(response);
  }

  // 充币记录
  async getDeposits(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/deposits');
    return this.unwrap<any[]>(response);
  }

  async getDeposit(id: number): Promise<any> {
    const response: AxiosResponse<any> = await this.api.get(`/deposits/${id}`);
    return this.unwrap<any>(response);
  }

  // 交易记录
  async getTransactions(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/transactions');
    return this.unwrap<any[]>(response);
  }

  // 币种配置
  async getCurrencies(): Promise<any[]> {
    const response: AxiosResponse<any[]> = await this.api.get('/currencies');
    return this.unwrap<any[]>(response);
  }

  async getSupportedChains(): Promise<string[]> {
    const response: AxiosResponse<any> = await this.api.get('/currencies/chains/supported');
    return this.unwrap<string[]>(response);
  }

  // 运维操作
  async triggerCollection(): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/ops/collection/trigger');
    return this.unwrap<any>(response);
  }

  async scanBlockOnce(): Promise<any> {
    const response: AxiosResponse<any> = await this.api.post('/ops/scanner/scan-once');
    return this.unwrap<any>(response);
  }
}

export const apiService = new ApiService();
export default apiService; 