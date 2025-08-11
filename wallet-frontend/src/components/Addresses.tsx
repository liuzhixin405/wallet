import React, { useEffect, useState } from 'react';
import { Box, Container, Paper, Typography, Button, List, ListItem, ListItemText, Select, MenuItem, FormControl, InputLabel, Alert } from '@mui/material';
import apiService from '../services/api';

interface AddressItem {
  id: number;
  address: string;
  chain_type: string;
  status: number;
  bind_time?: string;
  index_num: number;
  note?: string;
  created_time: string;
}

const Addresses: React.FC = () => {
  const [addresses, setAddresses] = useState<AddressItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [chainType, setChainType] = useState('Ethereum');
  const [chains, setChains] = useState<string[]>([]);

  const [success, setSuccess] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiService.getAddresses();
      setAddresses(data || []);
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to load addresses');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    (async () => {
      try {
        const ch = await apiService.getSupportedChains();
        setChains(ch || []);
        if (ch?.length) setChainType(ch[0]);
      } catch (e) {}
    })();
    load();
  }, []);

  const handleGenerate = async () => {
    setError('');
    setSuccess('');
    try {
      await apiService.generateAddress(chainType);
      setSuccess('Address generated');
      await load();
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to generate address');
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>
          Addresses
        </Typography>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}

        <Box sx={{ display: 'flex', gap: 2, mb: 2, alignItems: 'center' }}>
          <FormControl size="small">
            <InputLabel id="chain-type-label">Chain</InputLabel>
            <Select
              labelId="chain-type-label"
              value={chainType}
              label="Chain"
              onChange={(e) => setChainType(e.target.value as string)}
              sx={{ minWidth: 160 }}
            >
              {chains.map((ch) => (
                <MenuItem key={ch} value={ch}>{ch}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <Button variant="contained" onClick={handleGenerate} disabled={loading}>
            Generate Address
          </Button>
        </Box>

        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <List>
            {addresses.map((a) => (
              <ListItem key={a.id} divider>
                <ListItemText
                  primary={a.address}
                  secondary={`${a.chain_type} • created ${new Date(a.created_time).toLocaleString()}`}
                />
              </ListItem>
            ))}
            {!addresses.length && <Typography>No addresses yet</Typography>}
          </List>
        )}
      </Paper>
    </Container>
  );
};

export default Addresses; 