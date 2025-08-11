import axios, { AxiosInstance } from 'axios';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: 'http://localhost:8081/api/v1',
      timeout: 10000,
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
      (response) => {
        return response;
      },
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
  async login(username: string, password: string): Promise<any> {
    const response = await this.api.post('/auth/login', { username, password });
    return response.data;
  }

  async register(username: string, password: string): Promise<any> {
    const response = await this.api.post('/auth/register', { username, password });
    return response.data;
  }

  // 余额相关
  async getBalances(): Promise<any[]> {
    const response = await this.api.get('/balances');
    return response.data.data;
  }

  // 地址相关
  async getAddresses(): Promise<any[]> {
    const response = await this.api.get('/addresses');
    return response.data.data;
  }

  async generateAddress(chainType: string): Promise<any> {
    const response = await this.api.post('/addresses/generate', { chain_type: chainType });
    return response.data;
  }

  // 提现相关
  async getWithdraws(): Promise<any[]> {
    const response = await this.api.get('/withdraws');
    return response.data.data;
  }

  async createWithdraw(data: any): Promise<any> {
    const response = await this.api.post('/withdraws', data);
    return response.data;
  }

  // 充值相关
  async getDeposits(): Promise<any[]> {
    const response = await this.api.get('/deposits');
    return response.data.data;
  }

  // 交易相关
  async getTransactions(): Promise<any[]> {
    const response = await this.api.get('/transactions');
    return response.data.data;
  }

  // 币种配置相关
  async getCurrencies(): Promise<any[]> {
    const response = await this.api.get('/currencies');
    return response.data.data;
  }

  async getCurrencyBySymbol(symbol: string): Promise<any> {
    const response = await this.api.get(`/currencies/${symbol}`);
    return response.data.data;
  }

  async createCurrency(data: any): Promise<any> {
    const response = await this.api.post('/currencies', data);
    return response.data;
  }

  async updateCurrency(symbol: string, data: any): Promise<any> {
    const response = await this.api.put(`/currencies/${symbol}`, data);
    return response.data;
  }

  async deleteCurrency(symbol: string): Promise<any> {
    const response = await this.api.delete(`/currencies/${symbol}`);
    return response.data;
  }

  async enableCurrency(symbol: string): Promise<any> {
    const response = await this.api.post(`/currencies/${symbol}/enable`);
    return response.data;
  }

  async disableCurrency(symbol: string): Promise<any> {
    const response = await this.api.post(`/currencies/${symbol}/disable`);
    return response.data;
  }

  async getSupportedChains(): Promise<string[]> {
    const response = await this.api.get('/currencies/chains/supported');
    return response.data.data;
  }

  // 操作相关
  async triggerCollection(symbol: string, address: string): Promise<any> {
    const response = await this.api.post('/ops/collection/trigger', {
      symbol,
      address
    });
    return response.data;
  }

  async scanBlocks(symbol: string, startBlock: number, endBlock: number, addresses: string[]): Promise<any> {
    const response = await this.api.post('/ops/scanner/scan-blocks', {
      symbol,
      start_block: startBlock,
      end_block: endBlock,
      addresses
    });
    return response.data;
  }

  async scanBlockOnce(): Promise<any> {
    const response = await this.api.post('/ops/scanner/scan-once');
    return response.data;
  }

  async startScanner(): Promise<any> {
    const response = await this.api.post('/ops/scanner/start');
    return response.data;
  }

  async stopScanner(): Promise<any> {
    const response = await this.api.post('/ops/scanner/stop');
    return response.data;
  }

  async getScannerStatus(): Promise<any> {
    const response = await this.api.get('/ops/scanner/status');
    return response.data;
  }

  async startCollection(): Promise<any> {
    const response = await this.api.post('/ops/collection/start');
    return response.data;
  }

  async stopCollection(): Promise<any> {
    const response = await this.api.post('/ops/collection/stop');
    return response.data;
  }
}

export const apiService = new ApiService(); 