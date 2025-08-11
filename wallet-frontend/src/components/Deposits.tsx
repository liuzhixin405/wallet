import React, { useEffect, useState } from 'react';
import { Container, Paper, Typography, List, ListItem, ListItemText, Alert } from '@mui/material';
import apiService from '../services/api';

interface DepositItem {
  id: number;
  currency_symbol: string;
  chain_type: string;
  to_address: string;
  amount: number;
  created_time: string;
}

const Deposits: React.FC = () => {
  const [items, setItems] = useState<DepositItem[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiService.getDeposits();
      setItems(data || []);
    } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to load deposits');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>Deposit History</Typography>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <List>
            {items.map((d) => (
              <ListItem key={d.id} divider>
                <ListItemText
                  primary={`${d.currency_symbol} ${d.amount}`}
                  secondary={`${d.chain_type} • ${d.to_address} • ${new Date(d.created_time).toLocaleString()}`}
                />
              </ListItem>
            ))}
            {!items.length && <Typography>No deposits</Typography>}
          </List>
        )}
      </Paper>
    </Container>
  );
};

export default Deposits; 