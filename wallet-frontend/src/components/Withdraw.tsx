import React, { useState } from 'react';
import { Container, Paper, Typography, TextField, Button, Box, Alert, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import apiService from '../services/api';

const Withdraw: React.FC = () => {
  const [currency, setCurrency] = useState('ETH');
  const [chainType, setChainType] = useState('Ethereum');
  const [currencies, setCurrencies] = useState<Array<{ currency_symbol: string; chain_type: string }>>([]);
  const [chains, setChains] = useState<string[]>([]);

  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    (async () => {
      try {
        const [cur, ch] = await Promise.all([
          apiService.getCurrencies(),
          apiService.getSupportedChains(),
        ]);
        setCurrencies(cur?.map((c: any) => ({ currency_symbol: c.currency_symbol, chain_type: c.chain_type })) || []);
        setChains(ch || []);
        if (cur?.length) {
          setCurrency(cur[0].currency_symbol);
          setChainType(cur[0].chain_type);
        }
      } catch (e) {
        // ignore
      }
    })();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    const amt = parseFloat(amount);
    if (!toAddress || isNaN(amt) || amt <= 0) {
      setError('Invalid address or amount');
      return;
    }

    setLoading(true);
    try {
      await apiService.createWithdrawal({ currency_symbol: currency, chain_type: chainType, to_address: toAddress, amount: amt });
      setSuccess('Withdraw request created');
      setToAddress('');
      setAmount('');
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to create withdraw');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>Withdraw</Typography>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
        <Box component="form" onSubmit={handleSubmit} sx={{ display: 'grid', gap: 2 }}>
          <FormControl size="small">
            <InputLabel id="currency-label">Currency</InputLabel>
            <Select labelId="currency-label" label="Currency" value={currency} onChange={(e) => setCurrency(e.target.value as string)} sx={{ minWidth: 160 }}>
              {currencies.map((c) => (
                <MenuItem key={`${c.currency_symbol}-${c.chain_type}`} value={c.currency_symbol}>{c.currency_symbol}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <FormControl size="small">
            <InputLabel id="chain-label">Chain</InputLabel>
            <Select labelId="chain-label" label="Chain" value={chainType} onChange={(e) => setChainType(e.target.value as string)} sx={{ minWidth: 160 }}>
              {chains.map((ch) => (
                <MenuItem key={ch} value={ch}>{ch}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField label="To Address" value={toAddress} onChange={e => setToAddress(e.target.value)} required />
          <TextField label="Amount" value={amount} onChange={e => setAmount(e.target.value)} required />
          <Button type="submit" variant="contained" disabled={loading}>{loading ? 'Submitting...' : 'Submit'}</Button>
        </Box>
      </Paper>
    </Container>
  );
};

export default Withdraw; 