import React, { useEffect, useState } from 'react';
import { Container, Paper, Typography, List, ListItem, ListItemText, Alert } from '@mui/material';
import apiService from '../services/api';

interface TxItem {
  id: number;
  currency_symbol: string;
  chain_type: string;
  address: string;
  amount: number;
  created_time: string;
}

const Transactions: React.FC = () => {
  const [items, setItems] = useState<TxItem[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiService.getTransactions();
      setItems(data || []);
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to load transactions');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>Transactions</Typography>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <List>
            {items.map((t) => (
              <ListItem key={t.id} divider>
                <ListItemText
                  primary={`${t.currency_symbol} ${t.amount}`}
                  secondary={`${t.chain_type} • ${t.address} • ${new Date(t.created_time).toLocaleString()}`}
                />
              </ListItem>
            ))}
            {!items.length && <Typography>No transactions</Typography>}
          </List>
        )}
      </Paper>
    </Container>
  );
};

export default Transactions; 