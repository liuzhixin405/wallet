import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  CircularProgress,
  Divider,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';

import { apiService } from '../services/apiService';

interface Currency {
  symbol: string;
  chain_type: string;
  is_enabled: boolean;
}

interface Balance {
  address: string;
  balance: number;
  symbol: string;
}

const Tools: React.FC = () => {
  const [currencies, setCurrencies] = useState<Currency[]>([]);
  const [selectedCurrency, setSelectedCurrency] = useState<string>('');
  const [balances, setBalances] = useState<Balance[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // 归集相关状态
  const [collectionAddress, setCollectionAddress] = useState<string>('');
  const [collectionLoading, setCollectionLoading] = useState(false);

  // 扫块相关状态
  const [scanStartBlock, setScanStartBlock] = useState<string>('');
  const [scanEndBlock, setScanEndBlock] = useState<string>('');
  const [scanAddresses, setScanAddresses] = useState<string>('');
  const [scanLoading, setScanLoading] = useState(false);

  useEffect(() => {
    loadCurrencies();
  }, []);

  const loadCurrencies = async () => {
    try {
      const data = await apiService.getCurrencies();
      setCurrencies(data.filter((c: Currency) => c.is_enabled));
    } catch (error) {
      console.error('Failed to load currencies:', error);
    }
  };

  const loadBalances = async () => {
    if (!selectedCurrency) return;
    
    setLoading(true);
    try {
      const data = await apiService.getBalances();
      setBalances(data.filter((b: Balance) => b.symbol === selectedCurrency));
    } catch (error) {
      console.error('Failed to load balances:', error);
      setMessage({ type: 'error', text: '加载余额失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleCurrencyChange = (currency: string) => {
    setSelectedCurrency(currency);
    if (currency) {
      loadBalances();
    }
  };

  const handleCollection = async () => {
    if (!selectedCurrency || !collectionAddress) {
      setMessage({ type: 'error', text: '请选择币种并输入地址' });
      return;
    }

    setCollectionLoading(true);
    try {
      await apiService.triggerCollection(selectedCurrency, collectionAddress);
      setMessage({ type: 'success', text: '归集操作已触发' });
      setCollectionAddress('');
    } catch (error) {
      console.error('Collection failed:', error);
      setMessage({ type: 'error', text: '归集操作失败' });
    } finally {
      setCollectionLoading(false);
    }
  };

  const handleScanBlocks = async () => {
    if (!selectedCurrency || !scanStartBlock || !scanEndBlock || !scanAddresses) {
      setMessage({ type: 'error', text: '请填写完整的扫块参数' });
      return;
    }

    const startBlock = parseInt(scanStartBlock);
    const endBlock = parseInt(scanEndBlock);
    const addresses = scanAddresses.split(',').map(addr => addr.trim()).filter(addr => addr);

    if (isNaN(startBlock) || isNaN(endBlock) || startBlock >= endBlock) {
      setMessage({ type: 'error', text: '区块范围无效' });
      return;
    }

    if (addresses.length === 0) {
      setMessage({ type: 'error', text: '请至少输入一个地址' });
      return;
    }

    setScanLoading(true);
    try {
      await apiService.scanBlocks(selectedCurrency, startBlock, endBlock, addresses);
      setMessage({ type: 'success', text: '扫块操作已触发' });
      setScanStartBlock('');
      setScanEndBlock('');
      setScanAddresses('');
    } catch (error) {
      console.error('Scan blocks failed:', error);
      setMessage({ type: 'error', text: '扫块操作失败' });
    } finally {
      setScanLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        工具管理
      </Typography>

      {message && (
        <Alert severity={message.type} sx={{ mb: 2 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        {/* 币种选择 */}
        <Box>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                币种选择
              </Typography>
              <FormControl fullWidth>
                <InputLabel>选择币种</InputLabel>
                <Select
                  value={selectedCurrency}
                  label="选择币种"
                  onChange={(e) => handleCurrencyChange(e.target.value)}
                >
                  {currencies.map((currency) => (
                    <MenuItem key={currency.symbol} value={currency.symbol}>
                      {currency.symbol} ({currency.chain_type})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </CardContent>
          </Card>
        </Box>

        {/* 余额显示 */}
        {selectedCurrency && (
          <Box>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  {selectedCurrency} 余额
                </Typography>
                {loading ? (
                  <Box display="flex" justifyContent="center">
                    <CircularProgress />
                  </Box>
                ) : (
                  <TableContainer component={Paper}>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell>地址</TableCell>
                          <TableCell>余额</TableCell>
                          <TableCell>币种</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {balances.map((balance) => (
                          <TableRow key={balance.address}>
                            <TableCell>{balance.address}</TableCell>
                            <TableCell>{balance.balance}</TableCell>
                            <TableCell>{balance.symbol}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                )}
              </CardContent>
            </Card>
          </Box>
        )}

        {/* 手动归集 */}
        <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
          <Box sx={{ flex: 1, minWidth: 300 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                手动归集
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                将指定地址的资金归集到冷钱包
              </Typography>
              
              <TextField
                fullWidth
                label="归集地址"
                value={collectionAddress}
                onChange={(e) => setCollectionAddress(e.target.value)}
                placeholder="输入要归集的地址"
                sx={{ mb: 2 }}
              />
              
              <Button
                variant="contained"
                color="primary"
                onClick={handleCollection}
                disabled={!selectedCurrency || !collectionAddress || collectionLoading}
                fullWidth
              >
                {collectionLoading ? <CircularProgress size={24} /> : '执行归集'}
              </Button>
            </CardContent>
          </Card>
          </Box>

          {/* 手动扫块 */}
          <Box sx={{ flex: 1, minWidth: 300 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  手动扫块
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  扫描指定区块范围内的交易
                </Typography>
                
                <TextField
                  fullWidth
                  label="起始区块"
                  type="number"
                  value={scanStartBlock}
                  onChange={(e) => setScanStartBlock(e.target.value)}
                  placeholder="输入起始区块号"
                  sx={{ mb: 2 }}
                />
                
                <TextField
                  fullWidth
                  label="结束区块"
                  type="number"
                  value={scanEndBlock}
                  onChange={(e) => setScanEndBlock(e.target.value)}
                  placeholder="输入结束区块号"
                  sx={{ mb: 2 }}
                />
                
                <TextField
                  fullWidth
                  label="监控地址"
                  value={scanAddresses}
                  onChange={(e) => setScanAddresses(e.target.value)}
                  placeholder="输入地址，多个地址用逗号分隔"
                  multiline
                  rows={3}
                  sx={{ mb: 2 }}
                />
                
                <Button
                  variant="contained"
                  color="secondary"
                  onClick={handleScanBlocks}
                  disabled={!selectedCurrency || !scanStartBlock || !scanEndBlock || !scanAddresses || scanLoading}
                  fullWidth
                >
                  {scanLoading ? <CircularProgress size={24} /> : '执行扫块'}
                </Button>
              </CardContent>
            </Card>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};

export default Tools; 