export interface User {
  id: number;
  username: string;
  email?: string;
}

export interface Address {
  id: number;
  address: string;
  chain_type: string;
  status: number;
  bind_time?: string;
  index_num: number;
  note: string;
  created_time: string;
}

export interface Balance {
  id: number;
  currency_symbol: string;
  chain_type: string;
  protocol?: string;
  address: string;
  balance: number;
  frozen: number;
  total: number;
  created_time: string;
  updated_time: string;
}

export interface WithdrawRecord {
  id: number;
  currency_symbol: string;
  chain_type: string;
  protocol: string;
  user_id: number;
  from_address: string;
  to_address: string;
  txid?: string;
  amount: number;
  fee: number;
  total_amount: number;
  unique_id: string;
  status: number;
  block_height?: number;
  confirmations: number;
  is_internal: boolean;
  notify_status: boolean;
  fail_reason: string;
  remark?: string;
  confirmed_time?: string;
  created_at: string;
  updated_at: string;
  type?: number;
}

export interface DepositRecord {
  id: number;
  userid: number;
  currency_symbol: string;
  chain_type: string;
  protocol?: string;
  from_address: string;
  to_address: string;
  amount: number;
  fee: number;
  txid: string;
  unique_id: string;
  status: boolean;
  is_internal: boolean;
  confirmations: number;
  block_height?: number;
  notify_status: boolean;
  fail_reason: string;
  confirmed_time?: string;
  created_time: string;
  updated_time: string;
}

export interface CurrencyChainConfig {
  id: number;
  currency_symbol: string;
  currency_name: string;
  chain_type: string;
  chain_name: string;
  protocol: string;
  contract_address?: string;
  decimals: number;
  is_native: boolean;
  deposit_enabled: boolean;
  withdraw_enabled: boolean;
  min_deposit_amount: number;
  min_withdraw_amount: number;
  max_withdraw_amount: number;
  withdraw_fee: number;
  withdraw_confirms: number;
  deposit_confirms: number;
  rpc_url?: string;
  scan_url?: string;
  icon_url: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
} 